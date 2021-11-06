package requests

import (
	"crypto/tls"
	"golang.org/x/net/html"
	"net"
	"net/http"
	"time"
)

type Request struct {
	Url    string
	Cookie string
}

type Response struct {
	Body []byte
	Page *html.Node
}

func DefaultTransport() *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}
	return transport
}
