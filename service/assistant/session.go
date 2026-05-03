package assistant

import (
	"context"
	"encoding/json"
	"github.com/honeycombio/beeline-go"
	"github.com/pebble-dev/bobby-assistant/service/assistant/llm"
	"github.com/pebble-dev/bobby-assistant/service/assistant/persistence"
	"github.com/pebble-dev/bobby-assistant/service/assistant/quota"
	"github.com/pebble-dev/bobby-assistant/service/assistant/verifier"
	"github.com/pebble-dev/bobby-assistant/service/assistant/widgets"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/pebble-dev/bobby-assistant/service/assistant/config"
	"github.com/pebble-dev/bobby-assistant/service/assistant/functions"
	"github.com/pebble-dev/bobby-assistant/service/assistant/query"
	"github.com/redis/go-redis/v9"
	"nhooyr.io/websocket"
)

type PromptSession struct {
	conn             *websocket.Conn
	prompt           string
	userToken        string
	query            url.Values
	redis            *redis.Client
	threadId         uuid.UUID
	originalThreadId string
}

type QueryContext struct {
	values url.Values
}

func NewPromptSession(redisClient *redis.Client, rw http.ResponseWriter, r *http.Request) (*PromptSession, error) {
	prompt := r.URL.Query().Get("prompt")
	userToken := r.URL.Query().Get("token")
	originalThreadId := r.URL.Query().Get("threadId")
	c, err := websocket.Accept(rw, r, &websocket.AcceptOptions{
		OriginPatterns:     []string{"null"},
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	return &PromptSession{
		conn:             c,
		prompt:           prompt,
		userToken:        userToken,
		query:            r.URL.Query(),
		redis:            redisClient,
		threadId:         uuid.New(),
		originalThreadId: originalThreadId,
	}, nil
}

func newOpenAIClient() openai.Client {
	cfg := config.GetConfig()
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.LLMAPIKey),
	}
	if cfg.LLMBaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.LLMBaseURL))
	}
	return openai.NewClient(opts...)
}

func messagesToOpenAI(systemPrompt string, messages []*llm.ChatMessage) []openai.ChatCompletionMessageParamUnion {
	var result []openai.ChatCompletionMessageParamUnion
	result = append(result, openai.SystemMessage(systemPrompt))

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			result = append(result, openai.UserMessage(msg.Content))
		case "assistant":
			if msg.FunctionCall != nil {
				argsJSON, _ := json.Marshal(msg.FunctionCall.Args)
				result = append(result, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
							{
								OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
									ID: msg.FunctionCall.ID,
									Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
										Name:      msg.FunctionCall.Name,
										Arguments: string(argsJSON),
									},
								},
							},
						},
					},
				})
			} else {
				result = append(result, openai.AssistantMessage(msg.Content))
			}
		case "tool":
			resultJSON, _ := json.Marshal(msg.FunctionResponse.Response)
			result = append(result, openai.ToolMessage(string(resultJSON), msg.FunctionResponse.CallID))
		}
	}
	return result
}

