package hook

import (
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"net"
	"strings"
)

type ENCInfo struct {
	Email  string
	Phone  string
	Source string
}

func getWhois(domain string) (string, error) {
	defer func() {
		if x := recover(); x != nil {
			gologger.Errorf("[WHOIS] ERROR\n")
		}
	}()
	result, err := whois.Whois(domain)
	if err != nil {
		return "", err
	}
	res, err := whoisparser.Parse(result)
	if err == nil {
		return res.Registrant.Email, nil
	} else {
		return "", err
	}
}

func GetEnEmail(EnsInfosList map[string][][]interface{}, options *common.ENOptions) (DaList []ENCInfo) {
	//企业注册信息获取
	tmpMap := make(map[string]bool)
	enEmails := 0
	enPhones := 0
	for _, s := range EnsInfosList["enterprise_info"] {
		if utils.VerifyEmailFormat(s[4].(string)) {
			if ok := tmpMap[s[4].(string)]; !ok {
				DaList = append(DaList, ENCInfo{Email: s[4].(string), Source: "企业信息 " + s[0].(string)})
				enEmails++
				tmpMap[s[4].(string)] = true
			}
		}
		if !strings.Contains(s[3].(string), "*") {
			if ok := tmpMap[s[3].(string)]; !ok {
				DaList = append(DaList, ENCInfo{Phone: s[3].(string), Source: "企业信息 " + s[0].(string)})
				enPhones++
				tmpMap[s[3].(string)] = true
			}
		}
	}

	gologger.Infof("企业信息采集【%d】个邮箱【%d】个手机号\n", enEmails, enPhones)

	var icpList []string
	for _, s := range EnsInfosList["icp"] {
		icpList = append(icpList, s[2].(string))
	}
	icpList = utils.SetStr(icpList)
	for _, s := range icpList {
		whoisRes, err := getWhois(s)
		if err == nil && utils.VerifyEmailFormat(whoisRes) {
			DaList = append(DaList, ENCInfo{Email: whoisRes, Source: "WHOIS " + s})
		}
		if net.ParseIP(s) == nil {
			vpList := VPGetEmail(s, options)
			gologger.Infof("%s 获取邮箱 %d 个\n", s, len(vpList))
			for _, v := range vpList {
				TEM := v.Get("email").String()
				if ok := tmpMap[TEM]; !ok {
					DaList = append(DaList, ENCInfo{Email: TEM, Source: "在线接口 " + s})
					tmpMap[TEM] = true
				}
			}
		}
	}
	return DaList

}
