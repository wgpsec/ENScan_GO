package common

import (
	"crypto/tls"
	"github.com/imroc/req/v3"
	"time"
)

func NewClient(headers map[string]string, op *ENOptions) *req.Request {
	c := req.C()
	c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}) // 忽略证书验证
	c.SetTLSFingerprintChrome()
	c.SetCommonHeaders(headers)
	if op.GetENConfig().UserAgent != "" {
		c.SetUserAgent(op.GetENConfig().UserAgent)
	}
	if op.Proxy != "" {
		c.SetProxyURL(op.Proxy)
	}
	c.SetTimeout(time.Duration(op.TimeOut) * time.Minute)
	time.Sleep(time.Duration(op.GetDelayRTime()) * time.Second)
	return c.R()
}
