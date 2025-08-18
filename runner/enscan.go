package runner

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
)

// SearchByKeyWord 根据关键词筛选公司
func (j *EnJob) SearchByKeyWord(keyword string) (string, error) {
	enList, err := j.job.AdvanceFilter(keyword)
	enMap := j.job.GetENMap()["enterprise_info"]
	if err != nil {
		gologger.Error().Msg(err.Error())
		return "", err
	}
	gologger.Info().Msgf("关键词：“%s” 查询到 %d 个结果，默认选择第一个 \n", keyword, len(enList))
	//展示结果
	utils.TBS(append(enMap.KeyWord[:3], "PID"), append(enMap.Field[:3], enMap.Field[10]), "企业信息", enList)
	// 选择第一个的PID
	pid := enList[0].Get(enMap.Field[10]).String()
	gologger.Debug().Str("PID", pid).Msgf("搜索")
	return pid, nil
}

// getInfoById 根据查询的ID查询公司信息，并关联企业
func (j *EnJob) getInfoById(pid string, searchList []string) error {
	if pid == "" {
		gologger.Error().Msgf("获取PID为空！")
		return fmt.Errorf("获取PID为空")
	}
	enMap := j.job.GetENMap()
	// 基本信息获取
	enInfo := j.getCompanyInfoById(pid, searchList, "")
	enName := enInfo["enterprise_info"][0].Get(enMap["enterprise_info"].Field[0]).String()
	// 初始化需要深度获取的字段关系
	var dps []common.DPS
	for _, sk := range searchList {
		// 关联所有DPS数据信息
		dps = append(dps, j.getDPS(enInfo[sk], enName, 1, enMap, sk)...)
	}
	if len(dps) == 0 {
		j.closeCH()
		return nil
	}
	gologger.Info().Msgf("DPS长度：%d", len(dps))
	j.newTaskQueue(len(dps) * 2)
	j.StartWorkers()
	for _, dp := range dps {
		j.AddTask(DeepSearchTask{
			DPS:        dp,
			SearchList: searchList,
		})
	}
	j.wg.Wait()
	j.closeCH()
	return nil
}

// getCompanyInfoById 获取公司的详细的信息，包含下属列表信息
func (j *EnJob) getCompanyInfoById(pid string, searchList []string, ref string) map[string][]gjson.Result {
	enData := make(map[string][]gjson.Result)
	// 获取公司基本信息
	res, enMap := j.job.GetCompanyBaseInfoById(pid)
	gologger.Info().Msgf("正在获取⌈%s⌋信息", res.Get(j.job.GetENMap()["enterprise_info"].Field[0]))
	format, err := dataFormat(res, "enterprise_info", ref)
	if err != nil {
		gologger.Error().Msgf("格式化数据失败: %v", err)
	}
	enData["enterprise_info"] = append(enData["enterprise_info"], format)
	// 适配风鸟
	if res.Get("orderNo").String() != "" {
		pid = res.Get("orderNo").String()
	}
	// 批量获取信息
	for _, sk := range searchList {
		// 不支持这个搜索类型就跳过去
		if _, ok := enMap[sk]; !ok {
			continue
		}
		s := enMap[sk]
		// 没有这个数据就跳过去，提高速度
		if s.Total <= 0 || s.Api == "" {
			gologger.Debug().Str("type", sk).Msgf("判定 ⌈%s⌋ 为空，自动跳过获取\n数量：%d\nAPI：%s", s.Name, s.Total, s.Api)
			continue
		}
		listData, err := j.getInfoList(pid, s, sk, ref)
		if err != nil {
			gologger.Error().Msgf("尝试获取⌈%s⌋发生异常\n%v", s.Name, err)
			continue
		}
		enData[sk] = append(enData[sk], listData...)
	}

	// 实时存入一个查询的完整信息
	j.dataCh <- enData
	return enData
}

// getInfoList 获取列表信息
func (j *EnJob) getInfoList(pid string, em *common.EnsGo, sk string, ref string) (resData []gjson.Result, err error) {
	var listData []gjson.Result
	gologger.Info().Msgf("正在获取 ⌈%s⌋\n", em.Name)
	data, err := j.getInfoPage(pid, 1, em)
	if err != nil {
		// 如果第一页获取失败，就不继续了，判断直接失败
		return resData, err
	}
	// 如果一页能获取完就不翻页了
	if data.Size < data.Total && data.Size > 0 {
		pages := int((data.Total + data.Size - 1) / data.Size)
		for i := 2; i <= pages; i++ {
			gologger.Info().Msgf("正在获取 ⌈%s⌋ 第⌈%d/%d⌋页\n", em.Name, i, pages)
			d, e := j.getInfoPage(pid, i, em)
			if e != nil {
				// TODO 这里后续考虑加入重试机制，或者是等任务跑完可以再次尝试
				gologger.Error().Msgf("GET ⌈%s⌋ 第⌈%d⌋页失败\n", em.Name, i)
				continue
			}
			listData = append(listData, d.Data...)
		}
	}
	if len(listData) == 0 {
		return resData, err
	}
	for _, y := range listData {
		d, e := dataFormat(y, sk, ref)
		if e != nil {
			gologger.Error().Msgf("格式化数据失败: %v", err)
		}
		resData = append(resData, d)
	}

	// 展示数据
	utils.TBS(em.KeyWord, em.Field, em.Name, resData)
	return resData, err
}

