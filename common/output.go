package common

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ENSMap 单独结构
type ENSMap struct {
	Name    string
	Field   []string
	JField  []string
	KeyWord []string
	Only    string
}

// ENSMapLN 最终统一导出格式
var ENSMapLN = map[string]*ENSMap{
	"enterprise_info": {
		Name:    "企业信息",
		JField:  []string{"name", "legal_person", "status", "phone", "email", "registered_capital", "incorporation_date", "address", "scope", "reg_code", "pid"},
		KeyWord: []string{"企业名称", "法人代表", "经营状态", "电话", "邮箱", "注册资本", "成立日期", "注册地址", "经营范围", "统一社会信用代码", "PID"},
	},
	"icp": {
		Name:    "ICP备案",
		Only:    "domain",
		JField:  []string{"website_name", "website", "domain", "icp", "company_name"},
		KeyWord: []string{"网站名称", "网址", "域名", "网站备案/许可证号", "公司名称"},
	},
	"wx_app": {
		Name:    "微信小程序",
		JField:  []string{"name", "category", "logo", "qrcode", "read_num"},
		KeyWord: []string{"名称", "分类", "头像", "二维码", "阅读量"},
	},
	"wechat": {
		Name:    "微信公众号",
		JField:  []string{"name", "wechat_id", "description", "qrcode", "avatar"},
		KeyWord: []string{"名称", "ID", "简介", "二维码", "头像"},
	},
	"weibo": {
		Name:    "微博",
		JField:  []string{"name", "profile_url", "description", "avatar"},
		KeyWord: []string{"微博昵称", "链接", "简介", "头像"},
	},
	"supplier": {
		Name:    "供应商",
		JField:  []string{"name", "scale", "amount", "report_time", "data_source", "relation", "pid"},
		KeyWord: []string{"名称", "金额占比", "金额", "报告期/公开时间", "数据来源", "关联关系", "PID"},
	},
	"job": {
		Name:    "招聘",
		JField:  []string{"name", "education", "location", "publish_time", "salary"},
		KeyWord: []string{"招聘职位", "学历", "办公地点", "发布日期", "薪资"},
	},
	"invest": {
		Name:    "投资",
		JField:  []string{"name", "legal_person", "status", "scale", "pid"},
		KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "PID"},
	},
	"branch": {
		Name:    "分支机构",
		JField:  []string{"name", "legal_person", "status", "pid"},
		KeyWord: []string{"企业名称", "法人", "状态", "PID"},
	},
	"holds": {
		Name:    "控股企业",
		JField:  []string{"name", "legal_person", "status", "scale", "level", "pid"},
		KeyWord: []string{"企业名称", "法人", "状态", "投资比例", "持股层级", "PID"},
	},
	"app": {
		Name:    "APP",
		JField:  []string{"name", "category", "version", "update_at", "description", "logo", "bundle_id", "link", "market"},
		KeyWord: []string{"名称", "分类", "当前版本", "更新时间", "简介", "logo", "Bundle ID", "链接", "market"},
	},
	"copyright": {
		Name:    "软件著作权",
		JField:  []string{"name", "short_name", "category", "reg_num", "pub_type"},
		KeyWord: []string{"软件全称", "软件简称", "分类", "登记号", "权利取得方式"},
	},
	"partner": {
		Name:    "股东信息",
		JField:  []string{"name", "scale", "reg_cap", "pid"},
		KeyWord: []string{"股东名称", "持股比例", "认缴出资金额", "PID"},
	},
}

// InfoToMap 将输出的json转为统一map格式
func InfoToMap(infos map[string][]gjson.Result, enMap map[string]*EnsGo, extraInfo string) (res map[string][]map[string]string) {
	res = make(map[string][]map[string]string)
	for k, info := range infos {
		en := ENSMapLN[k].JField
		// 分类信息
		var data []map[string]string
		for _, v := range info {
			strData := make(map[string]string, len(enMap[k].Field)+1)
			// 获取字段值并转换为字符串
			for i, field := range enMap[k].Field {
				if i == len(enMap[k].Field)-1 && i >= len(en) {
					strData["from"] = v.Get(field).String()
				} else {
					strData[en[i]] = v.Get(field).String()
				}
			}
			// 添加额外信息,用于后期展示
			strData["extra"] = extraInfo
			data = append(data, strData)
		}
		// 信息全部存入
		res[k] = data
	}
	return res
}

func OutFileByEnInfo(data map[string][]map[string]string, name string, types string, dir string) (err error) {
	if dir == "!" {
		gologger.Debug().Str("设定DIR", dir).Msgf("不导出文件")
		return nil
	}
	gologger.Info().Msgf("%s 结果导出中", name)
	// 初始化导出环境
	_, err = os.Stat(dir)
	if err != nil {
		gologger.Info().Msgf("导出⌈%s⌋目录不存在，尝试创建\n", dir)
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			gologger.Debug().Str("dir", dir).Msgf(err.Error())
			return fmt.Errorf("【创建目录失败】\n %s \n", err.Error())
		}
	}
	// 判断文件名不要太长
	if len([]rune(name)) > 20 {
		name = string([]rune(name)[:20])
		gologger.Warning().Msgf("导出文件名过长，自动截断为⌈%s⌋", name)
	}
	fileUnix := time.Now().Format("2006-01-02") + "--" + strconv.FormatInt(time.Now().Unix(), 10)
	fileName := fmt.Sprintf("%s-%s.%s", name, fileUnix, types)
	savaPath := filepath.Join(dir, fileName)
	if types == "json" {
		jsonStr, err := json.Marshal(data)
		if err != nil {
			gologger.Debug().Msgf("原始格式\n %s \n", data)
			return fmt.Errorf("[JSON格式化数据失败]\n %s \n", err.Error())
		}
		err = os.WriteFile(savaPath, jsonStr, 0644)
		if err != nil {
			return fmt.Errorf("[JSON导出文件失败]\n%s", err.Error())
		}
	} else if types == "xlsx" {
		f := excelize.NewFile()
		for s, v := range data {
			em := ENSMapLN[s]
			exData := make([][]interface{}, len(v))
			for i, m := range v {
				if len(m) > 0 {
					for _, p := range append(em.JField, "from", "extra") {
						exData[i] = append(exData[i], m[p])
					}
				}
			}
			f, _ = utils.ExportExcel(em.Name, append(em.KeyWord, "数据关联", "补充信息"), exData, f)
		}
		f.DeleteSheet("Sheet1")
		if err := f.SaveAs(savaPath); err != nil {
			gologger.Fatal().Msgf("表格导出失败：%s", err)
		}
	} else {
		return fmt.Errorf("不支持的导出类型 %s", types)
	}
	gologger.Info().Msgf("导出成功⌈%s⌋", savaPath)
	return err
}
