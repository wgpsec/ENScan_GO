package common

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common/utils"
	"path/filepath"
	"regexp"
	"time"
)

var (
	BuiltAt   string
	GoVersion string
	GitAuthor string
	BuildSha  string
	GitTag    string
)

type ENOptions struct {
	KeyWord          string // Keyword of Search
	CompanyID        string // Company ID
	GroupID          string // Company ID
	InputFile        string // Scan Input File
	Output           string
	ScanType         string
	Proxy            string
	ISKeyPid         bool
	IsGroup          bool
	IsGetBranch      bool
	IsSearchBranch   bool
	InvestNum        float64
	DelayTime        int
	DelayMaxTime     int64
	TimeOut          int
	GetFlags         string
	Version          bool
	IsHold           bool
	IsSupplier       bool
	IsShow           bool
	GetField         []string
	GetType          []string
	IsDebug          bool
	IsJsonOutput     bool
	Deep             int
	IsMergeOut       bool   //聚合
	IsNoMerge        bool   //聚合
	OutPutType       string // 导出文件类型
	IsApiMode        bool
	IsMCPServer      bool
	McpPort          string // MCP服务器监听端口
	IsPlugins        bool   // 是否作为后置插件查询
	IsFast           bool   // 是否快速查询
	ENConfig         *ENConfig
	BranchFilter     string
	NameFilterRegexp *regexp.Regexp
}

// ENSearch 搜索必要的参数
// 暂时没想好如何使用，尤其是在多进程情况下
type ENSearch struct {
	KeyWord          string // Keyword of Search
	GetField         []string
	GetType          []string
	IsGetBranch      bool
	NameFilterRegexp *regexp.Regexp
	InvestNum        float64
	IsHold           bool
	IsSupplier       bool
	Deep             int
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
	Price        float32           // 价格
	Ex           []string          // 扩展字段
}

// ENsD 通用返回内容格式
type ENsD struct {
	KeyWord string // Keyword of Search
	Name    string // 企业
	Pid     string // PID
	Op      *ENOptions
}

type InfoPage struct {
	Total   int64
	Page    int64
	Size    int64
	HasNext bool
	Data    []gjson.Result
}

// DPS ENScan深度搜索包
type DPS struct {
	Name       string   `json:"name"`        // 企业名称
	Pid        string   `json:"pid"`         // 企业ID
	Ref        string   `json:"ref"`         // 关联原因
	Deep       int      `json:"deep"`        // 深度
	SK         string   `json:"type"`        // 搜索类型
	SearchList []string `json:"search_list"` // 深度搜索列表
}

func (h *ENOptions) GetDelayRTime() int64 {
	if h.DelayTime == -1 {
		return utils.RangeRand(1, 5)
	}
	if h.DelayTime != 0 {
		h.DelayMaxTime = int64(h.DelayTime)
		return int64(h.DelayTime)
	}
	return 0
}

func (h *ENOptions) GetENConfig() *ENConfig {
	return h.ENConfig
}

func (h *ENOptions) GetCookie(tpy string) (b string) {
	c := h.ENConfig.Cookies
	switch tpy {
	case "aqc":
		b = c.Aiqicha
	case "tyc":
		b = c.Tianyancha
	case "rb":
		b = c.RiskBird
	case "qcc":
		b = c.Qcc
	case "xlb":
		b = c.Xlb
	case "kc":
		b = c.KuaiCha
	case "qimai":
		b = c.QiMai
	}
	return b

}

type EnInfos struct {
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

var AbnormalStatus = []string{"注销", "吊销", "停业", "清算", "歇业", "关闭", "撤销", "迁出", "经营异常", "严重违法失信"}

// DefaultAllInfos 默认收集信息列表
var DefaultAllInfos = []string{"icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright"}
var DefaultInfos = []string{"icp", "weibo", "wechat", "app", "wx_app"}
var CanSearchAllInfos = []string{"enterprise_info", "icp", "weibo", "wechat", "app", "job", "wx_app", "copyright", "supplier", "invest", "branch", "holds", "partner"}
var DeepSearch = []string{"invest", "branch", "holds", "supplier"}
var ENSTypes = []string{"aqc", "xlb", "qcc", "tyc", "kc", "tycapi", "rb"}
var ENSApps = []string{"miit"}
var ScanTypeKeys = map[string]string{
	"aqc":     "爱企查",
	"qcc":     "企查查",
	"tyc":     "天眼查",
	"tycapi":  "天眼查",
	"xlb":     "小蓝本",
	"kc":      "快查",
	"rb":      "风鸟",
	"all":     "全部查询",
	"aldzs":   "阿拉丁",
	"coolapk": "酷安市场",
	"qimai":   "七麦数据",
	"chinaz":  "站长之家",
	"miit":    "miitICP",
}

// ENConfig YML配置文件，更改时注意变更 cfgYV 版本
type ENConfig struct {
	Version   float64 `yaml:"version"`
	UserAgent string  `yaml:"user_agent"` // 自定义 User-Agent
	Api       struct {
		Api string `yaml:"api"`
		Mcp string `yaml:"mcp"`
	}
	Cookies struct {
		Aldzs       string `yaml:"aldzs"`
		Xlb         string `yaml:"xlb"`
		Aiqicha     string `yaml:"aiqicha"`
		Qidian      string `yaml:"qidian"`
		KuaiCha     string `yaml:"kuaicha"`
		Tianyancha  string `yaml:"tianyancha"`
		Tycid       string `yaml:"tycid"`
		TycApiToken string `yaml:"tyc_api_token"`
		RiskBird    string `yaml:"risk_bird"`
		AuthToken   string `yaml:"auth_token"`
		Qcc         string `yaml:"qcc"`
		QccTid      string `yaml:"qcctid"`
		QiMai       string `yaml:"qimai"`
		ChinaZ      string `yaml:"chinaz"`
	}
	App struct {
		MiitApi string `yaml:"miit_api"`
	}
}

var cfgYName = filepath.Join(utils.GetConfigPath(), "config.yaml")
var cfgYV = 0.7
var configYaml = `version: 0.7
user_agent: ""			# 自定义 User-Agent（可设置为获取Cookie的浏览器）
app:
  miit_api: ''          # HG-ha的ICP_Query (非狼组维护 https://github.com/HG-ha/ICP_Query) 
api:
  api: ':31000'    # API监听地址
  mcp: 'http://localhost:8080'    # MCP SSE监听地址
cookies:
  aiqicha: ''           # 爱企查   Cookie
  tianyancha: ''        # 天眼查   Cookie
  tycid: ''        		# 天眼查   CApi ID(capi.tianyancha.com)
  auth_token: ''        # 天眼查   Token (capi.tianyancha.com)
  tyc_api_token: ''     # 天眼查   官方API Key(https://open.tianyancha.com)
  risk_bird: '' 		# 风鸟     Cookie
  qimai: ''             # 七麦数据 Cookie
`
