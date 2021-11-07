package aiqicha

/* Aiqicha By Keac
 * admin@wgpsec.org
 */
import (
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

func pageParseJson(content string) gjson.Result {
	//提取页面中的JSON字段
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
	}
	return gjson.Result{}
}

// GetEnInfoByPid 根据PID获取公司信息
// pid 爱企查pid参数
func GetEnInfoByPid(pid string) {
	//获取公司基本信息
	res := GetCompanyInfoById(pid)
	//查询对外投资详细信息
	if res.ensMap["invest"].total > 0 {
		res.investInfos = make(map[string]EnInfo)
		for _, t := range res.infos["invest"] {
			gologger.Infof("企业名称：%s 投资占比：%s\n", t.Get("entName"), t.Get("regRate"))
			//openStatus := t.Get("openStatus").String()
			//if openStatus == "注销" || openStatus == "吊销" {
			//	continue
			//}
			investNum := 0.00
			if t.Get("regRate").String() == "-" {
				investNum = -1
			} else {
				str := strings.Replace(t.Get("regRate").String(), "%", "", -1)
				investNum, _ = strconv.ParseFloat(str, 2)
			}
			if investNum >= 100 {
				n := GetCompanyInfoById(t.Get("pid").String())
				res.investInfos[t.Get("pid").String()] = n
			}
		}
	}

	//查询分支机构公司详细信息
	if res.ensMap["branch"].total > 0 {
		res.branchInfos = make(map[string]EnInfo)
		for _, t := range res.infos["branch"] {
			gologger.Infof("分支名称：%s 状态：%s\n", t.Get("entName"), t.Get("openStatus"))
			n := GetCompanyInfoById(t.Get("pid").String())
			res.branchInfos[t.Get("pid").String()] = n
		}
	}

	//导出 【2021.11.7】 暂时还不能一起投资的企业和分支机构的其他信息
	outPutExcelByEnInfo(res)

}

func outPutExcelByEnInfo(enInfo EnInfo) {
	f := excelize.NewFile()
	//Base info
	baseHeaders := []string{"信息", "值"}
	baseData := [][]interface{}{
		{"PID", enInfo.Pid},
		{"企业名称", enInfo.EntName},
		{"法人代表", enInfo.legalPerson},
		{"开业状态", enInfo.openStatus},
		{"电话", enInfo.telephone},
		{"邮箱", enInfo.email},
	}
	f, _ = utils.ExportExcel("基本信息", baseHeaders, baseData, f)

	for k, s := range enInfo.ensMap {
		if s.total > 0 && s.api != "" {
			gologger.Infof("正在导出%s\n", s.name)
			headers := s.keyWord
			var data [][]interface{}
			for _, y := range enInfo.infos[k] {
				results := gjson.GetMany(y.Raw, s.field...)
				var str []interface{}
				for _, s := range results {
					str = append(str, s.String())
				}
				data = append(data, str)
			}
			f, _ = utils.ExportExcel(s.name, headers, data, f)
		}
	}

	f.DeleteSheet("Sheet1")
	// Save spreadsheet by the given path.
	savaPath := "res/" +
		time.Now().Format("2006-01-02") +
		enInfo.EntName + strconv.FormatInt(time.Now().Unix(), 10) + ".xlsx"
	if err := f.SaveAs(savaPath); err != nil {
		gologger.Fatalf("导出失败：%s", err)
	}
	gologger.Infof("导出成功路径： %s\n", savaPath)

}

