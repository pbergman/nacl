package main

import (
	"fmt"
	"net/url"

	"github.com/BurntSushi/toml"
)

type Config struct {
	BaseUrl    *Uri              `toml:"base_url"`
	Signature  *SignatureRequest `toml:"signature"`
	PrivateKey *PrivateKey       `toml:"private_key"`
	Hosts      []string
	Ipv6       bool
}

func getConfig(prim toml.Primitive, meta *toml.MetaData) (*Config, error) {

	var config *Config

	if err := meta.PrimitiveDecode(prim, &config); err != nil {
		return nil, err
	}

	if 0 == len(config.Hosts) {
		return nil, fmt.Errorf("no records defined")
	}

	if nil == config.BaseUrl {
		uri, err := url.Parse("https://api.transip.nl/v6/")

		// should not happen...
		if err != nil {
			return nil, err
		}

		config.BaseUrl = &Uri{URL: *uri}
	}

	if nil == config.PrivateKey {
		return nil, fmt.Errorf("transip: missing private key for creating access tokens")
	}

	if nil == config.Signature {
		config.Signature = new(SignatureRequest)
	}

	if "" == config.Signature.Login {
		return nil, fmt.Errorf("transip: missing required user in config")
	}

	return config, nil
}
