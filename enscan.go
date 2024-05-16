package main

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
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
				if !runner.CurDone {
					gologger.Error().Msgf("任务未完成退出，自动保存文件！数据长度 %d", len(runner.TmpData))
					rdata := common.InfoToMap(runner.TmpData, runner.CurrJob.GetENMap(), "")
					err := common.OutFileByEnInfo(rdata, "意外退出保存文件", "xlsx", "outs")
					if err != nil {
						gologger.Error().Msgf(err.Error())
					}
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
