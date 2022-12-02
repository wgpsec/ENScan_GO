package qcc

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	urlTool "net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetEnInfoByPid 根据PID获取公司信息
func GetEnInfoByPid(options *common.ENOptions) (*common.EnInfos, map[string]*outputfile.ENSMap) {
	ensInfos := &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensOutMap := make(map[string]*outputfile.ENSMap)
	pid := ""
	groupID := ""
	if options.CompanyID == "" && !options.IsGroup {
		_, pid = SearchName(options)
	} else if options.GroupID == "" && options.IsGroup {
		var enNames []gjson.Result
		enNames, pid, groupID = searchGroupByName(options)
		enName := enNames[0].Get("GroupName").String()
		enName = strings.ReplaceAll(enName, "<em>", "")
		enName = strings.ReplaceAll(enName, "</em>", "")
		ensInfos.Name = enName
	} else if options.CompanyID != "" && !options.IsGroup {
		pid = options.CompanyID
	}
	outPutFlag := "company"
	if options.IsGroup {
		if groupID == "" {
			gologger.Errorf("没有获取到集团ID\n")
			return ensInfos, ensOutMap
		}
		outPutFlag = "all"
		gologger.Infof("查询集团ID %s\n", groupID)
		getGroupInfoById(groupID, options.Deep > 0, []string{}, ensInfos, options)
	} else {
		if pid == "" {
			gologger.Errorf("没有获取到PID\n")
			return ensInfos, ensOutMap
		}
		gologger.Infof("查询PID %s\n", pid)
		//获取公司信息
		getCompanyInfoById(pid, 1, true, "", options.GetField, ensInfos, options)
	}

	for k, v := range getENMap(outPutFlag, "") {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}
	outputfile.OutPutExcelByEnInfo(ensInfos, ensOutMap, options)
	return ensInfos, ensOutMap

}
func getCompanyInfoById(pid string, deep int, isEnDetail bool, inFrom string, searchList []string, ensInfo *common.EnInfos, options *common.ENOptions) {
	ensMap := getENMap("company", "")
	var CountInfo gjson.Result
	var DetailRes string
	if isEnDetail {
		enDes := "enterprise_info"
		//来源企业PK 可直接获取全部信息
		detailRes := getReq("https://www.qcc.com/api/company/getDetail?keyNo="+pid, "", options)
		DetailRes = detailRes
		CountInfo = gjson.Get(detailRes, "CountInfo")
		enJsonTMP, _ := sjson.Set(detailRes, "inFrom", inFrom)
		ensInfo.Infos[enDes] = append(ensInfo.Infos[enDes], gjson.Parse(enJsonTMP))
		rs := gjson.GetMany(detailRes, "Name", "Oper.Name", "ShortStatus", "ContactInfo.Email", "ContactInfo.PhoneNumber")
		ks := []string{"公司名称", "法人代表", "状态", "邮箱", "电话"}
		for k, v := range rs {
			fmt.Println(ks[k] + ":" + v.String())
		}
		for _, v := range gjson.Get(detailRes, "HisTelList").Array() {
			fmt.Println(v.Get("Tel"))
		}
		ensInfo.Name = gjson.Get(detailRes, "Name").String()

	}
	if deep <= options.Deep && options.InvestNum > 0 {
		if options.InvestNum >= 100 {
			ensMap["invest"].fids = "&fundedRatioLevel=6"
		} else if options.InvestNum >= 66 {
			ensMap["invest"].fids = "&fundedRatioLevel=5"
		} else if options.InvestNum >= 50 {
			ensMap["invest"].fids = "&fundedRatioLevel=4"
		} else if options.InvestNum >= 20 {
			ensMap["invest"].fids = "&fundedRatioLevel=3"
		} else if options.InvestNum >= 5 {
			ensMap["invest"].fids = "&fundedRatioLevel=2"
		}

		if options.IsInvestRd {
			// 间接持股 意义不大 indirect 去掉
			searchList = append(searchList, "holds", "contactrel")
		}

	}
	for _, v := range searchList {
		if v == "branch" && !options.IsGetBranch {
			continue
		}
		if _, ok := ensMap[v]; !ok {
			continue
		}

		if (CountInfo.Get(ensMap[v].gNum).Int() <= 0 && v != "partner") && isEnDetail {
			gologger.Infof("【QCC】%s 数量为空，自动跳过\n", ensMap[v].name)
			continue
		}
		gologger.Infof("查询 %s\n", v)
		var res []gjson.Result
		//分支机构信息直接可以从详情页获取
		if v == "branch" {
			res = gjson.Get(DetailRes, "Branches").Array()
		} else if v == "partner" {
			res = gjson.Get(DetailRes, "Partners").Array()
		} else {
			res = getInfoList(pid+ensMap[v].fids, ensMap[v].api, "", options)
		}
		//判断下网站备案，然后提取出来，留个坑看看有没有更好的解决方案
		if v == "icp" {
			var tmp []gjson.Result
			for _, y := range res {
				for _, o := range strings.Split(y.Get("DomainName").String(), ",") {
					value, _ := sjson.Set(y.Raw, "DomainName", o)
					tmp = append(tmp, gjson.Parse(value))
				}
			}
			res = tmp
		}
		for _, y := range res {
			value, _ := sjson.Set(y.Raw, "inFrom", inFrom)
			ensInfo.Infos[v] = append(ensInfo.Infos[v], gjson.Parse(value))
		}
		gologger.Infof("%d %s\n", len(res), v)
		if len(res) > 0 {
			//命令输出展示
			var data [][]string
			for _, y := range res {
				results := gjson.GetMany(y.Raw, ensMap[v].field...)
				var str []string
				for _, s := range results {
					str = append(str, s.String())
				}
				data = append(data, str)
			}
			common.TableShow(ensMap[v].keyWord, data, options)
		} else {
			gologger.Infof("【QCC】未查询到%s信息 \n", v)
		}
		if (v == "invest" || v == "holds" || v == "contactrel" || v == "branch") && deep <= options.Deep {
			if v == "branch" && !options.IsGetBranch {
				continue
			}

			for _, y := range res {
				if y.Get("Status").String() == "存续" || y.Get("Status").String() == "在业" {
					time.Sleep(time.Duration(1) * time.Second)
					gologger.Infof("查询公司信息 %s %s %s\n", y.Get("Name"), y.Get("Status"), y.Get("KeyNo"))
					beReason := fmt.Sprintf("%s %s %s %s",
						y.Get("Name"), ensMap[v].name, y.Get("FundedRatio"), y.Get("PercentTotal"))

					getCompanyInfoById(y.Get("KeyNo").String(), deep+1, false, beReason, searchList, ensInfo, options)
				}

			}
		}
	}
}