// GetCompanyInfoById 获取公司基本信息
func GetCompanyInfoById(pid string) EnInfo {
	var enInfo EnInfo
	enInfo.infos = make(map[string][]gjson.Result)
	urls := "https://aiqicha.baidu.com/company_detail_" + pid
	content := common.GetReq(urls)
	res := pageParseJson(string(content))
	//获取企业基本信息情况
	enInfo.Pid = res.Get("pid").String()
	enInfo.EntName = res.Get("entName").String()
	enInfo.legalPerson = res.Get("legalPerson").String()
	enInfo.openStatus = res.Get("openStatus").String()
	enInfo.telephone = res.Get("telephone").String()
	enInfo.email = res.Get("email").String()
	gologger.Infof("企业基本信息\n")
	data := [][]string{
		{"PID", enInfo.Pid},
		{"企业名称", enInfo.EntName},
		{"法人代表", enInfo.legalPerson},
		{"开业状态", enInfo.openStatus},
		{"电话", enInfo.telephone},
		{"邮箱", enInfo.email},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.AppendBulk(data)
	table.Render()

	//判断企业状态，不然就可以跳过了
	if enInfo.openStatus == "注销" || enInfo.openStatus == "吊销" {
		return enInfo
	}

	//获取企业信息
	enInfoUrl := "https://aiqicha.baidu.com/compdata/navigationListAjax?pid=" + pid
	enInfoRes := common.GetReq(enInfoUrl)
	ensInfoMap := make(map[string]*EnsGo)
	if gjson.Get(string(enInfoRes), "status").String() == "0" {
		data := gjson.Get(string(enInfoRes), "data").Array()
		for _, s := range data {
			for _, t := range s.Get("children").Array() {
				ensInfoMap[t.Get("id").String()] = &EnsGo{
					t.Get("name").String(),
					t.Get("total").Int(),
					t.Get("avaliable").Int(),
					"",
					[]string{},
					[]string{},
				}
			}
		}
	}

	//赋值API数据
	ensInfoMap["webRecord"].api = "detail/icpinfoAjax"
	ensInfoMap["webRecord"].field = []string{"domain", "siteName", "homeSite", "icpNo"}
	ensInfoMap["webRecord"].keyWord = []string{"域名", "站点名称", "首页", "ICP备案号"}

	ensInfoMap["appinfo"].api = "c/appinfoAjax"
	ensInfoMap["appinfo"].field = []string{"name", "classify", "logoWord", "logoBrief", "entName"}
	ensInfoMap["appinfo"].keyWord = []string{"APP名称", "分类", "LOGO文字", "描述", "所属公司"}

	ensInfoMap["microblog"].api = "c/microblogAjax"
	ensInfoMap["microblog"].field = []string{"nickname", "weiboLink", "logo"}
	ensInfoMap["microblog"].keyWord = []string{"微博昵称", "链接", "LOGO"}

	ensInfoMap["wechatoa"].api = "c/wechatoaAjax"
	ensInfoMap["wechatoa"].field = []string{"wechatName", "wechatId", "wechatIntruduction", "wechatLogo", "qrcode", "entName"}
	ensInfoMap["wechatoa"].keyWord = []string{"名称", "ID", "描述", "LOGO", "二维码", "归属公司"}

	ensInfoMap["enterprisejob"].api = "c/enterprisejobAjax"
	ensInfoMap["enterprisejob"].field = []string{"jobTitle", "location", "education", "publishDate", "desc"}
	ensInfoMap["enterprisejob"].keyWord = []string{"职位名称", "工作地点", "学历要求", "发布日期", "招聘描述"}

	ensInfoMap["copyright"].api = "detail/copyrightAjax"
	ensInfoMap["copyright"].field = []string{"softwareName", "shortName", "softwareType", "typeCode", "regDate"}
	ensInfoMap["copyright"].keyWord = []string{"软件名称", "软件简介", "分类", "行业", "日期"}

	ensInfoMap["supplier"].api = "c/supplierAjax"
	ensInfoMap["supplier"].field = []string{"supplier", "source", "principalNameClient", "cooperationDate"}
	ensInfoMap["supplier"].keyWord = []string{"供应商名称", "来源", "所属公司", "日期"}

	ensInfoMap["invest"].api = "detail/investajax" //对外投资
	ensInfoMap["invest"].field = []string{"entName", "openStatus", "regRate", "data"}
	ensInfoMap["invest"].keyWord = []string{"公司名称", "状态", "投资比例", "数据信息"}

	ensInfoMap["branch"].api = "detail/branchajax" //分支机构
	ensInfoMap["branch"].field = []string{"entName", "openStatus", "data"}
	ensInfoMap["branch"].keyWord = []string{"公司名称", "状态", "数据信息"}

	enInfo.ensMap = ensInfoMap

	//获取数据
	for k, s := range ensInfoMap {
		if s.total > 0 && s.api != "" {
			gologger.Infof("正在查询 %s\n", s.name)
			t := getInfoList(res.Get("pid").String(), s.api)

			//判断下网站备案，然后提取出来，留个坑看看有没有更好的解决方案
			if k == "webRecord" {
				var tmp []gjson.Result
				for _, y := range t {
					for _, o := range y.Get("domain").Array() {
						value, _ := sjson.Set(y.Raw, "domain", o.String())
						tmp = append(tmp, gjson.Parse(value))
					}
				}
				t = tmp
			}
			//保存数据
			enInfo.infos[k] = t

			//命令输出展示
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(ensInfoMap[k].keyWord)
			for _, y := range t {
				results := gjson.GetMany(y.Raw, ensInfoMap[k].field...)
				var str []string
				for _, s := range results {
					str = append(str, s.String())
				}
				table.Append(str)
			}
			table.Render()

		}
	}

	return enInfo

}

// getInfoList 获取信息列表
func getInfoList(pid string, types string) []gjson.Result {
	urls := "https://aiqicha.baidu.com/" + types + "?size=100&pid=" + pid
	content := common.GetReq(urls)
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
				content = common.GetReq(reqUrls)
				listData = append(listData, gjson.Get(string(content), "data.list").Array()...)
			}
		} else {
			listData = data.Get("list").Array()
		}
	}
	return listData

}

// SearchName 根据企业名称搜索信息
func SearchName(name string) []gjson.Result {
	urls := "https://aiqicha.baidu.com/s/advanceFilterAjax?q=" + name + "&p=1&s=10&t=0"
	content := common.GetReq(urls)
	enList := gjson.Get(string(content), "data.resultList").Array()
	gologger.Infof("关键词：“%s” 查询到 %d 个结果 \n", name, len(enList))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"PID", "企业名称", "法人代表"})
	for _, v := range enList {
		table.Append([]string{v.Get("pid").String(), v.Get("titleName").String(), v.Get("titleLegal").String()})
	}
	table.Render()
	return enList
}
