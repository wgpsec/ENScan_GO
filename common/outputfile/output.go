package outputfile

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"github.com/wgpsec/ENScan/db"
	"github.com/wgpsec/ENScan/internal/hook"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type ENSMap struct {
	Name    string
	Field   []string
	JField  []string
	KeyWord []string
	Only    string
}

var EnsInfosList = make(map[string][][]interface{})
var ENSMapList = make(map[string]*ENSMap)
var ENSMapLN = map[string]*ENSMap{
	"enterprise_info": {
		Name:    "企业信息",
		JField:  []string{"name", "legal_person", "status", "phone", "email", "registered_capital", "incorporation_date", "address", "scope", "reg_code", "pid"},
		KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
	},
	"icp": {
		Name:    "ICP信息",
		Only:    "domain",
		JField:  []string{"wesbite_name", "website", "domain", "icp", "company_name"},
		KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
	},
	"wx_app": {
		Name:    "微信小程序",
		JField:  []string{"name", "category", "logo", "qrcode", "read_num"},
		KeyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
	},
	"wechat": {
		Name:    "微信公众号",
		JField:  []string{"name", "wechat_id", "description", "qrcode", "avatar"},
		KeyWord: []string{"名称", "ID", "简介", "二维码", "头像"},
	},
	"weibo": {
		Name:    "微博",
		JField:  []string{"name", "profile_url", "description", "avatar"},
		KeyWord: []string{"微博昵称", "链接", "简介", "头像"},
	},
	"supplier": {
		Name:    "供应商",
		JField:  []string{"name", "scale", "amount", "report_time", "data_source", "relation", "pid"},
		KeyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
	},
	"job": {
		Name:    "招聘",
		JField:  []string{"name", "education", "location", "publish_time", "salary"},
		KeyWord: []string{"招聘职位", "学历", "办公地点", "发布日期", "薪资"},
	},
	"invest": {
		Name:    "投资",
		JField:  []string{"name", "legal_person", "status", "scale", "pid"},
		KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
	},
	"branch": {
		Name:    "分支机构",
		JField:  []string{"name", "legal_person", "status", "pid"},
		KeyWord: []string{"企业名称", "法人", "状态", "PID"},
	},
	"holds": {
		Name:    "控股企业",
		JField:  []string{"name", "legal_person", "status", "scale", "level", "pid"},
		KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
	},
	"app": {
		Name:    "应用",
		JField:  []string{"name", "category", "version", "update_at", "description", "logo", "bundle_id", "link", "market"},
		KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
	},
	"copyright": {
		Name:    "软件著作权",
		JField:  []string{"name", "short_name", "category", "reg_num", "pub_type"},
		KeyWord: []string{"软件全称", "软件简称", "分类", "登记号", "权利取得方式"},
	},
	"partner": {
		Name:    "股东信息",
		JField:  []string{"name", "scale", "reg_cap", "pid"},
		KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
	},
}
var EnJsonList = make(map[string][]map[string]interface{})

func GetEmails(f *excelize.File, options *common.ENOptions) {
	if options.IsEmailPro {
		headers := []string{"邮箱", "电话", "来源"}
		gologger.Infof("开始获取EMAIL信息\n")
		rData := hook.GetEnEmail(EnsInfosList, options)
		var data [][]interface{}
		for _, v := range rData {
			data = append(data, []interface{}{v.Email, v.Phone, v.Source})
		}
		f, _ = utils.ExportExcel("邮箱地址", headers, data, f)
	}
}

