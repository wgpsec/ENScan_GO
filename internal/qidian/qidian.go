package qidian

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
)

import (
	"github.com/wgpsec/ENScan/common/gologger"
)

type QD struct {
	Options *common.ENOptions
}

func (h *QD) AdvanceFilter() ([]gjson.Result, error) {
	url := "https://holmes.taobao.com/web/corp/customer/searchWithSummary"
	searchData := map[string]string{
		"pageNo":      "1",
		"pageSize":    "10",
		"keyword":     h.Options.KeyWord,
		"orderByType": "5",
	}
	searchJsonData, _ := json.Marshal(searchData)
	content := GetReq(url, string(searchJsonData), h.Options)
	res := gjson.Parse(content)
	enList := res.Get("data.data").Array()
	if len(enList) == 0 {
		gologger.Debug().Str("查询请求", h.Options.KeyWord).Msg(content)
		return enList, fmt.Errorf("【QD】没有查询到关键词 %s", h.Options.KeyWord)
	}
	return enList, nil
}

func (h *QD) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}
func (h *QD) GetEnsD() common.ENsD {
	ensD := common.ENsD{Name: h.Options.KeyWord, Pid: h.Options.CompanyID}
	return ensD
}
func (h *QD) GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo) {
	// 获取的格式不太好写，先放在这了，感兴趣的可以提PR
	gologger.Error().Msgf("【企典】功能尚未开发，感兴趣的师傅可以提交PR~")
	return gjson.Result{}, getENMap()
}
func (h *QD) GetEnInfoList(pid string, enMap *common.EnsGo) ([]gjson.Result, error) {
	listData := getInfoList(pid, enMap.Api, h.Options)
	return listData, nil
}
func getInfoList(pid string, api string, op *common.ENOptions) []gjson.Result {
	return []gjson.Result{}
}
