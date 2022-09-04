package tianyancha

import (
	"crypto/tls"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"github.com/robertkrimen/otto"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"
	"time"
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
	name           string
	total          int64
	available      int64
	api            string
	gNum           string
	tgNum          string
	sData          map[string]string
	gsData         string // get请求需要加的特殊参数
	rf             string //返回字段的json位置
	field          []string
	keyWord        []string
	PosiToTake     []int
	PosiToTaeS     [][]int
	NumOfEachGroup int //每组数量 总TR数除以行数计算，一条数据
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
			name:       "企业信息",
			field:      []string{"name", "legalPersonName", "regStatus", "phoneNumber", "email", "regCapitalAmount", "fromTime", "taxAddress", "businessScope", "creditCode", "id"},
			keyWord:    []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
			PosiToTaeS: [][]int{{0}, {1, 2}, {1, 4}, {1}, {2}, {3, 2}, {2, 2}, {10, 2}, {11, 2}, {5, 2}, {}},
		},
		"icp": {
			name: "ICP备案",
			api:  "cloud-intellectual-property/intellectualProperty/icpRecordList",
			gNum: "icpCount",
			//api:        "pagination/icp.xhtml",
			tgNum:      "knowledgeProperty.subItem.icpCount.num",
			rf:         "item",
			field:      []string{"webName", "webSite", "ym", "liscense", "companyName"},
			keyWord:    []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
			PosiToTake: []int{3, 4, 5, 6, 0},
		},
		"app": {
			name: "APP",
			api:  "cloud-business-state/v3/ar/appbkinfo",
			gNum: "productinfo",
			//api:            "pagination/product.xhtml",
			tgNum:      "manageStatus.subItem.productinfo.num",
			rf:         "items",
			field:      []string{"filterName", "classes", "", "", "brief", "icon", "", "", ""},
			keyWord:    []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
			PosiToTake: []int{2, 6, 0, 0, 5, 2, 0, 0, 0},
		},
		"weibo": {
			name: "微博",
			api:  "cloud-business-state/weibo/list",
			//api:        "pagination/weibo.xhtml",
			tgNum:      "manageStatus.subItem.weiboCount.num",
			gNum:       "weiboCount",
			rf:         "result",
			field:      []string{"name", "href", "info", "ico"},
			keyWord:    []string{"微博昵称", "链接", "简介", "logo"},
			PosiToTake: []int{2, 2, 4, 2},
		},
		"wechat": {
			name: "微信公众号",
			api:  "cloud-business-state/wechat/list",
			//api:        "pagination/wechat.xhtml",
			gNum:       "weChatCount",
			tgNum:      "manageStatus.subItem.weChatCount.num",
			rf:         "resultList",
			field:      []string{"title", "publicNum", "recommend", "codeImg", "titleImgURL"},
			keyWord:    []string{"名称", "ID", "描述", "二维码", "logo"},
			PosiToTake: []int{4, 5, 7, 6, 3},
		},
		"job": {
			name: "招聘信息",
			api:  "cloud-business-state/recruitment/list",
			//api:            "pagination/baipin.xhtml",
			tgNum:      "manageStatus.subItem.baipinCount.num",
			gNum:       "baipinCount",
			rf:         "list",
			field:      []string{"title", "education", "city", "startDate", "wapInfoPath"},
			keyWord:    []string{"招聘职位", "学历要求", "工作地点", "发布日期", "招聘描述"},
			PosiToTake: []int{3, 5, 7, 2, 0},
		},
		"copyright": {
			name: "软件著作权",
			//api:        "pagination/copyright.xhtml",
			api:        "cloud-intellectual-property/intellectualProperty/softwareCopyrightListV2",
			gNum:       "copyrightWorks",
			rf:         "items",
			tgNum:      "knowledgeProperty.subItem.cpoyRCount.num",
			field:      []string{"simplename", "fullname", "", "regnum", ""},
			keyWord:    []string{"软件名称", "软件简介", "分类", "登记号", "权利取得方式"},
			PosiToTake: []int{3, 4, 6, 5, 0},
		},
		"supplier": {
			name: "供应商",
			//api:        "pagination/supplies.xhtml",
			api:        "cloud-business-state/supply/summaryList",
			gNum:       "suppliesV2Count",
			tgNum:      "manageStatus.subItem.suppliesV2Count.num",
			rf:         "pageBean.result",
			gsData:     "&year=-100",
			field:      []string{"supplier_name", "ratio", "amt", "announcement_date", "dataSource", "relationship", "supplier_graphId"},
			keyWord:    []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
			PosiToTake: []int{2, 3, 4, 5, 6, 7, 2},
		},
		"invest": {
			name:  "投资信息",
			api:   "cloud-company-background/company/investListV2",
			tgNum: "backgroundItem.subItem.inverstCount.num",
			//api:        "pagination/invest.xhtml",
			gNum:       "inverstCount",
			rf:         "result",
			sData:      map[string]string{"category": "-100", "percentLevel": "-100", "province": "-100"},
			field:      []string{"name", "legalPersonName", "regStatus", "percent", "id"},
			keyWord:    []string{"企业名称", "法人", "状态", "投资比例", "PID"},
			PosiToTake: []int{2, 3, 7, 6, 2},
		},
		"holds": {
			name: "控股企业",
			api:  "cloud-equity-provider/v4/hold/companyholding",
			//api:            "pagination/companyholding.xhtml",
			gNum:           "finalInvestCount",
			tgNum:          "backgroundItem.subItem.finalInvestCount.num",
			rf:             "list",
			field:          []string{"name", "legalPersonName", "regStatus", "percent", "legalType", "cid"},
			keyWord:        []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
			PosiToTake:     []int{2, 0, 0, 5, 0, 4},
			NumOfEachGroup: 7,
		},
		"branch": {
			name: "分支信息",
			//api:        "pagination/branch.xhtml",
			api:        "cloud-company-background/company/branchList",
			tgNum:      "backgroundItem.subItem.branchCount.num",
			gNum:       "branchCount",
			field:      []string{"name", "legalPersonName", "regStatus", "id"},
			rf:         "result",
			keyWord:    []string{"企业名称", "法人", "状态", "PID"},
			PosiToTake: []int{4, 7, 10, 2},
		},
		"partner": {
			name: "股东信息",
			//api:        "pagination/holderCount.xhtml",
			api:        "cloud-company-background/companyV2/dim/holderForWeb",
			gNum:       "holderCount",
			tgNum:      "backgroundItem.subItem.holderCount.num",
			rf:         "result",
			sData:      map[string]string{"percentLevel": "-100", "sortField": "capitalAmount", "sortType": "-100"},
			field:      []string{"name", "finalBenefitShares", "amount", "id"},
			keyWord:    []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
			PosiToTake: []int{2, 3, 5, 2},
		},
	}

	for k, _ := range ensInfoMap {
		ensInfoMap[k].keyWord = append(ensInfoMap[k].keyWord, "数据关联")
		ensInfoMap[k].field = append(ensInfoMap[k].field, "inFrom")
	}
	return ensInfoMap
}

