package main

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/runner"
)

func main() {
	var enOptions common.ENOptions
	common.Flag(&enOptions)
	common.Parse(&enOptions)
	runner.RunEnumeration(&enOptions)
}
