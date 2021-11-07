package common

import (
	"flag"
	"github.com/wgpsec/ENScan/common/utils/gologger"
)

const banner = `

███████╗███╗   ██╗███████╗ ██████╗ █████╗ ███╗   ██╗    
██╔════╝████╗  ██║██╔════╝██╔════╝██╔══██╗████╗  ██║    
█████╗  ██╔██╗ ██║███████╗██║     ███████║██╔██╗ ██║    
██╔══╝  ██║╚██╗██║╚════██║██║     ██╔══██║██║╚██╗██║    
███████╗██║ ╚████║███████║╚██████╗██║  ██║██║ ╚████║    
╚══════╝╚═╝  ╚═══╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝

                           ENScan version: `

const Version = `0.0.1`

func Banner() {
	gologger.Printf("%s%s\n", banner, Version)
	gologger.Printf("\t\thttps://github.com/wgpsec/ENScan\n\n")
	gologger.Labelf("工具仅用于信息收集，请勿用于非法用途\n")
	gologger.Labelf("开发人员不承担任何责任，也不对任何滥用或损坏负责.\n")
}

func Flag(Info *ENOptions) {
	Banner()
	flag.StringVar(&Info.KeyWord, "n", "", "公司名称")
	flag.StringVar(&Info.CompanyID, "i", "", "公司ID")
	flag.StringVar(&Info.InputFile, "f", "", "包含公司ID的文件")
	flag.StringVar(&Info.CookieInfo, "c", "", "Cookie信息")
	flag.StringVar(&Info.ScanType, "type", "a", "收集API")
	flag.StringVar(&Info.Output, "o", "", "结果输出的文件(可选)")
	flag.BoolVar(&Info.IsGetBranch, "branch", false, "是否拿到分支机构详细信息，为了获取邮箱和人名信息等")
	flag.BoolVar(&Info.IsInvestRd, "invest-rd", false, "是否选出不清楚投资比例的（出现误报较高）")
	flag.IntVar(&Info.InvestNum, "invest-num", 0, "筛选投资比例，默认0为不筛选")
	flag.StringVar(&Info.GetFlags, "flags", "", "获取哪些字段信息")
	flag.BoolVar(&Info.Version, "v", false, "版本信息")
	flag.Parse()
}
