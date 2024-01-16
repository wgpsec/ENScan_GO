package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adjust/rmq/v4"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"github.com/wgpsec/ENScan/db"
	"github.com/wgpsec/ENScan/internal/aiqicha"
	"github.com/wgpsec/ENScan/internal/app/aldzs"
	"github.com/wgpsec/ENScan/internal/app/coolapk"
	"github.com/wgpsec/ENScan/internal/app/qimai"
	"github.com/wgpsec/ENScan/internal/chinaz"
	"github.com/wgpsec/ENScan/internal/tianyancha"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/net/context"
	"sync"
	"time"
)

// InfoDto web请求指定数据
type InfoDto struct {
	SearchList  []string
	OrgName     string
	RegCode     string
	Pid         string
	SearchType  []string
	InvestNum   int
	Depth       int
	Relate      string
	IsGetNext   bool
	IsDuplicate bool
	DuMap       map[string]map[string]bool
}

type TaskConsumer struct {
	QueueName string
}

// RunEnumeration 普通任务命令行模式，可批量导入文件查询
func RunEnumeration(options *common.ENOptions) {
	if options.InputFile != "" {
		res := utils.ReadFile(options.InputFile)
		if !options.IsMerge && !options.IsMergeOut {
			gologger.Infof("批量查询模式，自动开启合并，取消 --no-merge\n")
			options.IsMergeOut = true
		}
		time.Sleep(5 * time.Second)

		for k, v := range res {
			if v == "" {
				gologger.Errorf("【第%d条】关键词为空，自动跳过\n", k+1)
				continue
			}
			gologger.Infof("\n====================\n【第%d条】关键词 %s 查询中\n====================\n", k+1, v)
			if options.ISKeyPid {
				options.CompanyID = v
			} else {
				options.CompanyID = ""
				options.KeyWord = v
			}

			RunJob(options)
		}
		if options.IsMergeOut {
			outputfile.OutPutExcelByMergeEnInfo(options)
		}
	} else {
		RunJob(options)
	}
}

// Consume 任务队列消费者，调用启动任务
func (consumer *TaskConsumer) Consume(delivery rmq.Delivery) {
	var task *common.ENOptions
	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		// handle json error
		gologger.Errorf("JSON数据识别失败 %s \n", err)
		if err := delivery.Reject(); err != nil {
			gologger.Errorf("reject error %s \n", err)
		}
		return
	}
	gologger.Infof("任务接受成功【%s】查询中\n", task.KeyWord)
	common.Parse(task)
	RunJob(task)
	if err := delivery.Ack(); err != nil {
		// handle ack error
		gologger.Errorf("ack error %s \n", err)
	}
}

// Worker 注册任务消费者
func Worker(options *common.ENOptions) {
	if options.ClientMode != "" {
		taskQueue, err := db.RmqC.OpenQueue(options.ClientMode)
		err = taskQueue.StartConsuming(1, time.Second)
		taskConsumer := &TaskConsumer{}
		name, err := taskQueue.AddConsumer("task-consumer", taskConsumer)
		gologger.Infof("消费者注册成功Queue:【%s】Client:【%s】 \n", options.ClientMode, name)
		if err != nil {
			panic(err)
		}
	}
}

// SearchGo 查询入口判断
func SearchGo(types string, options *common.ENOptions) string {
	switch types {
	case "aqc":
		return aiqicha.SearchByName(options)
	case "tyc":
		return tianyancha.SearchByName(options)
	//case "qcc":
	//	return qcc.SearchByName(options)
	//case "xlb":
	//	return xiaolanben.SearchByName(options)
	default:
		return ""
	}
}

// SearchByKeyword 根据关键词查询，模糊搜索
func SearchByKeyword(keyword string, types string, options *common.ENOptions) string {
	if keyword != "" {
		options.KeyWord = keyword
		var keyList []string
		if types == "all" {
			for _, v := range []string{"aqc", "tyc", "qcc", "xlb"} {
				keyList = append(keyList, utils.DName(SearchGo(v, options)))
			}
		} else {
			keyList = append(keyList, utils.DName(SearchGo(types, options)))
		}

		keyList = utils.SetStr(keyList)
		if len(keyList) == 1 {
			return keyList[0]
		}
	}
	return ""
}

