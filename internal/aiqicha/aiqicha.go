package aiqicha

/* Aiqicha By Keac
 * admin@wgpsec.org
 */
import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	urlTool "net/url"
	"os"
	"strconv"
	"strings"
)

// pageParseJson 提取页面中的JSON字段
func pageParseJson(content string) gjson.Result {

	tag1 := "window.pageData ="
	tag2 := "window.isSpider ="
	//tag2 := "/* eslint-enable */</script><script data-app"
	idx1 := strings.Index(content, tag1)
	idx2 := strings.Index(content, tag2)
	if idx2 > idx1 {
		str := content[idx1+len(tag1) : idx2]
		str = strings.Replace(str, "\n", "", -1)
		str = strings.Replace(str, " ", "", -1)
		str = str[:len(str)-1]
		return gjson.Get(string(str), "result")
	} else {
		gologger.Errorf("无法解析信息错误信息%s\n", content)
	}
	return gjson.Result{}
}

// GetEnInfoByPid 根据PID获取公司信息
func GetEnInfoByPid(options *common.ENOptions) (*common.EnInfos, map[string]*outputfile.ENSMap) {
	pid := ""
	if options.CompanyID == "" {
		_, pid = SearchName(options)
	} else {
		pid = options.CompanyID
	}
	//获取公司信息
	ensInfos := &common.EnInfos{}
	ensOutMap := make(map[string]*outputfile.ENSMap)

	if pid == "" {
		gologger.Errorf("没有获取到PID\n")
		return ensInfos, ensOutMap
	}
	gologger.Infof("查询PID %s\n", pid)

	ensInfos.Infos = make(map[string][]gjson.Result)
	getCompanyInfoById(pid, 1, true, "", options.GetField, ensInfos, options)
	options.CompanyName = ensInfos.Name

	for k, v := range getENMap() {
		ensOutMap[k] = &outputfile.ENSMap{Name: v.name, Field: v.field, KeyWord: v.keyWord}
	}

	//outputfile.OutPutExcelByEnInfo(ensInfos, ensOutMap, options)
	return ensInfos, ensOutMap

}

