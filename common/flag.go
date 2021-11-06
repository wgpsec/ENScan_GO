package common

import (
	"flag"
	"github.com/wgpsec/ENScan/common/gologger"
)

const banner = `

███████╗███╗   ██╗███████╗ ██████╗ █████╗ ███╗   ██╗    
██╔════╝████╗  ██║██╔════╝██╔════╝██╔══██╗████╗  ██║    
█████╗  ██╔██╗ ██║███████╗██║     ███████║██╔██╗ ██║    
██╔══╝  ██║╚██╗██║╚════██║██║     ██╔══██║██║╚██╗██║    
███████╗██║ ╚████║███████║╚██████╗██║  ██║██║ ╚████║    
╚══════╝╚═╝  ╚═══╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝    

							v`

// Version is the current version of C
const Version = `0.0.1`

type Options struct {
	KeyWord   string
	CompanyID string // Target is a single URL/Domain to scan usng a template
	InputFile string // Targets specifies the targets to scan using templates.
}

func ParseOptions() *Options {
	options := &Options{}

	flag.StringVar(&options.KeyWord, "n", "", "公司名称")
	flag.StringVar(&options.CompanyID, "i", "", "公司ID号码")
	flag.StringVar(&options.InputFile, "f", "", "包含公司ID号码的文件")
	flag.Parse()
	showBanner()

	return options
}

func (options *Options) configureOutput() {
	// If the user desires verbose
	//output, show verbose output
	//if options.Verbose {
	gologger.MaxLevel = gologger.Verbose
	//}
	//if options.NoColor {
	//	gologger.UseColors = false
	//}
	//if options.Silent {
	//	gologger.MaxLevel = gologger.Silent
	//}
}

func showBanner() {
	gologger.Printf("%s%s\n", banner, Version)
	gologger.Printf("\t\thttps://github.com/wgpsec/ENScan\n\n")
	gologger.Labelf("请谨慎使用,您应对自己的行为负责\n")
	gologger.Labelf("开发人员不承担任何责任，也不对任何滥用或损坏负责.\n")
}
