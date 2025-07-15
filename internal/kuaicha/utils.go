package kuaicha

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"time"
)

func getENMap() map[string]*common.EnsGo {
	ensInfoMap := make(map[string]*common.EnsGo)
	ensInfoMap = map[string]*common.EnsGo{
		"enterprise_info": {
			Name:    "企业信息",
			Api:     "enterprise_info_api/V1/find_company_basic_info",
			Field:   []string{"name", "legal_person", "state", "telphone", "email", "reg_capital", "established_date", "corp_address", "operating_scope", "unified_social_credit_code", "org_id"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		//"icp": {
		//	Name:    "ICP备案",
		//	Api:     "detail/icpinfoAjax",
		//	Field:   []string{"siteName", "homeSite", "domain", "icpNo", ""},
		//	KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		//},
		//"app": {
		//	Name:    "APP",
		//	Api:     "c/appinfoAjax",
		//	Field:   []string{"name", "classify", "", "", "logoBrief", "logo", "", "", ""},
		//	KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		//},
		//"weibo": {
		//	Name:    "微博",
		//	Api:     "c/microblogAjax",
		//	Field:   []string{"nickname", "weiboLink", "brief", "logo"},
		//	KeyWord: []string{"微博昵称", "链接", "简介", "LOGO"},
		//},
		"wechat": {
			Name:    "微信公众号",
			Api:     "open/commercial/v1/wechat_public_number",
			Field:   []string{"wechat_public_number", "wechat_number", "introduction", "qr_code", ""},
			KeyWord: []string{"名称", "ID", "描述", "二维码", "LOGO"},
		},
		"job": {
			Name:    "招聘信息",
			Api:     "open/commercial/v1/require_info",
			Field:   []string{"position", "background", "location", "publish_time", "url"},
			KeyWord: []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
		},
		"copyright": {
			Name:    "软件著作权",
			Api:     "open/trademark/v1/software_info",
			Field:   []string{"software_full_name", "software_short_name", "type", "reg_num", ""},
			KeyWord: []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
		},
		"supplier": {
			Name:    "供应商",
			Api:     "open/commercial/v1/main_suppliers_extend",
			Field:   []string{"", "ratio", "amt", "publish_time", "", "supplier_to_customer_type", "orgid"},
			KeyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"invest": {
			Name:    "投资信息",
			Api:     "open/app/v1/pc_enterprise/invest_abroad/list",
			Field:   []string{"frgn_invest_corp_name", "legal_representative", "", "invest_ratio", "frgn_invest_corp_id"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
			Fids:    "&is_latest=1",
		},
		"holds": {
			Name:    "控股企业",
			Api:     "open/app/v1/pc_enterprise/hold_corp/list",
			Field:   []string{"hold_corp_name", "legal_representative", "", "hold_ratio", "", "hold_corp_code"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
		},
		"branch": {
			Name:    "分支信息",
			Api:     "open/app/v1/pc_enterprise/branch_office/list",
			Field:   []string{"org_name", "person_name", "", "org_id"},
			KeyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"partner": {
			Name:    "股东信息",
			Api:     "open/app/v1/pc_enterprise/shareholder/latest_announcement",
			Field:   []string{"shareholder_name", "shareholding_ratio", "share_held_num", "shareholder_id"},
			KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}
	for k := range ensInfoMap {
		ensInfoMap[k].KeyWord = append(ensInfoMap[k].KeyWord, "数据关联  ")
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "inFrom")
	}
	return ensInfoMap

}

func (h *KC) req(url string, data string) string {
	c := common.NewClient(map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43",
		"Accept":       "text/html, application/xhtml+xml, image/jxr, */*",
		"Content-Type": "application/json;charset=UTF-8",
		"Cookie":       h.Options.ENConfig.Cookies.KuaiCha,
		"Source":       "PC",
		"Referer":      "https://www.kuaicha365.com/search-result?",
	}, h.Options)

	method := "GET"
	if data != "" {
		method = "POST"
		c.SetBody(data)
	}

	resp, err := c.Send(method, url)

	if err != nil {
		gologger.Error().Msgf("【KC】请求发生错误， %s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return h.req(url, data)
	}
	if resp.IsSuccessState() {
		return resp.String()
	} else if resp.StatusCode == 403 {
		gologger.Error().Msgf("【KC】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 401 {
		gologger.Error().Msgf("【KC】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode == 302 {
		gologger.Error().Msgf("【KC】需要更新Cookie\n")
	} else if resp.StatusCode == 404 {
		gologger.Error().Msgf("【KC】请求错误 404 %s \n", url)
	} else {
		gologger.Error().Msgf("【KC】未知错误 %s\n", resp.StatusCode)
	}
	return ""
}
