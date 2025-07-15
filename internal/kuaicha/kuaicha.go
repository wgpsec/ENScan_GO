package kuaicha

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"math"
	"strconv"
	"strings"
)

type KC struct {
	Options *common.ENOptions
}

func (h *KC) AdvanceFilter(name string) ([]gjson.Result, error) {
	url := "https://www.kuaicha365.com/enterprise_info_app/V1/search/company_search_pc"
	searchStr := `{"search_conditions":[],"keyword":"` + name + `","page":1,"page_size":20}"}`
	content := h.req(url, searchStr)
	content = strings.ReplaceAll(content, "<em>", "⌈")
	content = strings.ReplaceAll(content, "</em>", "⌋")
	enList := gjson.Get(content, "data.list").Array()
	if len(enList) == 0 {
		gologger.Debug().Str("查询请求", name).Msg(content)
		return enList, fmt.Errorf("【KC】没有查询到关键词 ⌈%s⌋ \n", name)
	}
	return enList, nil
}

func (h *KC) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func (h *KC) GetEnsD() common.ENsD {
	ensD := common.ENsD{Name: h.Options.KeyWord, Pid: h.Options.CompanyID, Op: h.Options}
	return ensD
}

func (h *KC) GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo) {
	ensInfoMap := getENMap()
	detailRess := h.req("https://www.kuaicha365.com/open/app/v1/pc_enterprise/basic/info?org_id="+pid, "")
	detailRes := gjson.Get(detailRess, "data")
	// 匹配统计数据,需要从页面中提取数据来判断
	for k, _ := range ensInfoMap {
		ensInfoMap[k].Total = 1
	}
	return detailRes, ensInfoMap
}

func (h *KC) GetEnInfoList(pid string, enMap *common.EnsGo) ([]gjson.Result, error) {
	return h.getInfoList(pid+enMap.Fids, enMap.Api, h.Options), nil
}

func (h *KC) getInfoList(pid string, types string, options *common.ENOptions) []gjson.Result {
	urls := "https://www.kuaicha365.com/" + types
	if strings.Contains(types, "open/app/v1") {
		urls += "?page_size=10" + "&org_id=" + pid
	} else {
		urls += "?pageSize=10&&orgid=" + pid
	}
	content := h.req(urls+"&page=1", "")
	var listData []gjson.Result
	if gjson.Get(content, "status_code").String() == "0" {
		data := gjson.Get(content, "data")

		//判断是否多页，遍历获取所有数据
		pageCount := data.Get("total").Int()
		if data.Get("next_page").Bool() {
			for i := 1; int(math.Ceil(float64(pageCount)/float64(10))) >= i; i++ {
				gologger.Info().Msgf("当前：%s,%d\n", types, i)
				reqUrls := urls + "&page=" + strconv.Itoa(i)
				content = h.req(reqUrls, "")
				listData = append(listData, gjson.Get(content, "data.list").Array()...)
			}
		} else {
			listData = data.Get("list").Array()
		}
	}
	return listData
}
