package riskbird

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"strings"
	"time"
)

type RB struct {
	Options *common.ENOptions
	eid     string
}

func getENMap() map[string]*common.EnsGo {
	ensInfoMap := make(map[string]*common.EnsGo)
	ensInfoMap = map[string]*common.EnsGo{
		"search": {
			Name:    "企业信息",
			Api:     "newSearch",
			Field:   []string{"ENTNAME", "faren", "ENTSTATUS", "tels", "emails", "regConcat", "esDate", "dom", "", "UNISCID", "entid"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"enterprise_info": {
			Name:    "企业信息",
			Api:     "api/ent/query",
			Field:   []string{"entName", "personName", "entStatus", "telList", "emailList", "recConcat", "esDate", "yrAddress", "opScope", "uniscid", "entid"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"icp": {
			Name:    "ICP备案",
			Api:     "propertyIcp",
			GNum:    "propertyIcpCount",
			Field:   []string{"webname", "", "hostname", "icpnum", ""},
			KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			Name:    "APP",
			Api:     "propertyApp",
			GNum:    "propertyAppCount",
			Field:   []string{"appname", "", "", "updateDateAndroid", "brief", "iconUrl", "", "", "downloadCountLevel"},
			KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"wx_app": {
			Name:    "小程序",
			Api:     "propertyMiniprogram",
			GNum:    "propertyMiniprogramCount",
			Field:   []string{"name", "cate", "logo", "qrcode", ""},
			KeyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
		},
		"job": {
			Name:    "招聘信息",
			Api:     "job",
			GNum:    "jobCount",
			Field:   []string{"position", "education", "region", "pdate", "position"},
			KeyWord: []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
		},
		"copyright": {
			Name:    "软件著作权",
			Api:     "propertyCopyrightSoftware",
			GNum:    "propertyCopyrightSoftwareCount",
			Field:   []string{"sname", "sname", "", "snum", ""},
			KeyWord: []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
		},
		"invest": {
			Name:    "投资信息",
			Api:     "companyInvest",
			GNum:    "companyInvestCount",
			SData:   map[string]string{"category": "-100", "percentLevel": "-100", "province": "-100"},
			Field:   []string{"entName", "personName", "entStatus", "funderRatio", "entid"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
		},
		"branch": {
			Name:    "分支机构",
			GNum:    "companyBranchCount",
			Api:     "companyBranch",
			Field:   []string{"brName", "brPrincipal", "entStatus", "entid"},
			KeyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"partner": {
			Name:    "股东信息",
			Api:     "shareHolder",
			GNum:    "shareHolderCount",
			Field:   []string{"shaName", "fundedRatio", "subConAm", "shaId"},
			KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}

	for k, _ := range ensInfoMap {
		ensInfoMap[k].KeyWord = append(ensInfoMap[k].KeyWord, "数据关联")
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "ref")
	}
	return ensInfoMap
}

func (h *RB) req(url string, data string) string {
	c := common.NewClient(map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.60 Safari/537.36",
		"Accept":       "text/html,application/json,application/xhtml+xml, image/jxr, */*",
		"App-Device":   "WEB",
		"Content-Type": "application/json",
		"Cookie":       h.Options.GetCookie("rb"),
		"Origin":       "https://www.riskbird.com",
		"Referer":      "https://www.riskbird.com/ent/",
	}, h.Options)
	if !strings.Contains(url, "newSearch") {
		c.SetHeader("Xs-Content-Type", "application/json")
	}
	method := "GET"
	if data != "" {
		method = "POST"
		c.SetBody(data)
	}

	resp, err := c.Send(method, url)

	if err != nil {
		gologger.Error().Msgf("【RB】请求错误 %s 5秒后重试 【%s】\n", url, err)
		time.Sleep(5 * time.Second)
		return h.req(url, data)
	}
	if resp.StatusCode == 200 {
		rs := gjson.Parse(resp.String())
		if rs.Get("state").String() == "limit:auth" {
			gologger.Error().Msgf("【RB】您今日的查询次数已达到上限！请前往网站检查~ 30秒后重试\n")
			time.Sleep(30 * time.Second)
			return h.req(url, data)
		}
		return resp.String()
	} else if resp.StatusCode == 403 {
		gologger.Error().Msgf("【RB】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 401 {
		gologger.Error().Msgf("【RB】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode == 302 {
		gologger.Error().Msgf("【RB】需要更新Cookie\n")
	} else if resp.StatusCode == 404 {
		gologger.Error().Msgf("【RB】请求错误 404 %s \n", url)
	} else {
		gologger.Error().Msgf("【RB】未知错误 %s\n", resp.StatusCode)
		gologger.Debug().Msgf("【RB】\nURL:%s\nDATA:%s\n", url, data)
		gologger.Debug().Msgf("【RB】\n%s\n", resp.String())
	}
	return ""
}
