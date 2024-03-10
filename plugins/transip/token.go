package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type SignatureToken string

func NewSignatureToken(config *Config) (*SignatureToken, error) {

	config.Signature.Nonce = strconv.Itoa(int(time.Now().Unix()))
	payload, err := json.Marshal(config.Signature)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", config.BaseUrl.JoinPath("auth").String(), bytes.NewReader(payload))

	if err != nil {
		return nil, err
	}

	signature, err := config.PrivateKey.Sign(payload)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var body struct {
		Token SignatureToken `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body.Token, nil

}
