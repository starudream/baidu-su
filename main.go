package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-sdk/logx"
	"github.com/go-sdk/utilx/json"
	"github.com/robfig/cron/v3"
)

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Cron      string `json:"cron"`
	Timezone  string `json:"timezone"`
	Certs     []Cert `json:"certs"`
}

type Cert struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	CrtPath string `json:"crt_path"`
	KeyPath string `json:"key_path"`
}

type RespCertInfo struct {
	Info         string    `json:"info"`
	Issuer       string    `json:"issuer"`
	HostsContent string    `json:"hosts_content"`
	StartsOn     time.Time `json:"starts_on"`
	ExpiresOn    time.Time `json:"expires_on"`
}

var (
	config = Config{}

	c string
	d bool

	location *time.Location
	schedule cron.Schedule
)

func init() {
	flag.StringVar(&c, "config", "config.json", "config")
	flag.BoolVar(&d, "debug", strings.ToLower(os.Getenv("DEBUG")) == "true", "debug")
	flag.Parse()

	if !d {
		logx.SetLevel(logx.InfoLevel)
	}

	bs, err := ioutil.ReadFile(c)
	if err != nil {
		exit("[init] read file fail", err)
	}

	err = json.Unmarshal(bs, &config)
	if err != nil {
		exit("[init] decode file fail", err)
	}

	location, err = time.LoadLocation(config.Timezone)
	if err != nil {
		exit("[init] timezone parse fail", err)
	}
	schedule, err = cron.ParseStandard(config.Cron)
	if err != nil {
		exit("[init] cron parse fail", err)
	}

	logx.Debugf("[init] config: %s", json.MustMarshal(config))
}

func main() {
	for i := range config.Certs {
		go handle(i)
	}
	select {}
}

func handle(i int) {
	cert := config.Certs[i]

	c := cron.New(cron.WithLocation(location), cron.WithLogger(cron.VerbosePrintfLogger(&Log{i: i, name: cert.Name})))
	c.Schedule(schedule, cron.FuncJob(func() { do(i) }))
	c.Start()
}

func do(i int) {
	cert := config.Certs[i]

	bsCrt, err := ioutil.ReadFile(cert.CrtPath)
	if err != nil {
		exit("[%s:%d] read cert file fail", err)
		return
	}

	certBlock, _ := pem.Decode(bsCrt)
	if certBlock == nil {
		exit("[%s:%d] check cert format fail", nil)
		return
	}

	c, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		exit("[%s:%d] decode cert file fail", err)
		return
	}

	logx.Infof(
		"[%s:%d] local cert info: %s(%s) {%s}-{%s}",
		cert.Name, i, c.Subject.CommonName, c.Issuer.CommonName, c.NotBefore, c.NotAfter,
	)

	bsKey, err := ioutil.ReadFile(cert.KeyPath)
	if err != nil {
		exit("[%s:%d] read key file fail", err)
		return
	}

	logx.Infof("[%s:%d] cert check", cert.Name, i)

	resp, err := client().
		SetMethod(MethodGet).
		SetPath("v3/yjs/custom_certificates").
		SetParams(map[string]string{
			"domain": cert.Domain,
		}).
		Do()
	if err != nil {
		exit("[%s:%d] api call fail", err)
		return
	}

	certInfos := []*RespCertInfo{}
	err = json.Unmarshal(json.MustMarshal(resp.Result), &certInfos)
	if err != nil {
		exit("[%s:%d] decode cert info fail", err)
		return
	}

	renew, exist := false, false
	for _, info := range certInfos {
		if info.Info == cert.Name {
			exist = true
			logx.Infof(
				"[%s:%d] baidu su cert info: %s(%s) {%s}-{%s}",
				cert.Name, i, info.HostsContent, info.Issuer, info.StartsOn, info.ExpiresOn,
			)
			if c.NotAfter.Before(time.Now()) {
				logx.Infof("[%s:%d] cert has expired", cert.Name, i)
				return
			}
			if !c.NotAfter.After(info.ExpiresOn) {
				logx.Infof("[%s:%d] cert not expire", cert.Name, i)
				return
			}
			renew = true
		}
	}
	if !exist {
		renew = !exist
	}

	if renew {
		if exist {
			resp, err = client().
				SetMethod(MethodDelete).
				SetPath("v3/yjs/custom_certificates").
				SetParams(map[string]string{
					"domain": cert.Domain,
					"info":   cert.Name,
				}).
				Do()
			if err != nil {
				exit("[%s:%d] api call fail", err)
				return
			}
			logx.Infof("[%s:%d] cert deleted", cert.Name, i)
		}

		resp, err = client().
			SetMethod(MethodPost).
			SetPath("v3/yjs/custom_certificates").
			SetParams(map[string]string{
				"domain":      cert.Domain,
				"info":        cert.Name,
				"certificate": string(bsCrt),
				"private_key": string(bsKey),
			}).
			Do()
		if err != nil {
			exit("[%s:%d] api call fail", err)
			return
		}
		logx.Infof("[%s:%d] cert added", cert.Name, i)
	}
}

func client() *Client {
	return NewClient().SetAccessKey(config.AccessKey).SetSecretKey(config.SecretKey)
}
