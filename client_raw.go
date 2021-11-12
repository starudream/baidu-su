package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-sdk/lib/codec/json"
	"github.com/go-sdk/lib/log"
)

type ClientRaw struct {
	bduss  string
	method string
	path   string
	params map[string]string
}

type ZoneRaw struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

var zoneNameId = map[string]string{} // name -> id

func NewClientRaw() *ClientRaw {
	return &ClientRaw{
		params: map[string]string{},
	}
}

func (c *ClientRaw) SetBDUSS(bduss string) *ClientRaw {
	c.bduss = bduss
	return c
}

func (c *ClientRaw) SetMethod(method string) *ClientRaw {
	c.method = method
	return c
}

func (c *ClientRaw) SetPath(path string) *ClientRaw {
	c.path = path
	return c
}

func (c *ClientRaw) SetParams(params map[string]string) *ClientRaw {
	c.params = map[string]string{}
	for k, v := range params {
		c.params[k] = v
	}
	return c
}

func (c *ClientRaw) AddParams(kv ...string) *ClientRaw {
	if len(kv)%2 != 0 {
		return c
	}
	for i := 0; i < len(kv); i += 2 {
		c.params[kv[i]] = kv[i+1]
	}
	return c
}

func (c *ClientRaw) Do() (*Response, error) {
	m := c.method
	if c.method == http.MethodDelete {
		m = http.MethodPost
	}

	u := URLRaw + fmt.Sprintf("seed_%d", time.Now().UnixNano()/1e6) + "/" + c.path

	req, err := http.NewRequest(m, u, bytes.NewReader(json.MustMarshal(c.params)))
	if err != nil {
		return nil, fmt.Errorf("build request fail: %v", err)
	}

	req.AddCookie(&http.Cookie{Name: "BDUSS", Value: c.bduss})
	req.Header.Set("referer", "https://su.baidu.com/console/index.html")
	if c.method == http.MethodDelete {
		req.Header.Set("x-http-method-override", "DELETE")
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

	log.Debugf("[baidusu] method: %s, url: %s, req: %s, code: %d, resp: %s", m, u, json.MustMarshal(c.params), resp.StatusCode, RebuildResponse(bs))

	data := &Response{}
	err = json.Unmarshal(bs, data)
	if err != nil {
		return nil, fmt.Errorf("decode response fail: %v", err)
	}

	if !data.Success {
		return data, fmt.Errorf("raw call return error")
	}

	data.Raw = bs

	return data, nil
}

func (c *ClientRaw) Init() error {
	resp, err := c.GetZones("")
	if err != nil {
		return err
	}

	var zoneInfos []ZoneRaw
	err = json.Unmarshal(json.MustMarshal(resp.Result), &zoneInfos)
	if err != nil {
		return fmt.Errorf("decode zones fail: %v", err)
	}

	for _, info := range zoneInfos {
		zoneNameId[info.Name] = info.Id
	}
	return nil
}

func (c *ClientRaw) GetZones(_ string) (*Response, error) {
	return c.SetMethod(http.MethodGet).SetPath("api/su/zones?page=1&per_page=100").Do()
}

func (c *ClientRaw) GetCertificates(domain string) (*Response, error) {
	zone, err := c.getZoneId(domain)
	if err != nil {
		return nil, err
	}
	return c.SetMethod(http.MethodGet).SetPath("api/su/zones/" + zone + "/custom_certificates?page=1&per_page=100").Do()
}

func (c *ClientRaw) DeleteCertificates(domain, id, _ string) (*Response, error) {
	zone, err := c.getZoneId(domain)
	if err != nil {
		return nil, err
	}
	return c.SetMethod(http.MethodDelete).SetPath("api/su/zones/" + zone + "/custom_certificates/" + id).Do()
}

func (c *ClientRaw) PostCertificates(domain, name, crt, key string) (*Response, error) {
	zone, err := c.getZoneId(domain)
	if err != nil {
		return nil, err
	}
	return c.SetMethod(http.MethodPost).SetPath("api/su/zones/" + zone + "/custom_certificates").SetParams(map[string]string{"info": name, "certificate": crt, "private_key": key}).Do()
}

func (c *ClientRaw) getZoneId(name string) (string, error) {
	if v, ok := zoneNameId[name]; ok {
		return v, nil
	} else {
		return "", fmt.Errorf("not found domain id")
	}
}
