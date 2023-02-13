package tianyancha

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"golang.org/x/net/html"
	"os"
	"strconv"
	"strings"
	"time"
)

/* Tianyancha By Gungnir,Keac
 * admin@wgpsec.org
 */

func GetEnInfoByPid(options *common.ENOptions) (*common.EnInfos, map[string]*outputfile.ENSMap) {
	ensInfos := &common.EnInfos{}
	ensInfos.Infos = make(map[string][]gjson.Result)
	ensInfos.SType = "TYC"
	ensOutMap := make(map[string]*outputfile.ENSMap)
	for k, v := range getENMap() {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}
	pid := ""
	if options.CompanyID == "" {
		_, pid = SearchName(options)
	} else {
		pid = options.CompanyID
	}
	if pid == "" {
		gologger.Errorf("没有获取到PID\n")
		return ensInfos, ensOutMap
	}
	gologger.Infof("查询PID %s\n", pid)

	//ensInfos.Infoss = make(map[string][]map[string]string)
	//获取公司信息
	getCompanyInfoById(pid, 1, true, "", options.GetField, ensInfos, options)
	return ensInfos, ensOutMap

}

// getCompanyInfoById 获取公司基本信息
// pid 公司id
// isSearch 是否递归搜索信息【分支机构、对外投资信息】
// options options
func getCompanyInfoById(pid string, deep int, isEnDetail bool, inFrom string, searchList []string, ensInfo *common.EnInfos, options *common.ENOptions) {
	//获取初始化API数据
	ensMap := getENMap()
	var enCount gjson.Result
	var detailRes gjson.Result
	tmpEIS := make(map[string][]gjson.Result)
	//切换获取企业信息和统计的方式
	tds := false
	//基本信息

	//var res map[string]string
	//res, enCount = SearchBaseInfo(pid, ensInfoMap, options)
	//urls := "https://www.tianyancha.com/company/" + pid
	//body := GetReqReturnPage(urls, options)
	//提取页面的JS数据
	detailRes, enCount = SearchBaseInfo(pid, tds, options)
	enJsonTMP, _ := sjson.Set(detailRes.Raw, "inFrom", inFrom)
	//修复成立日期信息
	ts := time.UnixMilli(detailRes.Get("fromTime").Int())
	enJsonTMP, _ = sjson.Set(enJsonTMP, "fromTime", ts.Format("2006-01-02"))
	ensInfo.Infos["enterprise_info"] = append(ensInfo.Infos["enterprise_info"], gjson.Parse(enJsonTMP))
	//数量统计 API base_count
	//enCount = enBaseInfo.Get("props.pageProps.dehydratedState.queries").Array()[16].Get("state.data")
	if options.IsShow && isEnDetail {
		ensInfo.Name = detailRes.Get("name").String()
		rs := gjson.GetMany(detailRes.Raw, ensMap["enterprise_info"].field...)
		ks := ensMap["enterprise_info"].keyWord
		for k, v := range rs {
			fmt.Println(ks[k] + ":" + v.String())
		}
	}

	//获取数据
	for _, key := range searchList {
		if key == "branch" && !options.IsGetBranch {
			continue
		}
		if (key == "invest" || key == "partner" || key == "holds" || key == "branch" || key == "supplier") && (deep > options.Deep) {
			continue
		}

		if _, ok := ensMap[key]; !ok {
			continue
		}
		if enCount.Get(ensMap[key].gNum).Int() <= 0 && enCount.Get(ensMap[key].tgNum).Int() <= 0 {
			gologger.Infof("【TYC】%s 数量为空，自动跳过\n", ensMap[key].name)
			continue
		}

		s := ensMap[key]
		gologger.Infof("TYC查询 %s\n", s.name)

		res := getInfoList(pid, s.api, s, options)
		for _, y := range res {
			value, _ := sjson.Set(y.Raw, "inFrom", inFrom)
			ensInfo.Infos[key] = append(ensInfo.Infos[key], gjson.Parse(value))
			//存入临时数据
			tmpEIS[key] = append(tmpEIS[key], gjson.Parse(value))
		}

		if len(res) > 0 {
			//命令输出展示
			var data [][]string
			for _, y := range res {
				results := gjson.GetMany(y.Raw, ensMap[key].field...)
				var str []string
				for _, s := range results {
					str = append(str, s.String())
				}
				data = append(data, str)
			}
			common.TableShow(ensMap[key].keyWord, data, options)
		} else {
			gologger.Infof("【TYC】%s 数量为空\n", ensMap[key].name)
		}

	}

	//判断是否查询层级信息 deep
	if deep <= options.Deep {

		// 查询分支机构公司详细信息
		// 分支机构大于0 && 是否递归模式 && 参数是否开启查询
		if options.InvestNum > 0 {
			for _, t := range tmpEIS["invest"] {
				if t.Get("regStatus").String() == "注销" || t.Get("regStatus").String() == "吊销" {
					continue
				}
				gologger.Infof("企业名称：%s 投资占比：%s\n", t.Get("name"), t.Get("percent"))
				// 计算投资比例信息
				investNum := utils.FormatInvest(t.Get("percent").String())

				if investNum >= options.InvestNum {
					beReason := fmt.Sprintf("%s 投资【%d级】占比 %s", t.Get("name"), deep, t.Get("percent"))
					getCompanyInfoById(t.Get("id").String(), deep+1, false, beReason, options.GetField, ensInfo, options)
				}
			}

		}
		// 查询分支机构公司详细信息
		// 分支机构大于0 && 是否递归模式 && 参数是否开启查询
		// 不查询分支机构的分支机构信息
		if options.IsGetBranch && options.IsSearchBranch {
			for _, t := range tmpEIS["branch"] {
				if t.Get("inFrom").String() == "" {
					if t.Get("regStatus").String() == "注销" || t.Get("regStatus").String() == "吊销" {
						continue
					}
					gologger.Infof("分支名称：%s 状态：%s\n", t.Get("name"), t.Get("regStatus"))
					beReason := fmt.Sprintf("%s 分支机构", t.Get("entName"))
					getCompanyInfoById(t.Get("id").String(), 9999, false, beReason, searchList, ensInfo, options)
				}
			}
		}

		//查询控股公司
		// 不查询下层信息
		if options.IsHold {
			if len(tmpEIS["holds"]) == 0 {
				gologger.Infof("需要登陆才能查询控股公司！\n")
			} else {
				for _, t := range tmpEIS["holds"] {
					if t.Get("inFrom").String() == "" {
						gologger.Infof("控股公司：%s 状态：%s\n", t.Get("name"), t.Get("regStatus"))
						if t.Get("regStatus").String() == "注销" || t.Get("regStatus").String() == "吊销" {
							continue
						}
						beReason := fmt.Sprintf("%s 控股公司投资比例 %s", t.Get("name"), t.Get("percent"))
						getCompanyInfoById(t.Get("cid").String(), 9999, false, beReason, searchList, ensInfo, options)
					}
				}
			}
		}
		// 查询供应商
		// 不查询下层信息
		if options.IsSupplier {
			for _, t := range tmpEIS["supplier"] {
				if t.Get("inFrom").String() == "" {
					if t.Get("regStatus").String() == "注销" || t.Get("regStatus").String() == "吊销" {
						continue
					}
					gologger.Infof("供应商：%s 状态：%s\n", t.Get("supplier_name"), t.Get("regStatus"))
					beReason := fmt.Sprintf("%s 供应商", t.Get("supplier_name"))
					getCompanyInfoById(t.Get("supplier_graphId").String(), 9999, false, beReason, searchList, ensInfo, options)
				}
			}
		}

	}

	//return ensInfo.Infos
}

