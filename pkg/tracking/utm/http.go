package utm

import (
	"net/http"

	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/http/transport"
)

var emptyData = UTMData{}

func FromRequest(req *http.Request) (UTMData, error) {
	data := UTMData{
		Source:   req.URL.Query().Get("utm_source"),
		Medium:   req.URL.Query().Get("utm_medium"),
		Campaign: req.URL.Query().Get("utm_campaign"),
		Term:     req.URL.Query().Get("utm_term"),
		Content:  req.URL.Query().Get("utm_content"),
		Client:   req.URL.Query().Get("utm_partner_client"),
	}
	if data == emptyData {
		return emptyData, ErrNotFound
	}
	return data, nil
}

func AttachToRequest(data UTMData, req *http.Request) *http.Request {
	if data == emptyData {
		return req
	}
	q := req.URL.Query()
	q.Set("utm_source", data.Source)
	q.Set("utm_medium", data.Medium)
	q.Set("utm_campaign", data.Campaign)
	q.Set("utm_term", data.Term)
	q.Set("utm_content", data.Content)
	q.Set("utm_partner_client", data.Client)
	req.URL.RawQuery = q.Encode()
	return req
}

// Middleware attempts to attach utm data found in the request to the request context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := FromRequest(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			clientID, found := oauth2.ClientID(r.Context())
			if found && data.Client == "" {
				data.Client = clientID
			}
			ctx := ContextWithUTMData(r.Context(), data)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

var (
	_ http.RoundTripper               = (*RoundTripper)(nil)
	_ transport.ChainableRoundTripper = (*RoundTripper)(nil)
)

type RoundTripper struct {
	transport http.RoundTripper
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	data, found := FromContext(req.Context())
	if !found { // no utm data found, skip directly to next roundtripper
		return r.transport.RoundTrip(req)
	}
	newReq := cloneRequest(req)
	newReq = AttachToRequest(data, newReq)
	return r.transport.RoundTrip(newReq)
}

func (r *RoundTripper) Transport() http.RoundTripper {
	return r.transport
}

func (r *RoundTripper) SetTransport(tripper http.RoundTripper) {
	r.transport = tripper
}

// Taken from https://github.com/golang/oauth2/blob/9fd604954f58/transport.go#L32
// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
