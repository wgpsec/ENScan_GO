package main

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/runner"
)

func main() {
	var enOptions common.ENOptions
	common.Flag(&enOptions)
	enOptions.Parse()
	runner.RunEnumeration(&enOptions)
}
