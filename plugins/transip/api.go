package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pbergman/logger"
)

func NewApiClient(config *Config, logger *logger.Logger) (*ApiClient, error) {

	var client = &ApiClient{
		config: config,
		state:  new(ApiClientState),
	}

	client.client = &http.Client{
		Transport: &TransIpApiTransport{
			inner:  http.DefaultTransport,
			client: client,
			logger: logger,
		},
	}

	return client, nil
}

type ApiClientState struct {
	Token *SignatureToken
}

type ApiClient struct {
	config *Config
	client *http.Client
	state  *ApiClientState
}

func (a *ApiClient) do(method string, uri string, object any, entry *DNSEntry) error {

	var body io.Reader

	if nil != entry {
		body = new(bytes.Buffer)

		if err := json.NewEncoder(body.(*bytes.Buffer)).Encode(map[string]*DNSEntry{"dnsEntry": entry}); err != nil {
			return err
		}
	}

	request, err := http.NewRequest(method, uri, body)

	if err != nil {
		return err
	}

	resp, err := a.client.Do(request)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {

		if nil != object {
			if err := json.NewDecoder(resp.Body).Decode(object); err != nil {
				return err
			}
		}

		return nil
	}

	var data struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	return &ApiError{Code: resp.StatusCode, Message: data.Error}
}

func (a *ApiClient) GetDomains() ([]string, error) {
	var list = make([]string, 0)
	var body struct {
		Domains []*struct{ Name string } `json:"Domains"`
	}

	if err := a.do("GET", "domains", &body, nil); err != nil {
		return nil, err
	}

	for i, c := 0, len(body.Domains); i < c; i++ {
		list = append(list, body.Domains[i].Name)
	}

	return list, nil
}

func (a *ApiClient) DeleteDNSEntry(domain string, entry *DNSEntry) (*DNSEntry, error) {
	var body struct {
		Entry *DNSEntry `json:"dnsEntry"`
	}

	if err := a.do("DELETE", "domains/"+domain+"/dns", &body, entry); err != nil {
		return nil, err
	}

	return body.Entry, nil
}

func (a *ApiClient) UpdateDNSEntry(domain string, entry *DNSEntry) (*DNSEntry, error) {

	var body struct {
		Entry *DNSEntry `json:"dnsEntry"`
	}

	if err := a.do("PATCH", "domains/"+domain+"/dns", &body, entry); err != nil {
		return nil, err
	}

	return body.Entry, nil
}

func (a *ApiClient) GetDNSEntries(domain string) ([]*DNSEntry, error) {

	var body struct {
		Entries []*DNSEntry `json:"dnsEntries"`
	}

	if err := a.do("GET", "domains/"+domain+"/dns", &body, nil); err != nil {
		return nil, err
	}

	return body.Entries, nil
}
