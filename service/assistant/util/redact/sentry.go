package redact

import (
	"net/url"
	"strings"

	"github.com/getsentry/sentry-go"
)

func redactQuery(queryString string) string {
	vals, err := url.ParseQuery(queryString)
	if err != nil {
		return queryString
	}
	for key := range vals {
		vals.Set(key, "[redacted]")
	}
	return vals.Encode()
}

func cleanUrl(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.RawQuery = ""
	u.Fragment = ""
	if u.Host != "" {
		parts := strings.Split(u.Host, ".")
		if len(parts) > 2 {
			u.Host = strings.Join(parts[len(parts)-2:], ".")
		}
	}
	return u.String()
}

func BeforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if event.Request != nil {
		if event.Request.QueryString != "" {
			event.Request.QueryString = redactQuery(event.Request.QueryString)
		}
		if event.Request.URL != "" {
			event.Request.URL = cleanUrl(event.Request.URL)
		}
	}
	for i, breadcrumb := range event.Breadcrumbs {
		if breadcrumb.Data != nil {
			if q, ok := breadcrumb.Data["query"]; ok {
				if qs, ok := q.(string); ok {
					breadcrumb.Data["query"] = redactQuery(qs)
				}
			}
			if u, ok := breadcrumb.Data["url"]; ok {
				if us, ok := u.(string); ok {
					breadcrumb.Data["url"] = cleanUrl(us)
				}
			}
		}
		event.Breadcrumbs[i] = breadcrumb
	}
	return event
}
