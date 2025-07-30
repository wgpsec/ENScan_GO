package aiqicha

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"strings"
	"time"
)

func getENMap() map[string]*common.EnsGo {
	ensInfoMap := make(map[string]*common.EnsGo)
	ensInfoMap = map[string]*common.EnsGo{
		"enterprise_info": {
			Name:    "企业信息",
			Api:     "detail/basicAllDataAjax",
			Field:   []string{"entName", "legalPerson", "openStatus", "telephone", "email", "regCapital", "startDate", "regAddr", "scope", "taxNo", "pid"},
			KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"icp": {
			Name:    "ICP备案",
			Api:     "detail/icpinfoAjax",
			Field:   []string{"siteName", "homeSite", "domain", "icpNo", ""},
			KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			Name:    "APP",
			Api:     "c/appinfoAjax",
			Field:   []string{"name", "classify", "", "", "logoBrief", "logo", "", "", ""},
			KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"weibo": {
			Name:    "微博",
			Api:     "c/microblogAjax",
			Field:   []string{"nickname", "weiboLink", "brief", "logo"},
			KeyWord: []string{"微博昵称", "链接", "简介", "LOGO"},
		},
		"wechat": {
			Name:    "微信公众号",
			Api:     "c/wechatoaAjax",
			Field:   []string{"wechatName", "wechatId", "wechatIntruduction", "qrcode", "wechatLogo"},
			KeyWord: []string{"名称", "ID", "描述", "二维码", "LOGO"},
		},
		"job": {
			Name:    "招聘信息",
			Api:     "c/enterprisejobAjax",
			Field:   []string{"jobTitle", "education", "location", "publishDate", "desc"},
			KeyWord: []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
		},
		"copyright": {
			Name:    "软件著作权",
			Api:     "detail/copyrightAjax",
			Field:   []string{"softwareName", "shortName", "softwareType", "PubType", ""},
			KeyWord: []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
		},
		"supplier": {
			Name:    "供应商",
			Api:     "c/supplierAjax",
			Field:   []string{"supplier", "", "", "cooperationDate", "source", "", "supplierId"},
			KeyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"invest": {
			Name:    "投资信息",
			Api:     "detail/investajax",
			Field:   []string{"entName", "legalPerson", "openStatus", "regRate", "pid"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
		},
		"holds": {
			Name:    "控股企业",
			Api:     "detail/holdsAjax",
			Field:   []string{"entName", "", "", "proportion", "", "pid"},
			KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
		},
		"branch": {
			Name:    "分支信息",
			Api:     "detail/branchajax",
			Field:   []string{"entName", "legalPerson", "openStatus", "pid"},
			KeyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"partner": {
			Name:    "股东信息",
			Api:     "detail/sharesAjax",
			Field:   []string{"name", "subRate", "subMoney", "pid"},
			KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
	}
	for k := range ensInfoMap {
		ensInfoMap[k].KeyWord = append(ensInfoMap[k].KeyWord, "数据关联  ")
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "ref")
	}
	return ensInfoMap

}

func (h *AQC) req(url string) string {
	c := common.NewClient(map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43",
		"Accept":     "text/html, application/xhtml+xml, image/jxr, */*",
		"Cookie":     h.Options.GetCookie("aqc"),
		"Referer":    "https://aiqicha.baidu.com/",
	}, h.Options)
	resp, err := c.Get(url)

	if err != nil {
		gologger.Error().Msgf("【AQC】请求发生错误， %s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return h.req(url)
	}
	if resp.IsSuccessState() {
		if strings.Contains(resp.String(), "百度安全验证") {
			gologger.Error().Msgf("【AQC】需要安全验证，请打开浏览器进行验证后操作，10秒后重试！ %s \n", "https://aiqicha.baidu.com/")
			gologger.Debug().Msgf("URL:%s\n\n%s", url, resp.String())
			time.Sleep(10 * time.Second)
			return h.req(url)
		}
		return resp.String()
	} else if resp.StatusCode == 403 {
		gologger.Error().Msgf("【AQC】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode == 401 {
		gologger.Error().Msgf("【AQC】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode == 302 {
		gologger.Error().Msgf("【AQC】需要更新Cookie\n")
	} else if resp.StatusCode == 404 {
		gologger.Error().Msgf("【AQC】请求错误 404 %s \n", url)
	} else {
		gologger.Error().Msgf("【AQC】未知错误 %s\n", resp.StatusCode)
	}
	return ""
}

// pageParseJson 提取页面中的JSON字段
func pageParseJson(content string) (gjson.Result, error) {
	tag1 := "window.pageData ="
	tag2 := "window.isSpider ="
	//tag2 := "/* eslint-enable */</script><script data-app"
	idx1 := strings.Index(content, tag1)
	idx2 := strings.Index(content, tag2)
	if idx2 > idx1 {
		str := content[idx1+len(tag1) : idx2]
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Replace(str, " ", "", -1)
		str = str[:len(str)-1]
		return gjson.Get(string(str), "result"), nil
	} else {
		gologger.Error().Msgf("【AQC】无法解析页面数据，请开启Debug检查")
		gologger.Debug().Msgf("【AQC】页面返回数据\n————\n%s————\n", content)
	}
	return gjson.Result{}, fmt.Errorf("无法解析页面数据")
}

func transformNumber(input string, t int64) string {
	transformedStr := ""
	codes := make([]map[rune]rune, 3)
	codes[1] = map[rune]rune{
		'0': '0',
		'1': '1',
		'2': '2',
		'3': '3',
		'4': '5',
		'5': '4',
		'6': '7',
		'7': '6',
		'8': '9',
		'9': '8',
	}
	codes[2] = map[rune]rune{
		'0': '0',
		'1': '1',
		'2': '2',
		'3': '3',
		'4': '6',
		'5': '8',
		'6': '9',
		'7': '4',
		'8': '5',
		'9': '7',
	}
	for _, digitRune := range input {
		transformedStr += string(codes[t][digitRune])
	}
	return transformedStr
}

var enMapping = map[string]string{
	"webRecord":     "icp",
	"appinfo":       "app",
	"wechatoa":      "wechat",
	"enterprisejob": "job",
	"microblog":     "weibo",
	"hold":          "holds",
	"shareholders":  "partner",
}
