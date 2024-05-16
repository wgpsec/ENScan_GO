package common

import (
	"flag"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	yaml "gopkg.in/yaml.v2"
	"io"
	"os"
	"strings"
)

func Parse(options *ENOptions) {

	if options.KeyWord == "" && options.CompanyID == "" && options.InputFile == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	//DEBUG模式设定
	if options.IsDebug {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelDebug)
		gologger.Debug().Msgf("DEBUG 模式已开启\n")
	}

	//判断版本信息
	if options.Version {
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

	//初始化输出文件夹位置
	if options.Output == "" {
		options.Output = "outs"
	}
	if options.Output == "!" {
		gologger.Info().Msgf("当前模式不会导出文件信息！\n")
	}

	if options.Proxy != "" {
		gologger.Info().Msgf("代理已设定 ⌈%s⌋\n", options.Proxy)
	}

	if options.InputFile != "" {
		if ok := utils.FileExists(options.InputFile); !ok {
			gologger.Fatal().Msgf("未获取到文件⌈%s⌋请检查文件名是否正确\n", options.InputFile)
		}
	}

	//数据源判断 默认为爱企查
	if options.ScanType == "" && len(options.GetType) == 0 {
		options.ScanType = "aqc"
	}

	//如果是指定全部数据
	if options.ScanType == "all" {
		options.GetType = []string{"aqc", "tyc"}
		options.IsMergeOut = true
	} else if options.ScanType != "" {
		options.GetType = strings.Split(options.ScanType, ",")
	}

	options.GetType = utils.SetStr(options.GetType)
	var tmp []string
	for _, v := range options.GetType {
		if _, ok := ScanTypeKeys[v]; !ok {
			gologger.Error().Msgf("没有这个%s查询方式\n支持列表\n%s", v, ScanTypeKeys)
		} else {
			tmp = append(tmp, v)
		}
	}
	options.GetType = tmp

	// 判断获取数据字段信息
	options.GetField = utils.SetStr(options.GetField)
	if options.GetFlags == "" && len(options.GetField) == 0 {
		options.GetField = DefaultInfos
	} else if options.GetFlags == "all" {
		options.GetField = DefaultAllInfos
	} else if options.GetFlags != "" {
		options.GetField = strings.Split(options.GetFlags, ",")
		if len(options.GetField) <= 0 {
			gologger.Fatal().Msgf("没有获取字段信息！\n" + options.GetFlags)
		}
	}
	// 是否深度获取分支机构
	if options.IsSearchBranch {
		options.IsGetBranch = true
	}

	//是否获取分支机构
	if options.IsGetBranch {
		options.GetField = append(options.GetField, "branch")
	}
	// 投资信息如果不等于0，那就收集股东信息和对外投资信息
	if options.InvestNum != 0 {
		options.GetField = append(options.GetField, "invest")
		options.GetField = append(options.GetField, "partner")
		gologger.Info().Msgf("获取投资信息，将会获取⌈%d⌋级子公司", options.Deep)
	}
	// 控股信息（大部分需要VIP）
	if options.IsHold {
		options.GetField = append(options.GetField, "holds")
	}
	if options.IsSupplier {
		options.GetField = append(options.GetField, "supplier")
	}
	options.GetField = utils.SetStr(options.GetField)

	if options.IsMerge == true {
		gologger.Info().Msgf("批量查询文件将单独导出！\n")
	}
	options.IsMergeOut = !options.IsMerge
	options.GetField = utils.SetStr(options.GetField)
	options.ENConfig = conf
}
