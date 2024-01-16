package runner

import (
	"encoding/json"
	"errors"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"github.com/wgpsec/ENScan/db"
)

// GetWebInfo 获取信息
func GetWebInfo(infoDto InfoDto, reEnsList map[string][]map[string]interface{}, depths int, options *common.ENOptions) map[string][]map[string]interface{} {
	// 直接实时获取输出信息
	if options.IsOnline {
		RunJob(options)
		options.IsJsonOutput = true
		reEnsList = outputfile.OutPutExcelByMergeEnInfo(options)

		return reEnsList
	}
	return reEnsList
}

func WebJob(options *common.ENOptions) error {
	ch := make(chan *common.ENOptions)
	ch <- options
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

	return nil
}

// AddWebTask 添加扫描任务信息到任务队列
func AddWebTask(options *common.ENOptions) error {

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

	return nil
}
