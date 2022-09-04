package chinaz

import (
	"crypto/tls"
	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/robertkrimen/otto"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func GetEnInfoByPid(options *common.ENOptions) (ensInfos *common.EnInfos, ensOutMap map[string]*outputfile.ENSMap) {
	ensInfos = &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensOutMap = make(map[string]*outputfile.ENSMap)
	field := []string{"webName", "host", "host", "permit", "owner", "inFrom"}
	keyWord := []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称", "数据关联"}
	ensOutMap["icp"] = &outputfile.ENSMap{Name: "icp", Field: field, KeyWord: keyWord}
	gologger.Infof("ChinaZ API 查询 %s\n", options.KeyWord)
	enRes := getReq("https://icp.chinaz.com/"+url.QueryEscape(options.KeyWord), map[string]string{"kw": ""}, options)
	re := regexp.MustCompile(`var enkey = '(.+?)'`)
	rr := re.FindStringSubmatch(enRes)
	if len(rr) == 0 {
		gologger.Errorf("ChinaZ %s 签名计算失败\n", options.KeyWord)
		return
	}
	rSing := hSing(options.KeyWord, rr[1])
	token, _ := rSing.Get("token")
	enKey, _ := rSing.Get("enKey")
	randomNum, _ := rSing.Get("randomNum")
	data := map[string]string{
		"Kw":        options.KeyWord,
		"pageNo":    "1",
		"pageSize":  "20",
		"token":     token.String(),
		"enKey":     enKey.String(),
		"randomNum": randomNum.String(),
	}
	content := getReq("https://icp.chinaz.com/Home/PageData", data, options)
	pageCount := gjson.Get(content, "amount").Int()
	var resList []gjson.Result
	resList = append(resList, gjson.Get(content, "data").Array()...)
	if pageCount > 20 {
		for i := 2; int(pageCount/20) >= i-1; i++ {
			data["pageNo"] = strconv.Itoa(i)
			gologger.Infof("爬取中【%d/%d】\n", i, int(pageCount/20))
			ress := getReq("https://icp.chinaz.com/Home/PageData", data, options)
			resList = append(resList, gjson.Get(ress, "data").Array()...)
		}
	}

	ensInfos.Infos["icp"] = resList
	ensInfos.Name = options.KeyWord
	gologger.Infof("ChinaZ 查询到 %d 条数据\n", len(resList))
	if options.IsShow {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(keyWord)
		for _, v := range resList {
			res := gjson.GetMany(v.Raw, field...)
			var str []string
			for _, vv := range res {
				str = append(str, vv.String())

			}
			table.Append(str)
		}
		table.Render()
	}

	return ensInfos, ensOutMap
}

func getReq(url string, data map[string]string, options *common.ENOptions) string {
	//安全延时
	time.Sleep(time.Duration(options.DelayTime) * time.Second)

	//计算签名
	//构造ChinaZ请求
	client := resty.New()
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxy(options.Proxy)
	}
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.Header = http.Header{
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36"},
		"Content-Type": {"application/json;charset=UTF-8"},
		"Cookie":       {options.ENConfig.Cookies.ChinaZ},
		"Referer":      {"https://icp.chinaz.com/"},
	}
	clientR := client.R()
	if data["Kw"] != "" {
		clientR.Method = "POST"
		clientR.SetFormData(data)
	} else {
		clientR.Method = "GET"
	}

	clientR.URL = url
	resp, err := clientR.Send()
	if err != nil {
		gologger.Errorf("【ChinaZ】请求发生错误，%s 5秒后重试\n%s\n", url, err)
		time.Sleep(5 * time.Second)
		return getReq(url, data, options)
	}
	if resp.StatusCode() == 200 {
		if !strings.Contains(string(resp.Body()), "会员登录 - 企查查") {
			return string(resp.Body())
		} else {
			gologger.Errorf("【ChinaZ】需要登陆操作\n%s\n", err)
			return ""
		}

	} else if resp.StatusCode() == 403 {
		gologger.Errorf("【ChinaZ】ip被禁止访问网站，请更换ip\n")
	} else if resp.StatusCode() == 401 {
		gologger.Errorf("【ChinaZ】Cookie有问题或过期，请重新获取\n")
	} else if resp.StatusCode() == 301 {
		gologger.Errorf("【ChinaZ】需要更新Cookie\n")
	} else if resp.StatusCode() == 404 {
		gologger.Errorf("【ChinaZ】请求错误 404 %s\n", url)
	} else {
		gologger.Errorf("【ChinaZ】未知错误 %s\n", resp.StatusCode())
		return ""
	}

	if strings.Contains(string(resp.Body()), "使用该功能需要用户登录") {
		gologger.Errorf("【ChinaZ】Cookie有问题或过期，请重新获取\n")
	}
	return ""
}

