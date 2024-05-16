package runner

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	_interface "github.com/wgpsec/ENScan/interface"
)

func AdvanceFilter(job _interface.ENScan) string {
	enList, err := job.AdvanceFilter()
	enMap := job.GetENMap()["enterprise_info"]
	if err != nil {
		gologger.Error().Msg(err.Error())
	} else {
		gologger.Info().Msgf("关键词：“%s” 查询到 %d 个结果，默认选择第一个 \n", job.GetEnsD().Name, len(enList))
		//展示结果
		utils.TBS(append(enMap.KeyWord[:3], "PID"), append(enMap.Field[:3], enMap.Field[10]), enList)
		pid := enList[0].Get(enMap.Field[10]).String()
		gologger.Debug().Str("PID", pid).Msgf("搜索")
		return pid
	}
	return ""
}

func getInfoById(pid string, searchList []string, enJob *EnJob) (enInfo map[string][]gjson.Result) {
	job := enJob.job
	enMap := job.GetENMap()
	options := job.GetEnsD().Op
	// 基本信息获取
	enInfo = getCompanyInfoById(pid, "", searchList, job)
	enName := enInfo["enterprise_info"][0].Get(enMap["enterprise_info"].Field[0]).String()
	var ds []string
	for _, s := range searchList {
		if utils.IsInList(s, common.DeepSearch) {
			ds = append(ds, s)
		}
	}
	if len(ds) > 0 {
		gologger.Info().Msgf("深度搜索列表：%v", ds)
	}
	for _, sk := range ds {
		enSk := enMap[sk].Field
		pidName := enSk[len(enSk)-2]
		scaleName := enSk[3]
		association := enMap[sk].Name
		if len(enInfo[sk]) == 0 {
			gologger.Info().Str("type", sk).Msgf("【x】%s 数量为空，跳过搜索\n", association)
			continue
		}
		// 跳过分支机构搜索
		if sk == "branch" && !options.IsSearchBranch {
			continue
		}

		if sk == "invest" {
			dEnData := make(map[string][]gjson.Result)
			var iEnData [][]gjson.Result
			iEnData = append(iEnData, make([]gjson.Result, 0))
			// 投资信息赋值
			iEnData[0] = enInfo[sk]
			for i := 0; i < options.Deep; i++ {
				if len(iEnData[i]) <= 0 {
					break
				}
				var nextInK []gjson.Result
				for _, r := range iEnData[i] {
					tPid := r.Get(pidName).String()
					gologger.Debug().Str("PID", tPid).Str("PID NAME", pidName).Msgf("查询PID")
					// 计算投资比例判断是否符合
					investNum := utils.FormatInvest(r.Get(scaleName).String())
					if investNum <= options.InvestNum {
						continue
					}
					association = fmt.Sprintf("%s %d级 投资 %.2f", enName, i, investNum)
					gologger.Info().Msgf("%s", association)
					dEnData = getCompanyInfoById(tPid, association, searchList, job)
					// 保存当前数据
					for _, dr := range dEnData {
						enInfo[sk] = append(enJob.info[sk], dr...)
						// 全局存储，这里为了在过程中断时能存下来数据，所以临时存一份
						// 后续替换为 channel
						TmpData[sk] = append(TmpData[sk], dr...)
					}
					// 存下一层需要跑的信息
					nextInK = append(nextInK, dEnData[sk]...)
				}
				iEnData[i+1] = nextInK
			}

		} else {
			association = fmt.Sprintf("%s %s", enName, enMap[sk].KeyWord)
			gologger.Info().Msgf("%s", association)
			// 增加数据，该类型下的全部企业数据
			for _, r := range enInfo[sk] {
				tPid := r.Get(pidName).String()
				dEnData := make(map[string][]gjson.Result)
				dEnData = getCompanyInfoById(tPid, association, searchList, job)
				// 把查询完的一个企业按类别存起来
				for _, dr := range dEnData {
					enInfo[sk] = append(enInfo[sk], dr...)
					// 全局存储，这里为了在过程中断时能存下来数据，所以临时存一份
					// 后续替换为 channel
					TmpData[sk] = append(TmpData[sk], dr...)
				}
			}
		}
	}

	return enJob.info
}

func getCompanyInfoById(pid string, inFrom string, searchList []string, job _interface.ENScan) map[string][]gjson.Result {
	enData := make(map[string][]gjson.Result)
	res, enMap := job.GetCompanyBaseInfoById(pid)
	gologger.Info().Msgf("查询⌈%s⌋信息", res.Get(job.GetENMap()["enterprise_info"].Field[0]))
	// 增加企业信息
	enJsonTMP, _ := sjson.Set(res.Raw, "inFrom", inFrom)
	enData["enterprise_info"] = append(enData["enterprise_info"], gjson.Parse(enJsonTMP))
	// 批量获取信息
	for _, sk := range searchList {
		s := enMap[sk]
		// 不支持这个搜索类型就跳过去
		if _, ok := enMap[sk]; !ok {
			continue
		}
		gologger.Debug().Msgf("%d %s", s.Total, s.Name)
		// 没有这个数据就跳过去，提高速度
		if s.Total <= 0 || s.Api == "" {
			gologger.Info().Str("type", sk).Msgf("【X】%s 数量为空，跳过API\n", s.Name)
			continue
		}

		// 判断结束调用获取数据接口
		listData, err := job.GetEnInfoList(pid, enMap[sk])
		if err != nil {
			gologger.Error().Msg(err.Error())
		}

		// 添加来源信息，并把信息存储到数据里面
		for _, y := range listData {
			valueTmp, _ := sjson.Set(y.Raw, "inFrom", inFrom)
			gs := gjson.Parse(valueTmp)
			enData[sk] = append(enData[sk], gs)
			// 全局存储，这里为了在过程中断时能存下来数据，所以临时存一份
			// 后续替换为 channel
			TmpData[sk] = append(TmpData[sk], gs)
		}
		// 展示数据
		utils.TBS(s.KeyWord, s.Field, listData)
	}
	return enData
}
