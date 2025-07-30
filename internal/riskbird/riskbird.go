package riskbird

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"strconv"
)

func (h *RB) AdvanceFilter(name string) ([]gjson.Result, error) {
	urls := "https://www.riskbird.com/riskbird-api/newSearch"
	searchData := map[string]string{
		"searchKey":           name,
		"pageNo":              "1",
		"range":               "10",
		"referer":             "search",
		"queryType":           "1",
		"selectConditionData": "{\"status\":\"\",\"sort_field\":\"\"}",
	}

	marshal, err := json.Marshal(searchData)
	if err != nil {
		return nil, fmt.Errorf("【RB】关键词处理失败 %s", err.Error())
	}
	content := h.req(urls, string(marshal))
	enList := gjson.Get(content, "data.list").Array()
	if len(enList) == 0 {
		gologger.Debug().Str("查询请求", name).Msg(content)
		return enList, fmt.Errorf("【RB】没有查询到关键词 ⌈%s⌋", name)
	}
	return enList, nil
}

func (h *RB) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func (h *RB) GetEnsD() common.ENsD {
	ensD := common.ENsD{Name: h.Options.KeyWord, Pid: h.Options.CompanyID, Op: h.Options}
	return ensD
}

func (h *RB) GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo) {
	ensInfoMap := getENMap()
	detailRes, enCount := h.searchBaseInfo(pid)
	for k, v := range ensInfoMap {
		ensInfoMap[k].Total = enCount.Get(v.GNum).Int()
	}
	return detailRes, ensInfoMap
}

func (h *RB) GetInfoByPage(pid string, page int, em *common.EnsGo) (info common.InfoPage, err error) {
	urls := "https://www.riskbird.com/riskbird-api/companyInfo/list"
	li := map[string]string{
		"filterCnd":   "1",
		"page":        strconv.Itoa(page),
		"size":        "100",
		"orderNo":     pid,
		"extractType": em.Api,
		"sortField":   "",
	}
	marshal, err := json.Marshal(li)
	if err != nil {
		return info, err
	}
	content := gjson.Parse(h.req(urls, string(marshal)))
	if content.Get("code").String() != "20000" {
		return info, fmt.Errorf("【RB】获取数据失败 %s", content.Get("msg"))
	}
	info = common.InfoPage{
		Size:  100,
		Total: content.Get("data.totalCount").Int(),
		Data:  content.Get("data.apiData").Array(),
	}
	return info, err
}

func (h *RB) searchBaseInfo(pid string) (result gjson.Result, enBaseInfo gjson.Result) {
	r := gjson.Parse(h.req("https://www.riskbird.com/api/ent/query?entId="+pid, ""))
	result = r.Get("basicResult.apiData.list.jbxxInfo")
	enJsonTMP, _ := sjson.Set(result.Raw, "orderNo", r.Get("orderNo").String())
	enBaseInfo = r.Get("basicResult.apiData.count")
	return gjson.Parse(enJsonTMP), enBaseInfo
}