func pageParseJson(content string) (res gjson.Result) {
	content = strings.ReplaceAll(content, "var aa = ", "")
	return gjson.Parse(content)
}

func SearchBaseInfo(pid string, tds bool, options *common.ENOptions) (result gjson.Result, enBaseInfo gjson.Result) {
	urls := "https://www.tianyancha.com/company/" + pid
	body := GetReqReturnPage(urls, options)

	if tds {
		//htmlInfo := htmlquery.FindOne(body, "//*[@class=\"position-rel company-header-container\"]//script")
		//enBaseInfo = pageParseJson(htmlquery.InnerText(htmlInfo))
		result = gjson.Get(GetReq("https://capi.tianyancha.com/cloud-other-information/companyinfo/baseinfo/web?id="+pid, "", options), "data")
		fmt.Println(result.String())
	} else {
		htmlInfos := htmlquery.FindOne(body, "//*[@id=\"__NEXT_DATA__\"]")
		enInfo := gjson.Parse(htmlquery.InnerText(htmlInfos))
		enInfoD := enInfo.Get("props.pageProps.dehydratedState.queries").Array()
		result = enInfoD[0].Get("state.data.data")
		//数量统计 API base_count
		for i := 0; i < len(enInfoD); i++ {
			if enInfoD[i].Get("queryKey").String() == "base_count" {
				enBaseInfo = enInfoD[i].Get("state.data")
			}
		}
		//enBaseInfo = enInfo.Get("props.pageProps.dehydratedState.queries").Array()[11].Get("state.data")
	}

	return result, enBaseInfo
}

