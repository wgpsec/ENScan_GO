package hook

import (
	"crypto/tls"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func vGetReq(url string, data map[string]string, options *common.ENOptions) string {
	//安全延时
	time.Sleep(time.Duration(options.DelayTime) * time.Second)

	client := resty.New()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.Header = http.Header{
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36"},
		"Content-Type": {"application/json;charset=UTF-8"},
		"Cookie":       {options.ENConfig.Cookies.Veryvp},
		"Referer":      {"http://www.veryvp.com/"},
	}
	clientR := client.R()
	if data["key"] != "" {
		clientR.Method = "GET"
	} else {
		clientR.Method = "POST"
		clientR.SetFormData(data)
	}

	clientR.URL = url
	resp, err := clientR.Send()
	if err != nil {
		gologger.Errorf("【vp】请求发生错误，%s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return vGetReq(url, data, options)
	}
	if resp.StatusCode() == 200 {
		if !strings.Contains(string(resp.Body()), "登录") {
			return strings.ReplaceAll(string(resp.Body()), "\\", "")
		} else {
			gologger.Errorf("【vp】需要登陆操作\n%s\n", err)
			return ""
		}

	} else if resp.StatusCode() == 403 {
		gologger.Errorf("【vp】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode() == 401 {
		gologger.Errorf("【vp】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode() == 301 {
		gologger.Errorf("【vp】需要更新Cookie\n")
	} else if resp.StatusCode() == 404 {
		gologger.Errorf("【vp】请求错误 404 %s\n", url)
	} else if resp.StatusCode() == 500 {
		gologger.Errorf("【vp】请求数据错误，5S后重试\n%s\n", url)
		time.Sleep(5 * time.Second)
		return vGetReq(url, data, options)
	} else {
		gologger.Errorf("【vp】未知错误 %s\n", resp.StatusCode())
		return ""
	}

	if strings.Contains(string(resp.Body()), "使用该功能需要用户登录") {
		gologger.Errorf("【vp】Cookie有问题或过期，请重新获取\n")
	}
	return ""
}

func VPGetEmail(Key string, options *common.ENOptions) (resList []gjson.Result) {
	reqData := map[string]string{
		"Key":      Key,
		"PageNo":   "1",
		"Order":    "",
		"PageSize": "1000",
	}

	r := vGetReq("http://www.veryvp.com/SearchEmail/CreateSearch", map[string]string{"domain": Key}, options)
	if gjson.Get(r, "Type").Int() == 0 {
		time.Sleep(1 * time.Second)
	} else {
		gologger.Errorf("任务添加失败 %s", Key)
	}

	res := vGetReq("http://www.veryvp.com/SearchEmail/GetEmailList", reqData, options)
	resUn, _ := url.QueryUnescape(gjson.Get(res, "Data").String())
	resList = append(resList, gjson.Parse(resUn).Array()...)
	gologger.Infof("VP查询到%s条记录\n", gjson.Get(res, "RecordCount"))
	pages := gjson.Get(res, "PageCount").Int()
	//pages = 10
	for i := 2; i <= int(pages); i++ {
		reqData["PageNo"] = strconv.Itoa(i)
		resUn, _ := url.QueryUnescape(gjson.Get(res, "Data").String())
		resList = append(resList, gjson.Parse(resUn).Array()...)
		time.Sleep(1 * time.Second)
	}

	return resList
}
