package runner

import (
	"github.com/gin-gonic/gin"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
)

type webOp struct {
	OrgName  string  `form:"name" json:"name"`
	ScanType string  `form:"type" json:"type"`
	Filed    string  `form:"filed" json:"filed"`
	Depth    int     `form:"depth" json:"depth"`
	Invest   float64 `form:"invest" json:"invest"`
	Holds    bool    `form:"hold" json:"hold"`
	Supplier bool    `form:"supplier" json:"supplier"`
	Branch   bool    `form:"branch" json:"branch"`
}

func api(options *common.ENOptions) {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "OK",
		})
	})
	webInfo := func(c *gin.Context) {
		var w webOp
		err := c.ShouldBind(&w)
		if err != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "数据绑定异常！",
			})
			return
		}
		if w.OrgName == "" {
			c.JSON(400, gin.H{
				"code":    400,
				"message": "请输入查询条件！",
			})
			return
		}
		if w.Branch {
			options.IsGetBranch = true
		}
		options.KeyWord = w.OrgName
		options.GetFlags = w.Filed
		options.ScanType = w.ScanType
		options.InvestNum = w.Invest
		options.IsSupplier = w.Supplier
		options.IsHold = w.Holds
		options.Deep = w.Depth
		options.IsMergeOut = true
		common.Parse(options)
		data := RunJob(options)
		c.JSON(200, gin.H{
			"code":    200,
			"message": "ok",
			"data":    data,
		})
	}
	a := r.Group("/api")
	{
		a.GET("/info", webInfo)
		a.POST("/info", webInfo)
	}
	err := r.Run(":31000")
	if err != nil {
		gologger.Error().Msgf("API服务启动失败！")
		gologger.Fatal().Msgf(err.Error())
	}
}
