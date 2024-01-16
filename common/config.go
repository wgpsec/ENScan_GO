package common

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ENOptions struct {
	KeyWord        string // Keyword of Search
	CompanyID      string // Company ID
	GroupID        string // Company ID
	InputFile      string // Scan Input File
	Output         string
	CookieInfo     string
	ScanType       string
	Proxy          string
	ISKeyPid       bool
	IsGroup        bool
	IsGetBranch    bool
	IsSearchBranch bool
	IsInvestRd     bool
	IsEmailPro     bool
	InvestNum      float64
	DelayTime      int
	DelayMaxTime   int64
	TimeOut        int
	GetFlags       string
	Version        bool
	IsBiuCreate    bool
	IsHold         bool
	IsSupplier     bool
	IsShow         bool
	CompanyName    string
	GetField       []string
	GetType        []string
	IsDebug        bool
	Deep           int
	IsJsonOutput   bool
	IsApiMode      bool
	IsWebMode      bool
	IsMergeOut     bool   //聚合
	IsMerge        bool   //聚合
	ClientMode     string //客户端模式
	IsOnline       bool
	ENConfig       *ENConfig
}

func (h *ENOptions) GetDelayRTime() int64 {
	if h.DelayTime != 0 {
		h.DelayMaxTime = int64(h.DelayTime)
	}
	if h.DelayMaxTime == 0 {
		return 0
	}
	return utils.RangeRand(1, h.DelayMaxTime)
}

func (h *ENOptions) GetENConfig() *ENConfig {
	fmt.Println(h.KeyWord)
	return h.ENConfig
}

// ENConfig YML配置文件，更改时注意变更 cfgYV 版本
type ENConfig struct {
	Version float64 `yaml:"version"`
	Common  struct {
		Output string   `yaml:"output"`
		Field  []string `yaml:"field"`
	}
	Biu struct {
		Api      string   `yaml:"api"`
		Key      string   `yaml:"key"`
		Port     string   `yaml:"port"`
		IsPublic bool     `yaml:"is-public"`
		Tags     []string `yaml:"tags"`
	}
	Api struct {
		Server  string `yaml:"server"`
		Mongodb string `yaml:"mongodb"`
		Redis   string `yaml:"redis"`
	}
	Web struct {
		Port     string `yaml:"port"`
		Database struct {
			Type        string `yaml:"type"`
			Host        string `yaml:"host"`
			Port        int    `yaml:"port"`
			User        string `yaml:"user"`
			Password    string `yaml:"password"`
			Name        string `yaml:"name"`
			DBFile      string `yaml:"db_file"`
			TablePrefix string `yaml:"table_prefix"`
		}
	}
	Cookies struct {
		Aldzs      string `yaml:"aldzs"`
		Xlb        string `yaml:"xlb"`
		Aiqicha    string `yaml:"aiqicha"`
		Tianyancha string `yaml:"tianyancha"`
		Tycid      string `yaml:"tycid"`
		Qcc        string `yaml:"qcc"`
		QiMai      string `yaml:"qimai"`
		ChinaZ     string `yaml:"chinaz"`
		Veryvp     string `yaml:"veryvp"`
	}
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

type DBEnInfos struct {
	Id          primitive.ObjectID `bson:"_id"`
	Name        string
	RegCode     string
	InTime      time.Time
	InvestCount int
	InfoCount   map[string][]string
	Info        []map[string]interface{}
}

var ENSMapAQC = map[string]string{
	"webRecord":     "icp",
	"appinfo":       "app",
	"wechatoa":      "wechat",
	"enterprisejob": "job",
	"microblog":     "weibo",
	"hold":          "holds",
	"shareholders":  "partner",
}

// DefaultAllInfos 默认收集信息列表
var DefaultAllInfos = []string{"icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright"}
var DefaultInfos = []string{"icp", "weibo", "wechat", "app", "wx_app"}
var CanSearchAllInfos = []string{"enterprise_info", "icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright", "supplier", "invest", "branch", "holds", "partner"}

var ScanTypeKeys = map[string]string{
	"aqc":     "爱企查",
	"qcc":     "企查查",
	"tyc":     "天眼查",
	"xlb":     "小蓝本",
	"all":     "全部查询",
	"aldzs":   "阿拉丁",
	"coolapk": "酷安市场",
	"qimai":   "七麦数据",
	"chinaz":  "站长之家",
}

var ScanTypeKeyV = map[string]string{
	"爱企查":  "aqc",
	"企查查":  "qcc",
	"天眼查":  "tyc",
	"小蓝本":  "xlb",
	"阿拉丁":  "aldzs",
	"酷安市场": "coolapk",
	"七麦数据": "qimai",
	"站长之家": "chinaz",
}

// RequestTimeOut 请求超时设置
var RequestTimeOut = 30 * time.Second
var (
	BuiltAt   string
	GoVersion string
	GitAuthor string
	BuildSha  string
	GitTag    string
)
var cfgYName = utils.GetConfigPath() + "/config.yaml"
var cfgYV = 0.4
var configYaml = `version: 0.4
common:
  output: ""            # 导出文件位置
  field: [ ]			# 查询字段 如["website"]
web:
  port: "32000"
  database:
    type: "sqlite3"
    host: ""
    port: 0
    user: ""
    password: ""
    name: ""
    db_file: "enscan.db"
    table_prefix: "e_"
cookies:
  aiqicha: ''           # 爱企查   Cookie
  tianyancha: ''        # 天眼查   Cookie
  tycid: ''        		# 天眼查   CApi ID(capi.tianyancha.com)
  aldzs: ''             # 阿拉丁   TOKEN(see README)
  qimai: ''             # 七麦数据  Cookie
  chinaz: ''			# 站长之家  Cookie
  veryvp: '' 			# veryvp  Cookie
`
