package _interface

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
)

type ENScan interface {
	AdvanceFilter() ([]gjson.Result, error)
	GetENMap() map[string]*common.EnsGo
	GetEnsD() common.ENsD
	GetCompanyBaseInfoById(pid string) (gjson.Result, map[string]*common.EnsGo)
	GetEnInfoList(pid string, enMap *common.EnsGo) ([]gjson.Result, error)
}
