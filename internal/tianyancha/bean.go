package tianyancha

import (
	"github.com/antchfx/htmlquery"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"golang.org/x/net/html"
	"strings"
	"time"
)

func getENMap() map[string]*common.EnsGo {
	ensInfoMap := make(map[string]*common.EnsGo)
	ensInfoMap = map[string]*common.EnsGo{
		"enterprise_info": {
			Name:    "企业信息",
			Field:   []string{"name", "legalPersonName", "regStatus", "phoneNumber", "email", "regCapitalAmount", "fromTime", "taxAddress", "businessScope", "creditCode", "id"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"icp": {
			Name:    "ICP备案",
			Api:     "cloud-intellectual-property/intellectualProperty/icpRecordList",
			GNum:    "icpCount",
			Rf:      "item",
			Field:   []string{"webName", "webSite", "ym", "liscense", "companyName"},
			KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			Name:    "APP",
			Api:     "cloud-business-state/v3/ar/appbkinfo",
			GNum:    "productinfo",
			Rf:      "items",
			Field:   []string{"filterName", "classes", "", "", "brief", "icon", "", "", ""},
			KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"weibo": {
			Name:    "微博",
			Api:     "cloud-business-state/weibo/list",
			GNum:    "weiboCount",
			Rf:      "result",
			Field:   []string{"name", "href", "info", "ico"},
			KeyWord: []string{"微博昵称", "链接", "简介", "logo"},
		},
		"wechat": {
			Name:    "微信公众号",
			Api:     "cloud-business-state/wechat/list",
			GNum:    "weChatCount",
			Rf:      "resultList",
			Field:   []string{"title", "publicNum", "recommend", "codeImg", "titleImgURL"},
			KeyWord: []string{"名称", "ID", "描述", "二维码", "logo"},
		},
		"job": {
			Name:    "招聘信息",
			Api:     "cloud-business-state/recruitment/list",
			GNum:    "baipinCount",
			Rf:      "list",
			Field:   []string{"title", "education", "city", "startDate", "wapInfoPath"},
			KeyWord: []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
		},
		"copyright": {
			Name:    "软件著作权",
			Api:     "cloud-intellectual-property/intellectualProperty/softwareCopyrightListV2",
			GNum:    "copyrightWorks",
			Rf:      "items",
			Field:   []string{"simplename", "fullname", "", "regnum", ""},
			KeyWord: []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
		},
		"supplier": {
			Name:    "供应商",
			Api:     "cloud-business-state/supply/summaryList",
			GNum:    "suppliesV2Count",
			Rf:      "pageBean.result",
			GsData:  "&year=-100",
			Field:   []string{"supplier_name", "ratio", "amt", "announcement_date", "dataSource", "relationship", "supplier_graphId"},
			KeyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"invest": {
			Name:    "投资信息",
			Api:     "cloud-company-background/company/investListV2",
			GNum:    "inverstCount",
			Rf:      "result",
			SData:   map[string]string{"category": "-100", "percentLevel": "-100", "province": "-100"},
			Field:   []string{"name", "legalPersonName", "regStatus", "percent", "id"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
		},
		"holds": {
			Name:    "控股企业",
			Api:     "cloud-equity-provider/v4/hold/companyholding",
			GNum:    "finalInvestCount",
			Rf:      "list",
			Field:   []string{"name", "legalPersonName", "regStatus", "percent", "legalType", "cid"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
		},
		"branch": {
			Name:    "分支信息",
			Api:     "cloud-company-background/company/branchList",
			GNum:    "branchCount",
			Field:   []string{"name", "legalPersonName", "regStatus", "id"},
			Rf:      "result",
			KeyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"partner": {
			Name:    "股东信息",
			Api:     "cloud-company-background/companyV2/dim/holderForWeb",
			GNum:    "holderCount",
			Rf:      "result",
			SData:   map[string]string{"percentLevel": "-100", "sortField": "capitalAmount", "sortType": "-100"},
			Field:   []string{"name", "finalBenefitShares", "amount", "id"},
			KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}

	for k, _ := range ensInfoMap {
		ensInfoMap[k].KeyWord = append(ensInfoMap[k].KeyWord, "数据关联")
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "ref")
	}
	return ensInfoMap
}

func (h *TYC) req(url string, data string) string {
	c := common.NewClient(map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.60 Safari/537.36",
		"Accept":     "text/html,application/json,application/xhtml+xml, image/jxr, */*",
		"Version":    "TYC-Web",
		"Cookie":     h.Options.GetCookie("tyc"),
		"Origin":     "https://www.tianyancha.com",
		"Referer":    "https://www.tianyancha.com/",
	}, h.Options)
	if strings.Contains(url, "capi.tianyancha.com") {
		c.SetHeader("Content-Type", "application/json")
		//client.Header.Del("Cookie")
		c.SetHeader("X-Tycid", h.Options.ENConfig.Cookies.Tycid)
		c.SetHeader("X-Auth-Token", h.Options.ENConfig.Cookies.AuthToken)
	}

	method := "GET"
	if data != "" {
		method = "POST"
		c.SetBody(data)
	}

	resp, err := c.Send(method, url)

	if err != nil {
		gologger.Error().Msgf("【TYC】请求错误 %s 5秒后重试 【%s】\n", url, err)
		time.Sleep(5 * time.Second)
		return h.req(url, data)
	}
	if resp.StatusCode == 200 {
		if strings.Contains(resp.String(), "\"message\":\"mustlogin\"") {
			gologger.Error().Msgf("【TYC】需要登陆后尝试")
		}
		return resp.String()
	} else if resp.StatusCode == 403 {
		gologger.Error().Msgf("【TYC】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 401 {
		gologger.Error().Msgf("【TYC】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode == 302 {
		gologger.Error().Msgf("【TYC】需要更新Cookie\n")
	} else if resp.StatusCode == 404 {
		gologger.Error().Msgf("【TYC】请求错误 404 %s \n", url)
	} else if resp.StatusCode == 429 {
		gologger.Error().Msgf("【TYC】429请求被拦截，清打开链接滑动验证码，程序将在10秒后重试 %s \n", url)
		time.Sleep(10 * time.Second)
		return h.req(url, data)
	} else {
		gologger.Error().Msgf("【TYC】未知错误 %s\n", resp.StatusCode)
		gologger.Debug().Msgf("【TYC】\nURL:%s\nDATA:%s\n", url, data)
		gologger.Debug().Msgf("【TYC】\n%s\n", resp.String())
	}
	return ""
}

func (h *TYC) GetReqReturnPage(url string) *html.Node {
	body := h.req(url, "")
	if strings.Contains(body, "请输入中国大陆手机号") {
		gologger.Error().Msgf("[TYC] COOKIE检查失效，请检查COOKIE是否正确！\n")
	}
	if strings.Contains(body, "当前暂时无法访问") {
		gologger.Error().Msgf("[TYC] IP可能被拉黑！请使用代理尝试\n")
	}
	page, _ := htmlquery.Parse(strings.NewReader(body))
	return page
}
