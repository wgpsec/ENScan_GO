package common

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/wgpsec/ENScan/common/utils/gologger"
)

func GetReq(url string, options *ENOptions) string {
	client := resty.New()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}

	client.Header = http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43"},
		"Accept":     {"text/html, application/xhtml+xml, image/jxr, */*"},
		"Cookie":     {options.ENConfig.Cookies.Aiqicha},
		"Referer":    {"https://aiqicha.baidu.com/"},
	}
	resp, err := client.R().Get(url)

	if err != nil {
		if options.Proxy != "" {
			client.RemoveProxy()
		}
		gologger.Errorf("【AQC】请求发生错误， %s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return GetReq(url, options)
	}
	if resp.StatusCode() == 200 {
		return string(resp.Body())
	} else if resp.StatusCode() == 403 {
		gologger.Errorf("【AQC】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode() == 401 {
		gologger.Errorf("【AQC】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode() == 302 {
		gologger.Errorf("【AQC】需要更新Cookie\n")
	} else if resp.StatusCode() == 404 {
		gologger.Errorf("【AQC】请求错误 404 %s \n", url)
	} else {
		gologger.Errorf("【AQC】未知错误 %d\n", resp.StatusCode())
	}
	return ""
}
