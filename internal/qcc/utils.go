package qcc

import (
	"crypto/tls"
	"github.com/go-resty/resty/v2"
	"github.com/robertkrimen/otto"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"net/http"
	"strings"
	"time"
)

type EnsGo struct {
	name     string
	api      string
	fids     string
	gNum     string // 获取数量的json关键词 getDetail->CountInfo
	sData    map[string]string
	field    []string
	keyWord  []string
	typeInfo []string
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

func getENMap(searchType string, pid string) map[string]*EnsGo {
	resEnsMap := make(map[string]*EnsGo)
	cEnsMap := make(map[string]*EnsGo)
	gEnsMap := make(map[string]*EnsGo)

	cEnsMap = map[string]*EnsGo{
		"enterprise_info": {
			name:    "企业信息",
			api:     "company/getDetail",
			field:   []string{"Name", "Oper.Name", "ShortStatus", "ContactInfo.PhoneNumber", "ContactInfo.Email", "RegistCapi", "CheckDate", "Address", "Scope", "CreditCode", "_id"},
			keyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
		},
		"copyright": {
			name:    "软件著作权",
			api:     "datalist/rjzzqlist",
			gNum:    "SoftCRCount",
			field:   []string{"Name", "ShortName", "", "RegisterNo", "PubType"},
			keyWord: []string{"软件全称", "软件简称", "分类", "登记号", "权利取得方式"},
		},
		"icp": {
			name:    "ICP备案",
			api:     "datalist/websitelist",
			gNum:    "WebSiteCount",
			field:   []string{"WebsiteName", "HomeAddress", "DomainName", "WebrecordNo", "CompanyName"},
			keyWord: []string{"网站名称", "站点首页", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			name:    "APP信息",
			api:     "datalist/applist",
			gNum:    "APPCount",
			field:   []string{"AppName", "Category", "Version", "UpdateDate", "Description", "Logo", "", "", ""},
			keyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"wx_app": {
			name:    "小程序",
			api:     "datalist/wplist",
			gNum:    "WPCount",
			field:   []string{"Name", "Category", "AvatarUrl", "QrcodeUrl", "WechatReadAmount"},
			keyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
		},
		"wechat": {
			name:    "微信公众号",
			api:     "datalist/wechatlist",
			gNum:    "WeChatCount",
			field:   []string{"AccountName", "WechatAlias", "Description", "QrcodeUrl", "AvatarUrl"},
			keyWord: []string{"名称", "ID", "简介", "二维码", "头像"},
		},
		"weibo": {
			name:    "微博",
			api:     "datalist/weibolist",
			gNum:    "WeiBoCount",
			field:   []string{"Name", "DetailUrl", "Abstract", "ImageUrl"},
			keyWord: []string{"微博昵称", "链接", "简介", "头像"},
		},
		"job": {
			name:    "招聘",
			api:     "datalist/joblist",
			gNum:    "RecruitmentCount",
			field:   []string{"PositionName", "Education", "Address", "", "Salary"},
			keyWord: []string{"招聘职位", "学历", "办公地点", "发布日期", "薪资"},
		},
		"supplier": {
			name:    "供应商",
			api:     "datalist/supplierCustomer",
			gNum:    "SupplierCountV2", // 客户 CustomerCountV2
			fids:    "&dataType=0",     // 0 供应商 1 客户
			field:   []string{"CompanyName", "Proportion", "Quota", "ReportYear", "Source", "Relationship", "KeyNo"},
			keyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
		},
		"branch": {
			name:    "分支机构",
			api:     "datalist/branchelist",
			gNum:    "BranchCount",
			fids:    "&nodeName=Branches&pageIndex=1&sortField=shoulddate",
			field:   []string{"Name", "Oper.Name", "ShortStatus", "KeyNo"}, //ShortStatus 不兼容列表！
			keyWord: []string{"企业名称", "法人", "状态", "PID"},
		},
		"invest": { // 股权信息 参数 fundedRatioLevel 5 4 3 2 1
			name:    "对外投资信息",
			api:     "datalist/touzilist",
			gNum:    "InvesterCount",
			field:   []string{"Name", "OperName", "Status", "FundedRatio", "KeyNo"},
			keyWord: []string{"企业名称", "法人", "状态", "持股信息", "PID"},
		},
		"holds": {
			name:    "控制企业",
			api:     "datalist/holdcolist",
			gNum:    "HCCount",
			field:   []string{"Name", "OperName", "Status", "PercentTotal", "Level", "KeyNo"},
			keyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
		},
		"partner": {
			name:    "股东信息",
			api:     "datalist/holdcolist",
			field:   []string{"StockName", "StockPercent", "ShouldCapi", "KeyNo"},
			keyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
		},
		"indirect": {
			name:    "间接持股企业",
			api:     "datalist/indirectlist",
			gNum:    "IndirectInvestmentCount",
			field:   []string{"Name", "Oper.Name", "Status", "Level", "PercentTotal", "KeyNo"},
			keyWord: []string{"企业名称", "法人", "状态", "持股层级", "间接持股比例", "PID"},
		},
		"contactrel": {
			name:    "疑似关系",
			api:     "datalist/contactrel",
			gNum:    "ContactRelatedCount",
			field:   []string{"Name", "OperName", "Status", "RelatedReason", "RelatedTelephone", "KeyNo"},
			keyWord: []string{"企业名称", "法人", "状态", "疑似关联类型", "疑似关联详情", "PID"},
		},
	}
	gEnsMap = map[string]*EnsGo{
		"GroupCompanyList": {
			name:    "集团成员列表",
			api:     "group/getGroupCompanyList",
			field:   []string{"Name", "StockPercent", "OperName", "Status", "GW", "Email"},
			keyWord: []string{"企业名称", "持股比例", "法人", "状态", "官网", "邮箱"},
			sData: map[string]string{
				"flag":        "", //flag G_HX 核心企业 G_SS 上市公司
				"searchKey":   pid,
				"searchIndex": "flag",
				"statusCode":  "", //statusCode 20 续存 99 注销 10 在业 92 仍注册
				"searchType":  "0,1,3,4,5,10,11,12,20,21,22",
				"pageIndex":   "1",
			},
		},
		"PersonSummaryInfo": {
			name: "集团人员",
			api:  "group/GetPersonSummaryInfo",
			sData: map[string]string{
				"groupId":   pid,
				"pageIndex": "1",
				"pageSize":  "10",
				"type":      "1", //type 1 法人代表 2高管 3核心人员
			},
			field:   []string{"PersonInfo.Name"},
			keyWord: []string{"名字"},
		},
		"PartnerList": {
			name:    "投资方",
			api:     "group/getPartnerList",
			field:   []string{"PartnerName", "Oper.Name", "RegistCapi", "ShortStatus", "City"},
			keyWord: []string{"投资方", "法人代表", "注册资本", "状态", "所在地"},
			sData: map[string]string{
				"groupId":   pid,
				"pageIndex": "1",
				"pageSize":  "10",
				"type":      "0",
			},
		},
		"InvestmentList": {
			name:    "对外投资",
			api:     "group/GetInvestmentList",
			field:   []string{"InvestmentName", "Oper.Name", "RegistCapi", "ShortStatus", "City"},
			keyWord: []string{"被投资企业", "法人代表", "注册资本", "状态", "所在地"},
			sData: map[string]string{
				"groupId":     pid,
				"pageIndex":   "1",
				"pageSize":    "10",
				"shortStatus": "", //shortStatus 续存、在业等
			},
		},
		"BranchInfo": {
			name:    "分支机构",
			api:     "group/GetBranchInfo",
			field:   []string{"Name", "Count"},
			keyWord: []string{"成员公司", "分支机构数量"},
			sData: map[string]string{
				"groupId":   pid,
				"pageIndex": "1",
				"pageSize":  "10",
			},
		},
		"GroupRiskListByType": {
			name:    "经营风险",
			api:     "group/GetGroupRiskListByType",
			field:   []string{"Name", "Count"},
			keyWord: []string{"成员公司", "风险数量"},
			sData: map[string]string{
				"groupId":   pid,
				"pageIndex": "1",
				"type":      "2", // 2 被执行人 4 裁判文书 20 立案信息...
			},
		},
		"GroupBusinessListByType": {
			name:    "经营信息",
			api:     "group/GetGroupBusinessListByType",
			field:   []string{"Name", "Count"},
			keyWord: []string{"成员公司", "信息数量"},
			sData: map[string]string{
				"groupId":   pid,
				"pageIndex": "1",
				"type":      "114", // 102 招投标 114 供应商信息 115 客户信息 113 电信许可
			},
		},
		"OwnProductList": {
			name:    "产品服务",
			api:     "group/GetOwnProductList",
			field:   []string{"ProductName", "Round", "Location", "Description", "Tags", "RelatedName", "ProductLogo"},
			keyWord: []string{"产品名", "融资信息", "所属地", "描述", "标签", "关联企业", "产品图片"},
			sData: map[string]string{
				"groupId":   pid,
				"pageIndex": "1",
			},
		},
		"GroupPropertyInfo": {
			name:     "知识产权信息",
			api:      "group/GetGroupPropertyInfo",
			field:    []string{"Name", "Count", "KeyNo"},
			keyWord:  []string{"公司名称", "数量", "ID"},
			typeInfo: []string{"", "", "", "", "", "website", "wechat", "app", "wp", "weibo"},
			//typeInfo: []string{"", "", "", "", "", "website", "", "", "", ""},
			sData: map[string]string{
				"groupId":   pid,
				"type":      "6",
				"pageIndex": "1",
			},
		},
	}

	if searchType == "company" {
		resEnsMap = cEnsMap
	} else if searchType == "group" {
		resEnsMap = gEnsMap
	} else if searchType == "all" {
		resEnsMap = cEnsMap
		for k, v := range gEnsMap {
			resEnsMap[k] = v
		}
	}
	for k, _ := range resEnsMap {
		resEnsMap[k].keyWord = append(resEnsMap[k].keyWord, "信息关联")
		resEnsMap[k].field = append(resEnsMap[k].field, "inFrom")
	}
	return resEnsMap
}

// pageParseJson 提取页面中的JSON字段
func pageParseJson(content string) gjson.Result {

	tag1 := "window.__INITIAL_STATE__="
	tag2 := "(function()"
	//tag2 := "/* eslint-enable */</script><script data-app"
	idx1 := strings.Index(content, tag1)
	idx2 := strings.Index(content, tag2)
	if idx2 > idx1 {
		str := content[idx1+len(tag1) : idx2]
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Replace(str, " ", "", -1)
		str = str[:len(str)-1]
		return gjson.Parse(str)
	} else {
		gologger.Errorf("无法解析信息错误信息%s\n", content)
	}
	return gjson.Result{}
}

// getReq 企查查专属请求
// searchUrl 查询参数
func getReq(url string, data string, options *common.ENOptions) string {
	//安全延时
	time.Sleep(time.Duration(options.DelayTime) * time.Second)

	if data == "" {
		data = "{}"
	}

	//计算签名
	singKey := "s"
	singV := "s"
	if !strings.Contains(url, ".html") {
		searchUrl := strings.ReplaceAll(url, "https://www.qcc.com", "")
		resSing := hSing(searchUrl, data)
		singKey = resSing.Keys()[0]
		singValue, _ := resSing.Get(singKey)
		singV = singValue.String()
	}
	//构造QCC请求
	client := resty.New()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.Header = http.Header{
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36"},
		"Content-Type": {"application/json;charset=UTF-8"},
		"Cookie":       {options.ENConfig.Cookies.Qcc},
		singKey:        {singV},
		"Referer":      {"https://www.qcc.com/web/"},
	}
	clientR := client.R()
	if data == "{}" {
		clientR.Method = "GET"
	} else {
		clientR.Method = "POST"
		clientR.SetBody(data)
	}

	clientR.URL = url
	resp, err := clientR.Send()
	if err != nil {
		gologger.Errorf("【QCC】请求发生错误，%s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return getReq(url, data, options)
	}
	if resp.StatusCode() == 200 {
		if !strings.Contains(string(resp.Body()), "会员登录 - 企查查") {
			return string(resp.Body())
		} else {
			gologger.Errorf("【QCC】需要登陆操作\n%s\n", err)
			return ""
		}

	} else if resp.StatusCode() == 403 {
		gologger.Errorf("【QCC】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode() == 401 {
		gologger.Errorf("【QCC】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode() == 301 {
		gologger.Errorf("【QCC】需要更新Cookie\n")
	} else if resp.StatusCode() == 435 {
		gologger.Errorf("【QCC】签名验证失败\n")
		return ""
	} else if resp.StatusCode() == 412 {
		gologger.Errorf("【QCC】需要机器验证，请前往打开网站滑动验证，10秒后将会重试\n")
		time.Sleep(10 * time.Second)
		return getReq(url, data, options)
	} else if resp.StatusCode() == 404 {
		gologger.Errorf("【QCC】请求错误 404 %s\n", url)
	} else {
		gologger.Errorf("【QCC】未知错误 %s\n", resp.StatusCode())
		return ""
	}

	if strings.Contains(string(resp.Body()), "使用该功能需要用户登录") {
		gologger.Errorf("【QCC】Cookie有问题或过期，请重新获取\n")
	}
	return ""
}

// hSing QCC 计算签名
// url url地址
// data 查询的POST数据（没有就传空）
func hSing(url string, data string) *otto.Object {
	vm := otto.New()
	_, err := vm.Run(`
var window = {}, aaa;
!function (t) {
    function e(e) {
        for (var n, a, r = e[0], o = e[1], s = 0, c = []; s < r.length; s++)
            a = r[s],
            Object.prototype.hasOwnProperty.call(i, a) && i[a] && c.push(i[a][0]),
                i[a] = 0;
        for (n in o)
            Object.prototype.hasOwnProperty.call(o, n) && (t[n] = o[n]);
        for (l && l(e); c.length;)
            c.shift()()
    }

    var n = {}
        , a = {
        18: 0
    }
        , i = {
        18: 0
    };

    function r(e) {
        // console.log(e)
        if (n[e])
            return n[e].exports;
        var a = n[e] = {
            i: e,
            l: !1,
            exports: {}
        };
        return t[e].call(a.exports, a, a.exports, r),
            a.l = !0,
            a.exports
    }

    aaa = r;

    r.e = function (t) {
        var e = [];
        a[t] ? e.push(a[t]) : 0 !== a[t] && {
            0: 1,
            1: 1,
            4: 1,
            6: 1,
            7: 1,
            8: 1,
            9: 1,
            10: 1,
            11: 1,
            12: 1,
            13: 1,
            14: 1,
            15: 1,
            16: 1,
            20: 1,
            21: 1,
            22: 1,
            23: 1,
            24: 1,
            25: 1,
            26: 1,
            27: 1,
            28: 1,
            29: 1,
            30: 1,
            31: 1,
            32: 1,
            33: 1,
            34: 1,
            35: 1,
            36: 1,
            37: 1,
            38: 1,
            39: 1,
            40: 1,
            41: 1,
            42: 1,
            43: 1,
            44: 1,
            45: 1,
            46: 1,
            47: 1,
            48: 1,
            49: 1,
            50: 1,
            51: 1,
            52: 1,
            53: 1,
            54: 1,
            55: 1,
            56: 1,
            57: 1,
            58: 1,
            59: 1,
            60: 1,
            61: 1,
            62: 1,
            63: 1,
            64: 1,
            65: 1,
            66: 1,
            67: 1,
            68: 1
        }[t] && e.push(a[t] = new Promise((function (e, n) {
                for (var i = ({}[t] || t) + ".502689ef.css", o = r.p + i, s = document.getElementsByTagName("link"), c = 0; c < s.length; c++) {
                    var l = (d = s[c]).getAttribute("data-href") || d.getAttribute("href");
                    if ("stylesheet" === d.rel && (l === i || l === o))
                        return e()
                }
                var u = document.getElementsByTagName("style");
                for (c = 0; c < u.length; c++) {
                    var d;
                    if ((l = (d = u[c]).getAttribute("data-href")) === i || l === o)
                        return e()
                }
                var f = document.createElement("link");
                f.rel = "stylesheet",
                    f.type = "text/css",
                    f.onload = e,
                    f.onerror = function (e) {
                        var i = e && e.target && e.target.src || o
                            , r = new Error("Loading CSS chunk " + t + " failed.\n(" + i + ")");
                        r.code = "CSS_CHUNK_LOAD_FAILED",
                            r.request = i,
                            delete a[t],
                            f.parentNode.removeChild(f),
                            n(r)
                    }
                    ,
                    f.href = o,
                    document.getElementsByTagName("head")[0].appendChild(f)
            }
        )).then((function () {
                a[t] = 0
            }
        )));
        var n = i[t];
        if (0 !== n)
            if (n)
                e.push(n[2]);
            else {
                var o = new Promise((function (e, a) {
                        n = i[t] = [e, a]
                    }
                ));
                e.push(n[2] = o);
                var s, c = document.createElement("script");
                c.charset = "utf-8",
                    c.timeout = 120,
                r.nc && c.setAttribute("nonce", r.nc),
                    c.src = function (t) {
                        return r.p + "" + ({}[t] || t) + ".502689ef.js"
                    }(t);
                var l = new Error;
                s = function (e) {
                    c.onerror = c.onload = null,
                        clearTimeout(u);
                    var n = i[t];
                    if (0 !== n) {
                        if (n) {
                            var a = e && ("load" === e.type ? "missing" : e.type)
                                , r = e && e.target && e.target.src;
                            l.message = "Loading chunk " + t + " failed.\n(" + a + ": " + r + ")",
                                l.name = "ChunkLoadError",
                                l.type = a,
                                l.request = r,
                                n[1](l)
                        }
                        i[t] = void 0
                    }
                }
                ;
                var u = setTimeout((function () {
                        s({
                            type: "timeout",
                            target: c
                        })
                    }
                ), 12e4);
                c.onerror = c.onload = s,
                    document.head.appendChild(c)
            }
        return Promise.all(e)
    }
        ,
        r.m = t,
        r.c = n,
        r.d = function (t, e, n) {
            r.o(t, e) || Object.defineProperty(t, e, {
                enumerable: !0,
                get: n
            })
        }
        ,
        r.r = function (t) {
            "undefined" != typeof Symbol && Symbol.toStringTag && Object.defineProperty(t, Symbol.toStringTag, {
                value: "Module"
            }),
                Object.defineProperty(t, "__esModule", {
                    value: !0
                })
        }
        ,
        r.t = function (t, e) {
            if (1 & e && (t = r(t)),
            8 & e)
                return t;
            if (4 & e && "object" == typeof t && t && t.__esModule)
                return t;
            var n = Object.create(null);
            if (r.r(n),
                Object.defineProperty(n, "default", {
                    enumerable: !0,
                    value: t
                }),
            2 & e && "string" != typeof t)
                for (var a in t)
                    r.d(n, a, function (e) {
                        return t[e]
                    }
                        .bind(null, a));
            return n
        }
        ,
        r.n = function (t) {
            var e = t && t.__esModule ? function () {
                    return t.default
                }
                : function () {
                    return t
                }
            ;
            return r.d(e, "a", e),
                e
        }
        ,
        r.o = function (t, e) {
            return Object.prototype.hasOwnProperty.call(t, e)
        }
        ,
        r.p = "//qcc-static.qichacha.com/qcc/pc-web/prod-1.9.99/",
        r.oe = function (t) {
            throw t
        }
    ;
    var o = window.webpackJsonp = window.webpackJsonp || []
        , s = o.push.bind(o);
    o.push = e,
        o = o.slice();
    for (var c = 0; c < o.length; c++)
        e(o[c]);
    var l = s;
    // r(r.s = 6585)
}({
    4222: function (t, e, n) {
        var a;
        t.exports = (a = n(415),
            n(3117),
            n(4223),
            n(4224),
            a.HmacSHA512)
    },
    415: function (t, e, n) {
        var a;
        t.exports = (a = a || function (t, e) {
            var n = Object.create || function () {
                function t() {
                }

                return function (e) {
                    var n;
                    return t.prototype = e,
                        n = new t,
                        t.prototype = null,
                        n
                }
            }()
                , a = {}
                , i = a.lib = {}
                , r = i.Base = {
                extend: function (t) {
                    var e = n(this);
                    return t && e.mixIn(t),
                    e.hasOwnProperty("init") && this.init !== e.init || (e.init = function () {
                            e.$super.init.apply(this, arguments)
                        }
                    ),
                        e.init.prototype = e,
                        e.$super = this,
                        e
                },
                create: function () {
                    var t = this.extend();
                    return t.init.apply(t, arguments),
                        t
                },
                init: function () {
                },
                mixIn: function (t) {
                    for (var e in t)
                        t.hasOwnProperty(e) && (this[e] = t[e]);
                    t.hasOwnProperty("toString") && (this.toString = t.toString)
                },
                clone: function () {
                    return this.init.prototype.extend(this)
                }
            }
                , o = i.WordArray = r.extend({
                init: function (t, e) {
                    t = this.words = t || [],
                        this.sigBytes = null != e ? e : 4 * t.length
                },
                toString: function (t) {
                    return (t || c).stringify(this)
                },
                concat: function (t) {
                    var e = this.words
                        , n = t.words
                        , a = this.sigBytes
                        , i = t.sigBytes;
                    if (this.clamp(),
                    a % 4)
                        for (var r = 0; r < i; r++) {
                            var o = n[r >>> 2] >>> 24 - r % 4 * 8 & 255;
                            e[a + r >>> 2] |= o << 24 - (a + r) % 4 * 8
                        }
                    else
                        for (r = 0; r < i; r += 4)
                            e[a + r >>> 2] = n[r >>> 2];
                    return this.sigBytes += i,
                        this
                },
                clamp: function () {
                    var e = this.words
                        , n = this.sigBytes;
                    e[n >>> 2] &= 4294967295 << 32 - n % 4 * 8,
                        e.length = t.ceil(n / 4)
                },
                clone: function () {
                    var t = r.clone.call(this);
                    return t.words = this.words.slice(0),
                        t
                },
                random: function (e) {
                    for (var n, a = [], i = function (e) {
                        e = e;
                        var n = 987654321
                            , a = 4294967295;
                        return function () {
                            var i = ((n = 36969 * (65535 & n) + (n >> 16) & a) << 16) + (e = 18e3 * (65535 & e) + (e >> 16) & a) & a;
                            return i /= 4294967296,
                            (i += .5) * (t.random() > .5 ? 1 : -1)
                        }
                    }, r = 0; r < e; r += 4) {
                        var s = i(4294967296 * (n || t.random()));
                        n = 987654071 * s(),
                            a.push(4294967296 * s() | 0)
                    }
                    return new o.init(a, e)
                }
            })
                , s = a.enc = {}
                , c = s.Hex = {
                stringify: function (t) {
                    for (var e = t.words, n = t.sigBytes, a = [], i = 0; i < n; i++) {
                        var r = e[i >>> 2] >>> 24 - i % 4 * 8 & 255;
                        a.push((r >>> 4).toString(16)),
                            a.push((15 & r).toString(16))
                    }
                    return a.join("")
                },
                parse: function (t) {
                    for (var e = t.length, n = [], a = 0; a < e; a += 2)
                        n[a >>> 3] |= parseInt(t.substr(a, 2), 16) << 24 - a % 8 * 4;
                    return new o.init(n, e / 2)
                }
            }
                , l = s.Latin1 = {
                stringify: function (t) {
                    for (var e = t.words, n = t.sigBytes, a = [], i = 0; i < n; i++) {
                        var r = e[i >>> 2] >>> 24 - i % 4 * 8 & 255;
                        a.push(String.fromCharCode(r))
                    }
                    return a.join("")
                },
                parse: function (t) {
                    for (var e = t.length, n = [], a = 0; a < e; a++)
                        n[a >>> 2] |= (255 & t.charCodeAt(a)) << 24 - a % 4 * 8;
                    return new o.init(n, e)
                }
            }
                , u = s.Utf8 = {
                stringify: function (t) {
                    try {
                        return decodeURIComponent(escape(l.stringify(t)))
                    } catch (t) {
                        throw new Error("Malformed UTF-8 data")
                    }
                },
                parse: function (t) {
                    return l.parse(unescape(encodeURIComponent(t)))
                }
            }
                , d = i.BufferedBlockAlgorithm = r.extend({
                reset: function () {
                    this._data = new o.init,
                        this._nDataBytes = 0
                },
                _append: function (t) {
                    "string" == typeof t && (t = u.parse(t)),
                        this._data.concat(t),
                        this._nDataBytes += t.sigBytes
                },
                _process: function (e) {
                    var n = this._data
                        , a = n.words
                        , i = n.sigBytes
                        , r = this.blockSize
                        , s = i / (4 * r)
                        , c = (s = e ? t.ceil(s) : t.max((0 | s) - this._minBufferSize, 0)) * r
                        , l = t.min(4 * c, i);
                    if (c) {
                        for (var u = 0; u < c; u += r)
                            this._doProcessBlock(a, u);
                        var d = a.splice(0, c);
                        n.sigBytes -= l
                    }
                    return new o.init(d, l)
                },
                clone: function () {
                    var t = r.clone.call(this);
                    return t._data = this._data.clone(),
                        t
                },
                _minBufferSize: 0
            })
                , f = (i.Hasher = d.extend({
                cfg: r.extend(),
                init: function (t) {
                    this.cfg = this.cfg.extend(t),
                        this.reset()
                },
                reset: function () {
                    d.reset.call(this),
                        this._doReset()
                },
                update: function (t) {
                    return this._append(t),
                        this._process(),
                        this
                },
                finalize: function (t) {
                    return t && this._append(t),
                        this._doFinalize()
                },
                blockSize: 16,
                _createHelper: function (t) {
                    return function (e, n) {
                        return new t.init(n).finalize(e)
                    }
                },
                _createHmacHelper: function (t) {
                    return function (e, n) {
                        return new f.HMAC.init(t, n).finalize(e)
                    }
                }
            }),
                a.algo = {});
            return a
        }(Math),
            a)
    },
    3117: function (t, e, n) {
        var a, i, r, o, s, c;
        t.exports = (c = n(415),
            i = (a = c).lib,
            r = i.Base,
            o = i.WordArray,
            (s = a.x64 = {}).Word = r.extend({
                init: function (t, e) {
                    this.high = t,
                        this.low = e
                }
            }),
            s.WordArray = r.extend({
                init: function (t, e) {
                    t = this.words = t || [],
                        this.sigBytes = null != e ? e : 8 * t.length
                },
                toX32: function () {
                    for (var t = this.words, e = t.length, n = [], a = 0; a < e; a++) {
                        var i = t[a];
                        n.push(i.high),
                            n.push(i.low)
                    }
                    return o.create(n, this.sigBytes)
                },
                clone: function () {
                    for (var t = r.clone.call(this), e = t.words = this.words.slice(0), n = e.length, a = 0; a < n; a++)
                        e[a] = e[a].clone();
                    return t
                }
            }),
            c)
    },
    4223: function (t, e, n) {
        var a;
        t.exports = (a = n(415),
            n(3117),
            function () {
                var t = a
                    , e = t.lib.Hasher
                    , n = t.x64
                    , i = n.Word
                    , r = n.WordArray
                    , o = t.algo;

                function s() {
                    return i.create.apply(i, arguments)
                }

                var c = [s(1116352408, 3609767458), s(1899447441, 602891725), s(3049323471, 3964484399), s(3921009573, 2173295548), s(961987163, 4081628472), s(1508970993, 3053834265), s(2453635748, 2937671579), s(2870763221, 3664609560), s(3624381080, 2734883394), s(310598401, 1164996542), s(607225278, 1323610764), s(1426881987, 3590304994), s(1925078388, 4068182383), s(2162078206, 991336113), s(2614888103, 633803317), s(3248222580, 3479774868), s(3835390401, 2666613458), s(4022224774, 944711139), s(264347078, 2341262773), s(604807628, 2007800933), s(770255983, 1495990901), s(1249150122, 1856431235), s(1555081692, 3175218132), s(1996064986, 2198950837), s(2554220882, 3999719339), s(2821834349, 766784016), s(2952996808, 2566594879), s(3210313671, 3203337956), s(3336571891, 1034457026), s(3584528711, 2466948901), s(113926993, 3758326383), s(338241895, 168717936), s(666307205, 1188179964), s(773529912, 1546045734), s(1294757372, 1522805485), s(1396182291, 2643833823), s(1695183700, 2343527390), s(1986661051, 1014477480), s(2177026350, 1206759142), s(2456956037, 344077627), s(2730485921, 1290863460), s(2820302411, 3158454273), s(3259730800, 3505952657), s(3345764771, 106217008), s(3516065817, 3606008344), s(3600352804, 1432725776), s(4094571909, 1467031594), s(275423344, 851169720), s(430227734, 3100823752), s(506948616, 1363258195), s(659060556, 3750685593), s(883997877, 3785050280), s(958139571, 3318307427), s(1322822218, 3812723403), s(1537002063, 2003034995), s(1747873779, 3602036899), s(1955562222, 1575990012), s(2024104815, 1125592928), s(2227730452, 2716904306), s(2361852424, 442776044), s(2428436474, 593698344), s(2756734187, 3733110249), s(3204031479, 2999351573), s(3329325298, 3815920427), s(3391569614, 3928383900), s(3515267271, 566280711), s(3940187606, 3454069534), s(4118630271, 4000239992), s(116418474, 1914138554), s(174292421, 2731055270), s(289380356, 3203993006), s(460393269, 320620315), s(685471733, 587496836), s(852142971, 1086792851), s(1017036298, 365543100), s(1126000580, 2618297676), s(1288033470, 3409855158), s(1501505948, 4234509866), s(1607167915, 987167468), s(1816402316, 1246189591)]
                    , l = [];
                !function () {
                    for (var t = 0; t < 80; t++)
                        l[t] = s()
                }();
                var u = o.SHA512 = e.extend({
                    _doReset: function () {
                        this._hash = new r.init([new i.init(1779033703, 4089235720), new i.init(3144134277, 2227873595), new i.init(1013904242, 4271175723), new i.init(2773480762, 1595750129), new i.init(1359893119, 2917565137), new i.init(2600822924, 725511199), new i.init(528734635, 4215389547), new i.init(1541459225, 327033209)])
                    },
                    _doProcessBlock: function (t, e) {
                        for (var n = this._hash.words, a = n[0], i = n[1], r = n[2], o = n[3], s = n[4], u = n[5], d = n[6], f = n[7], p = a.high, h = a.low, v = i.high, m = i.low, g = r.high, y = r.low, b = o.high, _ = o.low, x = s.high, C = s.low, w = u.high, k = u.low, O = d.high, S = d.low, I = f.high, P = f.low, T = p, D = h, j = v, M = m, E = g, L = y, N = b, A = _, R = x, z = C, H = w, F = k, $ = O, V = S, B = I, q = P, K = 0; K < 80; K++) {
                            var G = l[K];
                            if (K < 16)
                                var W = G.high = 0 | t[e + 2 * K]
                                    , U = G.low = 0 | t[e + 2 * K + 1];
                            else {
                                var Y = l[K - 15]
                                    , X = Y.high
                                    , Z = Y.low
                                    , J = (X >>> 1 | Z << 31) ^ (X >>> 8 | Z << 24) ^ X >>> 7
                                    , Q = (Z >>> 1 | X << 31) ^ (Z >>> 8 | X << 24) ^ (Z >>> 7 | X << 25)
                                    , tt = l[K - 2]
                                    , et = tt.high
                                    , nt = tt.low
                                    , at = (et >>> 19 | nt << 13) ^ (et << 3 | nt >>> 29) ^ et >>> 6
                                    , it = (nt >>> 19 | et << 13) ^ (nt << 3 | et >>> 29) ^ (nt >>> 6 | et << 26)
                                    , rt = l[K - 7]
                                    , ot = rt.high
                                    , st = rt.low
                                    , ct = l[K - 16]
                                    , lt = ct.high
                                    , ut = ct.low;
                                W = (W = (W = J + ot + ((U = Q + st) >>> 0 < Q >>> 0 ? 1 : 0)) + at + ((U += it) >>> 0 < it >>> 0 ? 1 : 0)) + lt + ((U += ut) >>> 0 < ut >>> 0 ? 1 : 0),
                                    G.high = W,
                                    G.low = U
                            }
                            var dt, ft = R & H ^ ~R & $, pt = z & F ^ ~z & V, ht = T & j ^ T & E ^ j & E,
                                vt = D & M ^ D & L ^ M & L,
                                mt = (T >>> 28 | D << 4) ^ (T << 30 | D >>> 2) ^ (T << 25 | D >>> 7),
                                gt = (D >>> 28 | T << 4) ^ (D << 30 | T >>> 2) ^ (D << 25 | T >>> 7),
                                yt = (R >>> 14 | z << 18) ^ (R >>> 18 | z << 14) ^ (R << 23 | z >>> 9),
                                bt = (z >>> 14 | R << 18) ^ (z >>> 18 | R << 14) ^ (z << 23 | R >>> 9), _t = c[K],
                                xt = _t.high, Ct = _t.low, wt = B + yt + ((dt = q + bt) >>> 0 < q >>> 0 ? 1 : 0),
                                kt = gt + vt;
                            B = $,
                                q = V,
                                $ = H,
                                V = F,
                                H = R,
                                F = z,
                                R = N + (wt = (wt = (wt = wt + ft + ((dt += pt) >>> 0 < pt >>> 0 ? 1 : 0)) + xt + ((dt += Ct) >>> 0 < Ct >>> 0 ? 1 : 0)) + W + ((dt += U) >>> 0 < U >>> 0 ? 1 : 0)) + ((z = A + dt | 0) >>> 0 < A >>> 0 ? 1 : 0) | 0,
                                N = E,
                                A = L,
                                E = j,
                                L = M,
                                j = T,
                                M = D,
                                T = wt + (mt + ht + (kt >>> 0 < gt >>> 0 ? 1 : 0)) + ((D = dt + kt | 0) >>> 0 < dt >>> 0 ? 1 : 0) | 0
                        }
                        h = a.low = h + D,
                            a.high = p + T + (h >>> 0 < D >>> 0 ? 1 : 0),
                            m = i.low = m + M,
                            i.high = v + j + (m >>> 0 < M >>> 0 ? 1 : 0),
                            y = r.low = y + L,
                            r.high = g + E + (y >>> 0 < L >>> 0 ? 1 : 0),
                            _ = o.low = _ + A,
                            o.high = b + N + (_ >>> 0 < A >>> 0 ? 1 : 0),
                            C = s.low = C + z,
                            s.high = x + R + (C >>> 0 < z >>> 0 ? 1 : 0),
                            k = u.low = k + F,
                            u.high = w + H + (k >>> 0 < F >>> 0 ? 1 : 0),
                            S = d.low = S + V,
                            d.high = O + $ + (S >>> 0 < V >>> 0 ? 1 : 0),
                            P = f.low = P + q,
                            f.high = I + B + (P >>> 0 < q >>> 0 ? 1 : 0)
                    },
                    _doFinalize: function () {
                        var t = this._data
                            , e = t.words
                            , n = 8 * this._nDataBytes
                            , a = 8 * t.sigBytes;
                        return e[a >>> 5] |= 128 << 24 - a % 32,
                            e[30 + (a + 128 >>> 10 << 5)] = Math.floor(n / 4294967296),
                            e[31 + (a + 128 >>> 10 << 5)] = n,
                            t.sigBytes = 4 * e.length,
                            this._process(),
                            this._hash.toX32()
                    },
                    clone: function () {
                        var t = e.clone.call(this);
                        return t._hash = this._hash.clone(),
                            t
                    },
                    blockSize: 32
                });
                t.SHA512 = e._createHelper(u),
                    t.HmacSHA512 = e._createHmacHelper(u)
            }(),
            a.SHA512)
    },
    4224: function (t, e, n) {
        var a, i, r, o;
        t.exports = (a = n(415),
            r = (i = a).lib.Base,
            o = i.enc.Utf8,
            void (i.algo.HMAC = r.extend({
                init: function (t, e) {
                    t = this._hasher = new t.init,
                    "string" == typeof e && (e = o.parse(e));
                    var n = t.blockSize
                        , a = 4 * n;
                    e.sigBytes > a && (e = t.finalize(e)),
                        e.clamp();
                    for (var i = this._oKey = e.clone(), r = this._iKey = e.clone(), s = i.words, c = r.words, l = 0; l < n; l++)
                        s[l] ^= 1549556828,
                            c[l] ^= 909522486;
                    i.sigBytes = r.sigBytes = a,
                        this.reset()
                },
                reset: function () {
                    var t = this._hasher;
                    t.reset(),
                        t.update(this._iKey)
                },
                update: function (t) {
                    return this._hasher.update(t),
                        this
                },
                finalize: function (t) {
                    var e = this._hasher
                        , n = e.finalize(t);
                    return e.reset(),
                        e.finalize(this._oKey.clone().concat(n))
                }
            })))
    },
    3: function (t, e) {
        t.exports = function (t) {
            return t && t.__esModule ? t : {
                default: t
            }
        }
            ,
            t.exports.default = t.exports,
            t.exports.__esModule = !0
    }
})

var _n = 20, codes = {
    "0": "W",
    "1": "l",
    "2": "k",
    "3": "B",
    "4": "Q",
    "5": "g",
    "6": "f",
    "7": "i",
    "8": "i",
    "9": "r",
    "10": "v",
    "11": "6",
    "12": "A",
    "13": "K",
    "14": "N",
    "15": "k",
    "16": "4",
    "17": "L",
    "18": "1",
    "19": "8"
};

function r_default() {
    for (var t = arguments.length > 0 && void 0 !== arguments[0] ? arguments[0] : "/", e = t.toLowerCase(), n = e + e, a = "", r = 0; r < n.length; ++r) {
        var o = n[r].charCodeAt() % _n;
        a += codes[o]
    }
    return a
}

function i_default(t, e) {
    return aaa(4222)(t, e).toString()
}

function getName(e) {
    var t = arguments.length > 0 && void 0 !== arguments[0] ? arguments[0] : "/"
        , e = t.toLowerCase();
    return i_default(e, r_default(e)).toLowerCase().substr(10, 20)
}

function getValue(e, data) {
    var t = arguments.length > 0 && void 0 !== arguments[0] ? arguments[0] : "/"
        , e = arguments.length > 1 && void 0 !== arguments[1] ? arguments[1] : {}
        , n = t.toLowerCase()
        , a = JSON.stringify(e).toLowerCase();
    return (0,
        i_default)(n + n + a, (0,
        r_default)(n))
}

function getHeadersParam(e, data) {
    name = getName(e)
    value = getValue(e, data)
    item = {}
    item[name] = value
    // console.log(item)
    return item
}

`)
	if err != nil {
		return nil
	}
	object, _ := vm.Object("(" + data + ")")
	call, err := vm.Call("getHeadersParam", nil, url, object)
	if err != nil {
		return nil
	}
	return call.Object()
}