// RunJob 运行项目 添加新参数记得去Config添加
func RunJob(options *common.ENOptions) {
	if options.Proxy != "" {
		gologger.Infof("代理地址: %s\n", options.Proxy)
	}

	gologger.Infof("关键词:【%s|%s】数据源：%s 数据字段：%s\n", options.KeyWord, options.CompanyID, options.GetType, options.GetField)

	var wg sync.WaitGroup

	//爱企查
	if utils.IsInList("aqc", options.GetType) {
		if options.CompanyID == "" || (options.CompanyID != "" && utils.CheckPid(options.CompanyID) == "aqc") {
			wg.Add(1)
			go func() {
				//defer func() {
				//	if x := recover(); x != nil {
				//		gologger.Errorf("[QCC] ERROR: %v", x)
				//		wg.Done()
				//	}
				//}()
				//查询企业信息
				res, ensOutMap := aiqicha.GetEnInfoByPid(options)
				if options.IsMergeOut {
					//合并导出
					outputfile.MergeOutPut(res, ensOutMap, "爱企查", options)
				} else {
					//单独导出
					outputfile.OutPutExcelByEnInfo(res, ensOutMap, options)
				}
				//hook.BiuScan(res, options)
				wg.Done()
			}()
		}
	}

	//天眼查
	if utils.IsInList("tyc", options.GetType) {
		if options.CompanyID == "" || (options.CompanyID != "" && utils.CheckPid(options.CompanyID) == "tyc") {
			wg.Add(1)
			if options.ENConfig.Cookies.Tianyancha == "" || options.ENConfig.Cookies.Tycid == "" {
				gologger.Fatalf("【TYC】MUST LOGIN 请在配置文件补充天眼查COOKIE和tycId\n")
			}
			go func() {
				defer func() {
					if x := recover(); x != nil {
						gologger.Errorf("[TYC] ERROR: %v", x)
						wg.Done()
					}
				}()
				res, ensOutMap := tianyancha.GetEnInfoByPid(options)
				if options.IsMergeOut {
					outputfile.MergeOutPut(res, ensOutMap, "天眼查", options)
				} else {
					outputfile.OutPutExcelByEnInfo(res, ensOutMap, options)
				}
				//hook.BiuScan(res, options)
				wg.Done()
			}()
		}
	}

	// coolapk酷安应用市场查询
	if utils.IsInList("coolapk", options.GetType) {
		wg.Add(1)
		go func() {
			//defer func() {
			//	if x := recover(); x != nil {
			//		gologger.Errorf("[QCC] ERROR: %v", x)
			//		wg.Done()
			//	}
			//}()
			res, ensOutMap := coolapk.GetReq(options)
			if options.IsMergeOut {
				outputfile.MergeOutPut(res, ensOutMap, "酷安", options)
			} else {
				outputfile.OutPutExcelByEnInfo(res, ensOutMap, options)

			}
			wg.Done()
		}()
	}

	// ChinaZ查询
	if utils.IsInList("chinaz", options.GetType) {
		wg.Add(1)
		go func() {
			//defer func() {
			//	if x := recover(); x != nil {
			//		gologger.Errorf("[QCC] ERROR: %v", x)
			//		wg.Done()
			//	}
			//}()
			res, ensOutMap := chinaz.GetEnInfoByPid(options)
			if options.IsMergeOut {
				outputfile.MergeOutPut(res, ensOutMap, "站长之家", options)
			} else {
				outputfile.OutPutExcelByEnInfo(res, ensOutMap, options)
			}
			wg.Done()
		}()
	}

	// 七麦数据
	if utils.IsInList("qimai", options.GetType) {
		wg.Add(1)
		go func() {
			//defer func() {
			//	if x := recover(); x != nil {
			//		gologger.Errorf("[QCC] ERROR: %v", x)
			//		wg.Done()
			//	}
			//}()
			res, ensOutMap := qimai.GetInfoByKeyword(options)
			outputfile.OutPutExcelByEnInfo(res, ensOutMap, options)
			wg.Done()
		}()
	}

	// 微信小程序查询
	if utils.IsInList("aldzs", options.GetType) {
		wg.Add(1)
		options.CookieInfo = options.ENConfig.Cookies.Aldzs
		res, ensOutMap := aldzs.GetInfoByKeyword(options)
		if options.IsMergeOut {
			outputfile.MergeOutPut(res, ensOutMap, "阿拉丁指数", options)
		} else {
			outputfile.OutPutExcelByEnInfo(res, ensOutMap, options)
		}
		wg.Done()
	}

	wg.Wait()

	if !options.IsOnline {
		if options.IsWebMode {
			outputfile.OutPutXDBByMergeEnInfo(options)
		} else if options.IsMergeOut && options.InputFile == "" && !options.IsApiMode {
			// 如果不是API模式，而且不是批量文件形式查询 不是API 就合并导出到表格里面
			outputfile.OutPutExcelByMergeEnInfo(options)
		} else if options.IsApiMode {
			outputfile.OutPutJsonByMergeEnInfo(options)
		}
	}
}

