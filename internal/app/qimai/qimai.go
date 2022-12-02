package qimai

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"strconv"
)

func GetInfoByKeyword(options *common.ENOptions) (ensInfos *common.EnInfos, ensOutMap map[string]*outputfile.ENSMap) {
	ensInfos = &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensOutMap = make(map[string]*outputfile.ENSMap)
	ensInfos.Name = options.KeyWord
	params := map[string]string{
		"page":   "1",
		"search": options.KeyWord,
		"market": "6",
	}
	fmt.Println("233333")
	res := gjson.Parse(GetReq("search/android", params, options)).Get("appList").Array()
	if res[0].Get("company.id").Int() != 0 {
		fmt.Println(res[0].Get("company.name"))
		ensInfos.Infos = GetInfoByCompanyId(res[0].Get("company.id").Int(), options)
	}
	for k, v := range getENMap() {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}
	return ensInfos, ensOutMap
}

func GetInfoByCompanyId(companyId int64, options *common.ENOptions) (data map[string][]gjson.Result) {
	gologger.Infof("GetInfoByCompanyId: %d\n", companyId)
	data = map[string][]gjson.Result{}
	ensMap := getENMap()
	params := map[string]string{
		"id": strconv.Itoa(int(companyId)),
	}
	searchInfo := "enterprise_info"
	//gjson.GetMany(gjson.Get(GetReq(ensMap[searchInfo].api, params, options), "data").Raw, ensMap[searchInfo].field...)
	r, err := sjson.Set(gjson.Get(GetReq(ensMap[searchInfo].api, params, options), "data").Raw, "id", companyId)
	if err != nil {
		gologger.Errorf("Set pid error: %s", err.Error())
	}
	rs := gjson.Parse(r)
	data[searchInfo] = append(data[searchInfo], rs)
	params["page"] = "1"
	params["apptype"] = "2"
	searchInfo = "app"
	data[searchInfo] = append(data[searchInfo], getInfoList(ensMap["app"].api, params, options)...)
	//安卓
	params["page"] = "1"
	params["apptype"] = "3"
	data[searchInfo] = append(data[searchInfo], getInfoList(ensMap["app"].api, params, options)...)
	//命令输出展示
	var tdata [][]string
	for _, y := range data[searchInfo] {
		results := gjson.GetMany(y.Raw, ensMap[searchInfo].field...)
		var str []string
		for _, ss := range results {
			str = append(str, ss.String())
		}
		tdata = append(tdata, str)
	}
	common.TableShow(ensMap[searchInfo].keyWord, tdata, options)
	return data
}

func getInfoList(types string, params map[string]string, options *common.ENOptions) (listData []gjson.Result) {
	data := gjson.Parse(GetReq(types, params, options))
	if data.Get("code").String() == "10000" {
		getPath := "appList"
		getPage := "maxPage"
		if types == "company/getCompanyApplist" {
			if params["apptype"] == "2" {
				getPath = "ios"
			} else if params["apptype"] == "3" {
				getPath = "android"
			}
			getPage = getPath + "PageInfo.pageCount"
			getPath += "AppInfo"
			data = data.Get("data")
		}

		listData = append(listData, data.Get(getPath).Array()...)
		if data.Get(getPage).Int() <= 1 {
			return listData
		} else {
			for i := 2; i <= int(data.Get(getPage).Int()); i++ {
				gologger.Infof("getInfoList: %s %d\n", types, i)
				params["page"] = fmt.Sprintf("%d", i)
				listData = append(listData, gjson.Parse(GetReq(types, params, options)).Get("data."+getPath).Array()...)
			}
		}
		if len(listData) == 0 {
			gologger.Errorf("没有数据")
		}
	} else {
		gologger.Errorf("获取数据失败,请检查是否登陆\n")
		gologger.Debugf(data.Raw + "\n")
	}
	return listData
}
