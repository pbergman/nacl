package main

type DNSEntry struct {
	Name    string `json:"name"`
	Expire  int    `json:"expire"`
	Type    string `json:"type"`
	Content string `json:"content"`
}
