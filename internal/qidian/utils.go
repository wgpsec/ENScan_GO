package qidian

import (
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"time"
)

func GetReq(url string, data string, options *common.ENOptions) string {
	client := req.C()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxyURL(options.Proxy)
	}

	if data == "" {
		data = "{}"
	}

	client.SetCommonHeaders(map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43",
		"Accept":       "text/html, application/xhtml+xml, image/jxr, */*",
		"Content-Type": "application/json;charset=UTF-8",
		"Cookie":       options.ENConfig.Cookies.Qidian,
		"Referer":      "https://www.dingtalk.com/",
	})
	//加入随机延时
	time.Sleep(time.Duration(options.GetDelayRTime()) * time.Second)
	clientR := client.R()
	method := "GET"
	if data == "{}" {
		method = "GET"
	} else {
		method = "POST"
		clientR.SetBody(data)
	}

	resp, err := clientR.Send(method, url)

	if err != nil {
		if options.Proxy != "" {
			client.SetProxy(nil)
		}
		gologger.Error().Msgf("【qidian】请求发生错误， %s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return GetReq(url, data, options)
	}
	if resp.StatusCode == 200 {
		return resp.String()
	} else if resp.StatusCode == 403 {
		gologger.Error().Msgf("【qidian】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 401 {
		gologger.Error().Msgf("【qidian】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode == 302 {
		gologger.Error().Msgf("【qidian】需要更新Cookie\n")
	} else if resp.StatusCode == 404 {
		gologger.Error().Msgf("【qidian】请求错误 404 %s \n", url)
	} else {
		gologger.Error().Msgf("【qidian】未知错误 %s\n", resp.StatusCode)
	}
	return ""
}

type EnsGo struct {
	name         string
	api          string
	dataModuleId int
	fids         string
	gNum         string // 获取数量的json关键词 getDetail->CountInfo
	sData        map[string]string
	field        []string
	keyWord      []string
	typeInfo     []string
}
type EnInfos struct {
	Name        string
	Pid         string
	legalPerson string
	openStatus  string
	email       string
	telephone   string
	branchNum   int64
	investNum   int64
	Infos       map[string][]gjson.Result
}

func getENMap() map[string]*common.EnsGo {
	resEnsMap := map[string]*common.EnsGo{
		"enterprise_info": {
			Name:         "企业信息",
			DataModuleId: 748,
			Field:        []string{"Name", "Oper.Name", "ShortStatus", "ContactInfo.PhoneNumber", "ContactInfo.Email", "RegistCapi", "CheckDate", "Address", "Scope", "CreditCode", "_id"},
			KeyWord:      []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"copyright": {
			Name:         "软件著作权",
			DataModuleId: 481,
			Field:        []string{"Name", "ShortName", "", "RegisterNo", "PubType"},
			KeyWord:      []string{"软件全称", "软件简称", "分类", "登记号", "权利取得方式"},
		},
		"icp": {
			Name:         "ICP备案",
			DataModuleId: 512,
			Field:        []string{"WebsiteName", "HomeAddress", "DomainName", "WebrecordNo", "CompanyName"},
			KeyWord:      []string{"网站名称", "站点首页", "域名", "网站备案/许可证号", "公司名称"},
		},

		"supplier": {
			Name:         "招投标",
			DataModuleId: 477,
			Field:        []string{"CompanyName", "Proportion", "Quota", "ReportYear", "Source", "Relationship", "KeyNo"},
			KeyWord:      []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"branch": {
			Name:         "分支机构",
			DataModuleId: 485,
			Field:        []string{"Name", "Oper.Name", "ShortStatus", "KeyNo"}, //ShortStatus 不兼容列表！
			KeyWord:      []string{"企业名称", "法人", "状态", "PID"},
		},
		"invest": {
			Name:         "对外投资信息",
			DataModuleId: 453,
			Field:        []string{"Name", "OperName", "Status", "FundedRatio", "KeyNo"},
			KeyWord:      []string{"企业名称", "法人", "状态", "持股信息", "PID"},
		},
		"partner": {
			Name:         "股东信息",
			DataModuleId: 484,
			Field:        []string{"StockName", "StockPercent", "ShouldCapi", "KeyNo"},
			KeyWord:      []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}

	for k, _ := range resEnsMap {
		resEnsMap[k].KeyWord = append(resEnsMap[k].KeyWord, "信息关联")
		resEnsMap[k].Field = append(resEnsMap[k].Field, "inFrom")
	}
	return resEnsMap
}
