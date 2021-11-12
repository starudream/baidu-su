package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-sdk/lib/codec/json"
	"github.com/go-sdk/lib/cron"
	"github.com/go-sdk/lib/log"
)

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	BDUSS     string `json:"bduss"`
	Cron      string `json:"cron"`
	Certs     []Cert `json:"certs"`
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

var config = &Config{}

func StartJob() error {
	if err := initConfig(); err != nil {
		return err
	}

	c := cron.Default(nil)
	for i := 0; i < len(config.Certs); i++ {
		i := i
		t := config.Certs[i]
		c.Add(config.Cron, t.Name, func() { do(i, log.WithFields(log.Fields{"name": t.Name})) })
	}
	c.Start()

	return nil
}

func initConfig() error {
	bs, err := os.ReadFile(Path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, config)
	if err != nil {
		return err
	}

	if (config.AccessKey == "" || config.SecretKey == "") && config.BDUSS == "" {
		return fmt.Errorf("missing access key and secret key or bduss")
	}

	return nil
}

func do(i int, l *log.Entry) {
	cert := config.Certs[i]

	bsCrt, err := os.ReadFile(cert.CrtPath)
	if err != nil {
		l.Errorf("read file fail, %v", err)
		return
	}

	certBlock, _ := pem.Decode(bsCrt)
	if certBlock == nil {
		l.Errorf("check cert format fail, %v", err)
		return
	}

	c, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		l.Errorf("decode cert file fail, %v", err)
		return
	}

	l.Infof(
		"local cert info: %s(%s) {%s}-{%s}",
		c.Subject.CommonName, c.Issuer.CommonName, c.NotBefore, c.NotAfter,
	)

	bsKey, err := ioutil.ReadFile(cert.KeyPath)
	if err != nil {
		l.Errorf("read key file fail, %v", err)
		return
	}

	err = client().Init()
	if err != nil {
		l.Errorf("client init fail, %v", err)
		return
	}

	l.Infof("cert check")

	resp, err := client().GetCertificates(cert.Domain)
	if err != nil {
		l.Errorf("api call fail, %v", err)
		return
	}

	var certInfos []*RespCertInfo
	err = json.Unmarshal(json.MustMarshal(resp.Result), &certInfos)
	if err != nil {
		l.Errorf("decode cert info fail, %v", err)
		return
	}

	renew, exist, id := false, false, ""
	for _, info := range certInfos {
		if info.Info == cert.Name {
			exist, id = true, info.Id
			l.Infof(
				"baidu su cert info: %s(%s) {%s}-{%s}",
				info.HostsContent, info.Issuer, info.StartsOn, info.ExpiresOn,
			)
			if c.NotAfter.Before(time.Now()) {
				l.Infof("cert has expired")
				return
			}
			if !c.NotAfter.After(info.ExpiresOn) {
				l.Infof("cert not expire")
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
				l.Errorf("api call fail")
				return
			}
			l.Infof("cert deleted")
		}

		resp, err = client().PostCertificates(cert.Domain, cert.Name, string(bsCrt), string(bsKey))
		if err != nil {
			l.Errorf("api call fail")
			return
		}
		l.Infof("cert added")
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
