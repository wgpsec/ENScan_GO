package common

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"regexp"
	"strings"
)

func (op *ENOptions) Parse() {
	//DEBUG模式设定
	if op.IsDebug {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelDebug)
		gin.SetMode(gin.DebugMode)
		gologger.Debug().Msgf("DEBUG 模式已开启\n")
	}

	//判断版本信息
	if op.Version {
		gologger.Info().Msgf("Current Version: %s\n", GitTag)
		gologger.Info().Msgf("当前所需配置文件版本 V%.1f\n", cfgYV)
		if ok, _ := utils.PathExists(cfgYName); !ok {
			f, errs := os.Create(cfgYName) //创建文件
			_, errs = io.WriteString(f, configYaml)
			if errs != nil {
				gologger.Fatal().Msgf("配置文件创建失败 %s\n", errs)
			}
			gologger.Info().Msgf("配置文件生成成功！\n")
		}
		os.Exit(0)
	}

	// 配置文件检查
	if ok, _ := utils.PathExists(cfgYName); !ok {
		gologger.Fatal().Msgf("没有找到配置文件 %s 请先运行 -v 创建\n", cfgYName)
	}

	//加载配置信息~
	conf := new(ENConfig)
	yamlFile, err := os.ReadFile(cfgYName)
	if err != nil {
		gologger.Fatal().Msgf("配置文件解析错误 #%v ", err)
	}
	if err := yaml.Unmarshal(yamlFile, conf); err != nil {
		gologger.Fatal().Msgf("【配置文件加载失败】: %v", err)
	}
	if conf.Version < cfgYV {
		gologger.Fatal().Msgf("配置文件当前[V%.1f] 程序需要[V%.1f] 不匹配，请备份配置文件重新运行-v\n", conf.Version, cfgYV)
	}

	if op.KeyWord == "" && op.CompanyID == "" && op.InputFile == "" && !op.IsApiMode && !op.IsMCPServer {
		flag.PrintDefaults()
		os.Exit(0)
	}

	//初始化输出文件夹位置
	if op.Output == "" {
		op.Output = "outs"
	}

	if op.IsJsonOutput {
		op.OutPutType = "json"
	}

	if op.Output == "!" {
		gologger.Info().Msgf("当前模式不会导出文件信息！\n")
	}

	if op.Proxy != "" {
		gologger.Info().Msgf("代理已设定 ⌈%s⌋\n", op.Proxy)
	}

	if op.InputFile != "" {
		if ok := utils.FileExists(op.InputFile); !ok {
			gologger.Fatal().Msgf("未获取到文件⌈%s⌋请检查文件名是否正确\n", op.InputFile)
		}
	}
	if op.IsNoMerge {
		gologger.Info().Msgf("批量查询文件将单独导出！\n")
	}
	op.IsMergeOut = !op.IsNoMerge
	op.ENConfig = conf
	//数据源判断 默认为爱企查
	if op.ScanType == "" && len(op.GetType) == 0 {
		op.ScanType = "aqc"
	}

	//如果是指定全部数据
	if op.ScanType == "all" {
		op.GetType = ENSTypes
		op.IsMergeOut = true
	} else if op.ScanType != "" {
		op.GetType = strings.Split(op.ScanType, ",")
	}

	op.GetType = utils.SetStr(op.GetType)
	var tmp []string
	for _, v := range op.GetType {
		if _, ok := ScanTypeKeys[v]; !ok {
			gologger.Error().Msgf("没有这个%s查询方式\n支持列表\n%s", v, ScanTypeKeys)
		} else {
			tmp = append(tmp, v)
		}
	}
	op.GetType = tmp

	// 判断获取数据字段信息
	op.GetField = utils.SetStr(op.GetField)
	if op.GetFlags == "" && len(op.GetField) == 0 {
		op.GetField = DefaultInfos
	} else if op.GetFlags == "all" {
		op.GetField = DefaultAllInfos
	} else if op.GetFlags != "" {
		op.GetField = strings.Split(op.GetFlags, ",")
		if len(op.GetField) <= 0 {
			gologger.Fatal().Msgf("没有获取字段信息！\n" + op.GetFlags)
		}
	}

	// 是否深度获取分支机构
	if op.IsSearchBranch {
		op.IsGetBranch = true
	}

	if op.BranchFilter != "" {
		op.NameFilterRegexp = regexp.MustCompile(op.BranchFilter)
	}

	//是否获取分支机构
	if op.IsGetBranch {
		op.GetField = append(op.GetField, "branch")
	}
	// 投资信息如果不等于0，那就收集股东信息和对外投资信息
	if op.InvestNum != 0 {
		op.GetField = append(op.GetField, "invest")
		op.GetField = append(op.GetField, "partner")
		gologger.Info().Msgf("获取投资信息，将会获取⌈%d⌋级子公司", op.Deep)
	}
	// 控股信息（大部分需要VIP）
	if op.IsHold {
		op.GetField = append(op.GetField, "holds")
	}
	if op.IsSupplier {
		op.GetField = append(op.GetField, "supplier")
	}
	if op.Deep <= 0 {
		op.Deep = 1
	}
	op.GetField = utils.SetStr(op.GetField)

	op.GetField = utils.SetStr(op.GetField)

}