func (ps *PromptSession) Run(ctx context.Context) {
	ctx = query.ContextWith(ctx, ps.query)
	cfg := config.GetConfig()
	client := newOpenAIClient()

	var messages []*llm.ChatMessage
	messages = append(messages, &llm.ChatMessage{
		Role:    "user",
		Content: ps.prompt,
	})

	if ps.originalThreadId != "" {
		threadContext, err := ps.restoreContextFromThread(ctx, ps.originalThreadId)
		if err != nil {
			log.Printf("error restoring thread: %v\n", err)
			_ = ps.conn.Close(websocket.StatusInternalError, "Error restoring thread.")
			return
		}
		oldMessages := ps.restoreThread(threadContext)
		messages = append(oldMessages, messages...)
	}
	query.ThreadContextFromContext(ctx).ThreadId = ps.threadId
	user, err := quota.GetUserInfo(ctx, ps.userToken)
	if err != nil {
		log.Printf("get user info failed: %v\n", err)
		_ = ps.conn.Close(websocket.StatusInternalError, "get user info failed")
		return
	}
	beeline.AddField(ctx, "user_id", user.UserId)
	if !user.HasSubscription {
		beeline.AddField(ctx, "error", "no subscription")
		log.Printf("user %d has no subscription\n", user.UserId)
		_ = ps.conn.Close(websocket.StatusPolicyViolation, "You need an active Rebble subscription to use Bobby.")
		return
	}
	qt := quota.NewTracker(ps.redis, user.UserId)
	used, remaining, err := qt.GetQuota(ctx)
	if err != nil {
		log.Printf("get quota failed: %v\n", err)
		_ = ps.conn.Close(websocket.StatusInternalError, "Quota lookup failed.")
		return
	}
	if remaining < 1 {
		log.Printf("quota exceeded for user %d\n", user.UserId)
		_ = ps.conn.Close(websocket.StatusPolicyViolation, "You have exceeded your quota for this month.")
		return
	}
	log.Printf("user %d has used %d / %d credits\n", user.UserId, used, remaining)
	totalInputTokens := 0
	totalOutputTokens := 0
	iterations := 0
	for {
		cont, err := func() (bool, error) {
			ctx, span := beeline.StartSpan(ctx, "chat_iteration")
			defer span.Send()
			iterations++
			var tools []openai.ChatCompletionToolUnionParam
			if iterations <= 10 {
				tools = functions.GetFunctionDefinitionsForCapabilities(query.SupportedActionsFromContext(ctx))
			}
			systemPrompt := ps.generateSystemPrompt(ctx)
			streamCtx, streamSpan := beeline.StartSpan(ctx, "chat_stream")

			params := openai.ChatCompletionNewParams{
				Model:       openai.ChatModel(cfg.LLMModel),
				Messages:    messagesToOpenAI(systemPrompt, messages),
				Temperature: openai.Float(0.5),
			}
			if len(tools) > 0 {
				params.Tools = tools
			}

			stream := client.Chat.Completions.NewStreaming(streamCtx, params)

			var functionCall *llm.FunctionCall
			content := ""
			var currentToolCallID string
			var currentToolCallName string
			var currentToolCallArgs strings.Builder
			bufferedContent := ""
			leftTrimming := false
			var usagePromptTokens int
			var usageCompletionTokens int

		read_loop:
			for stream.Next() {
				evt := stream.Current()
				if evt.Usage.PromptTokens > 0 || evt.Usage.CompletionTokens > 0 {
					usagePromptTokens = int(evt.Usage.PromptTokens)
					usageCompletionTokens = int(evt.Usage.CompletionTokens)
				}
				if len(evt.Choices) == 0 {
					continue
				}
				choice := evt.Choices[0]
				ourContent := choice.Delta.Content

				if len(choice.Delta.ToolCalls) > 0 {
					for _, tc := range choice.Delta.ToolCalls {
						if tc.ID != "" {
							if currentToolCallID != "" {
								var args map[string]any
								_ = json.Unmarshal([]byte(currentToolCallArgs.String()), &args)
								functionCall = &llm.FunctionCall{
									ID:   currentToolCallID,
									Name: currentToolCallName,
									Args: args,
								}
							}
							currentToolCallID = tc.ID
							currentToolCallName = tc.Function.Name
							currentToolCallArgs.Reset()
						}
						currentToolCallArgs.WriteString(tc.Function.Arguments)
					}
				}

				if bufferedContent != "" {
					bufferedContent += ourContent
					closers := strings.Count(bufferedContent, "!>") + strings.Count(bufferedContent, "/>")
					if strings.Count(bufferedContent, "<!") != closers {
						continue
					} else {
						ourContent = bufferedContent
						bufferedContent = ""
					}
				} else {
					closers := strings.Count(ourContent, "!>") + strings.Count(ourContent, "/>")
					if strings.Count(ourContent, "<!") != closers {
						bufferedContent += ourContent
						continue
					}
				}
				if strings.TrimSpace(ourContent) != "" {
					streamContent := ourContent
					re := regexp.MustCompile(`(?s)\s*<!.+?[!/]>\s*`)
					widget := re.FindAllString(ourContent, -1)
					splitting := true
					if len(widget) > 0 {
						for _, w := range widget {
							processed, err := widgets.ProcessWidget(ctx, qt, w)
							replacement := ""
							if err != nil {
								log.Printf("process widget failed: %v\n", err)
								replacement = "(widget processing failed)"
							} else {
								jsoned, err := json.Marshal(processed)
								if err != nil {
									log.Printf("marshal widget failed: %v\n", err)
									replacement = "(widget processing failed)"
								} else {
									splitting = false
									replacement = "<<!!WIDGET:" + string(jsoned) + "!!>>"
								}
							}
							streamContent = strings.Replace(streamContent, w, replacement, 1)
							if strings.HasSuffix(streamContent, "!!>>") {
								leftTrimming = true
							}
						}
					}
					if leftTrimming {
						streamContent = strings.TrimLeft(streamContent, " \r\n\t")
					}
					if strings.TrimSpace(streamContent) != "" {
						var words []string
						if splitting {
							words = strings.Split(streamContent, " ")
							leftTrimming = false
						} else {
							words = []string{streamContent}
						}
						for i, w := range words {
							if i != len(words)-1 {
								w += " "
							}
							if err := ps.conn.Write(streamCtx, websocket.MessageText, []byte("c"+w)); err != nil {
								streamSpan.AddField("error", err)
								log.Printf("write to websocket failed: %v\n", err)
								break read_loop
							}
							time.Sleep(time.Millisecond * 40)
						}
					}
				}
				content += ourContent
			}
			if err := stream.Err(); err != nil {
				streamSpan.AddField("error", err)
				log.Printf("recv from LLM failed: %v\n", err)
				_ = ps.conn.Close(websocket.StatusInternalError, "Bobby is unavailable right now. Please try again in a few moments.")
				streamSpan.Send()
				return false, err
			}

			if currentToolCallID != "" {
				var args map[string]any
				_ = json.Unmarshal([]byte(currentToolCallArgs.String()), &args)
				functionCall = &llm.FunctionCall{
					ID:   currentToolCallID,
					Name: currentToolCallName,
					Args: args,
				}
			}

			streamSpan.Send()

			if usagePromptTokens > 0 {
				_, err = qt.ChargeInputQuota(ctx, usagePromptTokens, 0)
				if err != nil {
					log.Printf("charge input quota failed: %v\n", err)
				}
				totalInputTokens += usagePromptTokens
			}
			if usageCompletionTokens > 0 {
				_, err = qt.ChargeOutputQuota(ctx, usageCompletionTokens)
				if err != nil {
					log.Printf("charge output quota failed: %v\n", err)
				}
				totalOutputTokens += usageCompletionTokens
			}

			if len(strings.TrimSpace(content)) > 0 {
				messages = append(messages, &llm.ChatMessage{
					Role:    "assistant",
					Content: content,
				})
			}
			if functionCall != nil {
				messages = append(messages, &llm.ChatMessage{
					Role:         "assistant",
					FunctionCall: functionCall,
				})
				log.Printf("calling function %s\n", functionCall.Name)
				fnBytes, _ := json.Marshal(functionCall.Args)
				fnArgs := string(fnBytes)
				if err := ps.conn.Write(ctx, websocket.MessageText, []byte("f"+functions.SummariseFunction(functionCall.Name, fnArgs))); err != nil {
					log.Printf("write to websocket failed: %v\n", err)
					return false, err
				}
				var result string
				if functions.IsAction(functionCall.Name) {
					result, err = functions.CallAction(ctx, qt, functionCall.Name, fnArgs, ps.conn)
				} else {
					result, err = functions.CallFunction(ctx, qt, functionCall.Name, fnArgs)
				}
				if err != nil {
					log.Printf("call function failed: %v\n", err)
					result = "failed to call function: " + err.Error()
				}
				var mapResult map[string]any
				_ = json.Unmarshal([]byte(result), &mapResult)
				messages = append(messages, &llm.ChatMessage{
					Role: "tool",
					FunctionResponse: &llm.FunctionResponse{
						CallID:   functionCall.ID,
						Name:     functionCall.Name,
						Response: mapResult,
					},
				})
				return true, nil
			}
			return false, nil
		}()
		if err != nil {
			return
		}
		if !cont {
			log.Println("Stopping")
			break
		}
		log.Println("Going around again")
	}

	lies, err := verifier.FindLies(ctx, qt, messages)
	if err != nil {
		// Bobby doesn't usually lie, so this isn't worth killing the session over.
		log.Printf("find lies failed: %v\n", err)
	}
	if len(lies) > 0 {
		beeline.AddField(ctx, "lies", lies)
		log.Printf("lies detected: %v\n", lies)
		var formattedLies []string
		for _, l := range lies {
			switch l {
			case "alarm":
				formattedLies = append(formattedLies, "set an alarm")
			case "timer":
				formattedLies = append(formattedLies, "set a timer")
			case "reminder":
				formattedLies = append(formattedLies, "set a reminder")
			case "settings":
				formattedLies = append(formattedLies, "change any settings")
			}
		}
		prettyLies := strings.Join(formattedLies, ", ")
		if len(formattedLies) > 1 {
			prettyLies = strings.Join(formattedLies[:len(formattedLies)-1], ", ") + ", or " + formattedLies[len(formattedLies)-1]
		}
		message := "Bobby did not, in fact, " + prettyLies + "."
		if err := ps.conn.Write(ctx, websocket.MessageText, []byte("w"+message)); err != nil {
			log.Printf("write to websocket failed: %v\n", err)
		}
	}

	if err := ps.conn.Write(ctx, websocket.MessageText, []byte("d")); err != nil {
		log.Printf("write to websocket failed: %v\n", err)
	}

	beeline.AddField(ctx, "total_input_tokens", totalInputTokens)
	beeline.AddField(ctx, "total_output_tokens", totalOutputTokens)
	beeline.AddField(ctx, "total_cost", totalInputTokens*quota.InputTokenCredits+totalOutputTokens*quota.OutputTokenCredits)
	if err := ps.storeThread(ctx, messages); err != nil {
		log.Printf("store thread failed: %v\n", err)
		_ = ps.conn.Close(websocket.StatusInternalError, "store thread failed")
		return
	}
	if err := ps.conn.Write(ctx, websocket.MessageText, []byte("t"+ps.threadId.String())); err != nil {
		log.Printf("store thread ID failed: %s\n", err)
	}
	log.Println("Request handled successfully.")
	_ = ps.conn.Close(websocket.StatusNormalClosure, "")
}