func hSing(keyword string, enKey string) *otto.Object {
	vm := otto.New()
	_, err := vm.Run(`
var _0xodZ = 'jsjiami.com.v6'
    , _0xodZ_ = ['I_0xodZ']
    ,
    _0x42ef = [_0xodZ, 'W8KQacKheQ==', 'w6bDucOCQMOZJcOoJWfDpQ==', 'wrovw5sdRQ==', 'EHnCmCTCr8Ke', 'wqHCisOKbcK1', 'QF90w7Er', 'Pl7CjzvClg==', 'wozCtMO4YHI=', 'LsKZScO5LHg=', 'fMOcF8Oaw7TCvA==', 'MzEDwp5l', 'wrxKwpDDhXM=', 'Hxt2Rl8=', 'MzETJmnClg==', 'NG/CkTPClQ==', 'wonCiQsxEw==', 'woPDlcKXwpXDig==', 'wpzDgMKVwpvDisKb', 'BGDDtlrDig==', 'PDwCUsOu', 'wpFswrrDqyg=', 'w75twqNmCQ==', 'wrvDs8KdwpbDmQ==', 'wqxSwp3DnkRYw5bCv8Ofwr0=', 'd8KKwrXCiMOg', 'PsKESsO4GX4=', 'HC8EwpRFwp4UDDBM', 'DcK5w4/CsHs=', 'w6gTM3PCo1Q=', 'bcKOUcKjQw==', 'ZXfDvzk5', 'wrIMOzEz', 'UcOMYMOTwrk=', 'LCFCQEY=', 'L0DDp1HDvA==', 'PiE1Lks=', 'bx9/TGo=', 'VmDClHx7', 'aULCmX5X', 'TMKJwr7CkcOa', 'BRQoUcOP', 'bShJPSzCgsOBBcKl', 'bWbCm09J', 'fcOPK8OVw4U=', 'wp7DvEfDsykK', 'K8OPwp7Chg8=', 'w4zDtcOhP8Ko', 'w5DDhh50', 'J2TDoTrDhg==', 'VWpbS2M=', 'd0JWbXQ=', 'bFFgXG8=', 'EsKtbiFW', 'w5vDpMO3MsKU', 'Dx8mwrNv', 'HGkjw71hwrLCssOMwrc=', 'SUvDjRID', 'wqTDm3HDkh8=', 'MB9mChR9QcKWPVwuw4LCsB4NN8O2wqbCh8OvXyAkEMKNwq1OH1TCm2jDh8K6C8Okwowzw4swIMKwdwvCtAzCtTPCj8OcwoHDscKawrAGNsKgHFPDsMKxJ8KqNgfCs2DClX7DrmTCrMOVwpVCAsOHw6jCq1PCvMOWSC/CuDTDrMKxZ8KLJzjCmlHCmHHDsTN6YsOAXsKQwq7CnsO6w7vCiigKe8OJw7fCjhzCkTg6w4o+w7HDn8K1JDV9JsOsWnjCr0ZQw4tPw7F7fCbCjMOfw41rfMOnw743UcOJw6vDscK5wpprI2ALXXXDiRPDt23CksKdw5rCrzbDvhfCgcOCdwxJw5fCrjjCgj7Dggcuw5PCp01uwo/Dn8OzwqsWFSbCk8KYecO+w5FoPRtJwrEjZQ==', 'wqXDgMKIwrfDkQ==', 'DcKYw6DCu0s=', 'w4zCt8OBwq/CvQ==', 'wqANBBkC', 'wqnCucKfw5HDoQ==', 'L8K2bsKUXw==', 'KMOVwr7CszY=', 'w7/DhsKtPSc=', 'w67CucOcwpDCrw==', 'XHrCnmZb', 'OMKwTz5n', 'aytZUmY=', 'wp94wqLDqiA=', 'AMKVVcK5TQ==', 'D8OSCCAz', 'UWNxw7gU', 'w79NwrBAAA==', 'wo4+LsOgw6g=', 'W8K6JBUb', 'EG3DgWHDlw==', 'EU3CnzTCqQ==', 'wo0UIsO6w7k=', 'wqJswoXDoGM=', 'w6TDmsK3Pxo=', 'BwPCkgrDtg==', 'wrHCq8OqTsKJ', 'w5NFwrtFMw==', 'FcOMCis1IMKUwo3CiMOt', 'LRQbd8OW', 'ccOwPMOww6k=', 'c8K+MB8k', 'w77DicOdDsKK', 'w7pqwplZDg==', 'bWAZwoLDog==', 'wqzCrCkFJA==', 'HsKeWsONMg==', 'woQKHsOmw64=', 'DF/DkhrDhg==', 'HyZEU0Y=', 'wpXDiMKLwqrDtw==', 'w5rDqMO9AMKV', 'DwbCoDDDvQ==', 'WWVbw5E7', 'EsKdRsOOCg==', 'SVp3f38=', 'wozDucK4KsOtPkDCmBQpTTTDm2PDrU00MHIcN1Q=', 'd2YwwrjDkA==', 'TDdnTnM=', 'GsOBBT4CJw==', 'wplbwqLDnDhY', 'w4TDmcK+BzTDvD/DlWR5', 'aMKHw7Bzw6Q=', 'EyILwoFywpk=', 'w61QVsKAFQ==', 'wqHCjcK9w73Dlg==', 'w5HDkMOLNcK9', 'DcOwworCjhU=', 'CFBzIA==', 'J8KCXRtF', 'wp7DsMK9wp/Dnw==', 'wokZA8OFw7TDkw==', 'BkHDtjvDg0nCvcOtccOL', 'wqXChsO9b8K3', 'w6J1Q8K2IQ==', 'fcO2f8Oewr8=', 'wqbChcOBaVw=', 'w6vDusKYERY=', 'w40SHlHCvw==', 'NXjCtQbCsw==', 'GVh7AsKd', 'wrEww7Ynbw==', 'HjnCoB5c', 'w4bDu8OuS8OK', 'w6pKdcK9Iw==', 'w5PDpsKuNiE=', 'HjzCoSXDog==', 'akrDtBsL', 'wrHDqcKrwqTDqA==', 'BF5NB8KS', 'cm7Cr1FX', 'ScK9ZMK1Wg==', 'w4bClsOtwo3Clg==', 'ZVsOwrPDtg==', 'wpZNwp3Dohk=', 'KcO5wpbCgwU=', 'wpvClMKpw6zDlQ==', 'EUbDhD3Dsk/Ct8Ov', 'D0BpJ8Kgwoo=', 'aHMywrbDkMKP', 'w53ChcO+wpfCrA==', 'KA8RwoFP', 'G8KjwoRsw5U=', 'QsKXTcKETw==', 'BHfCsBDCnQ==', 'w6xuwr1mEQ==', 'OcKifMONCA==', 'KjItQsOFwpLCg8KHKcOT', 'w7R4woF8wrE=', 'wrDClMOrSsK/', 'HsKiwpB+w6U=', 'woA4w6k9Wg==', 'XWPDgSQG', 'AMK3ccKXeA==', 'wrjCpgcCNA==', 'w5VuwopMPA==', 'HUJJFsKc', 'OMKswpVcw54=', 'HMK0wrJLw6TDsjrDjQ==', 'wp/DqEvDpDIV', 'w6bDucOCQMObPg==', 'wrM2w6ABQQ==', 'ECIYLlc=', 'S2x4X14=', 'e8K9wrvCucOo', 'acO0FcOzw7U=', 'YnJ7SX4=', 'csKqw40fAg==', 'LkjChALCuA==', 'NMKoTcO5EA==', 'wonCogsaHQ==', 'w5lpbcKGKQ==', 'YXnCjFxp', 'DMOlwpbCiwACbA==', 'UsKJw6AvJDA=', 'wqrCqMONb8KMUDJFHMKE', 'Tmd9w6wSwqbDoMOCw4Z/wopW', 'HGjDlRjDgw==', 'wpNMwqPDlg9Ywo5UwoMgwrXDjg==', 'w6dnbsKFNA==', 'eMOtI8O6w60=', 'wozCtwM7Gg==', 'WcKLwr3ClcOtEQ3DusKDLkkQ', 'UcONEMOuw54SwrVLwqw9Lxg=', 'PMOpwpTClDE=', 'wqDCicOfU8K7', 'DGDDpAfDtA==', 'C8KzwoBNw5XDtDDDj8KTNw==', 'f8KfADQ9', 'wq/CssODcMKMVzdSHsKfwqdq', 'Z8K4wobCvcOD', 'wpbDl8KUwpHDvcKbSBDDrmzDl8Ow', 'wpLCsTgVGw==', 'dMKYScKlaw==', 'XMKOZMK6SsOHw6nCmcKmPcOGQA==', 'wp1qwpbDvCE=', 'XsO5MMOlw60=', 'w6wiB1PCug==', 'FTDCnixb', 'LygjXcOFwpXChsKQK8OIwqfCmA==', 'Jj4ywr5D', 'YsO2E8Ouw5w=', 'UsObScOowps=', 'BMKVfMOSHQ==', 'wqg1w6QUeg==', 'wp/CtcOzX3k=', 'TkFyWX8=', 'YH96w6MD', 'SsO8C8OZw4Y=', 'bWbClX1M', 'B8KpWcOuHg==', 'wrMnw7Exfw==', 'wqcXP8OOw6Q=', 'REBIV0k=', 'TmNyTHs=', 'wrwRBcOHw64=', 'ARIkR8OQ', 'BUHDplfDtw==', 'YFsEwqfDgw==', 'axd0f2w=', 'SMKBwrXCosOc', 'w4PDvMKHAxA=', 'wq3CjcO0a8Ko', 'w4XDsMOhBcK3', 'BS4Rwo5e', 'CCo5Q8Ot', 'w7t4wrBJwpk=', 'f8Kcw7s7Ow==', 'Km3DhiHDqA==', 'wqApN8OKw7A=', 'wqo4PMOKw6g=', 'LcKOwrtXw6Y=', 'MMOQwonCkwA=', 'BwojRMOn', 'Y8KJX8KgYg==', 'XEzDujQ5', 'O3vDhFPDrQ==', 'T8K+Kjoi', 'wrsxw700Yw==', 'HD8OfsOi', 'w6jDusOEdsOq', 'D8KdaMODOg==', 'KMKaej1U', 'wpfCjgUWBg==', 'asK4wpbCr8O9', 'J0LDuzPDpQ==', 'aidhXHc=', 'Jg02DXk=', 'G8OKwqrChi4=', 'wq3Cn8OsVmo=', 'wofCq8K1w6zDkw==', 'OMKLXcO7Fg==', 'cm8GwpnDkw==', 'wp3Cg8OrS2g=', 'wppIwr3DtHE=', 'Z8OXMcOow6I=', 'LhHCpAl8', 'wrnClhMYOA==', 'HcKbcwVf', 'acOxS8O0wpI=', 'LH3DkS3Dlg==', 'N8OALiwF', 'K8KRfiVO', 'w4V0Q8KxFg==', 'WMO1KcO/w5k=', 'TFoMwpPDvQ==', 'wrDCgcO+ZcKZ', 'C03DtD3Dpg==', 'ARYccsOf', 'TcKdwrjCm8OJ', 'wpcYB8OBw6c=', 'L0XDpVjDjg==', 'SRthRHY=', 'wqTDkFHDhwI=', 'P8K3wptew50=', 'C03CnyvCtQ==', 'wpFowqnDqBw=', 'w5hYRsKYFw==', 'GcK1e8KBaMODw73Cjg==', 'CyIdwpI=', 'w6FJwqhMwrUb', 'fcK3MSgB', 'woHDhsKywp7Dvw==', 'wpQfJMOAw4E=', 'AiHCthbDjg==', 'FsOKwoDClxs=', 'esKHw49Pw4U=', 'Z349wqPDp8KIw7jDpkPCkA==', 'w7J8wq9yOw==', 'cAp9QmrCtg==', 'KUjDjw/Dkw==', 'FSMpBH4=', 'gjsjiamiNPOwyu.cYowFVm.v6CuDgGb=='];
if (function (_0x42c482, _0x2caeef, _0x203733) {
    function _0xb3cddb(_0x31263b, _0x249f70, _0xac9d63, _0x289cd3, _0x1f8594, _0xe1af9c) {
        _0x249f70 = _0x249f70 >> 0x8,
            _0x1f8594 = 'po';
        var _0x61f1cf = 'shift'
            , _0x5675e3 = 'push'
            , _0xe1af9c = 'I';
        if (_0x249f70 < _0x31263b) {
            while (--_0x31263b) {
                _0x289cd3 = _0x42c482[_0x61f1cf]();
                if (_0x249f70 === _0x31263b && _0xe1af9c === 'I' && _0xe1af9c['length'] === 0x1) {
                    _0x249f70 = _0x289cd3,
                        _0xac9d63 = _0x42c482[_0x1f8594 + 'p']();
                } else if (_0x249f70 && _0xac9d63['replace'](/[gNPOwyuYwFVCuDgGb=]/g, '') === _0x249f70) {
                    _0x42c482[_0x5675e3](_0x289cd3);
                }
            }
            _0x42c482[_0x5675e3](_0x42c482[_0x61f1cf]());
        }
        return 0xf29ae;
    }
    ;
    return _0xb3cddb(++_0x2caeef, _0x203733) >> _0x2caeef ^ _0x203733;
}(_0x42ef, 0x11e, 0x11e00),
    _0x42ef) {
    _0xodZ_ = _0x42ef['length'] ^ 0x11e;
}
;

function _0xb861(_0x152687, _0x23ad4b) {
    _0x152687 = ~~'0x'['concat'](_0x152687['slice'](0x1));
    var _0x22b8b4 = _0x42ef[_0x152687];
    if (_0xb861['yHKjMp'] === undefined) {
        (function () {
            var _0x51a262 = typeof window !== 'undefined' ? window : typeof process === 'object' && typeof require === 'function' && typeof global === 'object' ? global : this;
            var _0x288f61 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
            _0x51a262['atob'] || (_0x51a262['atob'] = function (_0xb3176e) {
                    var _0x18f651 = String(_0xb3176e)['replace'](/=+$/, '');
                    for (var _0x37c04d = 0x0, _0x2f49d5, _0x43b8e6, _0xfdcaa = 0x0, _0x4ea504 = ''; _0x43b8e6 = _0x18f651['charAt'](_0xfdcaa++); ~_0x43b8e6 && (_0x2f49d5 = _0x37c04d % 0x4 ? _0x2f49d5 * 0x40 + _0x43b8e6 : _0x43b8e6,
                    _0x37c04d++ % 0x4) ? _0x4ea504 += String['fromCharCode'](0xff & _0x2f49d5 >> (-0x2 * _0x37c04d & 0x6)) : 0x0) {
                        _0x43b8e6 = _0x288f61['indexOf'](_0x43b8e6);
                    }
                    return _0x4ea504;
                }
            );
        }());

        function _0x5637d8(_0x4dbadd, _0x23ad4b) {
            var _0x37410b = [], _0x3b56ce = 0x0, _0x3c399d, _0x5dbb83 = '', _0x28d69d = '';
            _0x4dbadd = atob(_0x4dbadd);
            for (var _0x2998f3 = 0x0, _0x37ddd1 = _0x4dbadd['length']; _0x2998f3 < _0x37ddd1; _0x2998f3++) {
                _0x28d69d += '%' + ('00' + _0x4dbadd['charCodeAt'](_0x2998f3)['toString'](0x10))['slice'](-0x2);
            }
            _0x4dbadd = decodeURIComponent(_0x28d69d);
            for (var _0x3978e9 = 0x0; _0x3978e9 < 0x100; _0x3978e9++) {
                _0x37410b[_0x3978e9] = _0x3978e9;
            }
            for (_0x3978e9 = 0x0; _0x3978e9 < 0x100; _0x3978e9++) {
                _0x3b56ce = (_0x3b56ce + _0x37410b[_0x3978e9] + _0x23ad4b['charCodeAt'](_0x3978e9 % _0x23ad4b['length'])) % 0x100;
                _0x3c399d = _0x37410b[_0x3978e9];
                _0x37410b[_0x3978e9] = _0x37410b[_0x3b56ce];
                _0x37410b[_0x3b56ce] = _0x3c399d;
            }
            _0x3978e9 = 0x0;
            _0x3b56ce = 0x0;
            for (var _0x4ef890 = 0x0; _0x4ef890 < _0x4dbadd['length']; _0x4ef890++) {
                _0x3978e9 = (_0x3978e9 + 0x1) % 0x100;
                _0x3b56ce = (_0x3b56ce + _0x37410b[_0x3978e9]) % 0x100;
                _0x3c399d = _0x37410b[_0x3978e9];
                _0x37410b[_0x3978e9] = _0x37410b[_0x3b56ce];
                _0x37410b[_0x3b56ce] = _0x3c399d;
                _0x5dbb83 += String['fromCharCode'](_0x4dbadd['charCodeAt'](_0x4ef890) ^ _0x37410b[(_0x37410b[_0x3978e9] + _0x37410b[_0x3b56ce]) % 0x100]);
            }
            return _0x5dbb83;
        }

        _0xb861['XIxmKZ'] = _0x5637d8;
        _0xb861['aWSjpa'] = {};
        _0xb861['yHKjMp'] = !![];
    }
    var _0xa40d8d = _0xb861['aWSjpa'][_0x152687];
    if (_0xa40d8d === undefined) {
        if (_0xb861['fDLHrm'] === undefined) {
            _0xb861['fDLHrm'] = !![];
        }
        _0x22b8b4 = _0xb861['XIxmKZ'](_0x22b8b4, _0x23ad4b);
        _0xb861['aWSjpa'][_0x152687] = _0x22b8b4;
    } else {
        _0x22b8b4 = _0xa40d8d;
    }
    return _0x22b8b4;
}
;

function generateKey() {
    var _0x13b232 = {
        'jGVpu': function (_0x4c1d43, _0xf57a57) {
            return _0x4c1d43 + _0xf57a57;
        },
        'qcIbA': function (_0x1dd6d9, _0x480291) {
            return _0x1dd6d9 + _0x480291;
        },
        'hJfpz': function (_0x5e6e40, _0x207437) {
            return _0x5e6e40 + _0x207437;
        },
        'IZfsc': function (_0x518a60, _0x34343a, _0x21713d) {
            return _0x518a60(_0x34343a, _0x21713d);
        },
        'LaXFS': function (_0x4f7f5a, _0x221dae) {
            return _0x4f7f5a !== _0x221dae;
        },
        'JwTEc': 'wHHIZ',
        'albvp': function (_0x46bb88, _0x2dd934) {
            return _0x46bb88 < _0x2dd934;
        },
        'YPkKp': function (_0x206498, _0x397021) {
            return _0x206498 == _0x397021;
        },
        'zSyfb': function (_0x15774a, _0x3b0bb8) {
            return _0x15774a - _0x3b0bb8;
        },
        'BlXJB': function (_0x81b975, _0x4c3e69) {
            return _0x81b975(_0x4c3e69);
        },
        'BByxM': function (_0x44cfb9, _0x19040c) {
            return _0x44cfb9 + _0x19040c;
        },
        'DERsl': function (_0x6bd2ef, _0x178f7c) {
            return _0x6bd2ef + _0x178f7c;
        },
        'menGP': function (_0x2bc6fb, _0x5abfe1) {
            return _0x2bc6fb - _0x5abfe1;
        }
    };
    var _0x458db7 = $(_0xb861('U0', 'Ztb*'))[_0xb861('I1', '29sF')]();
    if (!_0x458db7)
        return '';
    var _0x112fbc = _0x458db7['split']('.');
    if (_0x112fbc[_0xb861('U2', 'b9h5')] != 0x4)
        return '';
    var _0x2c02a6 = _0x13b232[_0xb861('I3', ']NQ3')](_0x13b232[_0xb861('U4', 'hIIn')](_0x13b232[_0xb861('U5', 'b9C^')](_0x13b232['qcIbA'](_0x13b232[_0xb861('U6', 'aN)9')](_0x13b232[_0xb861('I7', 'dX^C')](_0x112fbc[0x3], '.'), _0x112fbc[0x2]), '.'), _0x112fbc[0x1]), '.'), _0x112fbc[0x0]);
    var _0xe15dff = _0x2c02a6[_0xb861('I8', '0QPn')]('.');
    var _0x5a6017 = '';
    var _0x14daa4 = '.'[_0xb861('I9', 'Oyg4')]();
    var _0x539c97 = _0x13b232[_0xb861('Ia', 'CZos')](getRandom, 0xa, 0x63);
    for (var _0x48fdbb = 0x0; _0x48fdbb < _0xe15dff[_0xb861('Ub', 'T!FO')]; _0x48fdbb++) {
        if (_0x13b232[_0xb861('Uc', '7%qN')](_0x13b232['JwTEc'], _0x13b232[_0xb861('Id', ']UsO')])) {
            utftext += String['fromCharCode'](c);
        } else {
            var _0x2e1303 = 0x0;
            for (var _0x21c78e = 0x0; _0x13b232[_0xb861('Ie', 'Ztb*')](_0x21c78e, _0xe15dff[_0x48fdbb]['length']); _0x21c78e++) {
                var _0x13b7af = _0xe15dff[_0x48fdbb]['charAt'](_0x21c78e);
                var _0x429253 = _0x13b7af[_0xb861('If', '^b4Y')]();
                _0x2e1303 = _0x2e1303 + _0x429253;
            }
            if (_0x13b232[_0xb861('I10', 'cD0l')](_0x48fdbb, _0x13b232['zSyfb'](_0xe15dff[_0xb861('U11', 'V4Vs')], 0x1)))
                _0x2e1303 = _0x13b232[_0xb861('U12', 'sS*@')](_0x2e1303, _0x539c97);
            else
                _0x2e1303 = _0x13b232[_0xb861('U13', '8U0q')](_0x2e1303, _0x14daa4) + _0x539c97;
            _0x5a6017 += _0x2e1303 + ',';
        }
    }
    return _0x13b232['BlXJB'](encodeURIComponent, _0x13b232[_0xb861('U14', 'V4Vs')](_0x13b232[_0xb861('U15', '0DyB')](_0x539c97, ','), _0x5a6017[_0xb861('U16', 'WItV')](0x0, _0x13b232['menGP'](_0x5a6017[_0xb861('U17', '[cQx')], 0x1))));
}

function generateHostKey(_0x4cdced) {
    var _0x3de45b = {
        'BoYng': function (_0x3b3255, _0x58db93, _0x5ea65d) {
            return _0x3b3255(_0x58db93, _0x5ea65d);
        },
        'VjvEI': function (_0xdf8463, _0x1c0d6d, _0x38407f, _0x23dcdf) {
            return _0xdf8463(_0x1c0d6d, _0x38407f, _0x23dcdf);
        },
        'BhHHg': function (_0x257ebf, _0x2c9125) {
            return _0x257ebf == _0x2c9125;
        },
        'ePBCP': function (_0x37daec, _0x5f35f5) {
            return _0x37daec - _0x5f35f5;
        },
        'HsgpN': function (_0x4c018e, _0x561458) {
            return _0x4c018e + _0x561458;
        },
        'PswKl': function (_0xa72e0b, _0x28ab9b) {
            return _0xa72e0b < _0x28ab9b;
        },
        'XLxhS': function (_0x4fbc03, _0x429fae) {
            return _0x4fbc03 !== _0x429fae;
        },
        'dRvPd': _0xb861('I18', '29sF'),
        'EKjgQ': 'IxdLM',
        'soXcj': function (_0x47902f, _0x196775) {
            return _0x47902f - _0x196775;
        },
        'WrZtJ': function (_0x85e933, _0x57f167) {
            return _0x85e933 + _0x57f167;
        },
        'auHoV': function (_0x2d76f2, _0x1908cb) {
            return _0x2d76f2 + _0x1908cb;
        },
        'lHymA': function (_0x3d34ed, _0x578126) {
            return _0x3d34ed + _0x578126;
        }
    };
    if (!_0x4cdced)
        return '';
    var _0x3963e4 = _0x4cdced[_0xb861('U19', 'VG7]')]('.');
    if (_0x3de45b['BhHHg'](_0x3963e4['length'], 0x0))
        return '';
    var _0x284409 = '';
    for (var _0x13bfe9 = _0x3de45b[_0xb861('I1a', '0uq&')](_0x3963e4[_0xb861('I1b', ']UsO')], 0x1); _0x13bfe9 >= 0x0; _0x13bfe9--) {
        _0x284409 += _0x3de45b[_0xb861('U1c', 'V4Vs')]('.', _0x3963e4[_0x13bfe9]);
    }
    _0x284409 = _0x284409[_0xb861('I1d', '^lJT')](0x1);
    var _0x61cacf = _0x284409[_0xb861('I1e', 'hIIn')]('.');
    var _0x5b617d = '';
    var _0xb9729f = '.'[_0xb861('I9', 'Oyg4')]();
    var _0x541bb8 = getRandom(0x64, 0x3e7);
    for (var _0x13bfe9 = 0x0; _0x3de45b['PswKl'](_0x13bfe9, _0x61cacf[_0xb861('U1f', 'hIIn')]); _0x13bfe9++) {
        if (_0x3de45b[_0xb861('U20', 'z&ze')](_0xb861('I21', 'r(gn'), _0x3de45b[_0xb861('I22', 'Ppa7')])) {
            var _0x93a987 = 0x0;
            for (var _0x66af51 = 0x0; _0x3de45b['PswKl'](_0x66af51, _0x61cacf[_0x13bfe9]['length']); _0x66af51++) {
                if (_0x3de45b['XLxhS'](_0x3de45b[_0xb861('I23', 'CZos')], _0xb861('U24', 'hIIn'))) {
                    var _0x9b49d8 = _0x61cacf[_0x13bfe9]['charAt'](_0x66af51);
                    var _0x43b205 = _0x9b49d8[_0xb861('I25', 'VG7]')]();
                    _0x93a987 = _0x3de45b[_0xb861('U26', 'dadA')](_0x93a987, _0x43b205);
                } else {
                    var _0x2fbfa0 = _0x61cacf[_0x13bfe9][_0xb861('U27', 'WItV')](_0x66af51);
                    var _0x155a10 = _0x2fbfa0[_0xb861('U28', '29sF')]();
                    _0x93a987 = _0x93a987 + _0x155a10;
                }
            }
            if (_0x3de45b[_0xb861('U29', '6VQd')](_0x13bfe9, _0x3de45b['soXcj'](_0x61cacf[_0xb861('U2a', 'jP67')], 0x1)))
                _0x93a987 = _0x3de45b[_0xb861('I2b', 'Ztb*')](_0x93a987, _0x541bb8);
            else
                _0x93a987 = _0x3de45b['WrZtJ'](_0x93a987 + _0xb9729f, _0x541bb8);
            _0x5b617d += _0x3de45b[_0xb861('I2c', '3IpG')](',', _0x93a987);
        } else {
            a = AEWbp14rxc_mGr51T5WIX(a, _0x3de45b[_0xb861('U2d', 'zxOA')](AEWbp14rxc_mGr51T5WIX, _0x3de45b[_0xb861('U2e', 'KWC!')](AEWbp14rxc_mGr51T5WIX, _0x3de45b[_0xb861('I2f', '0uq&')](AEWbp14rxc_H, b, c, d), x), ac));
            return AEWbp14rxc_mGr51T5WIX(_0x3de45b['BoYng'](AEWbp14rxc_RotateLeft, a, s), b);
        }
    }
    _0x5b617d = _0x5b617d[_0xb861('U30', 'z&ze')](0x1);
    return _0x3de45b[_0xb861('U31', ']UsO')](_0x3de45b['lHymA'](_0x541bb8, ','), _0x5b617d);
}

function generateWordKey(_0x40222e) {
    var _0x988bab = {
        'igNOb': function (_0x3b7c1e, _0x31fcac) {
            return _0x3b7c1e < _0x31fcac;
        },
        'LNdaI': function (_0x40568d, _0x587274) {
            return _0x40568d + _0x587274;
        }
    };
    var _0x21a9b8 = '6|3|0|5|2|1|4'[_0xb861('I32', 'T!FO')]('|')
        , _0x3b3266 = 0x0;
    while (!![]) {
        switch (_0x21a9b8[_0x3b3266++]) {
            case '0':
                var _0x43b587 = getRandom(0x64, 0x3e7);
                continue;
            case '1':
                _0x22f036 = _0x22f036[_0xb861('U33', 'Qj!^')](0x1);
                continue;
            case '2':
                for (var _0x1dee7f = 0x0; _0x988bab['igNOb'](_0x1dee7f, _0x5b8bb4[_0xb861('Ub', 'T!FO')]); _0x1dee7f++) {
                    var _0x2c5126 = _0x5b8bb4[_0x1dee7f]['charCodeAt']();
                    var _0x3ac27d = _0x988bab[_0xb861('I34', 'Qj!^')](_0x2c5126, _0x43b587);
                    _0x22f036 += _0x988bab['LNdaI'](',', _0x3ac27d);
                }
                continue;
            case '3':
                var _0x5b8bb4 = _0x40222e[_0xb861('I35', 'dadA')]('');
                continue;
            case '4':
                return _0x988bab[_0xb861('U36', 'r(gn')](_0x43b587, ',') + _0x22f036;
            case '5':
                var _0x22f036 = '';
                continue;
            case '6':
                if (!_0x40222e)
                    return '';
                continue;
        }
        break;
    }
}

function getRandom(_0x5882d9, _0x344a2b) {
    var _0x384a82 = {
        'HjfPW': function (_0x2e4925, _0x4016fa) {
            return _0x2e4925(_0x4016fa);
        },
        'JpTVX': function (_0x1bb7a7, _0x604e28) {
            return _0x1bb7a7 + _0x604e28;
        },
        'DlLkA': function (_0x2d7044, _0x13d36b) {
            return _0x2d7044 * _0x13d36b;
        },
        'UOxan': function (_0x528f75, _0x241181) {
            return _0x528f75 + _0x241181;
        }
    };
    var _0x2de225 = _0xb861('U37', ']UsO')['split']('|')
        , _0x4c808d = 0x0;
    while (!![]) {
        switch (_0x2de225[_0x4c808d++]) {
            case '0':
                var _0x33ffd4 = _0x384a82[_0xb861('I38', 'Qj!^')](parseInt, _0x384a82[_0xb861('U39', 'huea')](_0x384a82['DlLkA'](Math[_0xb861('U3a', 'uzRM')](), _0x384a82[_0xb861('U3b', 'dX^C')](_0x3a446c - _0x3de285, 0x1)), _0x3de285));
                continue;
            case '1':
                if (_0x3a446c < _0x3de285) {
                    _0x3a446c = _0x5882d9;
                    _0x3de285 = _0x344a2b;
                }
                continue;
            case '2':
                var _0x3a446c = _0x344a2b;
                continue;
            case '3':
                return _0x33ffd4;
            case '4':
                var _0x3de285 = _0x5882d9;
                continue;
        }
        break;
    }
}

function getRandomNum(_0x546604) {
    if (!_0x546604)
        return '';
    return _0x546604[_0xb861('U3c', 'kgEi')](',')[0x0];
}

function generateHostMD5Key(_0x11c798, _0x1e8f96) {
    var _0x51d917 = {
        'nCZtH': function (_0x5192f1, _0xf5b91, _0x45c4bf) {
            return _0x5192f1(_0xf5b91, _0x45c4bf);
        },
        'BMvsF': function (_0xc130a4, _0xe1fe0a) {
            return _0xc130a4 + _0xe1fe0a;
        },
        'SAApN': _0xb861('U3d', '^lJT')
    };
    return _0x51d917['nCZtH'](AEWbp14rxc_MD5, _0x51d917[_0xb861('I3e', '7%qN')](_0x11c798, _0x51d917[_0xb861('I3f', 'wmD8')]) + _0x1e8f96, 0x20);
}

function AEWbp14rxc_MD5(_0x4d95b9, _0x424a01) {
    var _0x1d95a6 = {
        'UesKo': function (_0x39fc71, _0x2a9abe) {
            return _0x39fc71 | _0x2a9abe;
        },
        'BIgCW': function (_0x151e31, _0x26f01b) {
            return _0x151e31 << _0x26f01b;
        },
        'tqSLt': function (_0x1a5306, _0x47cc53) {
            return _0x1a5306 >>> _0x47cc53;
        },
        'SQscS': function (_0x161bde, _0x458c75) {
            return _0x161bde - _0x458c75;
        },
        'mhBlB': function (_0x89e440, _0x57eecf) {
            return _0x89e440 / _0x57eecf;
        },
        'xkFSF': function (_0x4a4064, _0x1deee7) {
            return _0x4a4064 * _0x1deee7;
        },
        'WHtgI': function (_0x4400e3, _0x1bcda4) {
            return _0x4400e3 % _0x1bcda4;
        },
        'hcrDk': function (_0x17d9fc, _0x568d21) {
            return _0x17d9fc | _0x568d21;
        },
        'zyowB': function (_0x1a6115, _0x331b55) {
            return _0x1a6115 & _0x331b55;
        },
        'AJlHf': function (_0x277cc5, _0x2d2ddc) {
            return _0x277cc5 ^ _0x2d2ddc;
        },
        'PnfFV': function (_0x2ebbfc, _0x2c13a0) {
            return _0x2ebbfc ^ _0x2c13a0;
        },
        'bcPCJ': function (_0x456561, _0x14686d) {
            return _0x456561 ^ _0x14686d;
        },
        'yvcyE': function (_0xeb140f, _0x145d2a) {
            return _0xeb140f ^ _0x145d2a;
        },
        'fadOV': function (_0x5aed44, _0x5d019b) {
            return _0x5aed44 & _0x5d019b;
        },
        'VUXTW': function (_0x4e1b66, _0x34a4c5) {
            return _0x4e1b66 & _0x34a4c5;
        },
        'apSUU': function (_0x2ed2d7, _0x2f0e79) {
            return _0x2ed2d7 + _0x2f0e79;
        },
        'XwrHP': function (_0x2bbace, _0x5b9d16) {
            return _0x2bbace & _0x5b9d16;
        },
        'kgGMk': function (_0x58e6f6, _0x3b8c70) {
            return _0x58e6f6 & _0x3b8c70;
        },
        'REtnZ': function (_0x7333be, _0x455517) {
            return _0x7333be !== _0x455517;
        },
        'wDJwx': _0xb861('U40', 'wmD8'),
        'jFnQl': function (_0x282094, _0x4f63d5) {
            return _0x282094 & _0x4f63d5;
        },
        'DkyAX': function (_0xb2337b, _0x1ec7f1) {
            return _0xb2337b ^ _0x1ec7f1;
        },
        'mVyLd': function (_0x14bd2c, _0x5dee35) {
            return _0x14bd2c !== _0x5dee35;
        },
        'CkhJm': _0xb861('U41', 'wmD8'),
        'ivESF': function (_0x2907b0, _0x258dc6) {
            return _0x2907b0 ^ _0x258dc6;
        },
        'dNWGP': function (_0x28dd26, _0x2b7412) {
            return _0x28dd26 | _0x2b7412;
        },
        'aIEMi': function (_0x28340f, _0x59411a) {
            return _0x28340f & _0x59411a;
        },
        'ALPXV': function (_0x51fd87, _0x1f9756) {
            return _0x51fd87 & _0x1f9756;
        },
        'VIKWR': function (_0x82248e, _0x1c8d83) {
            return _0x82248e >> _0x1c8d83;
        },
        'OBrRg': 'CrqGj',
        'empVI': function (_0x538c56, _0x922bfa, _0x5d532c) {
            return _0x538c56(_0x922bfa, _0x5d532c);
        },
        'bhPTt': function (_0x318d2e, _0x15efd2, _0x1a7dd5, _0x1a52f8) {
            return _0x318d2e(_0x15efd2, _0x1a7dd5, _0x1a52f8);
        },
        'PXtkm': function (_0xdd1bac, _0x1197ef) {
            return _0xdd1bac < _0x1197ef;
        },
        'nUFca': function (_0x1b61bd, _0x2cb9cc) {
            return _0x1b61bd + _0x2cb9cc;
        },
        'xBXAN': _0xb861('U0', 'Ztb*'),
        'MwfKx': function (_0x1cb26b, _0x167635) {
            return _0x1cb26b(_0x167635);
        },
        'lFQrx': function (_0xf7860f, _0x429d54) {
            return _0xf7860f != _0x429d54;
        },
        'ntkzB': function (_0x5bc9d7, _0x11595a) {
            return _0x5bc9d7 + _0x11595a;
        },
        'LKGda': function (_0x105fc7, _0x5df6ac) {
            return _0x105fc7 + _0x5df6ac;
        },
        'IdCEh': function (_0x226bce, _0x229c31) {
            return _0x226bce + _0x229c31;
        },
        'qpIPj': function (_0x461e47, _0x57ef9e) {
            return _0x461e47 !== _0x57ef9e;
        },
        'OqmDR': _0xb861('I42', 'n)t!'),
        'ROFqZ': function (_0xbc3cab, _0x513899, _0xe07763) {
            return _0xbc3cab(_0x513899, _0xe07763);
        },
        'AOjYJ': function (_0x51623a, _0x344257, _0x113330, _0xf16dc2) {
            return _0x51623a(_0x344257, _0x113330, _0xf16dc2);
        },
        'fjphc': function (_0x41a4c5, _0x2f942e, _0x2ef89c) {
            return _0x41a4c5(_0x2f942e, _0x2ef89c);
        },
        'JqNgB': function (_0x252a1c, _0x524523) {
            return _0x252a1c === _0x524523;
        },
        'CjMyP': _0xb861('U43', 'kgEi'),
        'tWqCV': function (_0x1ccce8, _0xecc563, _0x4104cf) {
            return _0x1ccce8(_0xecc563, _0x4104cf);
        },
        'Issrx': function (_0x2e037d, _0xcaca30, _0x27b343, _0x2687ad) {
            return _0x2e037d(_0xcaca30, _0x27b343, _0x2687ad);
        },
        'nHCMd': function (_0x50b43c, _0x492339) {
            return _0x50b43c <= _0x492339;
        },
        'WbRNI': function (_0x5e961a, _0x3c2d12) {
            return _0x5e961a + _0x3c2d12;
        },
        'sAobS': function (_0x2d3ae2, _0x594cb1) {
            return _0x2d3ae2 !== _0x594cb1;
        },
        'aMRbR': _0xb861('U44', '29sF'),
        'kKPOS': '6|10|11|2|3|13|4|14|5|12|1|9|8|7|0',
        'sxeSC': function (_0x4cab8d, _0x541845) {
            return _0x4cab8d - _0x541845;
        },
        'IaXBH': function (_0x52ddf7, _0x1cb362) {
            return _0x52ddf7 - _0x1cb362;
        },
        'wambx': function (_0x374030, _0x12440f) {
            return _0x374030 << _0x12440f;
        },
        'yTGWp': function (_0x2f27bf, _0x481936) {
            return _0x2f27bf - _0x481936;
        },
        'vyqAs': function (_0x3d6081, _0x446d5e) {
            return _0x3d6081 | _0x446d5e;
        },
        'MGbds': function (_0x1d67d9, _0x27ccac) {
            return _0x1d67d9 + _0x27ccac;
        },
        'cGYko': function (_0xb570dc, _0x4cac0a) {
            return _0xb570dc % _0x4cac0a;
        },
        'Yavri': function (_0x2a4bf2, _0x564d9e) {
            return _0x2a4bf2(_0x564d9e);
        },
        'BCePB': 'dcReO',
        'awBBH': function (_0xb96108, _0x3a6371) {
            return _0xb96108 & _0x3a6371;
        },
        'PwtcH': function (_0x373adb, _0x15e869) {
            return _0x373adb * _0x15e869;
        },
        'PIPWt': _0xb861('I45', '8U0q'),
        'OveoJ': function (_0x4c5154, _0x389436) {
            return _0x4c5154 < _0x389436;
        },
        'DDiAF': function (_0x78182d, _0x4c7e9c) {
            return _0x78182d * _0x4c7e9c;
        },
        'yMlNu': function (_0x5bd271, _0x3efed7) {
            return _0x5bd271 === _0x3efed7;
        },
        'dYarS': 'vLRhH',
        'LFCWR': _0xb861('I46', '3IpG'),
        'yABQC': function (_0x52c695, _0x5c4d81) {
            return _0x52c695 < _0x5c4d81;
        },
        'kGkPt': function (_0x4abc5e, _0xe2b71d) {
            return _0x4abc5e | _0xe2b71d;
        },
        'hTZGm': function (_0x22974a, _0x384b7a) {
            return _0x22974a | _0x384b7a;
        },
        'vRail': function (_0x148786, _0x46fbee) {
            return _0x148786 >> _0x46fbee;
        },
        'wNvbP': function (_0x5775de, _0x33bc45) {
            return _0x5775de & _0x33bc45;
        },
        'wtKDM': function (_0x186d96, _0x360749) {
            return _0x186d96 >> _0x360749;
        },
        'YxHYL': function (_0x5f335b, _0x3b3262) {
            return _0x5f335b & _0x3b3262;
        },
        'hoglI': function (_0x3c96f8, _0x27fa2e) {
            return _0x3c96f8 < _0x27fa2e;
        },
        'XATEm': function (_0x2ddfaa, _0x226f12) {
            return _0x2ddfaa > _0x226f12;
        },
        'NdBrb': function (_0x203875, _0x127aed) {
            return _0x203875 >> _0x127aed;
        },
        'iFOfp': function (_0x41ac6d, _0x27eca1) {
            return _0x41ac6d & _0x27eca1;
        },
        'AxpUE': function (_0x2bc458, _0x3ec33c) {
            return _0x2bc458 >> _0x3ec33c;
        },
        'YyWXE': function (_0x638354, _0x45f19a) {
            return _0x638354 | _0x45f19a;
        },
        'UIlmA': function (_0x4b90b, _0x42f0c3) {
            return _0x4b90b & _0x42f0c3;
        },
        'KJTBO': function (_0x3a9704, _0x4c7d96) {
            return _0x3a9704 & _0x4c7d96;
        },
        'BirsP': function (_0x44f4e1, _0x89c3ce) {
            return _0x44f4e1(_0x89c3ce);
        },
        'iIsNt': _0xb861('U47', 'uzRM'),
        'WDYLg': _0xb861('U48', 'wmD8'),
        'HjhbR': function (_0x36c67f, _0x1b9546, _0x357784, _0x3a4f73, _0x952089, _0x2d98f2, _0x31a374, _0x551e54) {
            return _0x36c67f(_0x1b9546, _0x357784, _0x3a4f73, _0x952089, _0x2d98f2, _0x31a374, _0x551e54);
        },
        'ZErdF': function (_0x50d3af, _0x20813a) {
            return _0x50d3af + _0x20813a;
        },
        'PXAgJ': function (_0x8d0f84, _0x1c18f3, _0x361973, _0x5dc401, _0x61a14c, _0xb39c0, _0x48b70f, _0x4530aa) {
            return _0x8d0f84(_0x1c18f3, _0x361973, _0x5dc401, _0x61a14c, _0xb39c0, _0x48b70f, _0x4530aa);
        },
        'qwYYO': function (_0x1435ff, _0x1d637a) {
            return _0x1435ff + _0x1d637a;
        },
        'ceRzC': function (_0x5007ad, _0x23d749) {
            return _0x5007ad + _0x23d749;
        },
        'BkRld': function (_0x230e1b, _0x37ac2d) {
            return _0x230e1b + _0x37ac2d;
        },
        'Ymhen': function (_0x2c6467, _0x1116c2, _0x1b3a08, _0x35cdd7, _0x406c87, _0x1e8848, _0x38e452, _0x2fb474) {
            return _0x2c6467(_0x1116c2, _0x1b3a08, _0x35cdd7, _0x406c87, _0x1e8848, _0x38e452, _0x2fb474);
        },
        'HHhwV': function (_0x21660d, _0x2ba26a) {
            return _0x21660d + _0x2ba26a;
        },
        'dMXvg': function (_0x1f05b5, _0x536581) {
            return _0x1f05b5 + _0x536581;
        },
        'wxgZr': function (_0x530013, _0x584a4b, _0x54f386, _0x3dcd35, _0x4b30f6, _0x487cf2, _0x31b26f, _0x505edb) {
            return _0x530013(_0x584a4b, _0x54f386, _0x3dcd35, _0x4b30f6, _0x487cf2, _0x31b26f, _0x505edb);
        },
        'zulSk': function (_0x467d35, _0x5aea2f, _0x29bb94, _0x3fa3de, _0x1f9a2e, _0x458db5, _0x6b5a37, _0x4da6c8) {
            return _0x467d35(_0x5aea2f, _0x29bb94, _0x3fa3de, _0x1f9a2e, _0x458db5, _0x6b5a37, _0x4da6c8);
        },
        'zithX': function (_0x18765a, _0x3e2875) {
            return _0x18765a + _0x3e2875;
        },
        'Apusk': function (_0x20c217, _0x3fb417, _0x539086, _0x42c40d, _0x5a0cf8, _0x45d8c3, _0x3455fb, _0x54abbc) {
            return _0x20c217(_0x3fb417, _0x539086, _0x42c40d, _0x5a0cf8, _0x45d8c3, _0x3455fb, _0x54abbc);
        },
        'vTvbX': function (_0x644516, _0x2724b0) {
            return _0x644516 + _0x2724b0;
        },
        'ODQhh': function (_0x23be73, _0xd6d3e1) {
            return _0x23be73 + _0xd6d3e1;
        },
        'EUZhp': function (_0x4c1f61, _0x2e9935, _0x52f067, _0x1aeecd, _0x2b6659, _0x15d99a, _0x257d01, _0x29f1b0) {
            return _0x4c1f61(_0x2e9935, _0x52f067, _0x1aeecd, _0x2b6659, _0x15d99a, _0x257d01, _0x29f1b0);
        },
        'NPota': function (_0x2217a2, _0x265b3a, _0x220efe, _0xd406e9, _0x48bc6d, _0x181cdc, _0x475026, _0x5d0450) {
            return _0x2217a2(_0x265b3a, _0x220efe, _0xd406e9, _0x48bc6d, _0x181cdc, _0x475026, _0x5d0450);
        },
        'YuTwk': function (_0x131ebb, _0x403588) {
            return _0x131ebb + _0x403588;
        },
        'XNMbV': function (_0x52a546, _0x31141a, _0x3114f6, _0xef3dbc, _0x32700b, _0x3b3b0a, _0x36caa0, _0x1334eb) {
            return _0x52a546(_0x31141a, _0x3114f6, _0xef3dbc, _0x32700b, _0x3b3b0a, _0x36caa0, _0x1334eb);
        },
        'gWJat': function (_0x293e20, _0x70ed18) {
            return _0x293e20 + _0x70ed18;
        },
        'UeBNd': function (_0x117a72, _0xbe397a) {
            return _0x117a72 + _0xbe397a;
        },
        'mkgDp': function (_0x29efa7, _0x2ce901, _0x348423, _0x409f03, _0x102cbf, _0x345616, _0x4a9248, _0x5974e3) {
            return _0x29efa7(_0x2ce901, _0x348423, _0x409f03, _0x102cbf, _0x345616, _0x4a9248, _0x5974e3);
        },
        'RqCIb': function (_0x5b2119, _0x3ea076) {
            return _0x5b2119 + _0x3ea076;
        },
        'BoAmi': function (_0x449a56, _0x4817eb) {
            return _0x449a56 + _0x4817eb;
        },
        'vHryi': function (_0x5a7de6, _0x19c852) {
            return _0x5a7de6 + _0x19c852;
        },
        'UADWS': function (_0x5b0add, _0x3c6ff8, _0x4860c6, _0xc98735, _0x212374, _0x58e67f, _0x4be321, _0x1c0b9c) {
            return _0x5b0add(_0x3c6ff8, _0x4860c6, _0xc98735, _0x212374, _0x58e67f, _0x4be321, _0x1c0b9c);
        },
        'Bklze': function (_0x5bb055, _0xa7ed77, _0x3ab151, _0x25fc1f, _0x10d8c1, _0x354b08, _0x3ba996, _0x4c9d63) {
            return _0x5bb055(_0xa7ed77, _0x3ab151, _0x25fc1f, _0x10d8c1, _0x354b08, _0x3ba996, _0x4c9d63);
        },
        'yYKLd': function (_0x1c6022, _0x2dd945, _0x2ace93, _0x3020f6, _0x3657c8, _0x4cf15b, _0x59a6ca, _0x15bf09) {
            return _0x1c6022(_0x2dd945, _0x2ace93, _0x3020f6, _0x3657c8, _0x4cf15b, _0x59a6ca, _0x15bf09);
        },
        'eJLaO': function (_0x134157, _0x4ff20c, _0x1e3506, _0x51c31d, _0x3f9d94, _0x367669, _0xd1a577, _0x128ccc) {
            return _0x134157(_0x4ff20c, _0x1e3506, _0x51c31d, _0x3f9d94, _0x367669, _0xd1a577, _0x128ccc);
        },
        'enFEt': function (_0x2fab4e, _0x5330c0) {
            return _0x2fab4e + _0x5330c0;
        },
        'HsNrd': function (_0x3c5e45, _0x52d5d6, _0x1a9f8f, _0x45f29d, _0x1b2778, _0x4f1557, _0x5ee7e9, _0x354447) {
            return _0x3c5e45(_0x52d5d6, _0x1a9f8f, _0x45f29d, _0x1b2778, _0x4f1557, _0x5ee7e9, _0x354447);
        },
        'egvqN': function (_0x4311a1, _0x1b1ae5) {
            return _0x4311a1 + _0x1b1ae5;
        },
        'vyZHw': function (_0x43ccb1, _0x3c3134, _0x2b2ff1, _0x4da15b, _0x3ebfe7, _0xaeef8a, _0x49f626, _0x258a52) {
            return _0x43ccb1(_0x3c3134, _0x2b2ff1, _0x4da15b, _0x3ebfe7, _0xaeef8a, _0x49f626, _0x258a52);
        },
        'UrAXv': function (_0x3dad87, _0x3136a9) {
            return _0x3dad87 + _0x3136a9;
        },
        'wnHUb': function (_0x44cd33, _0x2968c9, _0x546357, _0x243551, _0x124982, _0x19476a, _0x4335db, _0x12f974) {
            return _0x44cd33(_0x2968c9, _0x546357, _0x243551, _0x124982, _0x19476a, _0x4335db, _0x12f974);
        },
        'zYJpb': function (_0x53b760, _0x3a914f) {
            return _0x53b760 + _0x3a914f;
        },
        'CsqJN': function (_0x13fb74, _0x515cb0) {
            return _0x13fb74 + _0x515cb0;
        },
        'AdEus': function (_0x1e4121, _0x400673) {
            return _0x1e4121 + _0x400673;
        },
        'zRrIL': function (_0x13cbde, _0x4b74ef, _0x1fcc26, _0x3d6e7b, _0x3d2bc5, _0x370ab3, _0x1c3548, _0x58ac0a) {
            return _0x13cbde(_0x4b74ef, _0x1fcc26, _0x3d6e7b, _0x3d2bc5, _0x370ab3, _0x1c3548, _0x58ac0a);
        },
        'ITFdV': function (_0x53e1e6, _0x4c078b, _0x511b37, _0x3c39f5, _0x74b86a, _0x14a3b7, _0x15f04a, _0x1f9f94) {
            return _0x53e1e6(_0x4c078b, _0x511b37, _0x3c39f5, _0x74b86a, _0x14a3b7, _0x15f04a, _0x1f9f94);
        },
        'HLPBY': function (_0x55bf5b, _0x3e82a6) {
            return _0x55bf5b + _0x3e82a6;
        },
        'yARxV': function (_0x21eb74, _0x5cbf4d, _0x373fad) {
            return _0x21eb74(_0x5cbf4d, _0x373fad);
        },
        'TOoln': function (_0xe1dd8e, _0x27e744, _0x291a23, _0x4d8232, _0x33878d, _0x4ded8c, _0x39b900, _0x47b695) {
            return _0xe1dd8e(_0x27e744, _0x291a23, _0x4d8232, _0x33878d, _0x4ded8c, _0x39b900, _0x47b695);
        },
        'ndctf': function (_0x2d00db, _0x1e101a, _0x44f8b8, _0x485f01, _0x272ae1, _0x2b6f1b, _0x51f2f1, _0xa9a094) {
            return _0x2d00db(_0x1e101a, _0x44f8b8, _0x485f01, _0x272ae1, _0x2b6f1b, _0x51f2f1, _0xa9a094);
        },
        'rdjcg': function (_0x95d830, _0x2f5730, _0x50a4d3, _0x5af3bb, _0x2ced20, _0x16ef4, _0x17f265, _0x3a9862) {
            return _0x95d830(_0x2f5730, _0x50a4d3, _0x5af3bb, _0x2ced20, _0x16ef4, _0x17f265, _0x3a9862);
        },
        'sikjW': function (_0x201070, _0x52f24e) {
            return _0x201070 == _0x52f24e;
        },
        'Utrah': function (_0x9e0cb, _0x413d17) {
            return _0x9e0cb + _0x413d17;
        },
        'HMxPD': function (_0x24f682, _0x21adcb) {
            return _0x24f682(_0x21adcb);
        },
        'WlzaK': function (_0x2f1867, _0x3c3a57) {
            return _0x2f1867(_0x3c3a57);
        },
        'wQihn': function (_0x3ac1f3, _0x3f1786) {
            return _0x3ac1f3(_0x3f1786);
        },
        'dVeSP': function (_0x2417ed, _0x1dd8dc) {
            return _0x2417ed(_0x1dd8dc);
        },
        'TxCMW': function (_0x3726f5, _0x1e40a9) {
            return _0x3726f5(_0x1e40a9);
        }
    };

    function _0x36d1e6(_0x1be843, _0xfbd5cf) {
        return _0x1d95a6[_0xb861('I49', 'hIIn')](_0x1d95a6[_0xb861('I4a', '6VQd')](_0x1be843, _0xfbd5cf), _0x1d95a6['tqSLt'](_0x1be843, _0x1d95a6['SQscS'](0x20, _0xfbd5cf)));
    }

    function _0x2fde77(_0x5ee1a5, _0x17dfc3) {
        var _0x540406 = {
            'kBCBh': function (_0x3d74cd, _0x41387b) {
                return _0x1d95a6[_0xb861('U4b', 'tu05')](_0x3d74cd, _0x41387b);
            },
            'LJCMo': function (_0xae1a2b, _0x25313e) {
                return _0x1d95a6['AJlHf'](_0xae1a2b, _0x25313e);
            },
            'LAOSN': function (_0x304ed1, _0x4708e5) {
                return _0x1d95a6[_0xb861('U4c', 'zxOA')](_0x304ed1, _0x4708e5);
            },
            'mQiwr': function (_0x4a5c85, _0x13948e) {
                return _0x1d95a6['bcPCJ'](_0x4a5c85, _0x13948e);
            },
            'hhOXy': function (_0x416c93, _0x258d44) {
                return _0x1d95a6['yvcyE'](_0x416c93, _0x258d44);
            },
            'NDFSU': function (_0x1292fb, _0x582378) {
                return _0x1d95a6['yvcyE'](_0x1292fb, _0x582378);
            }
        };
        var _0x352375, _0x107c7e, _0x338961, _0xb89d19, _0x5c5b5c;
        _0x338961 = _0x5ee1a5 & 0x80000000;
        _0xb89d19 = _0x1d95a6[_0xb861('I4d', 'Y4Pa')](_0x17dfc3, 0x80000000);
        _0x352375 = _0x1d95a6[_0xb861('U4e', ')QL5')](_0x5ee1a5, 0x40000000);
        _0x107c7e = _0x1d95a6[_0xb861('U4f', 'dX^C')](_0x17dfc3, 0x40000000);
        _0x5c5b5c = _0x1d95a6['apSUU'](_0x1d95a6[_0xb861('I50', 'q)1)')](_0x5ee1a5, 0x3fffffff), _0x1d95a6[_0xb861('I51', 'tu05')](_0x17dfc3, 0x3fffffff));
        if (_0x1d95a6['kgGMk'](_0x352375, _0x107c7e)) {
            return _0x1d95a6[_0xb861('U52', 'Qj!^')](_0x5c5b5c, 0x80000000) ^ _0x338961 ^ _0xb89d19;
        }
        if (_0x1d95a6['hcrDk'](_0x352375, _0x107c7e)) {
            if (_0x1d95a6[_0xb861('I53', 'n)t!')](_0x1d95a6[_0xb861('U54', 'T!FO')], 'jnalF')) {
                if (_0x1d95a6[_0xb861('I55', 'Ppa7')](_0x5c5b5c, 0x40000000)) {
                    return _0x1d95a6[_0xb861('I56', ')QL5')](_0x1d95a6[_0xb861('U57', 'HV0E')](_0x1d95a6['yvcyE'](_0x5c5b5c, 0xc0000000), _0x338961), _0xb89d19);
                } else {
                    return _0x1d95a6[_0xb861('U58', '8U0q')](_0x1d95a6[_0xb861('U59', 'CZos')](_0x5c5b5c ^ 0x40000000, _0x338961), _0xb89d19);
                }
            } else {
                if (_0x540406[_0xb861('I5a', 'b9C^')](_0x5c5b5c, 0x40000000)) {
                    return _0x540406[_0xb861('I5b', ']NQ3')](_0x540406[_0xb861('U5c', 'z&ze')](_0x540406[_0xb861('U5d', 'V4Vs')](_0x5c5b5c, 0xc0000000), _0x338961), _0xb89d19);
                } else {
                    return _0x540406[_0xb861('I5e', 'b9C^')](_0x540406['NDFSU'](_0x5c5b5c ^ 0x40000000, _0x338961), _0xb89d19);
                }
            }
        } else {
            if (_0x1d95a6[_0xb861('U5f', 'VG7]')](_0x1d95a6[_0xb861('U60', 'q)1)')], _0x1d95a6['CkhJm'])) {
                lWordCount = _0x1d95a6[_0xb861('I61', 'aN)9')](lByteCount - lByteCount % 0x4, 0x4);
                lBytePosition = _0x1d95a6[_0xb861('U62', 'sS*@')](_0x1d95a6['WHtgI'](lByteCount, 0x4), 0x8);
                lWordArray[lWordCount] = _0x1d95a6[_0xb861('U63', 'CZos')](lWordArray[lWordCount], _0x1d95a6['BIgCW'](_0x4d95b9[_0xb861('I64', 'HV0E')](lByteCount), lBytePosition));
                lByteCount++;
            } else {
                return _0x1d95a6['ivESF'](_0x5c5b5c ^ _0x338961, _0xb89d19);
            }
        }
    }

    function _0x2a6020(_0xcfb7e7, _0x147d69, _0x4aa818) {
        return _0x1d95a6[_0xb861('U65', 'r(gn')](_0x1d95a6[_0xb861('I66', '[cQx')](_0xcfb7e7, _0x147d69), ~_0xcfb7e7 & _0x4aa818);
    }

    function _0x3e2bc9(_0xc772bb, _0x3f0ec7, _0x26d77a) {
        return _0x1d95a6[_0xb861('U67', ']NQ3')](_0x1d95a6[_0xb861('U68', 'kgEi')](_0xc772bb, _0x26d77a), _0x1d95a6[_0xb861('U69', 'CZos')](_0x3f0ec7, ~_0x26d77a));
    }

    function _0x179f33(_0x1d0295, _0x1bec35, _0x2232da) {
        return _0x1d95a6['ivESF'](_0x1d95a6[_0xb861('U6a', 'Oyg4')](_0x1d0295, _0x1bec35), _0x2232da);
    }

    function _0x34197f(_0x57f619, _0x47a2c5, _0x1dd19e) {
        var _0x5ef5fd = {
            'avsDn': function (_0x57f619, _0x47a2c5) {
                return _0x1d95a6[_0xb861('I6b', '^lJT')](_0x57f619, _0x47a2c5);
            },
            'Pbviq': function (_0x57f619, _0x47a2c5) {
                return _0x57f619 & _0x47a2c5;
            },
            'Hgpou': function (_0x57f619, _0x47a2c5) {
                return _0x1d95a6['ALPXV'](_0x57f619, _0x47a2c5);
            }
        };
        if (_0xb861('I6c', 'WItV') !== _0x1d95a6['OBrRg']) {
            utftext += String['fromCharCode'](_0x5ef5fd[_0xb861('I6d', 'b9C^')](_0x4ad199, 0xc) | 0xe0);
            utftext += String['fromCharCode'](_0x5ef5fd['Pbviq'](_0x4ad199 >> 0x6, 0x3f) | 0x80);
            utftext += String['fromCharCode'](_0x5ef5fd['Hgpou'](_0x4ad199, 0x3f) | 0x80);
        } else {
            return _0x1d95a6[_0xb861('U6e', '7%qN')](_0x47a2c5, _0x57f619 | ~_0x1dd19e);
        }
    }

    function _0x51409c(_0x587983, _0x5dc7be, _0x30f496, _0x48cc91, _0x27c26b, _0x579f96, _0x390b74) {
        _0x587983 = _0x2fde77(_0x587983, _0x1d95a6['empVI'](_0x2fde77, _0x1d95a6[_0xb861('U6f', '0uq&')](_0x2fde77, _0x2a6020(_0x5dc7be, _0x30f496, _0x48cc91), _0x27c26b), _0x390b74));
        return _0x1d95a6[_0xb861('I70', 'hIIn')](_0x2fde77, _0x1d95a6[_0xb861('I71', 'kgEi')](_0x36d1e6, _0x587983, _0x579f96), _0x5dc7be);
    }
    ;

    function _0x2333e5(_0x55bf06, _0x32ec74, _0x109518, _0x41b97c, _0x2d023c, _0x588505, _0x23950a) {
        _0x55bf06 = _0x2fde77(_0x55bf06, _0x1d95a6['empVI'](_0x2fde77, _0x2fde77(_0x1d95a6['bhPTt'](_0x3e2bc9, _0x32ec74, _0x109518, _0x41b97c), _0x2d023c), _0x23950a));
        return _0x1d95a6['empVI'](_0x2fde77, _0x1d95a6[_0xb861('U72', 'aN)9')](_0x36d1e6, _0x55bf06, _0x588505), _0x32ec74);
    }
    ;

    function _0x4b4557(_0x1cf063, _0x4d4a32, _0x5489c1, _0x2d710c, _0x162889, _0x3343ad, _0x1a8801) {
        if (_0x1d95a6[_0xb861('I73', '8U0q')](_0x1d95a6[_0xb861('U74', 'WItV')], _0x1d95a6[_0xb861('I75', 'wmD8')])) {
            var _0x554128 = _0xb861('U76', 'kgEi')[_0xb861('I77', 'Oyg4')]('|')
                , _0x469b62 = 0x0;
            while (!![]) {
                switch (_0x554128[_0x469b62++]) {
                    case '0':
                        for (var _0x5be8a2 = 0x0; _0x1d95a6[_0xb861('I78', 'T!FO')](_0x5be8a2, _0x51cb57[_0xb861('I79', 'HV0E')]); _0x5be8a2++) {
                            var _0x4bb472 = 0x0;
                            for (var _0x2e77fd = 0x0; _0x2e77fd < _0x51cb57[_0x5be8a2][_0xb861('U7a', 'Ppa7')]; _0x2e77fd++) {
                                var _0x58055e = _0x51cb57[_0x5be8a2]['charAt'](_0x2e77fd);
                                var _0x2ea675 = _0x58055e[_0xb861('I7b', 'q)1)')]();
                                _0x4bb472 = _0x1d95a6[_0xb861('U7c', '0QPn')](_0x4bb472, _0x2ea675);
                            }
                            if (_0x5be8a2 == _0x51cb57[_0xb861('U7d', '29sF')] - 0x1)
                                _0x4bb472 = _0x1d95a6[_0xb861('U7e', 'UYsS')](_0x4bb472, _0x1760d5);
                            else
                                _0x4bb472 = _0x1d95a6[_0xb861('U7f', 'Y4Pa')](_0x4bb472 + _0x5c9426, _0x1760d5);
                            _0x4d98af += _0x1d95a6[_0xb861('U80', 'kgEi')](_0x4bb472, ',');
                        }
                        continue;
                    case '1':
                        var _0x51cb57 = _0x32e97f[_0xb861('I81', 'dX^C')]('.');
                        continue;
                    case '2':
                        var _0x1760d5 = _0x1d95a6['empVI'](getRandom, 0xa, 0x63);
                        continue;
                    case '3':
                        var _0x36fc7b = $(_0x1d95a6['xBXAN'])[_0xb861('I82', '6M^)')]();
                        continue;
                    case '4':
                        return _0x1d95a6[_0xb861('I83', 'n)t!')](encodeURIComponent, _0x1d95a6[_0xb861('I84', 'hIIn')](_0x1760d5, ',') + _0x4d98af['substr'](0x0, _0x1d95a6['SQscS'](_0x4d98af[_0xb861('I85', 'b9C^')], 0x1)));
                    case '5':
                        if (!_0x36fc7b)
                            return '';
                        continue;
                    case '6':
                        var _0x5c9426 = '.'[_0xb861('U86', '7%qN')]();
                        continue;
                    case '7':
                        var _0x4d98af = '';
                        continue;
                    case '8':
                        if (_0x1d95a6[_0xb861('U87', 'sS*@')](_0x39c0e5['length'], 0x4))
                            return '';
                        continue;
                    case '9':
                        var _0x32e97f = _0x1d95a6[_0xb861('I88', 'UYsS')](_0x1d95a6[_0xb861('U89', 'KWC!')](_0x1d95a6[_0xb861('I8a', '0DyB')](_0x1d95a6[_0xb861('I8b', 'q)1)')](_0x1d95a6[_0xb861('U8c', 'jP67')](_0x1d95a6[_0xb861('I8d', 'V4Vs')](_0x39c0e5[0x3], '.'), _0x39c0e5[0x2]), '.'), _0x39c0e5[0x1]), '.'), _0x39c0e5[0x0]);
                        continue;
                    case '10':
                        var _0x39c0e5 = _0x36fc7b[_0xb861('I81', 'dX^C')]('.');
                        continue;
                }
                break;
            }
        } else {
            _0x1cf063 = _0x1d95a6['empVI'](_0x2fde77, _0x1cf063, _0x1d95a6[_0xb861('U8e', '6M^)')](_0x2fde77, _0x1d95a6[_0xb861('I8f', 'cD0l')](_0x2fde77, _0x1d95a6['AOjYJ'](_0x179f33, _0x4d4a32, _0x5489c1, _0x2d710c), _0x162889), _0x1a8801));
            return _0x2fde77(_0x1d95a6['fjphc'](_0x36d1e6, _0x1cf063, _0x3343ad), _0x4d4a32);
        }
    }
    ;

    function _0x46feb6(_0x1f08be, _0x209d10, _0x1f2b2c, _0x25f9f8, _0xc69fde, _0x4323d7, _0x518906) {
        if (_0x1d95a6[_0xb861('I90', 'bqH]')](_0x1d95a6[_0xb861('U91', '^b4Y')], 'vLPYP')) {
            return lResult ^ lX8 ^ lY8;
        } else {
            _0x1f08be = _0x1d95a6[_0xb861('I92', 'UYsS')](_0x2fde77, _0x1f08be, _0x2fde77(_0x1d95a6[_0xb861('U93', 'q)1)')](_0x2fde77, _0x1d95a6['Issrx'](_0x34197f, _0x209d10, _0x1f2b2c, _0x25f9f8), _0xc69fde), _0x518906));
            return _0x2fde77(_0x1d95a6[_0xb861('U94', 'aN)9')](_0x36d1e6, _0x1f08be, _0x4323d7), _0x209d10);
        }
    }
    ;

    function _0x2c8b11(_0x4d95b9) {
        var _0x3c1ebd = {
            'csQYU': function (_0x24d985, _0x54c628) {
                return _0x1d95a6[_0xb861('U95', '3IpG')](_0x24d985, _0x54c628);
            },
            'Wypdd': function (_0x279d8a, _0x4dc9b6) {
                return _0x1d95a6[_0xb861('I96', 'hIIn')](_0x279d8a, _0x4dc9b6);
            },
            'TLRrb': function (_0x55126f, _0x4f0d34) {
                return _0x1d95a6[_0xb861('I97', '6M^)')](_0x55126f, _0x4f0d34);
            },
            'pOlGb': function (_0x40b55c, _0x537eaf) {
                return _0x1d95a6[_0xb861('U98', 'Qj!^')](_0x40b55c, _0x537eaf);
            }
        };
        if (_0x1d95a6[_0xb861('U99', 'Ztb*')](_0xb861('U9a', 'tu05'), _0x1d95a6[_0xb861('I9b', 'Oyg4')])) {
            var _0x473c2b = '', _0x3fe4f7 = '', _0x1c9be2, _0x461243;
            for (_0x461243 = 0x0; _0x3c1ebd[_0xb861('I9c', 'Ppa7')](_0x461243, 0x3); _0x461243++) {
                _0x1c9be2 = _0x3c1ebd[_0xb861('I9d', 'dX^C')](lValue >>> _0x3c1ebd[_0xb861('I9e', 'Y4Pa')](_0x461243, 0x8), 0xff);
                _0x3fe4f7 = '0' + _0x1c9be2[_0xb861('I9f', '7%qN')](0x10);
                _0x473c2b = _0x3c1ebd['pOlGb'](_0x473c2b, _0x3fe4f7[_0xb861('Ia0', '6M^)')](_0x3fe4f7[_0xb861('Ua1', 'Oyg4')] - 0x2, 0x2));
            }
            return _0x473c2b;
        } else {
            var _0x3714b7 = _0x1d95a6[_0xb861('Ua2', 'tu05')]['split']('|')
                , _0x119575 = 0x0;
            while (!![]) {
                switch (_0x3714b7[_0x119575++]) {
                    case '0':
                        return _0x4b428e;
                    case '1':
                        _0x237ebd = _0x1d95a6[_0xb861('Ia3', '29sF')](_0x122ec8, 0x4) * 0x8;
                        continue;
                    case '2':
                        var _0x266a6f = _0x1d95a6[_0xb861('Ua4', 'a%u&')](_0x2e2bef, _0x2e2bef % 0x40) / 0x40;
                        continue;
                    case '3':
                        var _0x20cb06 = _0x1d95a6[_0xb861('Ua5', 'Ztb*')](_0x266a6f + 0x1, 0x10);
                        continue;
                    case '4':
                        var _0x237ebd = 0x0;
                        continue;
                    case '5':
                        while (_0x122ec8 < _0x445628) {
                            _0x2ac4fc = _0x1d95a6['IaXBH'](_0x122ec8, _0x122ec8 % 0x4) / 0x4;
                            _0x237ebd = _0x1d95a6[_0xb861('Ua6', 'V4Vs')](_0x1d95a6[_0xb861('Ia7', 'CZos')](_0x122ec8, 0x4), 0x8);
                            _0x4b428e[_0x2ac4fc] = _0x1d95a6[_0xb861('Ia8', 'WItV')](_0x4b428e[_0x2ac4fc], _0x1d95a6['wambx'](_0x4d95b9[_0xb861('Ua9', 'r(gn')](_0x122ec8), _0x237ebd));
                            _0x122ec8++;
                        }
                        continue;
                    case '6':
                        var _0x2ac4fc;
                        continue;
                    case '7':
                        _0x4b428e[_0x1d95a6[_0xb861('Uaa', 'b9h5')](_0x20cb06, 0x1)] = _0x445628 >>> 0x1d;
                        continue;
                    case '8':
                        _0x4b428e[_0x1d95a6[_0xb861('Uab', 'sS*@')](_0x20cb06, 0x2)] = _0x445628 << 0x3;
                        continue;
                    case '9':
                        _0x4b428e[_0x2ac4fc] = _0x1d95a6[_0xb861('Uac', 'a%u&')](_0x4b428e[_0x2ac4fc], 0x80 << _0x237ebd);
                        continue;
                    case '10':
                        var _0x445628 = _0x4d95b9[_0xb861('Ub', 'T!FO')];
                        continue;
                    case '11':
                        var _0x2e2bef = _0x1d95a6['MGbds'](_0x445628, 0x8);
                        continue;
                    case '12':
                        _0x2ac4fc = _0x1d95a6['yTGWp'](_0x122ec8, _0x1d95a6[_0xb861('Uad', 'cD0l')](_0x122ec8, 0x4)) / 0x4;
                        continue;
                    case '13':
                        var _0x4b428e = _0x1d95a6[_0xb861('Uae', '3IpG')](Array, _0x1d95a6[_0xb861('Iaf', ')QL5')](_0x20cb06, 0x1));
                        continue;
                    case '14':
                        var _0x122ec8 = 0x0;
                        continue;
                }
                break;
            }
        }
    }
    ;

    function _0x39518b(_0x38b98c) {
        if (_0x1d95a6['JqNgB'](_0x1d95a6[_0xb861('Ub0', '^lJT')], _0x1d95a6['BCePB'])) {
            var _0x14234c = '', _0xa08f86 = '', _0x4bd88e, _0x2023dd;
            for (_0x2023dd = 0x0; _0x1d95a6[_0xb861('Ib1', 'CZos')](_0x2023dd, 0x3); _0x2023dd++) {
                _0x4bd88e = _0x1d95a6[_0xb861('Ib2', '6M^)')](_0x1d95a6['tqSLt'](_0x38b98c, _0x1d95a6[_0xb861('Ib3', 'a%u&')](_0x2023dd, 0x8)), 0xff);
                _0xa08f86 = _0x1d95a6['MGbds']('0', _0x4bd88e[_0xb861('Ub4', 'a%u&')](0x10));
                _0x14234c = _0x14234c + _0xa08f86[_0xb861('Ub5', 'uzRM')](_0x1d95a6['yTGWp'](_0xa08f86['length'], 0x2), 0x2);
            }
            return _0x14234c;
        } else {
            var _0x2edb9d = reverseIpArray[i][_0xb861('Ib6', '^b4Y')](j);
            var _0x4b5ed7 = _0x2edb9d['charCodeAt']();
            ipSum = ipSum + _0x4b5ed7;
        }
    }
    ;

    function _0x39fcbc(_0x4d95b9) {
        var _0x548ec4 = {
            'RTrAc': _0x1d95a6[_0xb861('Ib7', 'cD0l')],
            'iDfsH': function (_0x2f50a6, _0x34e2a9) {
                return _0x1d95a6[_0xb861('Ib8', ']UsO')](_0x2f50a6, _0x34e2a9);
            },
            'sGiHk': function (_0x4f475c, _0x471322) {
                return _0x1d95a6[_0xb861('Ub9', 'wmD8')](_0x4f475c, _0x471322);
            },
            'UIhSi': function (_0x27e8f5, _0xa6a6f2) {
                return _0x1d95a6[_0xb861('Uba', 'dadA')](_0x27e8f5, _0xa6a6f2);
            },
            'DuqCw': function (_0x5b9718, _0x58d643) {
                return _0x5b9718 - _0x58d643;
            }
        };
        if (_0x1d95a6[_0xb861('Ubb', '[cQx')](_0x1d95a6[_0xb861('Ubc', 'wmD8')], _0x1d95a6[_0xb861('Ubd', 'O]$J')])) {
            var _0x346bf8 = _0x548ec4[_0xb861('Ibe', 'V4Vs')][_0xb861('I8', '0QPn')]('|')
                , _0x2259b3 = 0x0;
            while (!![]) {
                switch (_0x346bf8[_0x2259b3++]) {
                    case '0':
                        if (_0x548ec4[_0xb861('Ibf', 'WItV')](_0x1cd5af, _0x1a6d7f)) {
                            _0x1cd5af = min;
                            _0x1a6d7f = max;
                        }
                        continue;
                    case '1':
                        var _0x1a6d7f = min;
                        continue;
                    case '2':
                        return _0x1dc30c;
                    case '3':
                        var _0x1dc30c = parseInt(_0x548ec4[_0xb861('Uc0', '^lJT')](_0x548ec4[_0xb861('Ic1', 'UYsS')](Math['random'](), _0x548ec4[_0xb861('Ic2', 'Qj!^')](_0x1cd5af, _0x1a6d7f) + 0x1), _0x1a6d7f));
                        continue;
                    case '4':
                        var _0x1cd5af = max;
                        continue;
                }
                break;
            }
        } else {
            _0x4d95b9 = _0x4d95b9[_0xb861('Uc3', 'dX^C')](/\r\n/g, '\x0a');
            var _0x3884f4 = '';
            for (var _0x49dcfa = 0x0; _0x1d95a6['yABQC'](_0x49dcfa, _0x4d95b9[_0xb861('Uc4', 'O]$J')]); _0x49dcfa++) {
                var _0x1275de = _0x4d95b9[_0xb861('Uc5', 'sS*@')](_0x49dcfa);
                if (_0x1d95a6['yABQC'](_0x1275de, 0x80)) {
                    _0x3884f4 += String[_0xb861('Uc6', '8U0q')](_0x1275de);
                } else if (_0x1275de > 0x7f && _0x1d95a6[_0xb861('Ic7', '7%qN')](_0x1275de, 0x800)) {
                    _0x3884f4 += String['fromCharCode'](_0x1275de >> 0x6 | 0xc0);
                    _0x3884f4 += String[_0xb861('Ic8', 'Ppa7')](_0x1d95a6[_0xb861('Ic9', 'UYsS')](_0x1275de & 0x3f, 0x80));
                } else {
                    _0x3884f4 += String[_0xb861('Uc6', '8U0q')](_0x1d95a6[_0xb861('Uca', '[cQx')](_0x1d95a6[_0xb861('Icb', '^lJT')](_0x1275de, 0xc), 0xe0));
                    _0x3884f4 += String[_0xb861('Ucc', 'dadA')](_0x1d95a6['wNvbP'](_0x1d95a6['wtKDM'](_0x1275de, 0x6), 0x3f) | 0x80);
                    _0x3884f4 += String[_0xb861('Ucd', 'huea')](_0x1d95a6['YxHYL'](_0x1275de, 0x3f) | 0x80);
                }
            }
            return _0x3884f4;
        }
    }
    ;var _0x2fba4c = Array();
    var _0x57c542, _0x216bdc, _0x3b26f6, _0x393bf7, _0x2cf07f, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75;
    var _0x246668 = 0x7
        , _0x5c5edb = 0xc
        , _0x51f6cf = 0x11
        , _0x886338 = 0x16;
    var _0x1695df = 0x5
        , _0x2a3a20 = 0x9
        , _0x127a1d = 0xe
        , _0x2c76c8 = 0x14;
    var _0x334ac8 = 0x4
        , _0x5a26dd = 0xb
        , _0x46df4d = 0x10
        , _0xecc592 = 0x17;
    var _0x12e133 = 0x6
        , _0x3b8ad1 = 0xa
        , _0x3c1ec9 = 0xf
        , _0x58aef = 0x15;
    _0x4d95b9 = _0x1d95a6[_0xb861('Uce', 'dX^C')](_0x39fcbc, _0x4d95b9);
    _0x2fba4c = _0x2c8b11(_0x4d95b9);
    _0x8f6b6f = 0x67452301;
    _0x46b135 = 0xefcdab89;
    _0x4ad199 = 0x98badcfe;
    _0x2e3c75 = 0x10325476;
    for (_0x57c542 = 0x0; _0x1d95a6['hoglI'](_0x57c542, _0x2fba4c[_0xb861('I85', 'b9C^')]); _0x57c542 += 0x10) {
        if (_0x1d95a6[_0xb861('Ucf', 'sS*@')] !== _0x1d95a6[_0xb861('Ud0', '7%qN')]) {
            var _0x10956a = _0x4d95b9[_0xb861('Ud1', 'a%u&')](n);
            if (_0x1d95a6[_0xb861('Ud2', ']NQ3')](_0x10956a, 0x80)) {
                utftext += String[_0xb861('Id3', 'sS*@')](_0x10956a);
            } else if (_0x1d95a6[_0xb861('Ud4', 'dadA')](_0x10956a, 0x7f) && _0x10956a < 0x800) {
                utftext += String[_0xb861('Id5', 'hIIn')](_0x1d95a6[_0xb861('Ud6', '^lJT')](_0x1d95a6[_0xb861('Ud7', 'Ztb*')](_0x10956a, 0x6), 0xc0));
                utftext += String[_0xb861('Ud8', 'Ztb*')](_0x1d95a6[_0xb861('Id9', 'Ppa7')](_0x1d95a6[_0xb861('Uda', 'huea')](_0x10956a, 0x3f), 0x80));
            } else {
                utftext += String['fromCharCode'](_0x1d95a6[_0xb861('Idb', 'jP67')](_0x1d95a6[_0xb861('Udc', 'bqH]')](_0x10956a, 0xc), 0xe0));
                utftext += String[_0xb861('Idd', 'r(gn')](_0x1d95a6[_0xb861('Ude', '29sF')](_0x1d95a6[_0xb861('Udf', 'huea')](_0x1d95a6[_0xb861('Ie0', 'KWC!')](_0x10956a, 0x6), 0x3f), 0x80));
                utftext += String[_0xb861('Ic8', 'Ppa7')](_0x1d95a6[_0xb861('Ie1', 'WItV')](_0x1d95a6[_0xb861('Ue2', 'cD0l')](_0x10956a, 0x3f), 0x80));
            }
        } else {
            var _0x1e97ee = _0x1d95a6[_0xb861('Ue3', '0DyB')]['split']('|')
                , _0x47117c = 0x0;
            while (!![]) {
                switch (_0x1e97ee[_0x47117c++]) {
                    case '0':
                        _0x46b135 = _0x1d95a6[_0xb861('Ie4', 'wmD8')](_0x46feb6, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6['ZErdF'](_0x57c542, 0x9)], _0x58aef, 0xeb86d391);
                        continue;
                    case '1':
                        _0x46b135 = _0x1d95a6[_0xb861('Ue5', '8U0q')](_0x2333e5, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x57c542 + 0x8], _0x2c76c8, 0x455a14ed);
                        continue;
                    case '2':
                        _0x4ad199 = _0x2333e5(_0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6[_0xb861('Ue6', '[cQx')](_0x57c542, 0x7)], _0x127a1d, 0x676f02d9);
                        continue;
                    case '3':
                        _0x4ad199 = _0x1d95a6['HjhbR'](_0x51409c, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x57c542 + 0x6], _0x51f6cf, 0xa8304613);
                        continue;
                    case '4':
                        _0x4ad199 = _0x1d95a6[_0xb861('Ie7', 'Qj!^')](_0x2333e5, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6[_0xb861('Ue8', 'WItV')](_0x57c542, 0x3)], _0x127a1d, 0xf4d50d87);
                        continue;
                    case '5':
                        _0x46b135 = _0x1d95a6[_0xb861('Ie9', 'cD0l')](_0x46feb6, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6['qwYYO'](_0x57c542, 0x1)], _0x58aef, 0x85845dd1);
                        continue;
                    case '6':
                        _0x4ad199 = _0x51409c(_0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6['ceRzC'](_0x57c542, 0xa)], _0x51f6cf, 0xffff5bb1);
                        continue;
                    case '7':
                        _0x8f6b6f = _0x51409c(_0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('Iea', 'b9C^')](_0x57c542, 0x0)], _0x246668, 0xd76aa478);
                        continue;
                    case '8':
                        _0x4ad199 = _0x1d95a6['Ymhen'](_0x2333e5, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6[_0xb861('Ueb', 'wmD8')](_0x57c542, 0xb)], _0x127a1d, 0x265e5a51);
                        continue;
                    case '9':
                        _0x8f6b6f = _0x4b4557(_0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x57c542 + 0x5], _0x334ac8, 0xfffa3942);
                        continue;
                    case '10':
                        _0x2e3c75 = _0x4b4557(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('Iec', 'wmD8')](_0x57c542, 0x0)], _0x5a26dd, 0xeaa127fa);
                        continue;
                    case '11':
                        _0x8f6b6f = _0x1d95a6[_0xb861('Ied', 'b9C^')](_0x4b4557, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('Uee', 'r(gn')](_0x57c542, 0x1)], _0x334ac8, 0xa4beea44);
                        continue;
                    case '12':
                        _0x2e3c75 = _0x1d95a6[_0xb861('Uef', 'z&ze')](_0x46feb6, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('Uf0', 'Oyg4')](_0x57c542, 0xf)], _0x3b8ad1, 0xfe2ce6e0);
                        continue;
                    case '13':
                        _0x2e3c75 = _0x2333e5(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0x6], _0x2a3a20, 0xc040b340);
                        continue;
                    case '14':
                        _0x46b135 = _0x1d95a6[_0xb861('Uf1', 'T!FO')](_0x51409c, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6['dMXvg'](_0x57c542, 0x7)], _0x886338, 0xfd469501);
                        continue;
                    case '15':
                        _0x46b135 = _0x1d95a6['tWqCV'](_0x2fde77, _0x46b135, _0x3b26f6);
                        continue;
                    case '16':
                        _0x4ad199 = _0x2fde77(_0x4ad199, _0x393bf7);
                        continue;
                    case '17':
                        _0x2e3c75 = _0x1d95a6[_0xb861('If2', 'dadA')](_0x2333e5, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0xe], _0x2a3a20, 0xc33707d6);
                        continue;
                    case '18':
                        _0x8f6b6f = _0x46feb6(_0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('Uf3', 'q)1)')](_0x57c542, 0xc)], _0x12e133, 0x655b59c3);
                        continue;
                    case '19':
                        _0x3b26f6 = _0x46b135;
                        continue;
                    case '20':
                        _0x2e3c75 = _0x2333e5(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('Uf4', 'sS*@')](_0x57c542, 0xa)], _0x2a3a20, 0x2441453);
                        continue;
                    case '21':
                        _0x4ad199 = _0x1d95a6['wxgZr'](_0x4b4557, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6['dMXvg'](_0x57c542, 0x3)], _0x46df4d, 0xd4ef3085);
                        continue;
                    case '22':
                        _0x2e3c75 = _0x1d95a6[_0xb861('Uf5', 'kgEi')](_0x46feb6, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6['zithX'](_0x57c542, 0xb)], _0x3b8ad1, 0xbd3af235);
                        continue;
                    case '23':
                        _0x8f6b6f = _0x2333e5(_0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('If6', '29sF')](_0x57c542, 0xd)], _0x1695df, 0xa9e3e905);
                        continue;
                    case '24':
                        _0x2e3c75 = _0x1d95a6[_0xb861('Uf7', 'r(gn')](_0x4b4557, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('Uf8', 'b9h5')](_0x57c542, 0xc)], _0x5a26dd, 0xe6db99e5);
                        continue;
                    case '25':
                        _0x8f6b6f = _0x1d95a6[_0xb861('If9', 'O]$J')](_0x46feb6, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6['ODQhh'](_0x57c542, 0x4)], _0x12e133, 0xf7537e82);
                        continue;
                    case '26':
                        _0x4ad199 = _0x4b4557(_0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6[_0xb861('Ufa', '7%qN')](_0x57c542, 0xf)], _0x46df4d, 0x1fa27cf8);
                        continue;
                    case '27':
                        _0x46b135 = _0x1d95a6[_0xb861('Ifb', 'b9C^')](_0x4b4557, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6[_0xb861('Ifc', 'b9C^')](_0x57c542, 0x6)], _0xecc592, 0x4881d05);
                        continue;
                    case '28':
                        _0x2e3c75 = _0x1d95a6[_0xb861('Ifd', 'a%u&')](_0x51409c, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0x5], _0x5c5edb, 0x4787c62a);
                        continue;
                    case '29':
                        _0x2e3c75 = _0x1d95a6[_0xb861('Ife', 'dX^C')](_0x2333e5, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6['ODQhh'](_0x57c542, 0x2)], _0x2a3a20, 0xfcefa3f8);
                        continue;
                    case '30':
                        _0x4ad199 = _0x1d95a6[_0xb861('Uff', 'r(gn')](_0x46feb6, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6[_0xb861('U100', 'Ztb*')](_0x57c542, 0xa)], _0x3c1ec9, 0xffeff47d);
                        continue;
                    case '31':
                        _0x46b135 = _0x1d95a6[_0xb861('U101', '3IpG')](_0x2333e5, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6[_0xb861('I102', 'z&ze')](_0x57c542, 0x4)], _0x2c76c8, 0xe7d3fbc8);
                        continue;
                    case '32':
                        _0x4ad199 = _0x1d95a6[_0xb861('I103', ']NQ3')](_0x46feb6, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x57c542 + 0x2], _0x3c1ec9, 0x2ad7d2bb);
                        continue;
                    case '33':
                        _0x4ad199 = _0x1d95a6[_0xb861('I104', 'cD0l')](_0x51409c, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6[_0xb861('U105', 'r(gn')](_0x57c542, 0xe)], _0x51f6cf, 0xa679438e);
                        continue;
                    case '34':
                        _0x2e3c75 = _0x46feb6(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0x3], _0x3b8ad1, 0x8f0ccc92);
                        continue;
                    case '35':
                        _0x8f6b6f = _0x1d95a6[_0xb861('U106', '^b4Y')](_0x51409c, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('U107', 'WItV')](_0x57c542, 0xc)], _0x246668, 0x6b901122);
                        continue;
                    case '36':
                        _0x8f6b6f = _0x1d95a6['mkgDp'](_0x2333e5, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('I108', 'n)t!')](_0x57c542, 0x5)], _0x1695df, 0xd62f105d);
                        continue;
                    case '37':
                        _0x2e3c75 = _0x51409c(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0x1], _0x5c5edb, 0xe8c7b756);
                        continue;
                    case '38':
                        _0x393bf7 = _0x4ad199;
                        continue;
                    case '39':
                        _0x4ad199 = _0x1d95a6[_0xb861('U109', '^lJT')](_0x4b4557, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6['vHryi'](_0x57c542, 0x7)], _0x46df4d, 0xf6bb4b60);
                        continue;
                    case '40':
                        _0x2e3c75 = _0x1d95a6[_0xb861('I10a', 'dadA')](_0x4b4557, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0x4], _0x5a26dd, 0x4bdecfa9);
                        continue;
                    case '41':
                        _0x46b135 = _0x1d95a6[_0xb861('I10b', '7%qN')](_0x2333e5, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6[_0xb861('I10c', 'T!FO')](_0x57c542, 0xc)], _0x2c76c8, 0x8d2a4c8a);
                        continue;
                    case '42':
                        _0x8f6b6f = _0x1d95a6[_0xb861('I10d', ']UsO')](_0x4b4557, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x57c542 + 0xd], _0x334ac8, 0x289b7ec6);
                        continue;
                    case '43':
                        _0x8f6b6f = _0x1d95a6[_0xb861('U10e', 'dX^C')](_0x46feb6, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('I10f', '0DyB')](_0x57c542, 0x8)], _0x12e133, 0x6fa87e4f);
                        continue;
                    case '44':
                        _0x8f6b6f = _0x1d95a6[_0xb861('U110', 'Y4Pa')](_0x51409c, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('U111', 'WItV')](_0x57c542, 0x8)], _0x246668, 0x698098d8);
                        continue;
                    case '45':
                        _0x2e3c75 = _0x1d95a6[_0xb861('I112', 'Oyg4')](_0x46feb6, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('U113', '0DyB')](_0x57c542, 0x7)], _0x3b8ad1, 0x432aff97);
                        continue;
                    case '46':
                        _0x2e3c75 = _0x51409c(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('I114', 'VG7]')](_0x57c542, 0xd)], _0x5c5edb, 0xfd987193);
                        continue;
                    case '47':
                        _0x46b135 = _0x2333e5(_0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x57c542 + 0x0], _0x2c76c8, 0xe9b6c7aa);
                        continue;
                    case '48':
                        _0x4ad199 = _0x1d95a6[_0xb861('U115', '[cQx')](_0x46feb6, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6['UrAXv'](_0x57c542, 0x6)], _0x3c1ec9, 0xa3014314);
                        continue;
                    case '49':
                        _0x2cf07f = _0x2e3c75;
                        continue;
                    case '50':
                        _0x46b135 = _0x51409c(_0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6[_0xb861('U116', 'bqH]')](_0x57c542, 0xb)], _0x886338, 0x895cd7be);
                        continue;
                    case '51':
                        _0x2e3c75 = _0x4b4557(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x57c542 + 0x8], _0x5a26dd, 0x8771f681);
                        continue;
                    case '52':
                        _0x8f6b6f = _0x2333e5(_0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('I117', '^lJT')](_0x57c542, 0x9)], _0x1695df, 0x21e1cde6);
                        continue;
                    case '53':
                        _0x8f6b6f = _0x1d95a6[_0xb861('U118', 'n)t!')](_0x2333e5, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x57c542 + 0x1], _0x1695df, 0xf61e2562);
                        continue;
                    case '54':
                        _0x4ad199 = _0x1d95a6[_0xb861('U115', '[cQx')](_0x4b4557, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x1d95a6['AdEus'](_0x57c542, 0xb)], _0x46df4d, 0x6d9d6122);
                        continue;
                    case '55':
                        _0x46b135 = _0x1d95a6[_0xb861('I119', 'KWC!')](_0x4b4557, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6['AdEus'](_0x57c542, 0xa)], _0xecc592, 0xbebfbc70);
                        continue;
                    case '56':
                        _0x46b135 = _0x1d95a6['ITFdV'](_0x4b4557, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x57c542 + 0xe], _0xecc592, 0xfde5380c);
                        continue;
                    case '57':
                        _0x46b135 = _0x1d95a6[_0xb861('I11a', '7%qN')](_0x46feb6, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x57c542 + 0x5], _0x58aef, 0xfc93a039);
                        continue;
                    case '58':
                        _0x46b135 = _0x1d95a6['ITFdV'](_0x4b4557, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6[_0xb861('I11b', 'HV0E')](_0x57c542, 0x2)], _0xecc592, 0xc4ac5665);
                        continue;
                    case '59':
                        _0x46b135 = _0x51409c(_0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x1d95a6[_0xb861('U11c', 'n)t!')](_0x57c542, 0x3)], _0x886338, 0xc1bdceee);
                        continue;
                    case '60':
                        _0x8f6b6f = _0x1d95a6[_0xb861('I11d', 'UYsS')](_0x4b4557, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('U11e', '[cQx')](_0x57c542, 0x9)], _0x334ac8, 0xd9d4d039);
                        continue;
                    case '61':
                        _0x8f6b6f = _0x1d95a6['yARxV'](_0x2fde77, _0x8f6b6f, _0x216bdc);
                        continue;
                    case '62':
                        _0x46b135 = _0x51409c(_0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x57c542 + 0xf], _0x886338, 0x49b40821);
                        continue;
                    case '63':
                        _0x8f6b6f = _0x1d95a6['TOoln'](_0x51409c, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x1d95a6[_0xb861('U11f', 'Oyg4')](_0x57c542, 0x4)], _0x246668, 0xf57c0faf);
                        continue;
                    case '64':
                        _0x216bdc = _0x8f6b6f;
                        continue;
                    case '65':
                        _0x2e3c75 = _0x1d95a6[_0xb861('I120', 'sS*@')](_0x2fde77, _0x2e3c75, _0x2cf07f);
                        continue;
                    case '66':
                        _0x4ad199 = _0x1d95a6[_0xb861('I121', '7%qN')](_0x2333e5, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x57c542 + 0xf], _0x127a1d, 0xd8a1e681);
                        continue;
                    case '67':
                        _0x46b135 = _0x1d95a6['ndctf'](_0x46feb6, _0x46b135, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x2fba4c[_0x57c542 + 0xd], _0x58aef, 0x4e0811a1);
                        continue;
                    case '68':
                        _0x4ad199 = _0x1d95a6['rdjcg'](_0x46feb6, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x57c542 + 0xe], _0x3c1ec9, 0xab9423a7);
                        continue;
                    case '69':
                        _0x2e3c75 = _0x51409c(_0x2e3c75, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2fba4c[_0x1d95a6[_0xb861('I122', 'r(gn')](_0x57c542, 0x9)], _0x5c5edb, 0x8b44f7af);
                        continue;
                    case '70':
                        _0x8f6b6f = _0x1d95a6[_0xb861('I123', 'dadA')](_0x46feb6, _0x8f6b6f, _0x46b135, _0x4ad199, _0x2e3c75, _0x2fba4c[_0x57c542 + 0x0], _0x12e133, 0xf4292244);
                        continue;
                    case '71':
                        _0x4ad199 = _0x1d95a6[_0xb861('I124', 'b9C^')](_0x51409c, _0x4ad199, _0x2e3c75, _0x8f6b6f, _0x46b135, _0x2fba4c[_0x57c542 + 0x2], _0x51f6cf, 0x242070db);
                        continue;
                }
                break;
            }
        }
    }
    if (_0x1d95a6[_0xb861('I125', 'z&ze')](_0x424a01, 0x20)) {
        return (_0x1d95a6[_0xb861('I126', 'T!FO')](_0x1d95a6[_0xb861('I127', 'uzRM')](_0x39518b, _0x8f6b6f) + _0x1d95a6[_0xb861('I128', 'a%u&')](_0x39518b, _0x46b135), _0x39518b(_0x4ad199)) + _0x1d95a6[_0xb861('U129', 'V4Vs')](_0x39518b, _0x2e3c75))['toLowerCase']();
    }
    return (_0x1d95a6[_0xb861('U12a', 'Ppa7')](_0x39518b, _0x46b135) + _0x1d95a6[_0xb861('I12b', 'UYsS')](_0x39518b, _0x4ad199))['toLowerCase']();
}
;_0xodZ = 'jsjiami.com.v6';
function getSing(keyword, enKey) {
    var s = {}
    var _0xe89c07 = generateWordKey(keyword);
    s["token"] = generateHostMD5Key(_0xe89c07, enKey);
    s["randomNum"] = getRandomNum(_0xe89c07);
    s["enKey"]=enKey
    return s
}
`)
	if err != nil {
		gologger.Errorf(err.Error())
		return nil
	}
	call, err := vm.Call("getSing", nil, keyword, enKey)

	if err != nil {
		gologger.Errorf(err.Error())
		return nil
	}
	res := call.Object()
	return res
}
