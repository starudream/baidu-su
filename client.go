package main

import (
	"github.com/go-sdk/lib/codec/json"
)

type Client interface {
	Init() error
	Do() (*Response, error)
	GetCertificates(domain string) (*Response, error)
	DeleteCertificates(domain, id, name string) (*Response, error)
	PostCertificates(domain, name, crt, key string) (*Response, error)
}

type Response struct {
	Success bool        `json:"success,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Raw     []byte      `json:"-"`
}

func RebuildResponse(bs []byte) string {
	var i interface{}
	err := json.Unmarshal(bs, &i)
	if err != nil {
		return string(bs)
	}
	return json.MustMarshalToString(i)
}
