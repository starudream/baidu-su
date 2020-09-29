package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-sdk/logx"
	"github.com/go-sdk/utilx/json"
	"github.com/robfig/cron/v3"
)

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	BDUSS     string `json:"bduss"`
	Cron      string `json:"cron"`
	Timezone  string `json:"timezone"`
	Certs     []Cert `json:"certs"`

	Path string `json:"-"`
	Help bool   `json:"-"`
}

type Cert struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	CrtPath string `json:"crt_path"`
	KeyPath string `json:"key_path"`
}

type RespCertInfo struct {
	Id           string    `json:"id"`
	Info         string    `json:"info"`
	Issuer       string    `json:"issuer"`
	HostsContent string    `json:"hosts_content"`
	StartsOn     time.Time `json:"starts_on"`
	ExpiresOn    time.Time `json:"expires_on"`
}

var (
	config = &Config{}

	location *time.Location
	schedule cron.Schedule
)

func init() {
	flag.StringVar(&config.Path, "config", "config.json", "config")
	flag.BoolVar(&config.Help, "help", false, "instructions for use")
	flag.Parse()

	if config.Help {
		flag.Usage()
		os.Exit(0)
	}

	bs, err := ioutil.ReadFile(config.Path)
	if err != nil {
		logx.WithField("err", err).Fatal("[init] read file fail")
	}

	err = json.Unmarshal(bs, config)
	if err != nil {
		logx.WithField("err", err).Fatal("[init] decode file fail")
	}

	logx.Infof("[config] %s", json.MustMarshal(config))

	location, err = time.LoadLocation(config.Timezone)
	if err != nil {
		logx.WithField("err", err).Fatal("[init] timezone parse fail")
	}
	schedule, err = cron.ParseStandard(config.Cron)
	if err != nil {
		logx.WithField("err", err).Fatal("[init] cron parse fail")
	}

	if (config.AccessKey == "" || config.SecretKey == "") && config.BDUSS == "" {
		logx.Fatal("[config] access or secret key and bduss is empty")
	}
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
		logx.WithField("err", err).Errorf("[%s:%d] read cert file fail", cert.Name, i)
		return
	}

	certBlock, _ := pem.Decode(bsCrt)
	if certBlock == nil {
		logx.Errorf("[%s:%d] check cert format fail", cert.Name, i)
		return
	}

	c, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		logx.WithField("err", err).Errorf("[%s:%d] decode cert file fail", cert.Name, i)
		return
	}

	logx.Infof(
		"[%s:%d] local cert info: %s(%s) {%s}-{%s}",
		cert.Name, i, c.Subject.CommonName, c.Issuer.CommonName, c.NotBefore, c.NotAfter,
	)

	bsKey, err := ioutil.ReadFile(cert.KeyPath)
	if err != nil {
		logx.WithField("err", err).Errorf("[%s:%d] read key file fail", cert.Name, i)
		return
	}

	err = client().Init()
	if err != nil {
		logx.WithField("err", err).Errorf("[%s:%d] client init fail", cert.Name, i)
		return
	}

	logx.Infof("[%s:%d] cert check", cert.Name, i)

	resp, err := client().GetCertificates(cert.Domain)
	if err != nil {
		logx.WithField("err", err).Errorf("[%s:%d] api call fail", cert.Name, i)
		return
	}

	certInfos := []*RespCertInfo{}
	err = json.Unmarshal(json.MustMarshal(resp.Result), &certInfos)
	if err != nil {
		logx.WithField("err", err).Errorf("[%s:%d] decode cert info fail", cert.Name, i)
		return
	}

	renew, exist, id := false, false, ""
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
			resp, err = client().DeleteCertificates(cert.Domain, id, cert.Name)
			if err != nil {
				logx.WithField("err", err).Errorf("[%s:%d] api call fail", cert.Name, i)
				return
			}
			logx.Infof("[%s:%d] cert deleted", cert.Name, i)
		}

		resp, err = client().PostCertificates(cert.Domain, cert.Name, string(bsCrt), string(bsKey))
		if err != nil {
			logx.WithField("err", err).Errorf("[%s:%d] api call fail", cert.Name, i)
			return
		}
		logx.Infof("[%s:%d] cert added", cert.Name, i)
	}
}

func client() Client {
	if config.AccessKey != "" && config.SecretKey != "" {
		return NewClient().SetAccessKey(config.AccessKey).SetSecretKey(config.SecretKey)
	} else if config.BDUSS != "" {
		return NewClientRaw().SetBDUSS(config.BDUSS)
	} else {
		return nil
	}
}
