package runner

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	_interface "github.com/wgpsec/ENScan/interface"
	"github.com/wgpsec/ENScan/internal/aiqicha"
	"github.com/wgpsec/ENScan/internal/tianyancha"
	"time"
)

type EnJob struct {
	info map[string][]gjson.Result
	job  _interface.ENScan
}

var TmpData = make(map[string][]gjson.Result)
var CurrJob _interface.ENScan
var CurDone = false

// RunEnumeration 普通任务命令行模式，可批量导入文件查询
func RunEnumeration(options *common.ENOptions) {
	if options.InputFile != "" {
		outFile := "xlsx"
		if options.IsJsonOutput {
			outFile = "json"
		}
		enDataList := make(map[string][]map[string]string)
		if !options.IsMerge && !options.IsMergeOut {
			gologger.Info().Msgf("批量查询已开启，如需单独输出企业请使用 --no-merge 取消合并\n")
			options.IsMergeOut = true
		}
		res := utils.ReadFile(options.InputFile)
		utils.SetStr(res)
		gologger.Info().Str("FileName", options.InputFile).Msgf("读取到 %d 条信息", len(res))
		time.Sleep(5 * time.Second)
		for k, v := range res {
			if v == "" {
				gologger.Error().Msgf("【第%d条】关键词为空，自动跳过\n", k+1)
				continue
			}
			gologger.Info().Msgf("\n⌈%d/%d⌋ 关键词：⌈%s⌋", k+1, len(res), v)
			if options.ISKeyPid {
				options.CompanyID = v
			} else {
				options.CompanyID = ""
				options.KeyWord = v
			}
			utils.MergeMap(RunJob(options), enDataList)
			err := common.OutFileByEnInfo(enDataList, options.KeyWord, outFile, options.Output)
			if err != nil {
				gologger.Error().Msgf(err.Error())
			}
			CurDone = false
		}
		CurDone = true
	} else {
		RunJob(options)
	}
}

// RunJob 运行项目 添加新参数记得去Config添加
func RunJob(options *common.ENOptions) map[string][]map[string]string {
	gologger.Info().Msgf("正在获取 ⌈%s%s⌋ 信息", options.KeyWord, options.CompanyID)
	gologger.Debug().Msgf("关键词：⌈%s⌋ PID：⌈%s⌋ 数据源：%s 数据字段：%s\n", options.KeyWord, options.CompanyID, options.GetType, options.GetField)
	jobs := map[string]_interface.ENScan{
		"aqc": &aiqicha.AQC{Options: options},
		"tyc": &tianyancha.TYC{Options: options},
	}
	enDataList := make(map[string][]map[string]string)
	outFile := "xlsx"
	if options.IsJsonOutput {
		outFile = "json"
	}
	for _, jobType := range options.GetType {
		var enJob EnJob
		job := jobs[jobType]
		enJob.job = job
		CurrJob = job
		// 搜索关键词id
		pid := ""
		if options.CompanyID != "" {
			pid = options.CompanyID
		} else {
			pid = AdvanceFilter(job)
		}
		data := getInfoById(pid, options.GetField, &enJob)
		rdata := common.InfoToMap(data, job.GetENMap(), "数据来源 "+jobType)
		if !options.IsMergeOut {
			err := common.OutFileByEnInfo(rdata, options.KeyWord, outFile, options.Output)
			if err != nil {
				gologger.Error().Msgf(err.Error())
			}
		} else {
			utils.MergeMap(rdata, enDataList)
		}
	}
	if options.IsMergeOut && options.InputFile == "" {
		err := common.OutFileByEnInfo(enDataList, options.KeyWord, outFile, options.Output)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}
	}
	CurDone = true
	return enDataList
}
