package common

import (
	"flag"
	"github.com/wgpsec/ENScan/common/gologger"
)

// mcpFlag is a custom flag type that supports optional values
// When used as --mcp (without value), it will be set to "true"
// When used as --mcp=:8080 or --mcp=http://..., it will be set to that value
type mcpFlag struct {
	serverAddr *string
	serverSet  *bool
}

func (m *mcpFlag) String() string {
	if m.serverAddr == nil {
		return ""
	}
	return *m.serverAddr
}

func (m *mcpFlag) Set(s string) error {
	// Validate pointers before dereferencing
	if m.serverAddr == nil || m.serverSet == nil {
		return nil
	}
	
	// When IsBoolFlag returns true and the flag is used without a value (e.g., --mcp alone),
	// Go's flag parser calls Set with "true" as the value.
	// We treat this special case as "use config default" by setting value to empty string.
	// For any other value (e.g., --mcp=:8080), we use that value directly.
	if s == "true" {
		*m.serverAddr = ""  // Empty means use config default
	} else {
		*m.serverAddr = s
	}
	*m.serverSet = true
	return nil
}

func (m *mcpFlag) IsBoolFlag() bool {
	return true
}

const banner = `

███████╗███╗   ██╗███████╗ ██████╗ █████╗ ███╗   ██╗
██╔════╝████╗  ██║██╔════╝██╔════╝██╔══██╗████╗  ██║
█████╗  ██╔██╗ ██║███████╗██║     ███████║██╔██╗ ██║
██╔══╝  ██║╚██╗██║╚════██║██║     ██╔══██║██║╚██╗██║
███████╗██║ ╚████║███████║╚██████╗██║  ██║██║ ╚████║
╚══════╝╚═╝  ╚═══╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝

`

func Banner() {
	gologger.Print().Msgf("%sBuilt At: %s\nGo Version: %s\nAuthor: %s\nBuild SHA: %s\nVersion: %s\n\n", banner, BuiltAt, GoVersion, GitAuthor, BuildSha, GitTag)
	gologger.Print().Msgf("https://github.com/wgpsec/ENScan_GO\n\n")
	gologger.Print().Msgf("工具仅用于信息收集，请勿用于非法用途\n")
	gologger.Print().Msgf("开发人员不承担任何责任，也不对任何滥用或损坏负责.\n")
}

func Flag(Info *ENOptions) {
	Banner()
	//指定参数
	flag.StringVar(&Info.KeyWord, "n", "", "关键词 eg 小米")
	flag.StringVar(&Info.CompanyID, "i", "", "公司PID")
	flag.StringVar(&Info.InputFile, "f", "", "批量查询，文本按行分隔")
	flag.StringVar(&Info.ScanType, "type", "aqc", "查询渠道，可多选")
	//查询参数指定
	flag.Float64Var(&Info.InvestNum, "invest", 0, "投资比例 eg 100")
	flag.StringVar(&Info.GetFlags, "field", "", "获取字段信息 eg icp")
	flag.IntVar(&Info.Deep, "deep", 1, "递归搜索n层公司")
	flag.BoolVar(&Info.IsHold, "hold", false, "是否查询控股公司")
	flag.BoolVar(&Info.IsSupplier, "supplier", false, "是否查询供应商信息")
	flag.BoolVar(&Info.IsGetBranch, "branch", false, "查询分支机构（分公司）信息")
	flag.BoolVar(&Info.IsSearchBranch, "is-branch", false, "深度查询分支机构信息（数量巨大）")
	flag.BoolVar(&Info.IsJsonOutput, "json", false, "json导出")
	flag.StringVar(&Info.Output, "out-dir", "", "结果输出的文件夹位置(默认为outs)")
	flag.StringVar(&Info.BranchFilter, "branch-filter", "", "提供一个正则表达式，名称匹配该正则的分支机构和子公司会被跳过")
	flag.StringVar(&Info.OutPutType, "out-type", "xlsx", "导出的文件后缀 默认xlsx")
	flag.BoolVar(&Info.IsDebug, "debug", false, "是否显示debug详细信息")
	flag.BoolVar(&Info.IsShow, "is-show", true, "是否展示信息输出")
	flag.BoolVar(&Info.IsFast, "is-fast", false, "跳过数量校验，直接开启查询")
	flag.BoolVar(&Info.IsPlugins, "is-plugin", false, "是否以插件功能运行，默认false")
	//其他设定
	flag.BoolVar(&Info.IsGroup, "is-group", false, "查询关键词为集团")
	flag.BoolVar(&Info.IsApiMode, "api", false, "API模式运行")
	flag.Var(&mcpFlag{serverAddr: &Info.MCPServer, serverSet: &Info.MCPServerSet}, "mcp", "MCP模式运行。用法: --mcp (使用配置文件默认值), --mcp=:8080 (自定义端口), --mcp=http://localhost:8080 (完整地址)")
	flag.BoolVar(&Info.ISKeyPid, "is-pid", false, "批量查询文件是否为公司PID")
	flag.IntVar(&Info.DelayTime, "delay", 0, "每个请求延迟（S）-1为随机延迟1-5S")
	flag.StringVar(&Info.Proxy, "proxy", "", "设置代理")
	flag.IntVar(&Info.TimeOut, "timeout", 1, "每个请求默认1（分钟）超时")
	flag.BoolVar(&Info.IsNoMerge, "no-merge", false, "开启后查询文件将单独导出")
	flag.BoolVar(&Info.Version, "v", false, "版本信息")
	flag.Parse()
}