func SearchBaseInfoByTables(pid string, ensMap map[string]*EnsGo, options *common.ENOptions) (result map[string]string, enInfoCout gjson.Result) {
	//var re = regexp.MustCompile(`(?m)placeholder="请输入公司名称、老板姓名、品牌名称等\"\s*value="(.*?)\"\/>`)
	defer func() {
		if x := recover(); x != nil {
			gologger.Errorf("[TYC] SearchBaseInfo panic: %v", x)
		}
	}()
	urls := "https://www.tianyancha.com/company/" + pid
	body := GetReqReturnPage(urls, options)
	htmlInfo := htmlquery.FindOne(body, "//*[@class=\"position-rel company-header-container\"]//script")
	enInfoCout = pageParseJson(htmlquery.InnerText(htmlInfo))
	htmlAll := htmlquery.Find(body, "//*[@id=\"_container_baseInfo\"]/table")
	result = make(map[string]string)
	isOrg := len(htmlquery.Find(htmlAll[0], "//tr"))
	var orgPs [][]int
	if isOrg == 5 {
		//兼容事业单位 社会组织
		orgPs = [][]int{{0}, {1, 2}, {1, 6}, {1}, {2}, {1, 4}, {3, 6}, {4, 2}, {5, 2}, {3, 4}, {}}
	} else {
		orgPs = ensMap["enterprise_info"].PosiToTaeS
	}

	for k, v := range orgPs {
		esf := ensMap["enterprise_info"].field[k]
		if esf == "pid" {
			result[esf] = pid
		}
		if len(v) == 1 {
			res := htmlquery.Find(body, "//*[@data-clipboard-target=\"#copyCompanyInfoThroughThisTag\"]")
			if len(res) > v[0] {
				result[esf] = htmlquery.InnerText(res[v[0]])
			}

		}
		if len(v) == 2 {
			expr := fmt.Sprintf("//tr[%d]/td[%d]", v[0], v[1])
			if esf == "legalPerson" && isOrg != 5 {
				expr += "//a"
			}
			htmlResTmp := htmlquery.Find(htmlAll[0], expr)
			if len(htmlResTmp) > 0 {
				result[esf] = htmlquery.InnerText(htmlResTmp[0])
			}
		}
	}

	return result, enInfoCout
}

func SearchName(options *common.ENOptions) ([]gjson.Result, string) {
	name := options.KeyWord
	//使用关键词推荐方法进行检索，会出现信息不对的情况
	//urls := "https://sp0.tianyancha.com/search/suggestV2.json?key=" + url.QueryEscape(name)
	urls := "https://capi.tianyancha.com/cloud-tempest/web/searchCompanyV3"
	searchData := map[string]string{
		"key":      name,
		"pageNum":  "1",
		"pageSize": "20",
		"referer":  "search",
		"sortType": "0",
		"word":     name,
	}
	marshal, err := json.Marshal(searchData)
	if err != nil {
		gologger.Errorf("[ERROR]TYC 关键词JSON错误")
		return nil, ""
	}
	content := GetReq(urls, string(marshal), options)
	enList := gjson.Get(content, "data.companyList").Array()

	if len(enList) == 0 {
		gologger.Errorf("没有查询到关键词 “%s” ", name)
		return enList, ""
	} else {
		gologger.Infof("关键词：“%s” 查询到 %d 个结果，默认选择第一个 \n", name, len(enList))
	}
	if options.IsShow {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PID", "企业名称"})
		for _, v := range enList {
			table.Append([]string{v.Get("id").String(), utils.DName(v.Get("name").String())})
		}
		table.Render()
	}
	return enList, enList[0].Get("id").String()
}

//func getKeyWord(page *html.Node, ensInfoMap *EnsGo) {
//	a, err := htmlquery.QueryAll(page, "//th")
//	if err != nil {
//		panic(`not a valid XPath expression.`)
//	}
//	for _, value := range a {
//		state := htmlquery.InnerText(value)
//		if state == "序号" {
//			continue
//		}
//		ensInfoMap.keyWord = append(ensInfoMap.keyWord, state)
//	}
//}

func JudgePageNumWithCookie(page *html.Node) int {
	list := htmlquery.Find(page, "//li")
	return len(list) - 1
}

