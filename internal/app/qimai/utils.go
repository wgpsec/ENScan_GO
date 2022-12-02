package qimai

import (
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"net/http"
	"sort"
	"strings"
	"time"
)

type EnsGo struct {
	name     string
	api      string
	fids     string
	params   map[string]string
	field    []string
	keyWord  []string
	typeInfo []string
}

type EnInfos struct {
	Name    string
	Pid     string
	RegCode string
	Infos   map[string][]gjson.Result
}

type GetData struct {
	id     string
	params map[string]string
}

func getENMap() map[string]*EnsGo {
	ensInfoMap := make(map[string]*EnsGo)
	ensInfoMap = map[string]*EnsGo{
		"enterprise_info": {
			name: "企业信息",
			api:  "company/getCompany",
			params: map[string]string{
				"id": "",
			},
			field:   []string{"name", "legal", "type", "telephone", "email", "registered", "creatTime", "address", "scope", "creditCode", "id"},
			keyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"app": {
			name:    "APP",
			api:     "company/getCompanyApplist",
			field:   []string{"appInfo.appName", "app_category", "", "time", "", "appInfo.icon", "", "", ""},
			keyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
	}
	for k, _ := range ensInfoMap {
		ensInfoMap[k].keyWord = append(ensInfoMap[k].keyWord, "数据关联  ")
		ensInfoMap[k].field = append(ensInfoMap[k].field, "inFrom")
	}
	return ensInfoMap
}

func sign(params string, url string) string {
	i := "xyz517cda96abcd"
	f := -utils.RangeRand(100, 10000)
	o := time.Now().Unix()*1000 - (f) - 1515125653845
	r := base64.StdEncoding.EncodeToString([]byte(params))
	r = fmt.Sprintf("%s@#%s@#%d@#1", r, url, o)
	e := len(r)
	n := len(i)
	ne := ""
	for t := 0; t < e; t++ {
		ne += string(rune(int(r[t]) ^ int(i[(t+10)%n])))
	}
	ne = base64.StdEncoding.EncodeToString([]byte(ne))
	return ne
}

func GetReq(url string, params map[string]string, options *common.ENOptions) string {
	client := resty.New()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}
	gologger.Debugf("[qimai] url: %s, params: %s\n", url, params)
	cookie := options.ENConfig.Cookies.QiMai
	cookie = strings.ReplaceAll(cookie, "syncd", "syncds")
	cookie = cookie + ";synct=1651890521.296; syncd=-552934"
	client.Header = http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43"},
		"Accept":     {"text/html, application/xhtml+xml, image/jxr, */*"},
		"Cookie":     {cookie},
		"Referer":    {"https://www.qimai.cn/"},
	}
	var par []string
	params["analysis"] = ""
	for _, v := range params {
		par = append(par, v)
	}
	sort.Strings(par)
	pts := strings.Join(par, "")
	analysis := sign(pts, "/"+url)
	params["analysis"] = analysis
	urls := "https://api.qimai.cn/" + url
	resp, err := client.R().SetQueryParams(params).Get(urls)
	gologger.Debugf("%s", resp)
	if err != nil {
		if options.Proxy != "" {
			client.RemoveProxy()
		}
		gologger.Errorf("请求发生错误，5秒后重试\n%s\n", err)
		time.Sleep(5 * time.Second)
		return GetReq(url, params, options)
	}
	if resp.StatusCode() == 200 {
		return string(resp.Body())
	} else if resp.StatusCode() == 403 {
		gologger.Errorf("ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode() == 401 {
		gologger.Errorf("Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode() == 302 {
		gologger.Errorf("需要更新Cookie\n")
	} else if resp.StatusCode() == 404 {
		gologger.Errorf("目标不存在\n")
	} else {
		gologger.Errorf("未知错误 %s\n", resp.StatusCode())
	}
	return ""
}
