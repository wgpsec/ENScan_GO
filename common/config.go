package common

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	BuiltAt   string
	GoVersion string
	GitAuthor string
	BuildSha  string
	GitTag    string
)

type ENOptions struct {
	KeyWord        string // Keyword of Search
	CompanyID      string // Company ID
	GroupID        string // Company ID
	InputFile      string // Scan Input File
	Output         string
	ScanType       string
	Proxy          string
	ISKeyPid       bool
	IsGroup        bool
	IsGetBranch    bool
	IsSearchBranch bool
	InvestNum      float64
	DelayTime      int
	DelayMaxTime   int64
	TimeOut        int
	GetFlags       string
	Version        bool
	IsHold         bool
	IsSupplier     bool
	IsShow         bool
	GetField       []string
	GetType        []string
	IsDebug        bool
	IsJsonOutput   bool
	Deep           int
	UPOutFile      string
	IsMergeOut     bool   //聚合
	IsNoMerge      bool   //聚合
	OutPutType     string // 导出文件类型
	IsApiMode      bool
	ENConfig       *ENConfig
	BranchFilter   string
}

// EnsGo EnScan 接口请求通用格式接口
type EnsGo struct {
	Name         string            // 接口名字
	Api          string            // API 地址
	Field        []string          // 获取的字段名称 看JSON
	KeyWord      []string          // 关键词
	Total        int64             // 统计梳理 AQC
	Available    int64             // 统计梳理 AQC
	GNum         string            // 获取数量的json关键词 getDetail->CountInfo QCC
	TypeInfo     []string          // 集团获取参数 QCC
	Fids         string            // 接口获取参数 QCC
	SData        map[string]string // 接口请求POST参数 QCC
	GsData       string            // get请求需要加的特殊参数 TYC
	Rf           string            // 返回数据关键词 TYC
	DataModuleId int               // 企点获取数据ID点
	AppParams    [2]string         // 插件需要获取的参数
}

// ENsD 通用返回内容格式
type ENsD struct {
	KeyWord string // Keyword of Search
	Name    string // 企业
	Pid     string // PID
	Op      *ENOptions
}

func (h *ENOptions) GetDelayRTime() int64 {
	if h.DelayTime == -1 {
		return utils.RangeRand(1, 5)
	}
	if h.DelayTime != 0 {
		h.DelayMaxTime = int64(h.DelayTime)
	}
	return 0
}

func (h *ENOptions) GetENConfig() *ENConfig {
	fmt.Println(h.KeyWord)
	return h.ENConfig
}

type EnInfos struct {
	Id          primitive.ObjectID `bson:"_id"`
	Search      string
	Name        string
	Pid         string
	LegalPerson string
	OpenStatus  string
	Email       string
	Telephone   string
	SType       string
	RegCode     string
	BranchNum   int64
	InvestNum   int64
	InTime      time.Time
	PidS        map[string]string
	Infos       map[string][]gjson.Result
	EnInfos     map[string][]map[string]interface{}
	EnInfo      []map[string]interface{}
}

// DefaultAllInfos 默认收集信息列表
var DefaultAllInfos = []string{"icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright"}
var DefaultInfos = []string{"icp", "weibo", "wechat", "app", "wx_app"}
var CanSearchAllInfos = []string{"enterprise_info", "icp", "weibo", "wechat", "app", "job", "wx_app", "copyright", "supplier", "invest", "branch", "holds", "partner"}
var DeepSearch = []string{"invest", "branch", "holds", "supplier"}
var ENSTypes = []string{"aqc", "tyc", "kc", "miit"}
var ScanTypeKeys = map[string]string{
	"aqc":     "爱企查",
	"qcc":     "企查查",
	"tyc":     "天眼查",
	"xlb":     "小蓝本",
	"kc":      "快查",
	"all":     "全部查询",
	"aldzs":   "阿拉丁",
	"coolapk": "酷安市场",
	"qimai":   "七麦数据",
	"chinaz":  "站长之家",
	"miit":    "miitICP",
}

// ENConfig YML配置文件，更改时注意变更 cfgYV 版本
type ENConfig struct {
	Version float64 `yaml:"version"`
	Cookies struct {
		Aldzs      string `yaml:"aldzs"`
		Aiqicha    string `yaml:"aiqicha"`
		Qidian     string `yaml:"qidian"`
		KuaiCha    string `yaml:"kuaicha"`
		Tianyancha string `yaml:"tianyancha"`
		Tycid      string `yaml:"tycid"`
		AuthToken  string `yaml:"auth_token"`
		QiMai      string `yaml:"qimai"`
	}
	App struct {
		MiitApi string `yaml:"miit_api"`
	}
}

var cfgYName = filepath.Join(utils.GetConfigPath(), "config.yaml")
var cfgYV = 0.5
var configYaml = `version: 0.5
app:
  miit_api: ''          # HG-ha的ICP_Query (非狼组维护，团队成员请使用内部版本)
cookies:
  aiqicha: ''           # 爱企查   Cookie
  tianyancha: ''        # 天眼查   Cookie
  tycid: ''        		# 天眼查   CApi ID(capi.tianyancha.com)
  qcc: ''               # 企查查   Cookie
  qcctid: '' 			# 企查查   TID console.log(window.tid)
  aldzs: ''             # 阿拉丁   Cookie
  xlb: ''               # 小蓝本   Token
  qimai: ''             # 七麦数据 Cookie
`