// Replace the narrow non-breaking space with a regular non-breaking space.
func fixUnsupportedCharacters(s string) string {
	return strings.ReplaceAll(s, "\u202f", "\u00a0")
}

func (ps *PromptSession) storeThread(ctx context.Context, messages []*llm.ChatMessage) error {
	ctx, span := beeline.StartSpan(ctx, "store_thread")
	defer span.Send()
	var toStore []persistence.SerializedMessage
	for _, m := range messages {
		if m.Role == "user" || m.Role == "assistant" {
			sm := persistence.SerializedMessage{
				Role:         m.Role,
				Content:      m.Content,
				FunctionCall: m.FunctionCall,
			}
			if sm.FunctionCall != nil || len(strings.TrimSpace(m.Content)) > 0 {
				toStore = append(toStore, sm)
			}
		} else if m.Role == "tool" && m.FunctionResponse != nil {
			fr := *m.FunctionResponse
			fnInfo := functions.GetFunctionRegistration(fr.Name)
			if fnInfo != nil && fnInfo.RedactOutputInChatHistory {
				fr.Response = map[string]any{"redacted": "redacted to reduce context size, call again if necessary"}
			}
			toStore = append(toStore, persistence.SerializedMessage{
				Role:             m.Role,
				FunctionResponse: &fr,
			})
		}
	}
	threadContext := query.ThreadContextFromContext(ctx)
	threadContext.Messages = toStore
	return persistence.StoreThread(ctx, ps.redis, threadContext)
}

func (ps *PromptSession) restoreContextFromThread(ctx context.Context, oldThreadId string) (*persistence.ThreadContext, error) {
	threadContext, err := persistence.LoadThread(ctx, ps.redis, oldThreadId)
	if err != nil {
		return nil, err
	}
	return threadContext, nil
}

func (ps *PromptSession) restoreThread(threadContext *persistence.ThreadContext) []*llm.ChatMessage {
	var result []*llm.ChatMessage
	for _, m := range threadContext.Messages {
		result = append(result, &llm.ChatMessage{
			Content:          m.Content,
			Role:             m.Role,
			FunctionCall:     m.FunctionCall,
			FunctionResponse: m.FunctionResponse,
		})
	}
	return result
}