// MergeOutPut 数据合并到MAP
func MergeOutPut(ensInfos *common.EnInfos, ensMap map[string]*ENSMap, info string, options *common.ENOptions) map[string][][]interface{} {
	gologger.Infof("%s【%s】信息合并\n", info, ensInfos.Name)

	//if ensInfos.SType == "TYC" {
	//	for k, s := range ensInfos.Infoss {
	//		ENSMapList[k] = ensMap[k]
	//		var data [][]interface{}
	//		for _, y := range s {
	//			var str []interface{}
	//			for _, t := range ensMap[k].Field {
	//				str = append(str, y[t])
	//			}
	//			if !options.IsApiMode {
	//				str = append(str, info+"【"+ensInfos.Name+"】")
	//			} else {
	//				str = append(str, info)
	//			}
	//			data = append(data, str)
	//		}
	//		EnsInfosList[k] = append(EnsInfosList[k], data...)
	//	}
	//} else {
	for k, s := range ensInfos.Infos {
		ENSMapList[k] = ensMap[k]
		var data [][]interface{}
		for _, y := range s {
			results := gjson.GetMany(y.Raw, ensMap[k].Field...)
			var str []interface{}
			for _, t := range results {
				str = append(str, t.String())
			}
			if !options.IsApiMode {
				str = append(str, info+"【"+ensInfos.Name+"】")
			} else {
				str = append(str, info)
			}
			data = append(data, str)
		}
		EnsInfosList[k] = append(EnsInfosList[k], data...)
	}

	//}
	return EnsInfosList
}