func getGroupInfoById(pid string, isSearch bool, searchList []string, ensInfo *common.EnInfos, options *common.ENOptions) {
	//fmt.Println(pageParseJson(getReq("https://www.qcc.com/web/groupdetail/"+pid+".html", "", options)))

	// 数据初始化
	ensMap := getENMap("group", pid)
	//需要查询展示的信息~
	if len(searchList) == 0 {
		//searchList = []string{"GroupCompanyList", "GroupPropertyInfo"}
		searchList = []string{"GroupPropertyInfo"}
	}
	for _, v := range searchList {
		gologger.Infof("正在查询 %s\n", ensMap[v].name)
		if v == "GroupPropertyInfo" {
			if len(ensMap[v].typeInfo) > 0 {
				for tK, tV := range ensMap[v].typeInfo {
					if tV != "" {
						gologger.Infof("查询 %s\n", tV)
						if tV == "website" {
							tV = "icp"
						}
						ensMap[v].sData["type"] = strconv.Itoa(tK + 1)
						data, _ := json.Marshal(ensMap[v].sData)
						res := getInfoList(pid, ensMap[v].api, string(data), options)
						fmt.Println(len(res))
						for _, rV := range res {
							//这块先放着等公司信息查询那边完成
							fmt.Println(gjson.GetMany(rV.Raw, ensMap[v].field...))
							if isSearch {
								getCompanyInfoById(rV.Get("KeyNo").String(), 1, false, "成员公司："+rV.Get("Name").String(), []string{tV}, ensInfo, options)
							}

						}
					}
				}

			}
		} else {
			data, _ := json.Marshal(ensMap[v].sData)
			res := getInfoList(pid, ensMap[v].api, string(data), options)
			//命令输出展示
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(ensMap[v].keyWord)
			ensInfo.Infos[v] = res
			for _, y := range res {
				results := gjson.GetMany(y.Raw, ensMap[v].field...)
				var str []string
				for _, s := range results {
					str = append(str, s.String())
				}
				table.Append(str)
			}
			table.Render()
		}
	}
}

