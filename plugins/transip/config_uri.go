package main

import "net/url"

type Uri struct {
	url.URL
}

func (u *Uri) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))

	if err != nil {
		return err
	}

	u.URL = *parsed

	return nil
}
