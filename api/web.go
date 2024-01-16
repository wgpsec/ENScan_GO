package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/outputfile"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"github.com/wgpsec/ENScan/db"
	"github.com/wgpsec/ENScan/runner"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

// RunWeb 轻量web模式
func RunWeb(options *common.ENOptions) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "OK",
		})
	})

	r.GET("/api/info", func(ginCtx *gin.Context) {
		//搜索参数
		orgname := ginCtx.Query("orgname")
		search := ginCtx.Query("search")
		types := ginCtx.Query("type")
		//筛选参数，查询参数
		field := ginCtx.Query("field")
		duplicate := ginCtx.Query("duplicate")

		depth, _ := strconv.Atoi(ginCtx.Query("depth"))
		InvestNum, _ := strconv.Atoi(ginCtx.Query("invest"))
		holds := ginCtx.Query("holds")
		supplier := ginCtx.Query("supplier")
		if ginCtx.Query("show") == "" {
			options.IsOnline = true
		}

		if ginCtx.Query("branch") == "true" {
			options.IsGetBranch = true
		} else {
			options.IsGetBranch = false
		}
		outputs := ginCtx.Query("output")
		if orgname == "" && search == "" {
			ginCtx.JSON(400, gin.H{
				"code":    400,
				"message": "orgname or search is empty",
			})
			return
		}
		if search != "" {
			if types == "" {
				types = "aqc"
			} else {
				if _, ok := common.ScanTypeKeys[types]; !ok {
					gologger.Errorf("没有这个%s查询方式\n支持列表\n%s", types, common.ScanTypeKeys)
					ginCtx.JSON(500, gin.H{
						"code":    500,
						"message": fmt.Sprintf("没有%s方式，支持列表：%s", types, common.ScanTypeKeys),
					})
					return
				}
			}
			orgname = runner.SearchByKeyword(search, types, options)
		}
		IsDuplicate := true
		if duplicate == "true" {
			IsDuplicate = true
		} else if duplicate == "false" {
			IsDuplicate = false
		}
		if types != "" {
			options.ScanType = types
		}
		options.GetFlags = field
		options.ScanType = "aqc"
		options.GetType = []string{}
		common.Parse(options)
		reEnsList := make(map[string][]map[string]interface{})
		searchList := []string{"enterprise_info"}
		searchList = append(searchList, options.GetField...)
		if holds == "true" {
			searchList = append(searchList, "holds")
		}
		if supplier == "true" {
			searchList = append(searchList, "supplier")
		}
		if InvestNum != 0 {
			searchList = append(searchList, "invest")
		}
		gologger.Debugf("searchList: %s\n", searchList)
		if depth == 0 {
			depth = 0
		}
		reEnsList = runner.GetWebInfo(runner.InfoDto{OrgName: orgname, SearchType: options.GetType, SearchList: searchList, InvestNum: InvestNum, Depth: depth, IsGetNext: false, IsDuplicate: IsDuplicate, DuMap: make(map[string]map[string]bool)}, reEnsList, 0, options)

		if len(reEnsList) > 0 {
			if outputs == "file" {
				ginCtx.Header("Content-Type", "application/octet-stream")
				ginCtx.Header("Content-Disposition", "attachment; filename="+orgname+".xlsx")
				ginCtx.Header("Content-Transfer-Encoding", "binary")
				//fmt.Println(reEnsList)
				err := outputfile.OutPutExcelByMergeJson(reEnsList, ginCtx.Writer)
				if err != nil {
					ginCtx.JSON(500, gin.H{
						"code":    500,
						"message": "导出失败",
					})
					return
				}
			} else {
				ginCtx.JSON(200, gin.H{
					"code":    200,
					"message": "OK",
					"data":    reEnsList,
					"columns": outputfile.ENSMapLN,
				})
				return
			}

		} else {

			ginCtx.JSON(200, gin.H{
				"code":    404,
				"message": fmt.Sprintf("没有查询到 %s", orgname),
			})
			return
		}

	})
	//先写死，之后再说~
	options.ENConfig.Web.Port = "3000"
	gologger.Infof("WEB 模式已开启 http://127.0.0.1:%s\n", options.ENConfig.Web.Port)
	err := r.Run(":" + options.ENConfig.Web.Port)
	if err != nil {
		gologger.Fatalf("web api run error: %v", err)
		return
	} else {
		gologger.Infof("web api run success\n\n")
	}
}

