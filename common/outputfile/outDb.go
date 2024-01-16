package outputfile

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"github.com/wgpsec/ENScan/db"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"io"
	"log"
	"time"
)

// OutPutXDBByMergeEnInfo 根据合并数据写入数据库
func OutPutXDBByMergeEnInfo(ENOptions *common.ENOptions) error {
	gologger.Infof("【%s】 数据写入结构数据库中\n", ENOptions.CompanyName)
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

	if ENOptions.IsWebMode {
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

		} else {
			gologger.Errorf("没有查询到任何企业数据，无法完成入库操作！\n")
		}

	}

	EnJsonList = make(map[string][]map[string]interface{})
	EnsInfosList = make(map[string][][]interface{})
	ENSMapList = make(map[string]*ENSMap)

	return nil
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
			gologger.Errorf("导出错误信//息 %s", k)
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
				data = append(data, str)
			}

			var err error
			f, err = utils.ExportExcel(ENSMapLN[k].Name, headers, data, f)
			if err != nil {
				gologger.Errorf(err.Error())
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
