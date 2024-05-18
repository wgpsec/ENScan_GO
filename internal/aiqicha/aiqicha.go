package aiqicha

/* Aiqicha By Keac
 * admin@wgpsec.org
 */
import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	urlTool "net/url"
	"strconv"
	"strings"
)

type AQC struct {
	Options *common.ENOptions
}

// AdvanceFilter 筛选过滤
func (h *AQC) AdvanceFilter() ([]gjson.Result, error) {
	name := h.Options.KeyWord
	//urls := "https://aiqicha.baidu.com/s?q=" + urlTool.QueryEscape(name) + "&t=0"
	urls := "https://aiqicha.baidu.com/s/advanceFilterAjax?q=" + urlTool.QueryEscape(name) + "&p=1&s=10&f={}"
	content := GetReq(urls, h.Options)
	content = strings.ReplaceAll(content, "<em>", "⌈")
	content = strings.ReplaceAll(content, "<\\/em>", "⌋")
	//rq := pageParseJson(content)
	enList := gjson.Get(content, "data.resultList").Array()
	ddw := gjson.Get(content, "ddw").Int()
	if len(enList) == 0 {
		gologger.Debug().Str("查询请求", name).Msg(content)
		return enList, fmt.Errorf("【AQC】没有查询到关键词 ⌈%s⌋", name)
	}
	// advanceFilterAjax 接口特殊处理
	for i, v := range enList {
		s, _ := sjson.Set(v.Raw, "pid", transformNumber(v.Get("pid").String(), ddw))
		enList[i] = gjson.Parse(s)
	}
	return enList, nil
}

func (h *AQC) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func (h *AQC) GetEnsD() common.ENsD {
	ensD := common.ENsD{Name: h.Options.KeyWord, Pid: h.Options.CompanyID}
	return ensD
}

func (h *AQC) GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo) {
	// 企业基本信息
	urls := "https://aiqicha.baidu.com/detail/basicAllDataAjax?pid=" + pid
	baseRes := GetReq(urls, h.Options)
	res := gjson.Get(baseRes, "data.basicData")
	// 修复没有pid的问题
	r, _ := sjson.Set(res.Raw, "pid", pid)
	res = gjson.Parse(r)
	//初始化ENMap
	ensInfoMap := getENMap()
	// 获取企业信息列表
	enInfoUrl := "https://aiqicha.baidu.com/compdata/navigationListAjax?pid=" + pid
	enInfoRes := GetReq(enInfoUrl, h.Options)
	// 初始化数量数据
	if gjson.Get(enInfoRes, "status").String() == "0" {
		for _, s := range gjson.Get(enInfoRes, "data").Array() {
			for _, t := range s.Get("children").Array() {
				resId := t.Get("id").String()
				// 判断内容是否在enscan支持范围内
				if _, ok := enMapping[resId]; ok {
					resId = enMapping[resId]
				}
				es := ensInfoMap[resId]
				if es == nil {
					es = &common.EnsGo{}
				}
				//gologger.Debug().Msgf("【AQC】数量" + t.Get("name").String() + "|" + t.Get("total").String() + "|" + t.Get("id").String())
				es.Name = t.Get("name").String()
				es.Total = t.Get("total").Int()
				es.Available = t.Get("avaliable").Int()
				ensInfoMap[t.Get("id").String()] = es
			}
		}
	} else {
		gologger.Error().Msg("初始化数量失败！")
		gologger.Debug().Str("pid", pid).Msgf("%s", enInfoRes)
	}
	return res, ensInfoMap
}

func (h *AQC) GetEnInfoList(pid string, enMap *common.EnsGo) ([]gjson.Result, error) {
	listData := getInfoList(pid, enMap.Api, h.Options)
	return listData, nil
}

// getInfoList 获取信息列表
func getInfoList(pid string, types string, options *common.ENOptions) []gjson.Result {
	urls := "https://aiqicha.baidu.com/" + types + "?pid=" + pid
	content := GetReq(urls, options)
	var listData []gjson.Result
	if gjson.Get(content, "status").String() == "0" {
		data := gjson.Get(content, "data")
		// 判断投资关系一个获取的特殊值
		if types == "relations/relationalMapAjax" {
			data = gjson.Get(content, "data.investRecordData")
		}

		//判断是否多页，遍历获取所有数据
		pageCount := data.Get("pageCount").Int()
		if pageCount > 1 {
			for i := 1; int(pageCount) >= i; i++ {
				gologger.Info().Msgf("当前：%s,%d\n", types, i)
				reqUrls := urls + "&p=" + strconv.Itoa(i)
				content = GetReq(reqUrls, options)
				listData = append(listData, gjson.Get(string(content), "data.list").Array()...)
			}
		} else {
			listData = data.Get("list").Array()
		}

		// 处理下ICP备案把他换成多行
		if types == "detail/icpinfoAjax" {
			var tmp []gjson.Result
			for _, y := range listData {
				for _, o := range y.Get("domain").Array() {
					valueTmp, _ := sjson.Set(y.Raw, "domain", o.String())
					valueTmp, _ = sjson.Set(valueTmp, "homeSite", y.Get("homeSite").Array()[0].String())
					tmp = append(tmp, gjson.Parse(valueTmp))
				}
			}
			listData = tmp
		}
	}
	return listData

}
