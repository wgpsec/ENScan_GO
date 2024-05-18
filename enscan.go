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
	var quitSig = make(chan os.Signal, 1)
	signal.Notify(quitSig, os.Interrupt, os.Kill)
	go func() {
		for {
			select {
			case <-quitSig:
				gologger.Error().Msgf("任务未完成退出，自动保存过程文件！")
				enDataList := make(map[string][]map[string]string)
				close(runner.EnCh)
				for ch := range runner.EnCh {
					utils.MergeMap(ch, enDataList)
				}
				err := common.OutFileByEnInfo(enDataList, "意外退出保存文件", "xlsx", "outs")
				if err != nil {
					gologger.Error().Msgf(err.Error())
				}
				log.Fatal("exit.by.signal")
			}
		}
	}()

	var enOptions common.ENOptions
	common.Flag(&enOptions)
	common.Parse(&enOptions)
	runner.RunEnumeration(&enOptions)

}
