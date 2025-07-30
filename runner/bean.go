package runner

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
)

// ENSOp 用于接收前端传递的参数
type ENSOp struct {
	OrgName  string  `form:"name" json:"name"`
	ScanType string  `form:"type" json:"type"`
	Filed    string  `form:"filed" json:"filed"`
	Depth    int     `form:"depth" json:"depth"`
	Invest   float64 `form:"invest" json:"invest"`
	Holds    bool    `form:"hold" json:"hold"`
	Supplier bool    `form:"supplier" json:"supplier"`
	Branch   bool    `form:"branch" json:"branch"`
}

func (q *ESJob) NewEnJob(name string) (n *EnJob) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, exists := q.enJob[name]; !exists {
		q.enJob[name] = &ENJob{
			DataCh: make(chan map[string][]gjson.Result, 10),
			Data:   make(map[string][]gjson.Result),
			Task:   []DeepSearchTask{},
			TaskCh: make(chan DeepSearchTask),
		}
	}
	n = &EnJob{
		dataCh:    q.enJob[name].DataCh,
		data:      q.enJob[name].Data,
		task:      &q.enJob[name].Task,
		taskCh:    q.enJob[name].TaskCh,
		total:     q.enJob[name].Total,
		processed: q.enJob[name].Processed,
	}
	return n
}

// getInfoPage 根据页面获取信息，兼容不同实现
func (j *EnJob) getInfoPage(pid string, page int, em *common.EnsGo) (common.InfoPage, error) {
	if j == nil {
		return common.InfoPage{}, fmt.Errorf("EnJob is nil")
	}
	if j.job != nil {
		return j.job.GetInfoByPage(pid, page, em)
	}
	if j.app != nil {
		return j.app.GetInfoByPage(pid, page, em)
	}
	return common.InfoPage{}, fmt.Errorf("no valid implementation")
}

func (j *EnJob) getENMap() map[string]*common.EnsGo {
	if j == nil {
		return map[string]*common.EnsGo{}
	}
	if j.job != nil {
		return j.job.GetENMap()
	}
	if j.app != nil {
		return j.app.GetENMap()
	}
	return map[string]*common.EnsGo{}
}

func (j *EnJob) startCH() {
	go func() {
		for {
			select {
			case data, ok := <-j.dataCh:
				if !ok {
					j.enWg.Done()
					return
				}
				if data != nil {
					j.mu.Lock()
					gologger.Debug().Msgf("startCH\nReceived data: %v\n", data)
					for key, newResults := range data {
						j.data[key] = append(j.data[key], newResults...)
					}
					j.mu.Unlock()
				}
			}
		}
	}()

}
func (j *EnJob) closeCH() {
	close(j.dataCh)
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.taskCh != nil {
		close(j.taskCh)
	}
	if j.Done != nil {
		close(j.Done)
	}
}

func (j *ESJob) closeCH() {
	close(j.ch)
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.Jobs != nil {
		close(j.Jobs)
	}
	if j.Done != nil {
		close(j.Done)
	}
}

type DeepSearchTask struct {
	common.DPS
	SearchList []string `json:"search_list"`
}

func (j *EnJob) newTaskQueue(capacity int) {
	j.taskCh = make(chan DeepSearchTask, capacity)
	j.Done = make(chan struct{})
}
func (j *EnJob) reTaskQueue() {
	for _, task := range *j.task {
		j.taskCh <- task
		j.wg.Add(1)
	}
}

func (q *EnJob) AddTask(task DeepSearchTask) {
	q.taskCh <- task
	q.mu.Lock()
	defer q.mu.Unlock()
	*q.task = append(*q.task, task)
	q.wg.Add(1)
	q.total++
}

func (q *EnJob) StartWorkers() {
	go func() {
		for {
			select {
			case task, ok := <-q.taskCh:
				if !ok {
					return
				}
				q.processTask(task)
				*q.task = (*q.task)[1:]
				q.mu.Lock()
				q.processed++
				q.mu.Unlock()
			}
		}
	}()

}