// AddTask 添加扫描任务信息到任务队列
func AddTask(options *common.ENOptions) error {
	if res, _ := db.RedisBb.Get("ENScanK:" + options.KeyWord).Result(); options.KeyWord != "" && res != "" {
		return errors.New(fmt.Sprintf("关键词%s任务已存在 %s", options.KeyWord, res))
	}
	if r2, _ := db.RedisBb.Get("ENScanK:" + options.CompanyID).Result(); options.CompanyID != "" && r2 != "" {
		return errors.New(fmt.Sprintf("PID %s任务已存在 %s", options.KeyWord, r2))
	}

	if options.KeyWord != "" {
		options.CompanyID = ""
	}
	if options.CompanyID != "" {
		r := utils.CheckPid(options.CompanyID)
		if r != "" {
			options.GetType = []string{r}
		} else {
			gologger.Errorf("PID %s %s NOT FOUND\n", options.CompanyID, options.ScanType)
			options.CompanyID = ""
		}
	}
	common.Parse(options)
	gologger.Infof("TASK %s %s %s %s ADD\n", options.KeyWord, options.CompanyID, options.GetType, options.GetField)
	if len(options.GetType) == 0 || !utils.CheckList(options.GetType) {
		return errors.New("未知的查询类型！")
	}

	taskQueue, err := db.RmqC.OpenQueue("tasks")
	taskBytes, err := json.Marshal(options)
	err = taskQueue.PublishBytes(taskBytes)
	if err != nil {
		gologger.Errorf("\n\nOpenQueue err: %s\n\n", err.Error())
		return errors.New("队列添加失败")
	}
	options.GetFlags = ""
	options.GetField = []string{}
	options.GetType = []string{}
	options.ScanType = ""

	//设置redis锁，防止任务重复添加多次
	if options.KeyWord != "" {
		db.RedisBb.Set("ENScanK:"+options.KeyWord, time.Now().Format("2006-01-02 15:04:05"), 30*time.Minute)
	}
	if options.CompanyID != "" {
		db.RedisBb.Set("ENScanK:"+options.CompanyID, time.Now().Format("2006-01-02 15:04:05"), 30*time.Minute)
	}
	return nil
}

