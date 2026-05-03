package functions

import (
	"context"
	"fmt"
	"log"

	"github.com/honeycombio/beeline-go"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"

	"github.com/pebble-dev/bobby-assistant/service/assistant/quota"
	"github.com/pebble-dev/bobby-assistant/service/assistant/util/currencies"
)

type CurrencyConversionRequest struct {
	Amount float64
	From   string
	To     string
}

type CurrencyConversionResponse struct {
	Amount   string // we return a nicely formatted string for the LLM's benefit
	Currency string
}

func init() {
	registerFunction(Registration{
		Definition: shared.FunctionDefinitionParam{
			Name:        "convert_currency",
			Description: openai.String("Convert an amount of one (real, non-crypto) currency to another. *Always* call this function to get exchange rates when doing currency conversion - never use memorised rates."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"amount": map[string]any{
						"type":        "number",
						"format":      "double",
						"description": "The amount of currency to convert.",
					},
					"from": map[string]any{
						"type":        "string",
						"description": "The currency code to convert from.",
					},
					"to": map[string]any{
						"type":        "string",
						"description": "The currency code to convert to.",
					},
				},
				"required": []string{"amount", "from", "to"},
			},
		},
		Fn:        convertCurrency,
		Thought:   convertCurrencyThought,
		InputType: CurrencyConversionRequest{},
	})
}

func convertCurrency(ctx context.Context, qt *quota.Tracker, input any) any {
	ctx, span := beeline.StartSpan(ctx, "convert_currency")
	defer span.Send()
	ccr := input.(*CurrencyConversionRequest)

	if !currencies.IsValidCurrency(ccr.From) {
		return Error{Error: "Unknown currency code " + ccr.From}
	}
	if !currencies.IsValidCurrency(ccr.To) {
		return Error{Error: "Unknown currency code " + ccr.To}
	}

	cdm := currencies.GetCurrencyDataManager()

	data, err := cdm.GetExchangeData(ctx, ccr.From)
	if err != nil {
		log.Printf("error getting currency data for %s/%s: %v", ccr.From, ccr.To, err)
		return Error{Error: err.Error()}
	}
	if data == nil {
		return Error{Error: "returned currency data is nil!?"}
	}

	rate, ok := data.ConversionRates[ccr.To]
	if !ok {
		return Error{Error: fmt.Sprintf("No currency conversion available from %s to %s", ccr.From, ccr.To)}
	}

	result := rate * ccr.Amount
	return &CurrencyConversionResponse{
		Amount:   fmt.Sprintf("%.2f", result),
		Currency: ccr.To,
	}
}

func convertCurrencyThought(i any) string {
	args := i.(*CurrencyConversionRequest)
	return fmt.Sprintf("Checking the %s/%s rate...", args.From, args.To)
}