// OutPutJsonByMergeEnInfo 根据合并数据写入数据库
func OutPutJsonByMergeEnInfo(ENOptions *common.ENOptions) error {
	gologger.Infof("【%s】 数据写入中\n", ENOptions.CompanyName)
	for k, s := range EnsInfosList {
		if _, ok := ENSMapLN[k]; ok {
			gologger.Infof("处理数据【%s】\n", ENSMapList[k].Name)
			for _, v := range s {
				var data = make(map[string]interface{})
				list := ENSMapLN[k].JField
				list = append(list, "relate")
				list = append(list, "source")
				for i, t := range v {
					data[list[i]] = t
				}
				EnJsonList[k] = append(EnJsonList[k], data)
			}
		} else {
			gologger.Errorf("导出错误信息 %s", k)
		}
	}

	if ENOptions.IsApiMode {
		if len(EnJsonList["enterprise_info"]) > 0 {
			//数据判断处理 几个数据源的企业名称是不是一样
			v0 := EnJsonList["enterprise_info"][0]
			for _, v := range EnJsonList["enterprise_info"] {
				v["name"] = utils.DName(v["name"].(string))
				v0["name"] = utils.DName(v0["name"].(string))
				if v["name"] != v0["name"] || v["reg_code"] != v0["reg_code"] {
					gologger.Errorf("【企业查询数据不一致，无法入库】\n")
					gologger.Errorf("%s %s %s | %s %s %s\n", v["source"], v["name"], v["reg_code"], v0["source"], v0["name"], v0["reg_code"])
					break
				}
			}

			database := db.MongoBb.Database("ENScan")

			//为更新企业数据的准备
			var updateEnId primitive.ObjectID
			ENCountUpdateOld := make(map[string][]string)

			for k, v := range EnJsonList {
				//连接不同的库
				collection := database.Collection(k)
				ens := common.DBEnInfos{
					Id:      primitive.NewObjectID(),
					Name:    utils.DName(v0["name"].(string)),
					RegCode: v0["reg_code"].(string),
					InTime:  time.Now(),
					Info:    v,
				}

				//根据关键词和统一社会信用码查询，防止错误
				var newEns common.DBEnInfos
				err := collection.FindOne(context.TODO(), bson.M{
					"name":    v0["name"],
					"regcode": v0["reg_code"],
				}).Decode(&newEns)

				if err != nil { //没有数据全新插入数据
					if k == "enterprise_info" {
						//数据标记，添加已经查询过的数据源和信息
						ens.InfoCount = make(map[string][]string)
						for _, VF := range ENOptions.GetField {
							ens.InfoCount[VF] = ENOptions.GetType
						}
					}
					for l, vs := range v {
						//对企业名称进行处理 爱企查的()
						if k == "enterprise_info" {
							v[l]["name"] = utils.DName(v[l]["name"].(string))
						}

						if k == "invest" { //对投资信息数据进行标记
							if len(vs) > 0 {
								ens.InvestCount = len(vs)
							} else {
								gologger.Infof("没有invest，标记为无数据\n")
								ens.InvestCount = -1
							}
							v[l]["scale"] = utils.FormatInvest(v[l]["scale"].(string))
						}
						//把数字格式转换
						if k == "holds" || k == "partner" || k == "supplier" {
							v[l]["scale"] = utils.FormatInvest(v[l]["scale"].(string))
						}
						//添加入库时间
						v[l]["intime"] = time.Now()
					}

					//查了投资信息，但是没结果 也是标记无数据
					if utils.IsInList("invest", ENOptions.GetField) && k == "enterprise_info" {
						if ok := EnJsonList["invest"]; len(ok) <= 0 {
							gologger.Infof("没有invest，标记为无数据\n")
							ens.InvestCount = -1
						}
					}

					insertOne, err := collection.InsertOne(context.TODO(), ens)
					if err != nil {
						gologger.Errorf("%s 入库失败\n", k)
					}
					gologger.Infof("%s 入库成功 %s\n", k, insertOne.InsertedID)
				} else {
					//遍历更新数据
					if len(v) > 0 {
						//保存旧数据，为了后面更新做个标记
						if k == "enterprise_info" {
							updateEnId = newEns.Id
							if newEns.InfoCount != nil {
								ENCountUpdateOld = newEns.InfoCount
							} else {
								ENCountUpdateOld = make(map[string][]string)
							}
						}
						var listStr []string

						//添加入库时间，处理数字格式的数据
						for l, vv := range v {
							if k == "enterprise_info" {
								v[l]["name"] = utils.DName(vv["name"].(string))
							}
							listStr = append(listStr, vv["source"].(string))
							v[l]["intime"] = time.Now()
							if k == "partner" || k == "holds" || k == "invest" || k == "supplier" {
								v[l]["scale"] = utils.FormatInvest(v[l]["scale"].(string))
							}
						}
						listStr = utils.SetStr(listStr)

						//判断从数据库里面拿出来的数据，和我要写入的数据源是不是一样，一样的数据源覆盖更新
						for _, cc := range newEns.Info {
							flag := true
							for _, sourceV := range listStr {
								if cc["source"] == sourceV {
									flag = false
									break
								}
							}
							if flag {
								v = append(v, cc)
							}
						}

						//标记下没有投资信息
						if utils.IsInList("invest", ENOptions.GetField) && k == "enterprise_info" {
							if ok := EnJsonList["invest"]; len(ok) == 0 {
								gologger.Infof("没有invest，标记为无数据\n")
								ens.InvestCount = -1
							}
						}
						//如果不是企业信息，就把完整的数据量更新下
						filter := bson.D{{"_id", newEns.Id}}
						//如果过滤的文档不存在，则插入新的文档
						opts := options.Update().SetUpsert(true)
						update := bson.D{{"$set", bson.M{"info": v}}}
						result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
						if err != nil {
							log.Fatal(err)
						}
						if result.MatchedCount != 0 {
							gologger.Infof("匹配%v条文档，更新%v条信息\n", result.MatchedCount, result.ModifiedCount)
						}
						if result.UpsertedCount != 0 {
							gologger.Infof("更新时找不到数据，插入 %v\n", result.UpsertedID)
						}
					}
				}

			}

			//单独更新下企业信息的数值，更新数量信息 防止添加任务时候重复查询，漏掉数据
			if !updateEnId.IsZero() {
				/*把查询的参数进行遍历去重，写入数据库
				  icp:[aqc,qcc,tyc]
				*/
				for _, VF := range ENOptions.GetField {
					ENCountUpdateOld[VF] = append(ENCountUpdateOld[VF], ENOptions.GetType...)
					ENCountUpdateOld[VF] = utils.SetStr(ENCountUpdateOld[VF])
				}
				filter := bson.D{{"_id", updateEnId}}
				opts := options.Update()
				collection := database.Collection("enterprise_info")
				update := bson.D{{"$set", bson.M{"infocount": ENCountUpdateOld}}}
				result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
				if err != nil {
					gologger.Errorf("插入失败")
				}
				if result.MatchedCount != 0 {
					gologger.Infof("【数据标记】匹配%v条文档，更新%v条信息\n", result.MatchedCount, result.ModifiedCount)
				}
				if result.UpsertedCount != 0 {
					gologger.Infof("更新时找不到数据，插入 %v\n", result.UpsertedID)
				}

			}

			// OLD DATA
			//collection := database.Collection("infos")
			//ens := common.EnInfos{
			//	Id:      primitive.NewObjectID(),
			//	Name:    EnJsonList["enterprise_info"][0]["name"].(string),
			//	RegCode: EnJsonList["enterprise_info"][0]["reg_code"].(string),
			//	EnInfos: EnJsonList,
			//}
			//var newEns common.EnInfos
			//errs := collection.FindOne(context.TODO(), bson.M{"name": EnJsonList["enterprise_info"][0]["name"].(string)}).Decode(&newEns)
			//if errs != nil {
			//	gologger.Errorf(errs.Error())
			//} else {
			//	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": newEns.Id})
			//	if err != nil {
			//		gologger.Errorf(err.Error())
			//		return err
			//	}
			//}
			//insertOne, err := collection.InsertOne(context.TODO(), ens)
			//if err != nil {
			//	gologger.Errorf(err.Error())
			//}
			//gologger.Infof("入库成功 %s", insertOne.InsertedID)

			//把之前的标记去掉，可以再次查询
			db.RedisBb.Del("ENScanK:" + ENOptions.KeyWord)
			db.RedisBb.Del("ENScanK:" + ENOptions.CompanyID)
		} else {
			gologger.Errorf("没有查询到任何企业数据，无法完成入库操作！\n")
		}
	}

	EnJsonList = make(map[string][]map[string]interface{})
	EnsInfosList = make(map[string][][]interface{})
	ENSMapList = make(map[string]*ENSMap)

	return nil
}

