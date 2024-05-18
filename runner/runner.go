package runner

import (
	"bufio"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	_interface "github.com/wgpsec/ENScan/interface"
	"github.com/wgpsec/ENScan/internal/aiqicha"
	"github.com/wgpsec/ENScan/internal/tianyancha"
	"os"
	"sync"
	"time"
)

type EnJob struct {
	info map[string][]gjson.Result
	job  _interface.ENScan
}

var EnCh chan map[string][]map[string]string

// RunEnumeration 普通任务命令行模式，可批量导入文件查询
func RunEnumeration(options *common.ENOptions) {
	if options.InputFile != "" {
		enDataList := make(map[string][]map[string]string)
		res := utils.ReadFileOutLine(options.InputFile)
		gologger.Info().Str("FileName", options.InputFile).Msgf("读取到 %d 条信息", len(res))
		time.Sleep(1 * time.Second)
		var wg sync.WaitGroup
		EnCh = make(chan map[string][]map[string]string, len(res))
		wg.Add(len(res))
		go func() {
			for k, v := range res {
				gologger.Info().Msgf("\n⌈%d/%d⌋ 关键词：⌈%s⌋", k+1, len(res), v)
				if options.ISKeyPid {
					options.CompanyID = v
				} else {
					options.CompanyID = ""
					options.KeyWord = v
				}
				jobRes := RunJob(options)
				utils.MergeMap(jobRes, enDataList)
				EnCh <- jobRes
				wg.Done()
			}
		}()
		go func() {
			if options.UPOutFile != "" {
				file, err := os.OpenFile(options.UPOutFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
				// 防止写入中文乱码
				_, err = file.WriteString("\xEF\xBB\xBF")
				if err != nil {
					gologger.Fatal().Str("写入文件名", options.UPOutFile).Msgf("写入文件操作失败！")
				}
				defer func() {
					if err = file.Close(); err != nil {
						gologger.Error().Msgf("关闭文件时出错: %v", err)
					}
				}()
				writer := bufio.NewWriter(file)
				for {
					select {
					case data := <-EnCh:
						_, err = writer.WriteString(common.OutStrByEnInfo(data, options.GetField[0]))
						err = writer.Flush()
						if err != nil {
							gologger.Error().Msgf("写入数据失败: %v", err)
						}
					}
				}
			}
		}()
		wg.Wait()
		err := common.OutFileByEnInfo(enDataList, options.KeyWord, options.OutPutType, options.Output)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}

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
	var wg sync.WaitGroup
	for _, jobType := range options.GetType {
		job := jobs[jobType]
		wg.Add(1)
		go func(jobType string) {
			defer func() {
				if x := recover(); x != nil {
					gologger.Error().Msgf("⌈%s⌋出现错误", jobType)
					wg.Done()
				}
			}()
			// 搜索关键词id
			pid := ""
			if options.CompanyID != "" {
				pid = options.CompanyID
			} else {
				pid = AdvanceFilter(job)
			}
			if pid != "" {
				data := getInfoById(pid, options.GetField, job)
				rdata := common.InfoToMap(data, job.GetENMap(), fmt.Sprintf("%s⌈%s⌋", jobType, options.KeyWord))
				if options.IsMergeOut {
					utils.MergeMap(rdata, enDataList)
				} else {
					err := common.OutFileByEnInfo(rdata, options.KeyWord, options.OutPutType, options.Output)
					if err != nil {
						gologger.Error().Msgf(err.Error())
					}
				}
			}
			defer wg.Done()
		}(jobType)
	}

	wg.Wait()
	if options.IsMergeOut && options.InputFile == "" {
		err := common.OutFileByEnInfo(enDataList, options.KeyWord, options.OutPutType, options.Output)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}
	}
	return enDataList
}
