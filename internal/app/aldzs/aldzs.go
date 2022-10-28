package aldzs

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils/gologger"
)

func getReq(searchType string, data map[string]string) gjson.Result {
	url := fmt.Sprintf("https://zhishuapi.aldwx.com/Main/action/%s", searchType)
	client := resty.New()
	client.SetTimeout(common.RequestTimeOut)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.Header = http.Header{
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"},
		"Accept":       {"text/html, application/xhtml+xml, image/jxr, */*"},
		"Content-Type": {"application/x-www-form-urlencoded; charset=UTF-8"},
		"Referer":      {"https://www.aldzs.com"},
	}
	resp, err := client.R().SetFormData(data).Post(url)
	if err != nil {
		fmt.Println(err)
	}
	res := gjson.Parse(string(resp.Body()))
	if res.Get("code").String() != "200" {
		gologger.Errorf("【aldzs】似乎出了点问题 %s \n", res.Get("msg"))
	}
	return res.Get("data")
}

func SearchByName(options *common.ENOptions) {
	keyword := options.KeyWord
	//拿到Token信息
	token := options.CookieInfo
	gologger.Infof("查询关键词 %s 的小程序\n", keyword)
	appList := getReq("Search/Search/search", map[string]string{
		"appName":    keyword,
		"page":       "1",
		"token":      token,
		"visit_type": "1",
	}).Array()
	if len(appList) == 0 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NO", "ID", "小程序名称", "所属公司", "描述"})
	for k, v := range appList {
		table.Append([]string{
			strconv.Itoa(k),
			v.Get("id").String(),
			v.Get("name").String(),
			v.Get("company").String(),
			v.Get("desc").String(),
		})
	}
	table.Render()
	//默认取第一个进行查询
	gologger.Infof("查询 %s 开发的相关小程序 【默认取100个】\n", appList[0].Get("company"))
	appKey := appList[0].Get("appKey").String()
	sAppList := getReq("Miniapp/App/sameBodyAppList", map[string]string{
		"appKey": appKey,
		"page":   "1",
		"size":   "100",
		"token":  token,
	}).Array()
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NO", "ID", "小程序名称", "描述"})
	for k, v := range sAppList {
		table.Append([]string{
			strconv.Itoa(k),
			v.Get("id").String(),
			v.Get("name").String(),
			v.Get("desc").String(),
		})
	}
	table.Render()
}