// OutPutExcelByMergeJson 合并导出从数据库提取的信息 流输出
// out 为输入的数据
func OutPutExcelByMergeJson(out map[string][]map[string]interface{}, w io.Writer) error {

	f := excelize.NewFile()
	gologger.Infof("JSON 导出中\n")
	for k, s := range out {
		if _, ok := ENSMapLN[k]; ok {
			gologger.Infof("正在导出%s\n", ENSMapLN[k].Name)
			headers := ENSMapLN[k].KeyWord
			headers = append(headers, "信息来源")
			headers = append(headers, "关联信息")
			headers = append(headers, "入库时间")
			var data [][]interface{}
			ENSMapLN[k].JField = append(ENSMapLN[k].JField, "source", "relate", "intime")
			if len(headers) != len(ENSMapLN[k].JField) {
				gologger.Errorf("len not ok %s", headers)
			}
			for _, ss := range s {
				var str []interface{}
				for _, sss := range ENSMapLN[k].JField {
					str = append(str, ss[sss])
				}
				fmt.Println(str)
				data = append(data, str)
			}

			var err error
			f, err = utils.ExportExcel(ENSMapLN[k].Name, headers, data, f)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
		} else {
			gologger.Errorf("导出错误信息 %s\n", k)
		}
	}
	f.DeleteSheet("Sheet1")

	EnJsonList = make(map[string][]map[string]interface{})
	EnsInfosList = make(map[string][][]interface{})
	ENSMapList = make(map[string]*ENSMap)

	_, err := f.WriteTo(w)
	if err != nil {
		gologger.Errorf("导出错误信息 %s\n", err.Error())
		return err
	}
	return nil

}

