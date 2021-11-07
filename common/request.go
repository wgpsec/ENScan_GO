package common

import (
	"github.com/wgpsec/ENScan/common/requests"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"io/ioutil"
	"net/http"
	"time"
)

func GetReq(url string, options *ENOptions) []byte {
	var transport = requests.DefaultTransport()
	var client = &http.Client{
		Transport: transport,
		//Timeout:       time.Duration(options.Timeout),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不进入重定向 */
		},
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"},
		"Accept":     {"text/html, application/xhtml+xml, image/jxr, */*"},
		"Cookie":     {options.CookieInfo},
		//"Accept-Encoding": {"gzip, deflate"},
		"Referer": {"https://www.baidu.com"},
	}
	resp, err := client.Do(req)
	if err != nil {
		gologger.Errorf("请求发生错误，5秒后重试\n%s\n", err)
		time.Sleep(5 * time.Second)
		return GetReq(url, options)
	}
	if resp.StatusCode == 403 {
		gologger.Fatalf("ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 401 {
		gologger.Fatalf("Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode == 302 {
		gologger.Fatalf("需要更新Cookie\n")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	//page, _ := htmlquery.Parse(strings.NewReader(string(body)))
	return body
}
