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

`

func Banner() {
	gologger.Printf("%sBuilt At: %s\nGo Version: %s\nAuthor: %s\nBuild SHA: %s\nVersion: %s\n\n", banner, BuiltAt, GoVersion, GitAuthor, BuildSha, GitTag)
	gologger.Printf("\t\thttps://github.com/wgpsec/ENScan\n\n")
	gologger.Labelf("工具仅用于信息收集，请勿用于非法用途\n")
	gologger.Labelf("开发人员不承担任何责任，也不对任何滥用或损坏负责.\n")
}

func Flag(Info *ENOptions) {
	Banner()
	flag.StringVar(&Info.KeyWord, "n", "", "关键词 eg 小米")
	flag.StringVar(&Info.CompanyID, "i", "", "公司PID")
	flag.StringVar(&Info.InputFile, "f", "", "批量查询，文本按行分隔")
	flag.StringVar(&Info.ScanType, "type", "aqc", "API类型 eg qcc")
	flag.StringVar(&Info.Output, "o", "", "结果输出的文件夹位置(可选)")
	flag.BoolVar(&Info.IsMergeOut, "is-merge", false, "合并导出")
	flag.BoolVar(&Info.IsJsonOutput, "json", false, "json导出")
	//查询参数指定
	flag.Float64Var(&Info.InvestNum, "invest", 0, "投资比例 eg 100")
	flag.StringVar(&Info.GetFlags, "field", "", "获取字段信息 eg icp")
	flag.IntVar(&Info.Deep, "deep", 1, "递归搜索n层公司")
	flag.BoolVar(&Info.IsHold, "hold", false, "是否查询控股公司")
	flag.BoolVar(&Info.IsSupplier, "supplier", false, "是否查询供应商信息")
	flag.BoolVar(&Info.IsGetBranch, "branch", false, "查询分支机构（分公司）信息")
	flag.BoolVar(&Info.IsSearchBranch, "is-branch", false, "深度查询分支机构信息（数量巨大）")
	//web api
	flag.BoolVar(&Info.IsWebMode, "web", false, "是否开启web")
	flag.BoolVar(&Info.IsApiMode, "api", false, "是否API模式")
	flag.StringVar(&Info.ClientMode, "client", "", "客户端模式通道 eg: task")
	flag.BoolVar(&Info.IsDebug, "debug", false, "是否显示debug详细信息")
	flag.BoolVar(&Info.IsShow, "is-show", true, "是否展示信息输出")
	//其他设定
	flag.BoolVar(&Info.IsInvestRd, "uncertain-invest", false, "包括未公示投资公司（无法确定占股比例）")
	flag.BoolVar(&Info.IsGroup, "is-group", false, "查询关键词为集团")
	flag.BoolVar(&Info.ISKeyPid, "is-pid", false, "批量查询文件是否为公司PID")
	flag.IntVar(&Info.DelayTime, "delay", 0, "填写最大延迟时间（秒）将会在1-n间随机延迟")
	flag.StringVar(&Info.Proxy, "proxy", "", "设置代理")
	flag.IntVar(&Info.TimeOut, "timeout", 1, "每个请求默认1（分钟）超时")
	flag.BoolVar(&Info.IsMerge, "no-merge", false, "批量查询【取消】合并导出")
	flag.BoolVar(&Info.Version, "v", false, "版本信息")
	flag.BoolVar(&Info.IsEmailPro, "email", false, "获取email信息")
	flag.Parse()
}
