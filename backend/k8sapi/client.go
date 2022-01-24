// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.
// Created at 2022/01/24 by Vincent Landgraf

package k8sapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/caarlos0/env"
	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/log"
)

// Client minimal client for the kubernetes API
type Client struct {
	Podname    string
	Namespace  string
	CACert     []byte
	Token      string
	cfg        *Config
	HttpClient *http.Client
}

// NewClient create new api client
func NewClient() (*Client, error) {
	var cl Client

	// lookup hostname (for pod update)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	cl.Podname = hostname

	// parse environment including secrets mounted by kubernetes
	err = env.Parse(&cl.cfg)
	if err != nil {
		return nil, err
	}

	caData, err := os.ReadFile(cl.cfg.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", cl.cfg.CACertFile, err)
	}
	cl.CACert = []byte(strings.TrimSpace(string(caData)))

	namespaceData, err := os.ReadFile(cl.cfg.NamespaceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", cl.cfg.NamespaceFile, err)
	}
	cl.Namespace = strings.TrimSpace(string(namespaceData))

	tokenData, err := os.ReadFile(cl.cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", cl.cfg.CACertFile, err)
	}
	cl.Token = strings.TrimSpace(string(tokenData))

	// add kubernetes api server cert
	chain := transport.NewDefaultTransportChain()
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(cl.CACert)
	if !ok {
		return nil, fmt.Errorf("failed to load kubernetes ca cert")
	}
	chain.Final(&http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	})
	cl.HttpClient.Transport = chain

	return &cl, nil
}

// SimpleRequest send a simple http request to kubernetes with the passed
// method, url and requestObj, decoding the result into responseObj
func (c *Client) SimpleRequest(ctx context.Context, method, url string, requestObj, responseObj interface{}) error {
	data, err := json.Marshal(requestObj)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json-patch+json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to do api request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body) // nolint: errcheck
		log.Ctx(ctx).Debug().Msgf("failed to do api request, due to: %s", string(body))
		return fmt.Errorf("k8s request failed with %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(responseObj)
}