func getInfoList(pid string, types string, s *EnsGo, options *common.ENOptions) (listData []gjson.Result) {
	data := ""
	if len(s.sData) != 0 {
		dataTmp, _ := json.Marshal(s.sData)
		data = string(dataTmp)
	}
	urls := "https://capi.tianyancha.com/" + types + "?_=" + strconv.Itoa(int(time.Now().Unix()))

	if data == "" {
		urls += "&pageSize=100&graphId=" + pid + "&id=" + pid + "&gid=" + pid + "&pageNum=1" + s.gsData
	} else {
		data, _ = sjson.Set(data, "gid", pid)
		data, _ = sjson.Set(data, "pageSize", 100)
		data, _ = sjson.Set(data, "pageNum", 1)
	}
	gologger.Debugf("[TYC] getInfoList %s\n", urls)
	content := GetReq(urls, data, options)
	if gjson.Get(content, "state").String() != "ok" {
		gologger.Errorf("[TYC] getInfoList %s\n", content)
		return listData
	}
	pageCount := 0
	pList := []string{"itemTotal", "count", "total", "pageBean.total"}
	for _, k := range gjson.GetMany(gjson.Get(content, "data").Raw, pList...) {
		if k.Int() != 0 {
			pageCount = int(k.Int())
		}
	}
	pats := "data." + s.rf

	listData = gjson.Get(content, pats).Array()
	if pageCount > 100 {
		urls = strings.ReplaceAll(urls, "&pageNum=1", "")
		for i := 2; int(pageCount/100) >= i-1; i++ {
			gologger.Infof("当前：%s,%d\n", types, i)
			reqUrls := urls
			if data == "" {
				reqUrls = urls + "&pageNum=" + strconv.Itoa(i)
			} else {
				data, _ = sjson.Set(data, "pageNum", i)
			}

			time.Sleep(time.Duration(options.GetDelayRTime()) * time.Second)
			content = GetReq(reqUrls, data, options)
			listData = append(listData, gjson.Get(content, pats).Array()...)
		}
	}

	//urls := "https://www.tianyancha.com/" + ensInfoMap.api + "?ps=30&id=" + pid
	//page := GetReqReturnPage(urls, options)
	//
	//List := getTb(page, ensInfoMap, 1)
	//page_num := JudgePageNumWithCookie(page)
	//if page_num > 1 {
	//	for i := 2; i <= page_num; i++ {
	//		urls = "https://www.tianyancha.com/" + ensInfoMap.api + "?ps=30&id=" + pid + "&pn=" + strconv.Itoa(i)
	//		page = GetReqReturnPage(urls, options)
	//		tmp_List := getTb(page, ensInfoMap, i)
	//		List = append(List, tmp_List...)
	//	}
	//}
	//

	return listData

}

func getInfoListByTable(pid string, ensInfoMap *EnsGo, options *common.ENOptions) []map[string]string {
	urls := "https://www.tianyancha.com/" + ensInfoMap.api + "?ps=30&id=" + pid
	page := GetReqReturnPage(urls, options)

	List := getTb(page, ensInfoMap, 1)
	page_num := JudgePageNumWithCookie(page)
	if page_num > 1 {
		for i := 2; i <= page_num; i++ {
			urls = "https://www.tianyancha.com/" + ensInfoMap.api + "?ps=30&id=" + pid + "&pn=" + strconv.Itoa(i)
			page = GetReqReturnPage(urls, options)
			tmp_List := getTb(page, ensInfoMap, i)
			List = append(List, tmp_List...)
		}
	}
	return List

}

func SearchByName(options *common.ENOptions) (enName string) {
	res, _ := SearchName(options)
	if len(res) > 0 {
		enName = res[0].Get("comName").String()
	}
	return enName
}

