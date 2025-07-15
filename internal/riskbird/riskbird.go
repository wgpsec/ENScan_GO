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

func (h *RB) GetEnInfoList(pid string, enMap *common.EnsGo) ([]gjson.Result, error) {
	listData := h.getInfoList(pid, enMap)
	return listData, nil
}

func (h *RB) getInfoList(eid string, s *common.EnsGo) (listData []gjson.Result) {
	urls := "https://www.riskbird.com/riskbird-api/companyInfo/list"
	li := map[string]string{
		"filterCnd":   "1",
		"page":        "1",
		"size":        "100",
		"orderNo":     eid,
		"extractType": s.Api,
		"sortField":   "",
	}
	marshal, err := json.Marshal(li)
	if err != nil {
		gologger.Error().Msgf("[RB]查询数据处理操作失败 %s", err.Error())
		return listData
	}
	data := string(marshal)
	gologger.Debug().Msgf("[RB] getInfoList %s\n", urls)
	content := gjson.Parse(h.req(urls, data))
	if content.Get("code").String() != "20000" {
		gologger.Error().Msgf("[RB] getInfoList %s\n", content.Get("msg"))
		return listData
	}
	listData = content.Get("data.apiData").Array()
	pc := content.Get("data.totalCount").Int()
	if pc > 100 {
		for i := 2; i <= int(pc/100+1); i++ {
			li["page"] = strconv.Itoa(i)
			marshal, err = json.Marshal(li)
			data = string(marshal)
			listData = append(listData, gjson.Get(h.req(urls, data), "data.totalCount").Array()...)
		}
	}
	return listData
}

func (h *RB) searchBaseInfo(pid string) (result gjson.Result, enBaseInfo gjson.Result) {
	r := gjson.Parse(h.req("https://www.riskbird.com/api/ent/query?entId="+pid, ""))
	result = r.Get("basicResult.apiData.list.jbxxInfo")
	enJsonTMP, _ := sjson.Set(result.Raw, "orderNo", r.Get("orderNo").String())
	enBaseInfo = r.Get("basicResult.apiData.count")
	return gjson.Parse(enJsonTMP), enBaseInfo
}
