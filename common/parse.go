package common

import (
	"flag"
	"io"
	"os"
	"strings"
	"time"

	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	yaml "gopkg.in/yaml.v2"
)

func Parse(options *ENOptions) {
	//判断版本信息
	if options.Version {
		gologger.Infof("Current Version: %s\n", GitTag)
		gologger.Infof("当前所需配置文件版本 V%.1f\n", cfgYV)
		if ok, _ := utils.PathExists(cfgYName); !ok {
			f, errs := os.Create(cfgYName) //创建文件
			_, errs = io.WriteString(f, configYaml)
			if errs != nil {
				gologger.Fatalf("配置文件创建失败 %s\n", errs)
			}
			gologger.Infof("配置文件生成成功\n")
		}
		os.Exit(0)
	}
	// 配置文件检查
	if ok, _ := utils.PathExists(cfgYName); !ok {
		gologger.Fatalf("没有找到配置文件 %s 请先运行 -v 创建\n", cfgYName)
	}

	//加载配置信息~
	conf := new(ENConfig)
	yamlFile, err := os.ReadFile(cfgYName)
	if err != nil {
		gologger.Fatalf("配置文件解析错误 #%v ", err)
	}
	if err := yaml.Unmarshal(yamlFile, conf); err != nil {
		gologger.Fatalf("【配置文件加载失败】: %v", err)
	}
	if conf.Version != cfgYV {
		gologger.Fatalf("配置文件当前[V%.1f] 程序需要[V%.1f] 不匹配，请备份配置文件重新运行-v\n", conf.Version, cfgYV)
	}
	//初始化输出文件夹位置
	if options.Output == "" && conf.Common.Output != "" {
		options.Output = conf.Common.Output
	} else if options.Output == "" {
		options.Output = "outs"
	}

	//DEBUG模式设定
	if options.IsDebug {
		gologger.MaxLevel = gologger.Debug
		gologger.Debugf("DEBUG 模式已开启\n")
	}

	if options.ClientMode != "" {
		options.IsApiMode = true
	}

	// 是否为API模式 加入基本参数判断
	if !options.IsApiMode && !options.IsWebMode {
		if options.KeyWord == "" && options.CompanyID == "" && options.InputFile == "" {
			flag.PrintDefaults()
			os.Exit(0)
		}
		if options.InputFile != "" {
			if ok := utils.FileExists(options.InputFile); !ok {
				gologger.Fatalf("没有输入文件 %s\n", options.InputFile)
			}
		}
	} else {
		options.IsShow = false
		options.IsMergeOut = true
		options.Deep = 1
	}

	if options.Output == "!" {
		gologger.Infof("当前模式不会导出文件信息！\n")
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
			gologger.Errorf("没有这个%s查询方式\n支持列表\n%s", v, ScanTypeKeys)
		} else {
			tmp = append(tmp, v)
		}
	}
	options.GetType = tmp

	//判断是否添加墨子任务
	if options.IsBiuCreate {
		if conf.Biu.Api == "" || conf.Biu.Key == "" {
			gologger.Fatalf("没有配置 墨子 API地址与Api key （请前往个人设置->安全设置中获取Api Key） \n")
		}
	}

	if len(conf.Biu.Tags) == 0 {
		conf.Biu.Tags = []string{"ENScan"}
	}

	// 判断获取数据字段信息
	options.GetField = utils.SetStr(options.GetField)
	if options.GetFlags == "" && len(conf.Common.Field) == 0 {
		if len(options.GetField) == 0 {
			options.GetField = DefaultInfos
		}
	} else if options.GetFlags == "all" {
		options.GetField = DefaultAllInfos
	} else {
		if len(conf.Common.Field) > 0 {
			options.GetField = conf.Common.Field
		}
		if options.GetFlags != "" {
			options.GetField = strings.Split(options.GetFlags, ",")
			if len(options.GetField) <= 0 {
				gologger.Fatalf("没有获取到字段信息 \n" + options.GetFlags)
			}

		}
	}
	//是否获取分支机构
	if options.IsGetBranch {
		options.GetField = append(options.GetField, "branch")
	}
	// 投资信息如果不等于0，那就收集股东信息和对外投资信息
	if options.InvestNum != 0 {
		options.GetField = append(options.GetField, "invest")
		options.GetField = append(options.GetField, "partner")
	}
	if options.IsHold {
		options.GetField = append(options.GetField, "holds")
	}
	if options.IsSupplier {
		options.GetField = append(options.GetField, "supplier")
	}
	options.GetField = utils.SetStr(options.GetField)
	// 判断是否在给定范围内，防止产生入库问题
	if options.IsApiMode {
		//var tmps []string
		//for _, v := range options.GetField {
		//	if _, ok := outputfile.ENSMapLN[v]; ok {
		//		tmps = append(tmps, v)
		//	} else {
		//		gologger.Debugf("%s不在范围内\n", v)
		//	}
		//}
		//options.GetType = tmps
	}

	if options.IsMerge == true {
		gologger.Infof("====已强制取消合并导出！====\n")
		options.IsMergeOut = false
	}

	options.GetField = utils.SetStr(options.GetField)

	options.ENConfig = conf
	//获取邮箱懒得写了就这样直接合并吧（
	if options.IsEmailPro {
		gologger.Infof("由于作者太懒，需开启合并后才能查询邮箱\n")

		if options.ENConfig.Cookies.Veryvp == "" {
			gologger.Fatalf("错误，Veryvp COOKIE为空\n")
		}
		time.Sleep(3 * time.Second)
		options.IsMergeOut = true
	}

}
