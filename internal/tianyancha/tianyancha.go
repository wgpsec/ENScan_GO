package tianyancha

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"strconv"
	"strings"
	"time"
)

type TYC struct {
	Options *common.ENOptions
}

func (h *TYC) AdvanceFilter(name string) ([]gjson.Result, error) {
	//使用关键词推荐方法进行检索，会出现信息不对的情况
	//urls := "https://sp0.tianyancha.com/search/suggestV2.json?key=" + url.QueryEscape(name)
	urls := "https://capi.tianyancha.com/cloud-tempest/web/searchCompanyV4"
	searchData := map[string]string{
		"key":      name,
		"pageNum":  "1",
		"pageSize": "20",
		"referer":  "search",
		"sortType": "0",
		"word":     name,
	}
	marshal, err := json.Marshal(searchData)
	if err != nil {
		return nil, fmt.Errorf("【TYC】关键词处理失败 %s", err.Error())
	}
	content := h.req(urls, string(marshal))
	content = strings.ReplaceAll(content, "<em>", "⌈")
	content = strings.ReplaceAll(content, "</em>", "⌋")
	enList := gjson.Get(content, "data.companyList").Array()

	if len(enList) == 0 {
		gologger.Debug().Str("【TYC】查询请求", name).Msg(content)
		return enList, fmt.Errorf("【TYC】没有查询到关键词 ⌈%s⌋", name)
	}
	return enList, nil
}

func (h *TYC) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func (h *TYC) GetEnsD() common.ENsD {
	ensD := common.ENsD{Name: h.Options.KeyWord, Pid: h.Options.CompanyID, Op: h.Options}
	return ensD
}

func (h *TYC) GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo) {
	ensInfoMap := getENMap()
	// 快速模式跳过企业基本信息
	if h.Options.IsFast {
		for k, _ := range ensInfoMap {
			ensInfoMap[k].Total = 1
		}
		return gjson.Result{}, ensInfoMap
	}
	detailRes, enCount := h.searchBaseInfo(pid, false, h.Options)
	//修复成立日期信息
	ts := time.UnixMilli(detailRes.Get("fromTime").Int())
	enJsonTMP, _ := sjson.Set(detailRes.Raw, "fromTime", ts.Format("2006-01-02"))
	// 匹配统计数据
	for k, v := range ensInfoMap {
		ensInfoMap[k].Total = enCount.Get(v.GNum).Int()
	}
	return gjson.Parse(enJsonTMP), ensInfoMap
}

func (h *TYC) GetInfoByPage(pid string, page int, em *common.EnsGo) (info common.InfoPage, err error) {
	sd := em.SData
	urls := "https://capi.tianyancha.com/" + em.Api + "?_=" + strconv.Itoa(int(time.Now().Unix()))
	if len(sd) > 0 {
		sd["gid"] = pid
		sd["pageSize"] = "100"
		sd["pageNum"] = strconv.Itoa(page)
		info.Page = 100
	} else {
		urls += "&pageSize=20&graphId=" + pid + "&id=" + pid + "&gid=" + pid + "&pageNum=" + strconv.Itoa(page) + em.GsData
		info.Page = 20
	}
	var m []byte
	m, err = json.Marshal(sd)
	if err != nil {
		return info, err
	}
	content := h.req(urls, string(m))
	if gjson.Get(content, "state").String() != "ok" {
		return info, fmt.Errorf("查询出现错误 %s\n", content)
	}
	pList := []string{"itemTotal", "count", "total", "pageBean.total"}
	for _, k := range gjson.GetMany(gjson.Get(content, "data").Raw, pList...) {
		if k.Int() != 0 {
			info.Total = k.Int()
		}
	}
	pats := "data." + em.Rf
	info.Data = gjson.Get(content, pats).Array()
	return info, err
}

// searchBaseInfo 获取基本信息（此操作容易触发验证）
func (h *TYC) searchBaseInfo(pid string, tds bool, options *common.ENOptions) (result gjson.Result, enBaseInfo gjson.Result) {
	// 这里没有获取统计信息的api，故从html获取
	if tds {
		//htmlInfo := htmlquery.FindOne(body, "//*[@class=\"position-rel company-header-container\"]//script")
		//enBaseInfo = pageParseJson(htmlquery.InnerText(htmlInfo))
		result = gjson.Get(h.req("https://capi.tianyancha.com/cloud-other-information/companyinfo/baseinfo/web?id="+pid, ""), "data")
		return result, gjson.Result{}
	} else {
		urls := "https://www.tianyancha.com/company/" + pid
		body := h.GetReqReturnPage(urls)
		htmlInfos := htmlquery.FindOne(body, "//*[@id=\"__NEXT_DATA__\"]")
		enInfo := gjson.Parse(htmlquery.InnerText(htmlInfos))
		enInfoD := enInfo.Get("props.pageProps.dehydratedState.queries").Array()
		result = enInfoD[0].Get("state.data.data")
		//数量统计 API base_count
		for i := 0; i < len(enInfoD); i++ {
			if enInfoD[i].Get("queryKey").String() == "base_count" {
				enBaseInfo = enInfoD[i].Get("state.data")
			}
		}
		//enBaseInfo = enInfo.Get("props.pageProps.dehydratedState.queries").Array()[11].Get("state.data")
	}
	return result, enBaseInfo
}