func getTb(page *html.Node, ensInfoMap *EnsGo, page_num int) []map[string]string {
	defer func() {
		if x := recover(); x != nil {
			gologger.Errorf("[TYC] getTb panic: %v", x)
		}
	}()
	var infoss []map[string]string
	exrps := "//body/table/tbody/tr"
	htmlAll, err := htmlquery.QueryAll(page, exrps)
	flag := false
	if len(htmlAll) == 0 {
		flag = true
		exrps = "//body/div/table/tbody/tr"
		htmlAll, err = htmlquery.QueryAll(page, exrps)
	}
	//doc := goquery.NewDocumentFromNode(page)
	if err != nil {
		panic(`not a valid XPath expression.`)
	}
	//doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
	//	s.Find("td").Each(func(ii int, t *goquery.Selection) {
	//		fmt.Println(t.Text())
	//	})
	//})

	for i := 0; i < len(htmlAll); i++ {
		result := make(map[string]string)
		for tmpNum := 0; tmpNum < len(ensInfoMap.field)-1; tmpNum++ {
			if (ensInfoMap.field[tmpNum] == "") || ensInfoMap.PosiToTake[tmpNum] == 0 {
				continue
			}
			expr := "//td"
			if flag {
				expr = "/td"
			}
			htmls, _ := htmlquery.QueryAll(htmlAll[i], expr)
			//fmt.Println(htmlquery.InnerText(htmlAll[i]))
			//if len(htmls) == 2 {
			//	fmt.Println(htmlquery.InnerText(htmls[0]))
			//}

			taskNum := ensInfoMap.PosiToTake[tmpNum] - 1
			if len(htmls) < taskNum {
				gologger.Errorf("[TYC] htmls len < taskNum, htmls len: %d, taskNum: %d", len(htmls), taskNum)
				continue
			}
			htmlAllS := htmls[taskNum]

			if ensInfoMap.field[tmpNum] == "logo" || ensInfoMap.field[tmpNum] == "qrcode" {
				htmlA := htmlquery.Find(htmlAllS, "//img")
				if len(htmlA) > 0 {
					result[ensInfoMap.field[tmpNum]] = htmlquery.SelectAttr(htmlA[0], "data-src")
				}
			} else if ensInfoMap.field[tmpNum] == "pid" || ensInfoMap.field[tmpNum] == "StockName" || ensInfoMap.field[tmpNum] == "legalPerson" || ensInfoMap.field[tmpNum] == "href" {
				htmlA := htmlquery.Find(htmlAllS, "//a")
				if len(htmlA) > 0 {
					if ensInfoMap.field[tmpNum] == "pid" || ensInfoMap.field[tmpNum] == "href" {
						if ensInfoMap.name == "股东信息" && strings.Contains(htmlquery.SelectAttr(htmlA[0], "href"), "human") {
							result[ensInfoMap.field[tmpNum]] = ""
							//result[ensInfoMap.field[tmpNum]] = strings.ReplaceAll(htmlquery.SelectAttr(htmlA[0], "href"), " https://www.tianyancha.com/human/", "")

						} else {
							result[ensInfoMap.field[tmpNum]] = strings.ReplaceAll(htmlquery.SelectAttr(htmlA[0], "href"), "https://www.tianyancha.com/company/", "")
						}

					} else {
						result[ensInfoMap.field[tmpNum]] = htmlquery.InnerText(htmlA[0])
					}
				}
			} else if ensInfoMap.name == "供应商" && (ensInfoMap.field[tmpNum] == "entName" || ensInfoMap.field[tmpNum] == "source") {

				htmlA := htmlquery.Find(htmlAllS, "//a")
				if ensInfoMap.field[tmpNum] == "entName" {
					if len(htmlA) > 0 {
						result[ensInfoMap.field[tmpNum]] = htmlquery.InnerText(htmlA[0])
					} else {
						str := htmlquery.InnerText(htmlAllS)
						str = strings.ReplaceAll(str, "查看全部", "")
						str = strings.ReplaceAll(str, "条采购数据", "")
						result[ensInfoMap.field[tmpNum]] = str
					}
				}
				if ensInfoMap.field[tmpNum] == "source" {
					result[ensInfoMap.field[tmpNum]] = htmlquery.InnerText(htmlAllS)
					if len(htmlA) > 0 {
						result[ensInfoMap.field[tmpNum]] += "https://www.tianyancha.com/" + htmlquery.SelectAttr(htmlA[0], "href")
					}
				}

			} else if ensInfoMap.name == "投资信息" && (ensInfoMap.field[tmpNum] == "entName") {
				htmlA := htmlquery.Find(htmlAllS, "//a")
				if len(htmlA) > 0 {
					result[ensInfoMap.field[tmpNum]] = htmlquery.InnerText(htmlA[0])
				} else {
					result[ensInfoMap.field[tmpNum]] = strings.ReplaceAll(htmlquery.InnerText(htmlAllS), "股权结构", "")
				}
			} else if ensInfoMap.name == "ICP备案" {
				result[ensInfoMap.field[tmpNum]] = strings.ReplaceAll(htmlquery.InnerText(htmlAllS), "该网站与ICP或年报备案网站一致", "")
			} else {
				txt := htmlquery.InnerText(htmlAllS)
				txt = strings.ReplaceAll(txt, "... 更多", "")
				result[ensInfoMap.field[tmpNum]] = txt
			}

			//result[ensInfoMap.field[tmp_num]] = htmlquery.InnerText(a[tmp_num+i*ensInfoMap.NumOfEachGroup])
		}
		infoss = append(infoss, result)
	}
	return infoss
}
