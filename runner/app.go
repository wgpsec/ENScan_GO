package runner

import (
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
)

// getAppById 直接使用关键词调用插件查询
func (j *EnJob) getAppByKeyWord(keyWord string, searchList []string, ref string) (e error) {
	enData := make(map[string][]gjson.Result)
	em := j.app.GetENMap()
	// 批量获取信息
	for _, sk := range searchList {
		// 不支持这个搜索类型就跳过去
		if _, ok := em[sk]; !ok {
			continue
		}
		s := em[sk]
		listData, err := j.getInfoList(keyWord, s, sk, ref)
		if err != nil {
			gologger.Error().Msgf("尝试获取⌈%s⌋发生异常\n%v", s.Name, err)
			e = err
			continue
		}
		enData[sk] = append(enData[sk], listData...)
	}
	gologger.Debug().Msgf("getAppByKeyWord\nReceived data: %v\n", enData)
	j.dataCh <- enData
	j.closeCH()
	return e
}

func (j *EnJob) getAppById(rdata map[string][]map[string]string, searchList []string) (enInfo map[string][]gjson.Result) {
	enData := make(map[string][]gjson.Result)
	enMap := j.app.GetENMap()
	for _, sk := range searchList {

		if _, ok := enMap[sk]; !ok {
			continue
		}
		s := enMap[sk]
		var enList []string
		// 获取需要的参数，比如企业名称、域名等
		// 0和1分别表示 ENSMapLN 的 key 和 value，定位出需要的数据
		// 暂时没遇到需要多种类型进行匹配的参数
		ap := s.AppParams
		for _, ens := range rdata[ap[0]] {
			enList = append(enList, ens[ap[1]])
		}
		// 对获取的目标进行去重
		utils.SetStr(enList)
		gologger.Info().Msgf("共获取到【%d】条，开始执行插件获取信息", len(enList))
		for i, v := range enList {
			gologger.Info().Msgf("正在获取第【%d】条数据 【%s】", i+1, v)
			listData := j.app.GetInfoList(v, sk)
			enData[sk] = append(enData[sk], listData...)
			utils.TBS(s.KeyWord, s.Field, s.Name, listData)
		}
	}
	return enData
}
