package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sdk/lib/codec/json"
	"github.com/go-sdk/lib/crypto"
	"github.com/go-sdk/lib/log"
)

type ClientApi struct {
	accessKey string
	secretKey string
	method    string
	path      string
	headers   map[string]string
	params    map[string]string
}

func NewClient() *ClientApi {
	return &ClientApi{
		headers: map[string]string{},
		params:  map[string]string{},
	}
}

func (c *ClientApi) SetAccessKey(accessKey string) *ClientApi {
	c.accessKey = accessKey
	return c
}

func (c *ClientApi) SetSecretKey(secretKey string) *ClientApi {
	c.secretKey = secretKey
	return c
}

func (c *ClientApi) SetMethod(method string) *ClientApi {
	c.method = method
	return c
}

func (c *ClientApi) SetPath(path string) *ClientApi {
	c.path = path
	return c
}

func (c *ClientApi) SetParams(params map[string]string) *ClientApi {
	c.params = map[string]string{}
	for k, v := range params {
		c.params[k] = v
	}
	return c
}

func (c *ClientApi) AddParams(kv ...string) *ClientApi {
	if len(kv)%2 != 0 {
		return c
	}
	for i := 0; i < len(kv); i += 2 {
		c.params[kv[i]] = kv[i+1]
	}
	return c
}

func (c *ClientApi) Do() (*Response, error) {
	c.headers = map[string]string{
		"X-Auth-Access-Key":       c.accessKey,
		"X-Auth-Nonce":            crypto.RandString(24, crypto.CharsetLetterLower),
		"X-Auth-Path-Info":        c.path,
		"X-Auth-Signature-Method": SignatureMethod,
		"X-Auth-Timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
	}

	c.headers["X-Auth-Sign"] = c.sign(c.build(c.headers)+"&"+c.build(c.params), c.secretKey)

	req, err := http.NewRequest(c.method, URLApi+c.path, bytes.NewReader(json.MustMarshal(c.params)))
	if err != nil {
		return nil, fmt.Errorf("build request fail: %v", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: time.Second * 30}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request fail: %v", err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response fail: %v", err)
	}

	log.Debugf("[baidusu] method: %s, url: %s, req: %s, code: %d, resp: %s", c.method, URLApi+c.path, json.MustMarshal(c.params), resp.StatusCode, RebuildResponse(bs))

	data := &Response{}
	err = json.Unmarshal(bs, data)
	if err != nil {
		return nil, fmt.Errorf("decode response fail: %v", err)
	}

	if !data.Success {
		return data, fmt.Errorf("api call return error")
	}

	data.Raw = bs

	return data, nil
}

func (c *ClientApi) Init() error {
	return nil
}

func (c *ClientApi) GetCertificates(domain string) (*Response, error) {
	return c.SetMethod(http.MethodGet).SetPath("v3/yjs/custom_certificates").SetParams(map[string]string{"domain": domain}).Do()
}

func (c *ClientApi) DeleteCertificates(domain, _, name string) (*Response, error) {
	return c.SetMethod(http.MethodDelete).SetPath("v3/yjs/custom_certificates").SetParams(map[string]string{"domain": domain, "info": name}).Do()
}

func (c *ClientApi) PostCertificates(domain, name, crt, key string) (*Response, error) {
	return c.SetMethod(http.MethodPost).SetPath("v3/yjs/custom_certificates").SetParams(map[string]string{"domain": domain, "info": name, "certificate": crt, "private_key": key}).Do()
}

func (c *ClientApi) build(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	bb := strings.Builder{}
	for _, k := range ks {
		bb.WriteString(k)
		bb.WriteString("=")
		bb.WriteString(m[k])
		bb.WriteString("&")
	}
	return bb.String()[:bb.Len()-1]
}

func (c *ClientApi) sign(content, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(content))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
