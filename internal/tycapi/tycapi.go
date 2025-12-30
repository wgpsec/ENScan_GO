package tycapi

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"net/url"
	"strconv"
	"time"
)

type TycAPI struct {
	Options *common.ENOptions
}

func (h *TycAPI) AdvanceFilter(name string) ([]gjson.Result, error) {
	em := getENMap()
	urls := em["search"].Api + "?word=" + url.QueryEscape(name)
	content := h.req(urls)
	enList := gjson.Get(content, "items.items").Array()
	if len(enList) == 0 {
		gologger.Debug().Str("【TYC-API】查询请求", name).Msg(content)
		return enList, fmt.Errorf("【TYC-API】没有查询到关键词 ⌈%s⌋", name)
	}
	return enList, nil
}

func (h *TycAPI) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func (h *TycAPI) GetEnsD() common.ENsD {
	ensD := common.ENsD{Name: h.Options.KeyWord, Pid: h.Options.CompanyID, Op: h.Options}
	return ensD
}

func (h *TycAPI) GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo) {
	ensInfoMap := getENMap()
	detailRes := h.searchBaseInfo(pid)
	//修复成立日期信息
	ts := time.UnixMilli(detailRes.Get("fromTime").Int())
	enJsonTMP, _ := sjson.Set(detailRes.Raw, "fromTime", ts.Format("2006-01-02"))
	// 匹配统计数据,API没有该接口，所以全部标记为存在
	for k, _ := range ensInfoMap {
		ensInfoMap[k].Total = 1
	}
	return gjson.Parse(enJsonTMP), ensInfoMap
}

func (h *TycAPI) GetInfoByPage(pid string, page int, em *common.EnsGo) (info common.InfoPage, err error) {
	uv := em.Api + "?keyword=" + pid + "&page=" + strconv.Itoa(page)
	res := gjson.Get(h.req(uv), "result")
	items := res.Get("items").Array()
	info = common.InfoPage{
		Size:  int64(len(items)),
		Total: res.Get("total").Int(),
		Data:  items,
	}
	return info, err
}

func (h *TycAPI) getInfo(pid string, s *common.EnsGo) (ef map[string][]gjson.Result) {
	u := s.Api + "?keyword=" + pid
	res := gjson.Get(h.req(u), "result")
	for k, v := range s.SData {
		ef[v] = res.Get(k).Array()
	}
	return ef
}

// searchBaseInfo 获取企业基本信息
func (h *TycAPI) searchBaseInfo(pid string) (result gjson.Result) {
	result = gjson.Get(h.req(getENMap()["enterprise_info_normal"].Api+"?keyword="+pid), "result")
	return result
}
