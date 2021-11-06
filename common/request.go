package common

import (
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/requests"
	"io/ioutil"
	"net/http"
	"time"
)

func GetReq(url string) []byte {
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
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"},
		"Accept":     []string{"text/html, application/xhtml+xml, image/jxr, */*"},
		"Cookie":     []string{""},
		//"Accept-Encoding": []string{"gzip, deflate"},
		"Referer": []string{"https://www.baidu.com"},
	}
	resp, err := client.Do(req)
	if err != nil {
		gologger.Fatalf("请求发生错误，请检查网络连接\n%s\n", err)
		time.Sleep(5)
		GetReq(url)
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