func GetReq(url string, data string, options *common.ENOptions) string {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}
	client.Header = http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43"},
		"Accept":     {"text/html,application/json,application/xhtml+xml, image/jxr, */*"},
		"Version":    {"TYC-Web"},
		"Cookie":     {options.ENConfig.Cookies.Tianyancha},
		"Origin":     {"https://www.tianyancha.com"},
		"Referer":    {"https://www.tianyancha.com/"},
	}

	if strings.Contains(url, "capi.tianyancha.com") {
		client.Header.Set("Content-Type", "application/json")
		client.Header.Del("Cookie")
		client.Header.Set("X-Tycid", "11111111111111ec111111ac10")
		client.Header.Set("X-Auth-Token", "eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIxMzAwMzY5Njkz11111Q3OTg5NiwiZXhwIjoxNjU3MDcxODk2fQ.1")
	}

	clientR := client.R()
	if data == "" {
		clientR.Method = "GET"
	} else {
		clientR.Method = "POST"
		clientR.SetBody(data)
	}
	clientR.URL = url
	resp, err := clientR.Send()

	//暂时没法直接算出Cookie信息等之后再看看吧
	if options.ENConfig.Cookies.Tianyancha == "" {
		re := regexp.MustCompile(`arg1='([\w\s]+)';`)
		rr := re.FindAllStringSubmatch(string(resp.Body()), 1)
		if len(rr) > 0 {
			str := rr[0][1]
			client.SetCookies(append(resp.Cookies(), &http.Cookie{Name: "acw_sc__v2", Value: str}))
		}
		gologger.Infof("【TYC】计算反爬获取Cookie成功 %s\n")
		resp, _ = clientR.Send()
	}

	if err != nil {
		if options.Proxy != "" {
			client.RemoveProxy()
		}

		gologger.Errorf("【TYC】请求错误 %s 5秒后重试 【%s】\n", url, err)
		if err.Error() == "unexpected EOF" {
			UpCookie(string(resp.Body()), options)

		}
		time.Sleep(5 * time.Second)
		return GetReq(url, data, options)
	}
	if resp.StatusCode() == 200 {
		return string(resp.Body())
	} else if resp.StatusCode() == 403 {
		gologger.Errorf("【TYC】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode() == 401 {
		gologger.Errorf("【TYC】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode() == 302 {
		gologger.Errorf("【TYC】需要更新Cookie\n")
	} else if resp.StatusCode() == 404 {
		gologger.Errorf("【TYC】请求错误 404 %s \n", url)
	} else {
		gologger.Errorf("【TYC】未知错误 %s\n", resp.StatusCode())
	}
	return ""
}

func GetReqReturnPage(url string, options *common.ENOptions) *html.Node {
	body := GetReq(url, "", options)
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
					gologger.Infof("【TYC】反爬计算签名成功！\n")
					options.ENConfig.Cookies.Tianyancha = strings.ReplaceAll(options.ENConfig.Cookies.Tianyancha, str2, SingAwcSCV2(str))
				} else {
					gologger.Errorf("【TYC】反爬Cookie存在问题\n")
				}
			}
		} else {
			gologger.Infof("【TYC】未登录反爬计算签名成功！\n")
			options.ENConfig.Cookies.Tianyancha = SingAwcSCV2(str)
		}
	} else {
		gologger.Errorf("【TYC】反爬存在问题\n")
	}
}

//SingAwcSCV2 acw_sc__v2
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