func RunApiWeb(options *common.ENOptions) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "OK",
		})
	})

	r.GET("/status", func(c *gin.Context) {
		queues, err := db.RmqC.GetOpenQueues()
		stats, err := db.RmqC.CollectStats(queues)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(stats)
		c.JSON(200, gin.H{
			"code":    200,
			"message": "OK",
			"data":    stats,
		})
	})
	r.POST("/api/info", func(ginCtx *gin.Context) {
		token := ginCtx.PostForm("orgname")
		if token == "" {
			ginCtx.JSON(400, gin.H{
				"code":    400,
				"message": "orgname is empty",
			})
			return
		}
		if ginCtx.PostForm("update") == "" {
			var result common.DBEnInfos
			collection := db.MongoBb.Database("ENScan").Collection("enterprise_info")
			err := collection.FindOne(context.TODO(), bson.M{"name": token}).Decode(&result)
			if err == nil {
				if result.InTime.Sub(time.Now()).Hours() < 720 {
					ginCtx.JSON(200, gin.H{
						"code":    400,
						"message": "30DAY",
					})
					return
				}
			} else {
				gologger.Errorf("NO")
			}
		}

		options.ScanType = ginCtx.PostForm("type")

		if ginCtx.PostForm("invest_rd") == "true" {
			options.IsInvestRd = true
		}
		if ginCtx.PostForm("branch") == "true" {
			options.IsGetBranch = true
		}
		//if ginCtx.PostForm("is_branch") == "true" {
		//	options.IsSearchBranch = true
		//}
		if ginCtx.PostForm("proxy") != "" {
			options.Proxy = ginCtx.PostForm("proxy")
		}
		field := ginCtx.PostForm("field")
		options.GetFlags = field
		options.DelayMaxTime = 5
		options.InvestNum = utils.FormatInvest(ginCtx.PostForm("invest"))

		options.KeyWord = token
		err := runner.AddTask(options)
		if err != nil {
			ginCtx.JSON(200, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}
		ginCtx.JSON(200, gin.H{
			"code":    200,
			"message": "OK",
			"data":    "Add OK",
		})

	})
	r.GET("/api/stockchart", func(ginCtx *gin.Context) {
		orgname := ginCtx.Query("orgname")
		search := ginCtx.Query("search")
		//默认去重
		duplicate := ginCtx.Query("duplicate")
		IsDuplicate := true
		if duplicate == "true" {
			IsDuplicate = true
		} else if duplicate == "false" {
			IsDuplicate = false
		}
		if orgname == "" && search == "" {
			ginCtx.JSON(400, gin.H{
				"code":    400,
				"message": "orgname or search is empty",
			})
			return
		}
		if search != "" {
			orgname = runner.SearchByKeyword(search, options.ScanType, options)
		}
		options.GetFlags = "invest,partner"
		options.InvestNum = -1
		common.Parse(options)
		reEnsList := make(map[string][]map[string]interface{})
		reEnsList = runner.GetInfo(runner.InfoDto{OrgName: orgname, SearchList: []string{"enterprise_info", "partner", "invest"}, InvestNum: -1, Depth: 0, IsGetNext: true, IsDuplicate: IsDuplicate, DuMap: make(map[string]map[string]bool)}, reEnsList, 0, options)
		if len(reEnsList) > 0 {
			ginCtx.JSON(200, gin.H{
				"code":    200,
				"message": "OK",
				"data":    reEnsList,
			})
		} else {
			ginCtx.JSON(200, gin.H{
				"code":    404,
				"message": orgname + " NO DATA",
			})
		}
	})
	r.GET("/api/info", func(ginCtx *gin.Context) {
		//搜索参数
		orgname := ginCtx.Query("orgname")
		search := ginCtx.Query("search")
		types := ginCtx.Query("type")
		//筛选参数，查询参数
		field := ginCtx.Query("field")
		duplicate := ginCtx.Query("duplicate")

		depth, _ := strconv.Atoi(ginCtx.Query("depth"))
		InvestNum, _ := strconv.Atoi(ginCtx.Query("invest"))
		holds := ginCtx.Query("holds")
		supplier := ginCtx.Query("supplier")

		if ginCtx.Query("branch") == "true" {
			options.IsGetBranch = true
		} else {
			options.IsGetBranch = false
		}
		outputs := ginCtx.Query("output")
		if orgname == "" && search == "" {
			ginCtx.JSON(400, gin.H{
				"code":    400,
				"message": "orgname or search is empty",
			})
			return
		}
		if search != "" {
			if types == "" {
				types = "aqc"
			} else {
				if _, ok := common.ScanTypeKeys[types]; !ok {
					gologger.Errorf("没有这个%s查询方式\n支持列表\n%s", types, common.ScanTypeKeys)
					ginCtx.JSON(500, gin.H{
						"code":    500,
						"message": fmt.Sprintf("没有%s方式，支持列表：%s", types, common.ScanTypeKeys),
					})
					return
				}
			}
			orgname = runner.SearchByKeyword(search, types, options)
		}
		IsDuplicate := true
		if duplicate == "true" {
			IsDuplicate = true
		} else if duplicate == "false" {
			IsDuplicate = false
		}
		if types != "" {
			options.ScanType = types
		}
		options.GetFlags = field
		options.ScanType = "all"
		options.GetType = []string{}
		common.Parse(options)
		reEnsList := make(map[string][]map[string]interface{})
		searchList := []string{"enterprise_info"}
		searchList = append(searchList, options.GetField...)
		if holds == "true" {
			searchList = append(searchList, "holds")
		}
		if supplier == "true" {
			searchList = append(searchList, "supplier")
		}
		if InvestNum != 0 {
			searchList = append(searchList, "invest")
		}
		gologger.Debugf("searchList: %s\n", searchList)
		if depth == 0 {
			depth = 0
		}
		reEnsList = runner.GetInfo(runner.InfoDto{OrgName: orgname, SearchType: options.GetType, SearchList: searchList, InvestNum: InvestNum, Depth: depth, IsGetNext: false, IsDuplicate: IsDuplicate, DuMap: make(map[string]map[string]bool)}, reEnsList, 0, options)

		if len(reEnsList) > 0 {
			if outputs == "file" {
				ginCtx.Header("Content-Type", "application/octet-stream")
				ginCtx.Header("Content-Disposition", "attachment; filename="+orgname+".xlsx")
				ginCtx.Header("Content-Transfer-Encoding", "binary")
				//fmt.Println(reEnsList)
				err := outputfile.OutPutExcelByMergeJson(reEnsList, ginCtx.Writer)
				if err != nil {
					ginCtx.JSON(500, gin.H{
						"code":    500,
						"message": "导出失败",
					})
					return
				}
			} else {
				ginCtx.JSON(200, gin.H{
					"code":    200,
					"message": "OK",
					"data":    reEnsList,
					"columns": outputfile.ENSMapLN,
				})
				return
			}

		} else {
			if res, _ := db.RedisBb.Get("ENScanK:" + orgname).Result(); orgname != "" && res != "" {
				ginCtx.JSON(200, gin.H{
					"code":    000,
					"inTime":  res,
					"message": fmt.Sprintf("关键词 %s 于 %s 入库，正在队列查询，请刷新重试", orgname, res),
				})
				return
			}

			ginCtx.JSON(200, gin.H{
				"code":    404,
				"message": fmt.Sprintf("没有查询到 %s,任务队列已添加", orgname),
			})
			return
		}

	})

	gologger.Infof("API 模式已开启 :31000\n")
	err := r.Run(":31000")
	if err != nil {
		gologger.Fatalf("web api run error: %v", err)
		return
	} else {
		gologger.Infof("web api run success\n\n")
	}
}
