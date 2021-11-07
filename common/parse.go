package common

import (
	"flag"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"os"
)

func Parse(options *ENOptions) {
	if options.Version {
		gologger.Infof("Current Version: %s\n", Version)
		os.Exit(0)
	}
	if options.KeyWord == "" && options.CompanyID == "" && options.InputFile == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

}