// GetInfo 获取信息
func GetInfo(infoDto InfoDto, reEnsList map[string][]map[string]interface{}, depths int, options *common.ENOptions) map[string][]map[string]interface{} {
	InvestCount := 0                     //如果是-1那就没有投资信息
	ENCount := make(map[string][]string) //判断是否有查询过 防止重复查询
	for _, v := range infoDto.SearchList {
		//如果是超过了深度查询，就不再查询
		if (v == "partner" || v == "holds" || v == "branch") && depths > 0 {
			options.IsGetBranch = false
			continue
		}
		if (v == "invest" || v == "supplier") && depths >= infoDto.Depth && infoDto.Depth != 0 {
			continue
		}
		var result common.DBEnInfos
		filter := bson.M{"name": utils.DName(infoDto.OrgName)}
		collection := db.MongoBb.Database("ENScan").Collection(v)
		err := collection.FindOne(context.TODO(), filter).Decode(&result)
		//初始化下查询数据，用于判断是否有数据
		if v == "enterprise_info" {
			InvestCount = result.InvestCount
			ENCount = result.InfoCount
		}
		if err != nil { //没有从数据库里面查到数据，那就要看看是不是要添加下任务
			//gologger.Debugf("NO ENInfo: %s %s %s %d\n", v, infoDto.OrgName, infoDto.Pid, depths)
			if infoDto.OrgName != "" {
				if (v == "invest" || v == "holds") && InvestCount == -1 {
					continue
				}
				//优先根据公司PID去查 ，判断下PID是哪个数据源
				checkPids := ""
				if infoDto.Pid != "" {
					options.CompanyID = infoDto.Pid
					checkPids = utils.CheckPid(options.CompanyID)
				} else {
					options.KeyWord = infoDto.OrgName
				}
				//判断下查询数据 如果说这个数据已经查过，就把他去掉不要查
				var tmp []string
				if checkPids == "" {
					for _, vss := range infoDto.SearchType {
						if !utils.IsInList(vss, ENCount[v]) {
							tmp = append(tmp, vss)
						}
					}
				} else {
					if !utils.IsInList(checkPids, ENCount[v]) {
						tmp = append(tmp, checkPids)
					}
				}
				if v == "branch" {
					tmp = utils.DelInList("xlb", tmp)
				}
				if len(tmp) > 0 {
					options.GetField = append(options.GetField, v)
					options.GetType = tmp
					options.ScanType = ""
					if utils.CheckList(options.GetType) {
						gologger.Debugf("NE Search ADD: %s %s %s %d\n", v, infoDto.OrgName, infoDto.Pid, depths)
						_ = AddTask(options)
					}
				}
			}

		} else {
			// 如果查询到数据开始遍历处理 todo 【需要优化】会造成大量数据查询，待优化
			for _, v2 := range result.Info {

				//数据去重，从数据库遍历里面加进去判断，如果数据已经有了就给去掉
				if infoDto.IsDuplicate {
					if infoDto.DuMap[v] == nil {
						infoDto.DuMap[v] = make(map[string]bool)
					}
					ots := outputfile.ENSMapLN[v].JField[0]
					if outputfile.ENSMapLN[v].Only != "" {
						ots = outputfile.ENSMapLN[v].Only
					}
					if !infoDto.DuMap[v][utils.DName(v2[ots].(string))] {
						infoDto.DuMap[v][utils.DName(v2[ots].(string))] = true
					} else {
						continue
					}
				}

				//如果投资小于设定值 就不要继续查了
				if v == "invest" {
					if int(v2["scale"].(float64)) < infoDto.InvestNum {
						continue
					}
				}

				//判断数据是否需要下钻，数据量一旦大也会卡 22.5.16 优化了下，判断合集数量
				if (v == "invest" || v == "partner") && infoDto.IsGetNext {
					cos := db.MongoBb.Database("ENScan").Collection(v)
					if res, errCos := cos.CountDocuments(context.TODO(), bson.M{"name": utils.DName(v2["name"].(string))}); res > 0 && errCos == nil {
						v2["is_drill"] = true
					} else {
						v2["is_drill"] = false
					}
				}

				//数据信息关联，为什么会把这个数据拿进来进行说明
				if infoDto.Relate != "" {
					v2["relate"] = infoDto.Relate
				}
				reEnsList[v] = append(reEnsList[v], v2)

			}

			// 有投资信息，开始往下层查孙公司信息 todo 【需要优化】数据量一大非常非常卡
			if (infoDto.InvestNum != 0 && v == "invest" || v == "holds" || v == "branch" || v == "supplier") && depths < infoDto.Depth && InvestCount != -1 {
				if len(result.Info) > 0 {
					depths++ //遍历深度增加
					//遍历信息，然后查公司
					for _, vv := range result.Info {
						// 计算投资比例信息
						investNum := 0
						if v == "invest" {
							investNum = int(vv["scale"].(float64))
						}
						openStatus := ""
						if v != "supplier" {
							openStatus = vv["status"].(string)
						}

						//判断下投资的额度是不是符合设定，如果是就开始往下面钻
						if (openStatus != "注销" && openStatus != "吊销") && (investNum >= infoDto.InvestNum || v == "holds" || v == "branch" || v == "supplier") {
							var rs common.DBEnInfos
							filters := bson.M{"name": utils.DName(vv["name"].(string))}
							coll := db.MongoBb.Database("ENScan").Collection("enterprise_info")
							errs := coll.FindOne(context.TODO(), filters).Decode(&rs)
							//如果没查到这个公司信息，判断下是否要获取下数据
							if errs != nil {
								gologger.Debugf("【深度搜索】数据为空 %s %s %s %d\n", vv["name"], vv["pid"], v, depths)
								//判断下拿过来的这个数据PID是哪个数据源 如果没法判断就默认直接查关键词
								stype := utils.CheckPid(vv["pid"].(string))
								if stype != "" {
									options.ScanType = stype
									options.CompanyID = vv["pid"].(string)
									options.KeyWord = ""
								} else {
									options.KeyWord = vv["name"].(string)
								}
								//如果是查控股 就不需要再去查子公司的控股信息了
								if v == "holds" || v == "branch" || v == "supplier" {
									options.GetField = utils.DelInList("holds", options.GetField)
									options.GetField = utils.DelInList("branch", options.GetField)
									options.GetField = utils.DelInList("supplier", options.GetField)
									options.IsGetBranch = false
								}
								_ = AddTask(options)
								options.ScanType = ""
							} else {
								//开始下钻查询信息
								beReason := fmt.Sprintf("%s 投资【%d级】占比 %f%%", vv["name"], depths, vv["scale"])
								if v == "holds" {
									beReason = fmt.Sprintf("%s 控股【%d级】占比 %f%%", vv["name"], depths, vv["scale"])
								}
								if v == "branch" {
									beReason = fmt.Sprintf("%s 分支机构【%d级】", vv["name"], depths)
								}
								if v == "supplier" {
									beReason = fmt.Sprintf("%s 供应商【%d级】占比 %f%%", vv["name"], depths, vv["scale"])
								}
								infoDto.Relate = beReason
								infoDto.Pid = vv["pid"].(string)
								infoDto.OrgName = vv["name"].(string)
								GetInfo(infoDto, reEnsList, depths, options)
							}
						}
					}
				}
			}
		}

	}
	return reEnsList
}
