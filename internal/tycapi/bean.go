package tycapi

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"time"
)

func getENMap() map[string]*common.EnsGo {
	ensInfoMap := make(map[string]*common.EnsGo)
	ensInfoMap = map[string]*common.EnsGo{
		"cb_ic": {
			// 工商信息混合获取接口
			Name:  "工商信息",
			Api:   "cb/ic/2.0",
			Price: 1,
			SData: map[string]string{
				"branchList":      "branch",  // 分支机构
				"investList":      "invest",  // 投资信息
				"shareHolderList": "partner", // 股东信息
			},
			Field:   []string{"name", "legalPersonName", "regStatus", "phoneNumber", "email", "regCapital", "estiblishTime", "base", "businessScope", "creditCode", "id"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"cb_ipr": {
			// 知识产权混合获取接口
			Name:  "知识产权",
			Api:   "cb/ic/2.0",
			Price: 1.5,
			SData: map[string]string{
				"copyRegList": "copyright", // 软件著作权
				"icpList":     "icp",       // 域名信息
			},
		},
		"search": {
			Name:    "搜索",
			Api:     "search/2.0",
			Price:   0.01,
			Field:   ensInfoMap["enterprise_info"].Field,
			KeyWord: ensInfoMap["enterprise_info"].KeyWord,
		},
		"enterprise_info_normal": {
			Name:    "企业信息（基本）",
			Api:     "ic/baseinfo/normal",
			Price:   0.15,
			Field:   ensInfoMap["enterprise_info"].Field,
			KeyWord: ensInfoMap["enterprise_info"].KeyWord,
		},
		"enterprise_info": {
			Name:    "企业信息",
			Api:     "ic/baseinfoV2/2.0",
			Price:   0.2,
			Field:   []string{"name", "legalPersonName", "regStatus", "phoneNumber", "email", "regCapital", "estiblishTime", "base", "regLocation", "creditCode", "id"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"icp": {
			Name:    "ICP备案",
			Api:     "ipr/icp/3.0",
			Price:   0.1,
			Field:   []string{"webName", "webSite", "ym", "liscense", "companyName"},
			KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			Name:    "APP",
			Api:     "m/appbkInfo/2.0",
			Price:   0.2,
			Field:   []string{"name", "classes", "", "", "brief", "icon", "", "", ""},
			KeyWord: []string{"名称", "领域", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"weibo": {
			Name:    "微博",
			Api:     "m/weibo/2.0",
			Price:   0.2,
			Field:   []string{"name", "href", "info", "ico"},
			KeyWord: []string{"微博昵称", "链接", "简介", "logo"},
		},
		"wechat": {
			Name:    "微信公众号",
			Api:     "ipr/publicWeChat/2.0",
			Price:   0.2,
			Field:   []string{"title", "publicNum", "recommend", "codeImg", "titleImgURL"},
			KeyWord: []string{"名称", "ID", "描述", "二维码", "logo"},
		},
		"job": {
			Name:    "招聘信息",
			Api:     "ipr/publicWeChat/2.0",
			Price:   0.2,
			Field:   []string{"title", "education", "city", "startDate", "wapInfoPath"},
			KeyWord: []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
		},
		"copyright": {
			Name:    "软件著作权",
			Api:     "ipr/copyReg/2.0",
			Price:   0.1,
			Field:   []string{"simplename", "fullname", "catnum", "regnum", ""},
			KeyWord: []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
		},
		"supplier": {
			Name:    "供应商",
			Api:     "m/supply/2.0",
			Price:   0.2,
			Field:   []string{"supplier_name", "ratio", "amt", "announcement_date", "dataSource", "relationship", "supplier_graphId"},
			KeyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"invest": {
			Name:    "投资信息",
			Api:     "ic/inverst/2.0",
			Price:   0.15,
			Field:   []string{"name", "legalPersonName", "regStatus", "percent", "id"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
		},
		"holds": {
			Name:    "控股企业",
			Api:     "v4/open/companyholding",
			Price:   1,
			Field:   []string{"name", "legalPersonName", "regStatus", "percent", "legalType", "cid"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
		},
		"branch": {
			Name:    "分支信息",
			Api:     "ic/branch/2.0",
			Price:   0.15,
			Field:   []string{"name", "legalPersonName", "regStatus", "id"},
			KeyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"partner": {
			Name:    "股东信息",
			Api:     "ic/holder/2.0",
			Price:   0.15,
			Field:   []string{"name", "capital.percent", "capital.amomon", "id"},
			KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}
	base := "http://open.api.tianyancha.com/services/"
	for k, _ := range ensInfoMap {
		ensInfoMap[k].KeyWord = append(ensInfoMap[k].KeyWord, "数据关联")
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "inFrom")
		if k != "holds" {
			base += "open/"
		}
		ensInfoMap[k].Api = base + ensInfoMap[k].Api
	}
	return ensInfoMap
}

var reqMap = map[int]string{
	0:      "请求成功",
	300000: "经查无结果",
	300001: "请求失败",
	300002: "账号失效",
	300003: "账号过期",
	300004: "访问频率过快",
	300005: "无权限访问此api",
	300006: "余额不足",
	300007: "剩余次数不足",
	300008: "缺少必要参数",
	300009: "账号信息有误",
	300010: "URL不存在",
	300011: "此IP无权限访问此api",
	300012: "报告生成中",
}

func (h *TycAPI) req(url string) string {
	c := common.NewClient(map[string]string{
		"User-Agent":    "ENScanGO/" + common.GitTag,
		"Accept":        "text/html,application/json,application/xhtml+xml, image/jxr, */*",
		"Authorization": h.Options.ENConfig.Cookies.TycApiToken,
	}, h.Options)
	resp, err := c.Get(url)
	if err != nil {
		gologger.Error().Msgf("【TYC-API】请求错误 %s 5秒后重试 【%s】\n", url, err)
		time.Sleep(5 * time.Second)
		return h.req(url)
	}
	if resp.StatusCode == 200 {
		rs := gjson.Parse(resp.String())
		if rs.Get("error_code").Int() != 0 {
			if rs.Get("error_code").Int() == 300004 {
				time.Sleep(3 * time.Second)
				gologger.Error().Msgf("【TYC-API】访问频率过快，3秒后重试 %s\n", url)
				return h.req(url)
			}
			gologger.Error().Msgf("【TYC-API】错误 %s %s\n", rs.Get("error_code").String(), rs.Get("reason").String())
		}
		return resp.String()
	} else if resp.StatusCode == 403 {
		gologger.Error().Msgf("【TYC-API】403 IP被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 404 {
		gologger.Error().Msgf("【TYC-API】请求错误 404 %s \n", url)
	} else {
		gologger.Error().Msgf("【TYC-API】未知错误 %d\n", resp.StatusCode)
		gologger.Debug().Msgf("【TYC-API】\nURL:%s\n\n", url)
		gologger.Debug().Msgf("【TYC-API】\n%s\n", resp.String())
	}
	return ""
}
