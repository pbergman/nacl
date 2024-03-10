package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pbergman/logger"
)

type TransIpApiTransport struct {
	inner  http.RoundTripper
	client *ApiClient
	logger *logger.Logger
}

func (t *TransIpApiTransport) getAuthorization(force bool) (string, error) {
	if t.client.state.Token == nil || force {
		token, err := NewSignatureToken(t.client.config)

		if err != nil {
			return "", err
		}

		t.client.state.Token = token
	}
	return "Bearer " + string(*t.client.state.Token), nil
}

func (t *TransIpApiTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.URL.Scheme == "" {
		req.URL = t.client.config.BaseUrl.JoinPath(strings.Split(req.URL.Path, "/")...)
	}
	for i, c := 0, 5; i < c; i++ {
		auth, err := t.getAuthorization(i > 0)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", auth)
		req.Header.Set("Content-Type", "application/json")

		resp, err = t.inner.RoundTrip(req)

		if err != nil {
			t.logger.Error(fmt.Sprintf("%s %s %s", req.Method, req.URL.Path, req.Proto))
			return nil, err
		}

		t.logger.Debug(fmt.Sprintf("%s %s %s %d", req.Method, req.URL.Path, req.Proto, resp.StatusCode))

		if resp.StatusCode != 401 {
			break
		}
	}
	return resp, err
}
