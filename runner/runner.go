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
	"github.com/wgpsec/ENScan/internal/app/miit"
	"github.com/wgpsec/ENScan/internal/kuaicha"
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
			// 文件实时写入
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
		if !options.IsNoMerge {
			err := common.OutFileByEnInfo(enDataList, options.KeyWord, options.OutPutType, options.Output)
			if err != nil {
				gologger.Error().Msgf(err.Error())
			}
		}

	} else if options.IsApiMode {
		api(options)
	} else if options.IsMCPServer {
		mcpServer(options)
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
		"kc":  &kuaicha.KC{Options: options},
	}
	apps := map[string]_interface.App{
		"miit": &miit.Miit{Options: options},
	}
	enDataList := make(map[string][]map[string]string)
	var wg sync.WaitGroup
	// 等待全部执行完成再结束
	wg.Add(len(options.GetType))
	for _, jobType := range options.GetType {
		// 跳过插件类型
		if _, ok := jobs[jobType]; !ok {
			continue
		}
		job := jobs[jobType]

		go func(jobType string) {
			defer func() {
				if x := recover(); x != nil {
					gologger.Error().Msgf("⌈%s⌋异常出现错误", jobType)
					gologger.Error().Msgf("%v", x)
					wg.Done()
				}
			}()
			// 搜索关键词id，如果PID为空那么根据关键词去搜索拿到PID
			pid := ""
			if options.CompanyID != "" {
				pid = options.CompanyID
			} else {
				pid = AdvanceFilter(job)
			}
			if pid != "" {
				// 获取企业信息，通过查询到的信息
				data := getInfoById(pid, options.GetField, job)
				// 把获取的公司信息合并统一格式MAP方便后续调用
				rdata := common.InfoToMap(data, job.GetENMap(), fmt.Sprintf("%s⌈%s⌋", jobType, options.KeyWord))

				// 引入插件功能，根据输出的信息来进一步拓展信息
				for _, appType := range options.GetType {
					if _, ok := apps[appType]; !ok {
						continue
					}
					plData := getAppById(rdata, options.GetField, apps[appType])
					plList := common.InfoToMap(plData, apps[appType].GetENMap(), fmt.Sprintf("%s⌈%s⌋", appType, options.KeyWord))
					// 把插件数据合并进去
					utils.MergeMap(plList, rdata)
					// 插件执行完成也需要减数
					wg.Done()
				}

				// 处理完成数据进行导出
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
	if options.IsMergeOut && options.InputFile == "" && !options.IsApiMode && !options.IsMCPServer {
		err := common.OutFileByEnInfo(enDataList, options.KeyWord, options.OutPutType, options.Output)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}
	}
	return enDataList
}
