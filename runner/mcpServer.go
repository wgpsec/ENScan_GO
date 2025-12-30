package runner

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	"golang.org/x/net/context"
	"log"
	"strconv"
)

// getInfoPro 顾名思义，PRO方法
func getInfoPro(q *ESJob, tpy string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var w webOp
		err := request.BindArguments(&w)
		if err != nil {
			return nil, err
		}
		job := q.getENJob(w.ScanType)
		em := job.job.GetENMap()
		var jsonBytes []byte

		switch tpy {
		case "advance_filter":
			res, e := job.job.AdvanceFilter(w.OrgName)
			if e != nil {
				return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), e
			}
			jsonBytes, err = json.Marshal(common.OriginalToMapList(res, em["enterprise_info"]))
		case "get_ensd":
			res := job.job.GetEnsD()
			jsonBytes, err = json.Marshal(res)
			if err != nil {
				return mcp.NewToolResultText(string(jsonBytes)), err
			}
		case "get_base_info":
			res, _ := job.job.GetCompanyBaseInfoById(w.OrgName)
			jsonBytes, err = json.Marshal(common.OriginalToMap(res, em["enterprise_info"]))
		case "get_page":
			filed := em[w.Filed]
			page, err := strconv.Atoi(w.Page)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
			}
			res, e := job.job.GetInfoByPage(w.OrgName, page, filed)
			if e != nil {
				return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
			}
			jsonBytes, err = json.Marshal(InfoPage{
				res.Total,
				res.Page,
				res.Size,
				res.HasNext,
				common.OriginalToMapList(res.Data, filed)})

		}
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// getInfoByKeyword 根据关键词查企业信息
func getInfoByKeyword(q *ESJob) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var w webOp
		err := request.BindArguments(&w)
		if err != nil {
			return nil, err
		}
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
		q.AddTask(options.KeyWord)
		q.wg.Wait()

		jsonBytes, err := json.Marshal(q.OutDataByEnInfo())
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

func McpServer(options *common.ENOptions) {
	s := server.NewMCPServer(
		"EnScan",
		common.GitTag,
	)
	enTask := NewENTaskQueue(1, options)
	s.AddTool(mcp.NewTool("根据关键词详细信息",
		mcp.WithDescription("根据关键词搜索企业列表"),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("API类型"),
			mcp.Enum("aqc", "rb"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("企业搜索结果的关键词"),
		),
	), getInfoPro(enTask, "advance_filter"))

	s.AddTool(mcp.NewTool("获取企业转换MAP",
		mcp.WithDescription("获取企业转换MAP"),
	), getInfoPro(enTask, "get_ensd"))

	s.AddTool(mcp.NewTool("根据PID获取企业基本信息",
		mcp.WithDescription("根据PID获取企业基本信息"),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("API类型"),
			mcp.Enum("aqc", "tyc", "rb"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("企业pid"),
		),
	), getInfoPro(enTask, "get_base_info"))

	s.AddTool(mcp.NewTool("根据pid分页获取信息",
		mcp.WithDescription("根据pid依据分类和页码获取对应的信息"),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("API类型"),
			mcp.Enum("aqc", "rb"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("企业pid"),
		),
		mcp.WithString("filed",
			mcp.Required(),
			mcp.Description("分类，只允许一个"),
			mcp.Enum("icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright"),
		),
		mcp.WithString("page",
			mcp.Required(),
			mcp.Description("页码信息"),
		),
	), getInfoPro(enTask, "get_page"))

	s.AddTool(mcp.NewTool("根据关键词详细信息",
		mcp.WithDescription("根据关键词搜索企业的icp备案、微博、微信、app、微博、招聘、微信小程序、版权信息"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("企业搜索结果的关键词"),
		),
		mcp.WithString("filed",
			mcp.Description("获取信息类别多个类别需要以,分隔"),
			mcp.Enum("icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright"),
		),
		mcp.WithString("type",
			mcp.Description("API类型"),
			mcp.Enum("aqc", "rb"),
		),
		mcp.WithString("invest",
			mcp.Description("投资比例，选择大于%几的对外投资公司"),
		),
		mcp.WithString("depth",
			mcp.Description("投资搜索深度，几级的子公司"),
		),
		mcp.WithString("branch",
			mcp.Description("是否搜索分支机构"),
		),
	), getInfoByKeyword(enTask))
	enTask.StartENWorkers()
	
	// 确定使用的端口：优先使用命令行参数，其次使用配置文件
	var port string
	var baseURL string
	if options.McpPort != "" {
		// 使用命令行指定的端口
		port = options.McpPort
		// 验证端口号是否有效
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			gologger.Error().Msgf("MCP服务启动失败！")
			gologger.Fatal().Msgf("无效的端口号: %s (端口号必须在 1-65535 之间)", port)
		}
		baseURL = "http://localhost:" + port
		gologger.Info().Msgf("使用命令行指定的端口: %s", port)
	} else {
		// 使用配置文件中的URL和端口
		baseURL = options.ENConfig.Api.Mcp
		var err error
		port, err = utils.ExtractPortString(baseURL)
		if err != nil {
			gologger.Error().Msgf("MCP服务启动失败！")
			gologger.Fatal().Msgf(err.Error())
		}
		gologger.Info().Msgf("使用配置文件中的端口: %s", port)
	}
	
	sseServer := server.NewSSEServer(s, server.WithBaseURL(baseURL))
	gologger.Info().Msgf("SSE server listening on :" + port)
	if err := sseServer.Start(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