// getCompanyInfoById 获取公司基本信息
// pid 公司id
// isSearch 是否递归搜索信息【分支机构、对外投资信息】
// options options
func getCompanyInfoById(pid string, deep int, isEnDetail bool, inFrom string, searchList []string, ensInfo *common.EnInfos, options *common.ENOptions) {

	// 获取初始化API数据
	ensInfoMap := getENMap()
	// 企业基本信息获取

	urls := "https://aiqicha.baidu.com/company_detail_" + pid
	res := pageParseJson(common.GetReq(urls, options))
	//获取企业基本信息情况
	enDes := "enterprise_info"
	enJsonTMP, _ := sjson.Set(res.Raw, "inFrom", inFrom)
	ensInfo.Infos[enDes] = append(ensInfo.Infos[enDes], gjson.Parse(enJsonTMP))
	tmpEIS := make(map[string][]gjson.Result)
	if isEnDetail {
		ensInfo.Pid = res.Get("pid").String()
		ensInfo.Name = res.Get("entName").String()
		ensInfo.LegalPerson = res.Get("legalPerson").String()
		ensInfo.OpenStatus = res.Get("openStatus").String()
		ensInfo.Telephone = res.Get("telephone").String()
		ensInfo.Email = res.Get("email").String()
		ensInfo.RegCode = res.Get("taxNo").String()

		gologger.Infof("企业基本信息\n")
		data := [][]string{
			{"PID", ensInfo.Pid},
			{"企业名称", ensInfo.Name},
			{"法人代表", ensInfo.LegalPerson},
			{"开业状态", ensInfo.OpenStatus},
			{"电话", ensInfo.Telephone},
			{"邮箱", ensInfo.Email},
			{"统一社会信用代码", ensInfo.RegCode},
		}
		common.TableShow([]string{}, data, options)
	}

	// 判断企业状态，如果是异常情况就可以跳过了
	if ensInfo.OpenStatus == "注销" || ensInfo.OpenStatus == "吊销" {

	}

	// 获取企业信息列表
	enInfoUrl := "https://aiqicha.baidu.com/compdata/navigationListAjax?pid=" + pid
	enInfoRes := common.GetReq(enInfoUrl, options)

	// 初始化数量数据
	if gjson.Get(enInfoRes, "status").String() == "0" {
		for _, s := range gjson.Get(enInfoRes, "data").Array() {
			for _, t := range s.Get("children").Array() {
				resId := t.Get("id").String()
				if _, ok := common.ENSMapAQC[resId]; ok {
					resId = common.ENSMapAQC[resId]
				}
				es := ensInfoMap[resId]
				if es == nil {
					es = &EnsGo{}
				}
				//fmt.Println(t.Get("name").String() + "|" + t.Get("id").String())
				es.name = t.Get("name").String()
				es.total = t.Get("total").Int()
				es.available = t.Get("avaliable").Int()
				ensInfoMap[t.Get("id").String()] = es
			}
		}
	}

	//获取数据
	for _, k := range searchList {
		if _, ok := ensInfoMap[k]; ok {
			s := ensInfoMap[k]
			if s.total > 0 && s.api != "" {
				if k == "branch" && !options.IsGetBranch {
					continue
				}
				if (k == "invest" || k == "partner" || k == "supplier" || k == "branch" || k == "holds") && (deep > options.Deep) {
					continue
				}
				gologger.Infof("AQC查询 %s\n", s.name)
				t := getInfoList(pid, s.api, options)
				//判断下网站备案，然后提取出来，处理下数据
				if k == "icp" {
					var tmp []gjson.Result
					for _, y := range t {
						for _, o := range y.Get("domain").Array() {
							valueTmp, _ := sjson.Set(y.Raw, "domain", o.String())
							valueTmp, _ = sjson.Set(valueTmp, "homeSite", y.Get("homeSite").Array()[0].String())
							tmp = append(tmp, gjson.Parse(valueTmp))
						}
					}
					t = tmp
				}

				// 添加来源信息，并把信息存储到数据里面
				for _, y := range t {
					valueTmp, _ := sjson.Set(y.Raw, "inFrom", inFrom)
					ensInfo.Infos[k] = append(ensInfo.Infos[k], gjson.Parse(valueTmp))
					//存入临时数据
					tmpEIS[k] = append(tmpEIS[k], gjson.Parse(valueTmp))
				}

				//命令输出展示
				var data [][]string
				for _, y := range t {
					results := gjson.GetMany(y.Raw, ensInfoMap[k].field...)
					var str []string
					for _, ss := range results {
						str = append(str, ss.String())
					}
					data = append(data, str)
				}
				common.TableShow(ensInfoMap[k].keyWord, data, options)
			}
		}
	}
	//判断是否查询层级信息 deep
	if deep <= options.Deep {
		// 查询对外投资详细信息
		// 对外投资>0 && 是否递归 && 参数投资信息大于0
		if ensInfoMap["invest"].total > 0 && options.InvestNum > 0 {
			for _, t := range tmpEIS["invest"] {
				gologger.Infof("企业名称：%s 投资【%d级】占比：%s\n", t.Get("entName"), deep, t.Get("regRate"))
				openStatus := t.Get("openStatus").String()
				if openStatus == "注销" || openStatus == "吊销" {
					continue
				}
				// 计算投资比例信息
				investNum := utils.FormatInvest(t.Get("regRate").String())
				// 如果达到设定要求就开始获取信息
				if investNum >= options.InvestNum {
					beReason := fmt.Sprintf("%s 投资【%d级】占比 %s", t.Get("entName"), deep, t.Get("regRate"))
					getCompanyInfoById(t.Get("pid").String(), deep+1, false, beReason, searchList, ensInfo, options)
				}
			}
		}

		// 查询分支机构公司详细信息
		// 分支机构大于0 && 是否递归模式 && 参数是否开启查询
		// 不查询分支机构的分支机构信息
		if ensInfoMap["branch"].total > 0 && options.IsGetBranch && options.IsSearchBranch {
			for _, t := range tmpEIS["branch"] {
				if t.Get("inFrom").String() == "" {
					openStatus := t.Get("openStatus").String()
					if openStatus == "注销" || openStatus == "吊销" {
						continue
					}
					gologger.Infof("分支名称：%s 状态：%s\n", t.Get("entName"), t.Get("openStatus"))
					beReason := fmt.Sprintf("%s 分支机构", t.Get("entName"))
					getCompanyInfoById(t.Get("pid").String(), -1, false, beReason, searchList, ensInfo, options)
				}
			}
		}

		//查询控股公司
		// 不查询下层信息
		if ensInfoMap["holds"].total > 0 && options.IsHold {
			if len(tmpEIS["holds"]) == 0 {
				gologger.Infof("【无控股信息】，需要账号开通【超级会员】！\n")
			} else {
				for _, t := range tmpEIS["holds"] {
					if t.Get("inFrom").String() == "" {
						openStatus := t.Get("openStatus").String()
						gologger.Infof("控股公司：%s 状态：%s\n", t.Get("entName"), t.Get("openStatus"))
						if openStatus == "注销" || openStatus == "吊销" {
							continue
						}
						beReason := fmt.Sprintf("%s 控股公司投资比例 %s", t.Get("entName"), t.Get("proportion"))
						getCompanyInfoById(t.Get("pid").String(), -1, false, beReason, searchList, ensInfo, options)
					}
				}
			}
		}

		// 查询供应商
		// 不查询下层信息
		if ensInfoMap["supplier"].total > 0 && options.IsSupplier {
			for _, t := range tmpEIS["supplier"] {
				if t.Get("inFrom").String() == "" {
					openStatus := t.Get("openStatus").String()
					gologger.Infof("供应商：%s 状态：%s\n", t.Get("supplier"), t.Get("openStatus"))
					if openStatus == "注销" || openStatus == "吊销" {
						continue
					}
					beReason := fmt.Sprintf("%s 供应商", t.Get("supplier"))
					getCompanyInfoById(t.Get("supplierId").String(), -1, false, beReason, searchList, ensInfo, options)
				}
			}
		}
	}

}

