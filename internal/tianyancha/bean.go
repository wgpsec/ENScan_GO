package tianyancha

import (
	"crypto/tls"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/imroc/req/v3"
	"github.com/robertkrimen/otto"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"golang.org/x/net/html"
	"regexp"
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
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "inFrom")
	}
	return ensInfoMap
}

func GetReq(url string, data string, options *common.ENOptions) string {
	client := req.C()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	client.SetTLSFingerprintChrome()
	if options.Proxy != "" {
		client.SetProxyURL(options.Proxy)
	}
	client.SetCommonHeaders(map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.60 Safari/537.36",
		"Accept":     "text/html,application/json,application/xhtml+xml, image/jxr, */*",
		"Version":    "TYC-Web",
		"Cookie":     options.ENConfig.Cookies.Tianyancha,
		"Origin":     "https://www.tianyancha.com",
		"Referer":    "https://www.tianyancha.com/",
	})
	clientR := client.R()

	if strings.Contains(url, "capi.tianyancha.com") {
		clientR.SetHeader("Content-Type", "application/json")
		//client.Header.Del("Cookie")
		clientR.SetHeader("X-Tycid", options.ENConfig.Cookies.Tycid)
		//client.Header.Set("X-Auth-Token", "")
	}
	//加延迟1S
	//强制延时1s
	time.Sleep(1 * time.Second)
	time.Sleep(time.Duration(options.GetDelayRTime()) * time.Second)

	method := "GET"
	if data == "" {
		method = "GET"
	} else {
		method = "POST"
		clientR.SetBody(data)
	}

	resp, err := clientR.Send(method, url)

	//暂时没法直接算出Cookie信息等之后再看看吧
	//if options.ENConfig.Cookies.Tianyancha == "" {
	//	re := regexp.MustCompile(`arg1='([\w\s]+)';`)
	//	rr := re.FindAllStringSubmatch(resp.String(), 1)
	//	if len(rr) > 0 {
	//		str := rr[0][1]
	//		client.R().SetCookies(append(resp.Cookies(), &http.Cookie{Name: "acw_sc__v2", Value: str}))
	//	}
	//	gologger.Info().Msgf("【TYC】计算反爬获取Cookie成功 %s\n")
	//	resp, _ = clientR.Send(method, url)
	//}

	if err != nil {
		if options.Proxy != "" {
			client.SetProxy(nil)
		}

		gologger.Error().Msgf("【TYC】请求错误 %s 5秒后重试 【%s】\n", url, err)
		if err.Error() == "unexpected EOF" {
			UpCookie(resp.String(), options)

		}
		time.Sleep(5 * time.Second)
		return GetReq(url, data, options)
	}
	if resp.StatusCode == 200 {
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
		return GetReq(url, data, options)

	} else {
		gologger.Error().Msgf("【TYC】未知错误 %s\n", resp.StatusCode)
		gologger.Debug().Msgf("【TYC】\nURL:%s\nDATA:%s\n", url, data)
		gologger.Debug().Msgf("【TYC】\n%s\n", resp.String())
	}
	return ""
}

func GetReqReturnPage(url string, options *common.ENOptions) *html.Node {
	body := GetReq(url, "", options)
	if strings.Contains(body, "请输入中国大陆手机号") {
		gologger.Error().Msgf("[TYC] COOKIE检查失效，请检查COOKIE是否正确！\n")
	}
	if strings.Contains(body, "当前暂时无法访问") {
		gologger.Error().Msgf("[TYC] IP可能被拉黑！请使用代理尝试\n")
	}
	page, _ := htmlquery.Parse(strings.NewReader(body))
	return page
}

func UpCookie(res string, options *common.ENOptions) {
	re := regexp.MustCompile(`arg1='([\w\s]+)';`)
	rr := re.FindAllStringSubmatch(res, 1)
	str := rr[0][1]
	if str != "" {
		if options.ENConfig.Cookies.Tianyancha != "" {
			re = regexp.MustCompile(`acw_sc__v2=([\w\s]+)`)
			rr = re.FindAllStringSubmatch(options.ENConfig.Cookies.Tianyancha, 1)
			if len(rr) > 0 {
				str2 := rr[0][1]
				if str2 != "" {
					gologger.Info().Msgf("【TYC】反爬计算签名成功！\n")
					options.ENConfig.Cookies.Tianyancha = strings.ReplaceAll(options.ENConfig.Cookies.Tianyancha, str2, SingAwcSCV2(str))
				} else {
					gologger.Error().Msgf("【TYC】反爬Cookie存在问题\n")
				}
			}
		} else {
			gologger.Info().Msgf("【TYC】未登录反爬计算签名成功！\n")
			options.ENConfig.Cookies.Tianyancha = SingAwcSCV2(str)
		}
	} else {
		gologger.Error().Msgf("【TYC】反爬存在问题\n")
	}
}

// SingAwcSCV2 acw_sc__v2
func SingAwcSCV2(tt string) string {
	vm := otto.New()
	_, err := vm.Run(`
function s2 (t1,t) {
    var str = "";
    for (var i = 0; i < t1.length && i < t.length; i += 2) {
        var a = parseInt(t1.slice(i, i + 2), 16);
        var b = parseInt(t.slice(i, i + 2), 16);
        var c = (a ^ b).toString(16);
        if (c.length == 1) {
            c = "0" + c;
        }
        str += c;
    }
    return str;
}
 function s1 (tt) {
    var listStr = [
        0xf, 0x23, 0x1d, 0x18, 0x21, 0x10, 0x1, 0x26, 0xa, 0x9, 0x13, 0x1f, 0x28,
        0x1b, 0x16, 0x17, 0x19, 0xd, 0x6, 0xb, 0x27, 0x12, 0x14, 0x8, 0xe, 0x15,
        0x20, 0x1a, 0x2, 0x1e, 0x7, 0x4, 0x11, 0x5, 0x3, 0x1c, 0x22, 0x25, 0xc,
        0x24
    ];
    var litss = [];
    var a = "";
    for (var i = 0; i < tt.length; i++) {
        var b = tt[i];
        for (var t = 0; t < listStr.length; t++) {
            if (listStr[t] == i + 1) {
                litss[t] = b;
            }
        }
    }
    a = litss.join("");
    return a;
}
function sign(arg1,num) {
    return s2(s1(arg1),num);
}

`)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	call, err := vm.Call("sign", nil, tt, "3000176000856006061501533003690027800375")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	res := call.String()
	return res
}
