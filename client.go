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

	"github.com/go-sdk/logx"
	"github.com/go-sdk/utilx/json"
)

type Client struct {
	accessKey string
	secretKey string
	method    string
	path      string
	headers   map[string]string
	params    map[string]string
}

type Resp struct {
	Success bool          `json:"success,omitempty"`
	Errors  []interface{} `json:"errors,omitempty"`
	Message []interface{} `json:"message,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Raw     []byte        `json:"-"`
}

func NewClient() *Client {
	return &Client{
		headers: map[string]string{},
		params:  map[string]string{},
	}
}

func (c *Client) SetAccessKey(accessKey string) *Client {
	c.accessKey = accessKey
	return c
}

func (c *Client) SetSecretKey(secretKey string) *Client {
	c.secretKey = secretKey
	return c
}

func (c *Client) SetMethod(method string) *Client {
	c.method = method
	return c
}

func (c *Client) SetPath(path string) *Client {
	c.path = path
	return c
}

func (c *Client) SetParams(params map[string]string) *Client {
	c.params = map[string]string{}
	for k, v := range params {
		c.params[k] = v
	}
	return c
}

func (c *Client) AddParams(kv ...string) *Client {
	if len(kv)%2 != 0 {
		return c
	}
	for i := 0; i < len(kv); i += 2 {
		c.params[kv[i]] = kv[i+1]
	}
	return c
}

func (c *Client) Do() (*Resp, error) {
	c.headers = map[string]string{
		"X-Auth-Access-Key":       c.accessKey,
		"X-Auth-Nonce":            nonce(24),
		"X-Auth-Path-Info":        c.path,
		"X-Auth-Signature-Method": SignatureMethod,
		"X-Auth-Timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
	}

	c.headers["X-Auth-Sign"] = sign(build(c.headers)+"&"+build(c.params), c.secretKey)

	req, err := http.NewRequest(c.method, URL+c.path, bytes.NewReader(json.MustMarshal(c.params)))
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

	data := &Resp{}
	err = json.Unmarshal(bs, data)
	if err != nil {
		return nil, fmt.Errorf("decode response fail: %v", err)
	}

	logx.Debugf("[baidusu] response: %s", json.MustMarshal(data))

	if !data.Success {
		return data, fmt.Errorf("api call return error")
	}

	data.Raw = bs

	return data, nil
}

func build(m map[string]string) string {
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

func sign(content, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(content))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
