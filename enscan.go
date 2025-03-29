package main

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/runner"
	"log"
	"os"
	"os/signal"
)

func main() {
	var enOptions common.ENOptions
	common.Flag(&enOptions)
	common.Parse(&enOptions)
	var quitSig = make(chan os.Signal, 1)
	signal.Notify(quitSig, os.Interrupt, os.Kill)
	go func() {
		for {
			select {
			case <-quitSig:
				if !enOptions.IsApiMode && !enOptions.IsMCPServer {
					gologger.Error().Msgf("任务未完成退出，将自动保存过程文件！")
					enDataList := make(map[string][]map[string]string)
					close(runner.EnCh)
					if len(runner.EnCh) > 0 {
						for ch := range runner.EnCh {
							utils.MergeMap(ch, enDataList)
						}
						err := common.OutFileByEnInfo(enDataList, "进程退出保存文件", "xlsx", "outs")
						if err != nil {
							gologger.Error().Msgf(err.Error())
						}
					}
				}
				log.Fatal("exit.by.signal")
			}
		}
	}()
	runner.RunEnumeration(&enOptions)

}