// getInfoList 获取信息列表
func getInfoList(pid string, types string, options *common.ENOptions) []gjson.Result {
	urls := "https://aiqicha.baidu.com/" + types + "?pid=" + pid
	content := common.GetReq(urls, options)
	var listData []gjson.Result
	if gjson.Get(string(content), "status").String() == "0" {
		data := gjson.Get(string(content), "data")
		//判断一个获取的特殊值
		if types == "relations/relationalMapAjax" {
			data = gjson.Get(string(content), "data.investRecordData")
		}
		//判断是否多页，遍历获取所有数据
		pageCount := data.Get("pageCount").Int()
		if pageCount > 1 {
			for i := 1; int(pageCount) >= i; i++ {
				gologger.Infof("当前：%s,%d\n", types, i)
				reqUrls := urls + "&p=" + strconv.Itoa(i)
				content = common.GetReq(reqUrls, options)
				listData = append(listData, gjson.Get(string(content), "data.list").Array()...)
			}
		} else {
			listData = data.Get("list").Array()
		}
	}
	return listData

}

// SearchName 根据企业名称搜索信息
func SearchName(options *common.ENOptions) ([]gjson.Result, string) {
	name := options.KeyWord

	urls := "https://aiqicha.baidu.com/s?q=" + urlTool.QueryEscape(name) + "&t=0"
	content := common.GetReq(urls, options)
	rq := pageParseJson(content)
	enList := rq.Get("resultList").Array()
	if len(enList) == 0 {
		if options.IsDebug {
			gologger.Debugf("【查询错误信息】\n%s\n", content)
		}
		gologger.Errorf("没有查询到关键词 “%s” \n", name)
		return enList, ""
	} else {
		gologger.Infof("关键词：“%s” 查询到 %d 个结果，默认选择第一个 \n", name, len(enList))
	}
	if options.IsShow {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PID", "企业名称", "法人代表", "社会统一信用代码"})
		for _, v := range enList {
			table.Append([]string{v.Get("pid").String(), v.Get("titleName").String(), v.Get("titleLegal").String(), v.Get("regNo").String()})
		}
		table.Render()
	}
	return enList, enList[0].Get("pid").String()
}

func SearchByName(options *common.ENOptions) (enName string) {
	res, _ := SearchName(options)
	if len(res) > 0 {
		enName = res[0].Get("titleName").String()
	}
	return enName
}
