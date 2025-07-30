package miit

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	urlTool "net/url"
	"strconv"
	"time"
)

type Miit struct {
	Options *common.ENOptions
}

func (h *Miit) GetInfoList(keyword string, types string) []gjson.Result {
	enMap := getENMap()
	return h.getInfoList(keyword, enMap[types].Api, h.Options)
}

func (h *Miit) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func (h *Miit) GetInfoByPage(keyword string, page int, em *common.EnsGo) (info common.InfoPage, err error) {
	url := h.Options.ENConfig.App.MiitApi + "/query/" + em.Api + "?page=" + strconv.Itoa(page) + "&search=" + urlTool.QueryEscape(keyword)
	content := h.getReq(url+"&page=1", "")
	var listData []gjson.Result
	data := gjson.Get(content, "data")
	listData = data.Get("list").Array()
	info = common.InfoPage{
		Total:   data.Get("total").Int(),
		Size:    data.Get("pages").Int(),
		HasNext: data.Get("hasNextPage").Bool(),
		Data:    listData,
	}
	return info, err
}

func (h *Miit) getInfoList(keyword string, types string, options *common.ENOptions) []gjson.Result {
	url := options.ENConfig.App.MiitApi + "/query/" + types + "?page=1&search=" + urlTool.QueryEscape(keyword)
	content := h.getReq(url+"&page=1", "")
	var listData []gjson.Result
	data := gjson.Get(content, "data")
	listData = data.Get("list").Array()
	if data.Get("hasNextPage").Bool() {
		for i := 2; int(data.Get("pages").Int()) >= i; i++ {
			gologger.Info().Msgf("当前：%s,%d\n", types, i)
			reqUrls := url + "&page=" + strconv.Itoa(i)
			// 强制延时！！特别容易被封
			time.Sleep(5 * time.Second)
			content = h.getReq(reqUrls, "")
			listData = append(listData, gjson.Get(content, "data.list").Array()...)
		}
	}
	return listData
}
