package aiqicha

import (
	"github.com/tidwall/gjson"
)

type EnBen struct {
	Pid           string `json:"pid"`
	EntName       string `json:"entName"`
	EntType       string `json:"entType"`
	ValidityFrom  string `json:"validityFrom"`
	Domicile      string `json:"domicile"`
	EntLogo       string `json:"entLogo"`
	OpenStatus    string `json:"openStatus"`
	LegalPerson   string `json:"legalPerson"`
	LogoWord      string `json:"logoWord"`
	TitleName     string `json:"titleName"`
	TitleLegal    string `json:"titleLegal"`
	TitleDomicile string `json:"titleDomicile"`
	RegCap        string `json:"regCap"`
	Scope         string `json:"scope"`
	RegNo         string `json:"regNo"`
	PersonTitle   string `json:"personTitle"`
	PersonID      string `json:"personId"`
}

type EnsGo struct {
	name      string
	total     int64
	available int64
	api       string   //API 地址
	gNum      string   //判断数量大小的关键词
	field     []string //获取的字段名称 看JSON
	keyWord   []string //关键词
}

type EnInfo struct {
	Pid         string `json:"pid"`
	EntName     string `json:"entName"`
	legalPerson string
	openStatus  string
	email       string
	telephone   string
	branchNum   int64
	investNum   int64
	//info
	Infos  map[string][]gjson.Result
	ensMap map[string]*EnsGo
	//other
	investInfos map[string]EnInfo
	branchInfos map[string]EnInfo
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

func getENMap() map[string]*EnsGo {
	ensInfoMap := make(map[string]*EnsGo)
	ensInfoMap = map[string]*EnsGo{
		"enterprise_info": {
			name:    "企业信息",
			field:   []string{"entName", "legalPerson", "openStatus", "telephone", "email", "regCapital", "startDate", "regAddr", "scope", "taxNo", "pid"},
			keyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"icp": {
			name:    "ICP备案",
			api:     "detail/icpinfoAjax",
			field:   []string{"siteName", "homeSite", "domain", "icpNo", ""},
			keyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			name:    "APP",
			api:     "c/appinfoAjax",
			field:   []string{"name", "classify", "", "", "logoBrief", "logo", "", "", ""},
			keyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"weibo": {
			name:    "微博",
			api:     "c/microblogAjax",
			field:   []string{"nickname", "weiboLink", "brief", "logo"},
			keyWord: []string{"微博昵称", "链接", "简介", "LOGO"},
		},
		"wechat": {
			name:    "微信公众号",
			api:     "c/wechatoaAjax",
			field:   []string{"wechatName", "wechatId", "wechatIntruduction", "qrcode", "wechatLogo"},
			keyWord: []string{"名称", "ID", "描述", "二维码", "LOGO"},
		},
		"job": {
			name:    "招聘信息",
			api:     "c/enterprisejobAjax",
			field:   []string{"jobTitle", "education", "location", "publishDate", "desc"},
			keyWord: []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
		},
		"copyright": {
			name:    "软件著作权",
			api:     "detail/copyrightAjax",
			field:   []string{"softwareName", "shortName", "softwareType", "PubType", ""},
			keyWord: []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
		},
		"supplier": {
			name:    "供应商",
			api:     "c/supplierAjax",
			field:   []string{"supplier", "", "", "cooperationDate", "source", "", "supplierId"},
			keyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"invest": {
			name:    "投资信息",
			api:     "detail/investajax",
			field:   []string{"entName", "legalPerson", "openStatus", "regRate", "pid"},
			keyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
		},
		"holds": {
			name:    "控股企业",
			api:     "detail/holdsAjax",
			field:   []string{"entName", "", "", "proportion", "", "pid"},
			keyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
		},
		"branch": {
			name:    "分支信息",
			api:     "detail/branchajax",
			field:   []string{"entName", "legalPerson", "openStatus", "pid"},
			keyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"partner": {
			name:    "股东信息",
			api:     "detail/sharesAjax",
			field:   []string{"name", "subRate", "subMoney", "pid"},
			keyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}
	for k, _ := range ensInfoMap {
		ensInfoMap[k].keyWord = append(ensInfoMap[k].keyWord, "数据关联  ")
		ensInfoMap[k].field = append(ensInfoMap[k].field, "inFrom")
	}
	return ensInfoMap

}