// processTask 处理企业关系任务，进行关联查询
func (j *EnJob) processTask(task DeepSearchTask) {
	gologger.Info().Msgf("【%d,%d】正在获取⌈%s⌋信息，关联原因 %s", j.processed, j.total, task.Name, task.Ref)
	data := j.getCompanyInfoById(task.Pid, task.SearchList, task.Ref)
	j.wg.Done()
	// 如果已经到了对应层级就不需要跑了
	if task.Deep >= j.job.GetEnsD().Op.Deep {
		return
	}
	gologger.Info().Msgf("⌈%s⌋深度搜索到第⌈%d⌋层", task.Name, task.Deep)
	// 根据返回结果继续筛选进行关联
	dps := j.getDPS(data[task.SK], task.Name, task.Deep+1, j.job.GetENMap(), task.SK)
	for _, dp := range dps {
		j.AddTask(DeepSearchTask{
			dp,
			task.SearchList,
		})
	}
}

// getDPS 根据List和规则筛选需要深度搜索的规则
func (j *EnJob) getDPS(list []gjson.Result, ref string, deep int, ems map[string]*common.EnsGo, sk string) (dpList []common.DPS) {
	op := j.job.GetEnsD().Op
	if op.Deep <= 1 {
		return
	}
	// 前置判断
	if len(list) == 0 {
		return
	}
	// 跳过非深度搜索字段
	if !utils.IsInList(sk, common.DeepSearch) {
		return
	}
	// 初始化判断变量

	em := ems[sk]
	nPid := em.Field[len(em.Field)-2]
	nEnName := em.Field[0]
	nStatus := em.Field[2]
	nScale := em.Field[3]
	// 初始化关联原因
	association := fmt.Sprintf("%s %s", em.Name, ref)
	gologger.Info().Msgf("%s", association)
	// 增加数据，该类型下的全部企业数据
	for _, r := range list {
		// 如果不深度搜索分支机构就跳过
		if sk == "branch" && !op.IsSearchBranch {
			continue
		}
		tName := r.Get(nEnName).String()
		// 正则过滤器匹配
		if op.NameFilterRegexp != nil && op.NameFilterRegexp.MatchString(tName) {
			gologger.Info().Msgf("根据过滤器跳过 [%s]", tName)
			continue
		}
		// 判断企业状态
		tStatus := r.Get(nStatus).String()
		if utils.IsInList(tStatus, common.AbnormalStatus) {
			gologger.Info().Msgf("根据状态跳过[%s]的企业[%s] ", tStatus, tName)
			continue
		}
		// 判断计算投资比例
		if sk == "invest" {
			investNum := utils.FormatInvest(r.Get(nScale).String())
			if investNum < op.InvestNum {
				continue
			}
			association = fmt.Sprintf("%s ⌈%d⌋级投资⌈%.2f%%⌋-%s", tName, deep, investNum, ref)
		}
		// 关联返回企业的DPS数据
		gologger.Debug().Msgf("关联⌈%s⌋企业信息，关联原因 %s", tName, association)
		pid := r.Get(nPid).String()
		dpList = append(dpList, common.DPS{
			Name: tName,
			Pid:  pid,
			Ref:  association,
			Deep: deep,
			SK:   sk,
		})
	}
	return dpList
}

func (q *ESJob) InfoToMap(j *EnJob, ref string) (res map[string][]map[string]string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	gologger.Debug().Msgf("InfoToMap\nReceived data: %v\n", j.data)
	res = common.InfoToMap(j.data, j.getENMap(), ref)
	j.data = map[string][]gjson.Result{}
	return res
}

// ListDataFormat 对list数据进行格式化处理
func dataFormat(data gjson.Result, typ string, ref string) (res gjson.Result, e error) {
	// 对这条数据加入关联原因，为什么会关联到这个数
	valueTmp, err := sjson.Set(data.Raw, "ref", ref)
	if err != nil {
		return res, err
	}
	// 对其他特殊渠道信息进行处理
	if typ == "icp" {
		// TODO 需要完善下针对多个域名在一行的情况，根据，分隔
	}
	// 处理完成恢复返回
	res = gjson.Parse(valueTmp)

	return res, nil
}