// OutPutExcelByMergeEnInfo 根据合并信息导出表格
func OutPutExcelByMergeEnInfo(options *common.ENOptions) {

	tmp := options.Output
	_, err := os.Stat(tmp)
	if err != nil {
		gologger.Infof("【%s】目录不存在，自动创建\n", tmp)
		err := os.Mkdir(tmp, os.ModePerm)
		if err != nil {
			gologger.Fatalf("缺少%s文件夹，并且创建失败！", tmp)
		}
	}
	// Save spreadsheet by the given path.
	fileName := ""
	if len([]rune(options.CompanyName)) > 20 {
		fileName = options.KeyWord
	} else {
		fileName = options.CompanyName
	}
	savaPath := tmp + "/【合并】" + fileName + "--" + time.Now().Format("2006-01-02") + "--" + strconv.FormatInt(time.Now().Unix(), 10)

	if options.IsJsonOutput {
		savaPath += ".json"
		jsonData := map[string][]map[string]interface{}{}
		for k, s := range EnsInfosList {
			//if _, ok := ENSMapList[k]; ok {
			for _, s1 := range s {
				tmps := map[string]interface{}{}
				for k1, v1 := range ENSMapLN[k].JField {
					tmps[v1] = s1[k1]
				}
				jsonData[k] = append(jsonData[k], tmps)
			}

			//}
		}

		jsonStr, err := json.Marshal(jsonData)
		if err != nil {
			gologger.Fatalf("JSON导出失败：%s", err)
		}
		err = ioutil.WriteFile(savaPath,
			jsonStr, 0644)
		if err != nil {
			gologger.Errorf("文件写入失败 %v", err)
		}
	} else {
		savaPath += ".xlsx"
		f := excelize.NewFile()
		gologger.Infof("【%s】导出中\n", options.CompanyName)

		for k, s := range EnsInfosList {
			if _, ok := ENSMapList[k]; ok {
				gologger.Infof("正在导出%s\n", ENSMapList[k].Name)
				headers := ENSMapList[k].KeyWord
				headers = append(headers, "查询信息")
				data := s
				f, _ = utils.ExportExcel(ENSMapList[k].Name, headers, data, f)
			} else {
				gologger.Errorf("导出错误信息 %s\n", k)
			}
		}
		GetEmails(f, options)

		f.DeleteSheet("Sheet1")

		if err := f.SaveAs(savaPath); err != nil {
			gologger.Fatalf("导出失败：%s", err)
		}
	}

	gologger.Infof("导出成功路径： %s\n", savaPath)
	EnJsonList = make(map[string][]map[string]interface{})
	EnsInfosList = make(map[string][][]interface{})
	ENSMapList = make(map[string]*ENSMap)

}

// OutPutExcelByEnInfo 直接导出单独表格信息
func OutPutExcelByEnInfo(ensInfos *common.EnInfos, ensMap map[string]*ENSMap, options *common.ENOptions) {
	if ensInfos.Name == "" {
		ensInfos.Name = options.KeyWord
	}
	if ensInfos.Name != "" && !options.IsApiMode {
		//初始化导出目录
		tmp := options.Output
		_, err := os.Stat(tmp)
		if err != nil {
			gologger.Infof("【%s】目录不存在，自动创建\n", tmp)
			err := os.Mkdir(tmp, os.ModePerm)
			if err != nil {
				gologger.Fatalf("缺少%s文件夹，并且创建失败！", tmp)
			}
		}
		// 修复导出文件名过长的问题
		fileName := ""
		if len([]rune(ensInfos.Name)) > 20 {
			fileName = options.KeyWord
		} else {
			fileName = ensInfos.Name
		}
		savaPath := tmp + "/" + fileName + "--" + time.Now().Format("2006-01-02") + "--" + strconv.FormatInt(time.Now().Unix(), 10)
		if options.IsJsonOutput {
			savaPath += ".json"
			jsonData := map[string][]map[string]interface{}{}
			for k1, y := range ensInfos.Infos {
				for _, s := range y {
					jsonData[k1] = append(jsonData[k1], s.Value().(map[string]interface{}))
				}
			}
			jsonStr, err := json.Marshal(jsonData)
			if err != nil {
				gologger.Fatalf("JSON导出失败：%s", err)
			}
			err = ioutil.WriteFile(savaPath,
				jsonStr, 0644)
			if err != nil {
				gologger.Errorf("文件写入失败 %v", err)
			}
		} else {
			savaPath += ".xlsx"
			// 导出表格信息
			f := excelize.NewFile()
			gologger.Infof("【%s】导出中\n", ensInfos.Name)
			for k, s := range ensInfos.Infos {
				gologger.Infof("正在导出%s\n", ensMap[k].Name)
				headers := ensMap[k].KeyWord
				var data [][]interface{}
				for _, y := range s {
					var str []interface{}
					results := gjson.GetMany(y.Raw, ensMap[k].Field...)
					for _, t := range results {
						str = append(str, t.String())
					}
					data = append(data, str)
				}
				f, _ = utils.ExportExcel(ensMap[k].Name, headers, data, f)
			}
			f.DeleteSheet("Sheet1")
			if err := f.SaveAs(savaPath); err != nil {
				gologger.Fatalf("表格导出失败：%s", err)
			}
		}
		gologger.Infof("导出成功路径： %s\n", savaPath)
	} else {
		gologger.Errorf("无法导出，公司名不存在")
	}
}
