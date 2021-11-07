package runner

import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/internal/aiqicha"
)

func RunEnumeration(options *common.ENOptions) {
	if options.ScanType == "a" {
		aiqicha.GetEnInfoByPid(options)
	}

}
