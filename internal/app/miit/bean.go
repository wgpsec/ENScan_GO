package miit

import (
	"crypto/tls"
	"github.com/imroc/req/v3"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"time"
)

func getENMap() map[string]*common.EnsGo {
	ensInfoMap := make(map[string]*common.EnsGo)
	ensInfoMap = map[string]*common.EnsGo{
		"icp": {
			Name:    "ICP备案",
			Api:     "web",
			Field:   []string{"", "domain", "domain", "serviceLicence", "unitName"},
			KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
		},
		"app": {
			Name:    "APP",
			Api:     "app",
			Field:   []string{"serviceName", "serviceType", "version", "updateRecordTime", "", "", "", "", ""},
			KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
		},
		"wx_app": {
			Name:    "小程序",
			Api:     "miniapp",
			Field:   []string{"serviceName", "serviceType", "", "", ""},
			KeyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
		},
		"fastapp": {
			Name:    "快应用",
			Api:     "fastapp",
			Field:   []string{"serviceName", "serviceType", "", "", ""},
			KeyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
		},
	}
	for k := range ensInfoMap {
		// 获取插件需要的信息
		ensInfoMap[k].AppParams = [2]string{"enterprise_info", "name"}
		ensInfoMap[k].KeyWord = append(ensInfoMap[k].KeyWord, "数据关联  ")
		ensInfoMap[k].Field = append(ensInfoMap[k].Field, "inFrom")
	}
	return ensInfoMap
}

func getReq(url string, data string, options *common.ENOptions) string {
	client := req.C()
	client.SetTLSFingerprintChrome()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetTimeout(time.Duration(options.TimeOut) * time.Minute)
	if options.Proxy != "" {
		client.SetProxyURL(options.Proxy)
	}
	client.SetCommonHeaders(map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36 Edg/98.0.1108.43",
		"Accept":       "text/html, application/xhtml+xml, image/jxr, */*",
		"Content-Type": "application/json;charset=UTF-8",
	})
	//加入随机延时
	time.Sleep(time.Duration(options.GetDelayRTime()) * time.Second)
	clientR := client.R()

	method := "GET"
	if data == "" {
		method = "GET"
	} else {
		method = "POST"
		clientR.SetBody(data)
	}

	resp, err := clientR.Send(method, url)

	if err != nil {
		if options.Proxy != "" {
			client.SetProxy(nil)
		}
		gologger.Error().Msgf("【miit】请求发生错误， %s 10秒后重试\n%s\n", url, err)
		time.Sleep(10 * time.Second)
		return getReq(url, data, options)
	}
	if resp.IsSuccessState() {
		return resp.String()
	} else {
		gologger.Error().Msgf("【miit】未知错误 %s\n", resp.StatusCode)
	}
	return ""
}
