package _interface

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
)

type COMMON interface {
	GetENMap() map[string]*common.EnsGo
	GetInfoByPage(pid string, page int, em *common.EnsGo) (info common.InfoPage, err error)
}

type ENScan interface {
	// AdvanceFilter 筛选公司
	AdvanceFilter(name string) ([]gjson.Result, error)
	GetEnsD() common.ENsD
	GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo)
	COMMON
}

type App interface {
	GetInfoList(keyword string, types string) []gjson.Result
	COMMON
}
