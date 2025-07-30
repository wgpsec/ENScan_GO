package runner

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"strconv"
)

type webOp struct {
	OrgName   string  `form:"name" json:"name"`
	ScanType  string  `form:"type" json:"type"`
	Filed     string  `form:"filed" json:"filed"`
	Depth     int     `form:"depth" json:"depth"`
	Invest    float64 `form:"invest" json:"invest"`
	Holds     bool    `form:"hold" json:"hold"`
	Supplier  bool    `form:"supplier" json:"supplier"`
	Branch    bool    `form:"branch" json:"branch"`
	IsRefresh bool    `form:"refresh" json:"refresh"`
	Page      string  `form:"page" json:"page"`
}

func api(options *common.ENOptions) {
	r := gin.Default()
	enApiData := make(map[string]map[string][]map[string]string)
	enApiStatus := make(map[string]bool)
	enTask := NewENTaskQueue(1, options)
	getEnInfo := func(c *gin.Context) {
		w, err := bindWebOp(c)
		if err != nil {
			c.JSON(400, gin.H{
				"code":    400,
				"message": err.Error(),
			})
		}
		var rdata map[string][]map[string]string
		message := "查询数据中，请稍等..."
		if is := enApiStatus[w.OrgName]; !is || w.IsRefresh {
			if data, ok := enApiData[w.OrgName]; !ok || w.IsRefresh {
				enApiStatus[w.OrgName] = true
				rdata = enTask.getInfoByApi(w)
				fmt.Println(rdata)
				message = "查询成功"
				enApiStatus[w.OrgName] = false
				enApiData[w.OrgName] = rdata
			} else {
				rdata = data
				message = "来源缓存数据"
				enApiStatus[w.OrgName] = false
			}
		}

		c.JSON(200, gin.H{
			"code":    200,
			"message": message,
			"data":    rdata,
		})
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "ENSCAN IS OK!",
		})
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "ENSCAN IS OK!",
			"data":    enTask.getStatus(),
		})
	})
	getPro := func(c *gin.Context) {
		h, code, err := advApi(c, enTask, c.Param("type"))
		if err != nil {
			return
		}
		c.JSON(code, h)
	}

	a := r.Group("/api")
	{
		a.GET("/info", getEnInfo)
		a.POST("/info", getEnInfo)
		a.GET("/pro/:type", getPro)

	}
	enTask.StartENWorkers()
	err := r.Run(":31000")
	if err != nil {
		gologger.Error().Msgf("API服务启动失败！")
		gologger.Fatal().Msgf(err.Error())
	}
}

type ENStatus struct {
	JobProcessed int         `json:"processed"`
	JobTotal     int         `json:"total"`
	JobItems     []ENJobTask `json:"items"`
	Tasks        []ENTask    `json:"tasks"`
}
type ENTask struct {
	TaskName      string           `json:"name"`
	TaskProcessed int              `json:"processed"`
	TaskTotal     int              `json:"total"`
	Items         []DeepSearchTask `json:"items"`
}

type InfoPage struct {
	Total   int64               `json:"total"`
	Page    int64               `json:"page"`
	Size    int64               `json:"size"`
	HasNext bool                `json:"has_next"`
	List    []map[string]string `json:"list"`
}

func (q *ESJob) getStatus() (status ENStatus) {
	status.JobProcessed = q.processed
	status.JobTotal = q.total
	status.JobItems = q.items
	var taskList []ENTask
	for s, task := range q.enJob {
		taskList = append(taskList, ENTask{
			TaskName:      s,
			TaskProcessed: task.Processed,
			TaskTotal:     task.Total,
			Items:         task.Task,
		})
	}
	status.Tasks = taskList
	return status
}
func (q *ESJob) getInfoByApi(w webOp) map[string][]map[string]string {
	options := q.op
	options.KeyWord = w.OrgName
	options.GetFlags = w.Filed
	options.ScanType = w.ScanType
	options.InvestNum = w.Invest
	options.IsSupplier = w.Supplier
	options.IsHold = w.Holds
	options.Deep = w.Depth
	options.IsMergeOut = true
	options.Parse()

	q.AddTask(w.OrgName)
	q.wg.Wait()
	return q.OutDataByEnInfo()
}

func bindWebOp(c *gin.Context) (w webOp, err error) {
	err = c.ShouldBind(&w)
	if err != nil {
		return w, fmt.Errorf("绑定失败")
	}
	if w.OrgName == "" {
		return w, fmt.Errorf("请输入查询条件")
	}
	return w, nil
}

func advApi(c *gin.Context, enTask *ESJob, tpy string) (r gin.H, code int, err error) {
	w, err := bindWebOp(c)
	if err != nil {
		return r, 400, err
	}
	job := enTask.getENJob(w.ScanType)
	em := job.job.GetENMap()
	switch tpy {
	case "advance_filter":
		res, err := job.job.AdvanceFilter(w.OrgName)
		if err != nil {
			return r, 500, err
		}
		r = gin.H{
			"code":    200,
			"message": "查询成功",
			"data":    common.OriginalToMapList(res, em["enterprise_info"]),
		}
	case "get_ensd":
		res := job.job.GetEnsD()
		r = gin.H{
			"code":    200,
			"message": "查询成功",
			"data":    res,
		}
	case "get_base_info":
		res, e := job.job.GetCompanyBaseInfoById(w.OrgName)
		r = gin.H{
			"code":    200,
			"message": "查询成功",
			"data":    common.OriginalToMap(res, em["enterprise_info"]),
			"em":      e,
		}
	case "get_page":
		filed := em[c.Query("filed")]
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			return r, 500, err
		}
		res, err := job.job.GetInfoByPage(w.OrgName, page, filed)
		if err != nil {
			return r, 500, err
		}
		r = gin.H{
			"code":    200,
			"message": "查询成功",
			"data": InfoPage{
				res.Total,
				res.Page,
				res.Size,
				res.HasNext,
				common.OriginalToMapList(res.Data, filed),
			},
		}
	default:
		return r, 400, fmt.Errorf("接口不存在")
	}
	return r, 200, nil
}
