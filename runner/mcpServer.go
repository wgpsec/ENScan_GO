package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	"github.com/wgpsec/ENScan/internal/aiqicha"
	"golang.org/x/net/context"
	"log"
)

func helloSearchById(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	var enOptions common.ENOptions
	enOptions.IsMCPServer = true
	pid, ok := arguments["pid"].(string)
	if !ok {
		return nil, errors.New("pid must be a string")
	}
	filed, ok := arguments["filed"].(string)
	if ok {
		enOptions.GetFlags = filed
	}
	common.Parse(&enOptions)
	enOptions.CompanyID = pid
	enOptions.IsMergeOut = true
	data := RunJob(&enOptions)
	r, err := json.Marshal(data)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
	}
	return mcp.NewToolResultText(fmt.Sprintf("%s", r)), nil
}
func helloSearchListByOgrName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	keyWord, ok := arguments["OrgName"].(string)
	if !ok {
		return nil, errors.New("orgName must be a string")
	}
	var enOptions common.ENOptions
	enOptions.IsMCPServer = true
	common.Parse(&enOptions)
	enOptions.KeyWord = keyWord
	job := &aiqicha.AQC{Options: &enOptions}
	enList, err := job.AdvanceFilter()
	enMap := job.GetENMap()["enterprise_info"]
	if err != nil {
		gologger.Error().Msg(err.Error())
		return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
	} else {
		utils.TBS(append(enMap.KeyWord[:3], "PID"), append(enMap.Field[:3], enMap.Field[10]), "企业信息", enList)
		plList := common.InfoToMap(map[string][]gjson.Result{
			"enterprise_info": enList,
		}, job.GetENMap(), "")
		r, err := json.Marshal(plList["enterprise_info"])
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("处理异常！")), err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%s", r)), nil

	}
}

func mcpServer(options *common.ENOptions) {
	s := server.NewMCPServer(
		"EnScan",
		"1.0.0",
	)
	// Add tool

	// Add tool handler
	s.AddTool(mcp.NewTool("根据PID详细信息",
		mcp.WithDescription("根据pid搜索企业的icp备案、微博、微信、app、微博、招聘、微信小程序、版权信息"),

		mcp.WithString("pid",
			mcp.Required(),
			mcp.Description("企业搜索结果的PID"),
		),
		mcp.WithString("filed",
			mcp.Description("获取信息类别多个类别需要以,分隔"),
			mcp.Enum("icp", "weibo", "wechat", "app", "weibo", "job", "wx_app", "copyright"),
		),
	), helloSearchById)
	s.AddTool(mcp.NewTool("关键词匹配企业列表",
		mcp.WithDescription("根据关键词搜索匹配企业列表"),
		mcp.WithString("OrgName",
			mcp.Required(),
			mcp.Description("企业名称"),
		),
	), helloSearchListByOgrName)

	sseServer := server.NewSSEServer(s, server.WithBaseURL("http://localhost:8080"))
	gologger.Info().Msgf("SSE server listening on :8080")
	if err := sseServer.Start(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