// getInfoList 获取信息列表
func getInfoList(pid string, types string, data string, options *common.ENOptions) []gjson.Result {
	var listData []gjson.Result
	urls := "https://www.qcc.com/api/" + types
	if data == "" {
		urls += "?keyNo=" + pid
	}
	content := getReq(urls, data, options)
	if data == "" {
		pageInfo := gjson.Get(content, "pageInfo")
		//判断是否多页，遍历获取所有数据
		pageCount := pageInfo.Get("total").Int()
		pageSize := pageInfo.Get("pageSize").Int()
		pats := "data"
		if types == "datalist/holdcolist" || types == "datalist/indirectlist" {
			pats = "data.Names"
		}
		if pageCount > pageSize {
			for i := 1; int(pageCount/pageSize) >= i-1; i++ {
				gologger.Infof("当前：%s,%d\n", types, i)
				reqUrls := urls + "&pageIndex=" + strconv.Itoa(i)
				time.Sleep(time.Duration(options.GetDelayRTime()) * time.Second)
				content = getReq(reqUrls, "", options)
				listData = append(listData, gjson.Get(content, pats).Array()...)
			}
		} else {
			listData = gjson.Get(content, pats).Array()
		}
	} else {
		pageInfo := gjson.Get(content, "Paging")
		pageCount := pageInfo.Get("TotalRecords").Int()
		pageSize := pageInfo.Get("PageSize").Int()
		if pageCount > pageSize {
			for i := 1; int(pageCount/pageSize) >= i; i++ {
				gologger.Infof("当前：%s,%d\n", types, i)
				data, _ = sjson.Set(data, "pageIndex", i)
				time.Sleep(time.Duration(options.GetDelayRTime()) * time.Second)
				content = getReq(urls, data, options)
				listData = append(listData, gjson.Get(content, "Result").Array()...)
			}
		} else {
			listData = gjson.Get(content, "Result").Array()
		}
	}
	return listData
}

func searchGroupByName(options *common.ENOptions) ([]gjson.Result, string, string) {
	name := options.KeyWord
	url := "https://www.qcc.com/api/bigsearch/getTrzGroup?pageIndex=1&pageSize=20&searchKey=" + urlTool.QueryEscape(name)
	res := getReq(url, "", options)
	enList := gjson.Get(res, "Result").Array()
	if len(enList) == 0 {
		gologger.Errorf("没有查询到关键词 “%s”\n\n ", name)
		return enList, "", ""
	} else {
		gologger.Infof("关键词：“%s” 查询到 %d 个结果，默认选择第一个 \n", name, len(enList))
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "集团名称", "主公司", "疑似控制人", "成员公司数量"})
	for _, v := range enList {
		enName := v.Get("GroupName").String()
		CompanyName := v.Get("CompanyName").String()
		enName = strings.ReplaceAll(enName, "<em>", "")
		enName = strings.ReplaceAll(enName, "</em>", "")
		CompanyName = strings.ReplaceAll(CompanyName, "<em>", "")
		CompanyName = strings.ReplaceAll(CompanyName, "</em>", "")
		table.Append([]string{v.Get("Id").String(), enName, CompanyName, v.Get("ActualController.V").String(), v.Get("GroupCountInfo.n").String()})
	}
	table.Render()
	return enList, enList[0].Get("KeyNo").String(), enList[0].Get("Id").String()
}

func SearchName(options *common.ENOptions) ([]gjson.Result, string) {
	name := options.KeyWord
	url := "https://www.qcc.com/api/search/searchMulti"
	searchStr := `{"pageIndex":1,"pageSize":20,"searchKey":"` + name + `"}`
	res := getReq(url, searchStr, options)
	enList := gjson.Get(res, "Result").Array()
	if len(enList) == 0 {
		gologger.Errorf("没有查询到关键词 “%s” ", name)
		return enList, ""
	} else {
		gologger.Infof("关键词：“%s” 查询到 %d 个结果，默认选择第一个 \n", name, len(enList))
	}
	if options.IsShow {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PID", "企业名称", "法人代表", "社会统一信用代码"})
		for _, v := range enList {
			enName := v.Get("Name").String()
			enName = strings.ReplaceAll(enName, "<em>", "")
			enName = strings.ReplaceAll(enName, "</em>", "")
			table.Append([]string{v.Get("KeyNo").String(), enName, v.Get("OperName").String(), v.Get("CreditCode").String()})
		}
		table.Render()
	}
	return enList, enList[0].Get("KeyNo").String()
}

func SearchByName(options *common.ENOptions) (enName string) {
	res, _ := SearchName(options)
	if len(res) > 0 {
		enName = res[0].Get("Name").String()
		enName = strings.ReplaceAll(enName, "<em>", "")
		enName = strings.ReplaceAll(enName, "</em>", "")
	}
	return enName
}
