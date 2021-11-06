package main

import (
	"github.com/wgpsec/ENScan/internal/aiqicha"
)

func main() {
	//options := common.ParseOptions()

	aiqicha.GetEnInfoByPid(aiqicha.SearchName("小米")[0].Get("pid").String())
}
