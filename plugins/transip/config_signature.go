package main

type SignatureRequest struct {
	Login          string `json:"login" toml:"login"`
	Nonce          string `json:"nonce"`
	ReadOnly       bool   `json:"read_only,omitempty" toml:"read_only"`
	ExpirationTime string `json:"expiration_time,omitempty" toml:"expiration"`
	Label          string `json:"label,omitempty" toml:"label"`
	GlobalKey      bool   `json:"global_key" toml:"global_key"`
}
