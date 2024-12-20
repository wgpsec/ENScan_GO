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
	return getInfoList(keyword, enMap[types].Api, h.Options)
}

func (h *Miit) GetENMap() map[string]*common.EnsGo {
	return getENMap()
}

func getInfoList(keyword string, types string, options *common.ENOptions) []gjson.Result {
	url := options.ENConfig.App.MiitApi + "/query/" + types + "?page=1&search=" + urlTool.QueryEscape(keyword)
	content := getReq(url+"&page=1", "", options)
	var listData []gjson.Result
	data := gjson.Get(content, "data")
	listData = data.Get("list").Array()
	if data.Get("hasNextPage").Bool() {
		for i := 2; int(data.Get("pages").Int()) >= i; i++ {
			gologger.Info().Msgf("当前：%s,%d\n", types, i)
			reqUrls := url + "&page=" + strconv.Itoa(i)
			// 强制延时！！特别容易被封
			time.Sleep(5 * time.Second)
			content = getReq(reqUrls, "", options)
			listData = append(listData, gjson.Get(content, "data.list").Array()...)
		}
	}
	return listData
}
