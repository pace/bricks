package utm

import (
	"bytes"
	"context"
	"fmt"
)

type ctxKey struct{}

var key = ctxKey{}

// https://en.wikipedia.org/wiki/UTM_parameters
type UTMData struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string

	Client string // Unofficial pace utm parameter
}

func (d UTMData) ToMap() map[string]string {
	return map[string]string{
		"utm_source":         d.Source,
		"utm_medium":         d.Medium,
		"utm_campaign":       d.Campaign,
		"utm_term":           d.Term,
		"utm_content":        d.Content,
		"utm_partner_client": d.Client,
	}
}

func (d UTMData) ToString() string {
	b := new(bytes.Buffer)

	fmt.Fprintf(b, "utm_source=\"%s\"\n", d.Source)
	fmt.Fprintf(b, "utm_medium=\"%s\"\n", d.Medium)
	fmt.Fprintf(b, "utm_campaign=\"%s\"\n", d.Campaign)
	fmt.Fprintf(b, "utm_term=\"%s\"\n", d.Term)
	fmt.Fprintf(b, "utm_content=\"%s\"\n", d.Content)
	fmt.Fprintf(b, "utm_partner_client=\"%s\"\n", d.Client)

	return b.String()
}

func FromMap(m map[string]string) UTMData {
	return UTMData{
		Source:   m["utm_source"],
		Medium:   m["utm_medium"],
		Campaign: m["utm_campaign"],
		Term:     m["utm_term"],
		Content:  m["utm_content"],
		Client:   m["utm_partner_client"],
	}
}

func ContextWithUTMData(parentCtx context.Context, data UTMData) context.Context {
	return context.WithValue(parentCtx, key, data)
}

func FromContext(ctx context.Context) (UTMData, bool) {
	val := ctx.Value(key)
	data, found := val.(UTMData)
	return data, found
}
