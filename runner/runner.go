package runner

import (
	"encoding/gob"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/gologger"
	"github.com/wgpsec/ENScan/common/utils"
	_interface "github.com/wgpsec/ENScan/interface"
	"github.com/wgpsec/ENScan/internal/aiqicha"
	"github.com/wgpsec/ENScan/internal/app/miit"
	"github.com/wgpsec/ENScan/internal/kuaicha"
	"github.com/wgpsec/ENScan/internal/riskbird"
	"github.com/wgpsec/ENScan/internal/tianyancha"
	"github.com/wgpsec/ENScan/internal/tycapi"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type EnJob struct {
	job       _interface.ENScan
	app       _interface.App
	tpy       string
	Done      chan struct{}
	total     int
	processed int
	wg        sync.WaitGroup
	enWg      sync.WaitGroup
	mu        sync.Mutex
	task      *[]DeepSearchTask
	data      map[string][]gjson.Result
	taskCh    chan DeepSearchTask
	dataCh    chan map[string][]gjson.Result
}
type ENJobTask struct {
	Keyword string `json:"keyword"`
	Typ     string `json:"type"`
	Ref     string `json:"ref"`
}

type ENSCacheDTO struct {
	JobName string
	Date    string
	Tasks   []ENJobTask
	Data    map[string][]map[string]string
	Jobs    map[string]CacheableENJob
}

type ESJob struct {
	JobName   string
	Jobs      chan ENJobTask
	items     []ENJobTask
	ch        chan map[string][]map[string]string
	dt        map[string][]map[string]string
	Done      chan struct{}
	wg        sync.WaitGroup
	esWg      sync.WaitGroup
	mu        sync.Mutex
	total     int
	processed int
	op        *common.ENOptions
	gobPath   string
	enJobs    map[string]_interface.ENScan
	enApps    map[string]_interface.App
	enJob     map[string]*ENJob
}

type CacheableENJob struct {
	Task []DeepSearchTask          `json:"task"`
	Data map[string][]gjson.Result `json:"data"` // 将 gjson.Result 转换为字符串存储
}

type ENJob struct {
	Task      []DeepSearchTask
	Data      map[string][]gjson.Result
	TaskCh    chan DeepSearchTask
	DataCh    chan map[string][]gjson.Result
	Total     int
	Processed int
}

// RunEnumeration 普通任务命令行模式，可批量导入文件查询
func RunEnumeration(options *common.ENOptions) {
	if options.InputFile != "" {
		res := utils.ReadFileOutLine(options.InputFile)
		gologger.Info().Str("FileName", options.InputFile).Msgf("读取到 %d 条信息\n", len(res))
		gologger.Info().Msgf("任务将在5秒后开始\n")
		time.Sleep(5 * time.Second)
		esjob := NewENTaskQueue(len(res), options)
		esjob.StartENWorkers()
		err := esjob.loadCacheFromGob()
		if err == nil {
			gologger.Info().Msgf("使用缓存队列进行任务，如需重新开始请删除缓存文件，缓存大小: %d", len(esjob.dt))
		} else {
			for _, v := range res {
				esjob.AddTask(v)
			}
		}
		// 批量查询任务前需要加载缓存
		esjob.wg.Wait()
		// 等待数据处理完成
		esjob.closeCH()
		esjob.esWg.Wait()
		_ = esjob.doneCacheGob()
		//如果没有指定不合并就导出
		if !options.IsNoMerge {
			err := esjob.OutFileByEnInfo(options.InputFile + "批量查询任务结果")
			if err != nil {
				gologger.Error().Msgf(err.Error())
			}
		}

	} else if options.IsApiMode {
		api(options)
	} else if options.IsMCPServer {
		McpServer(options)
	} else {
		esjob := NewENTaskQueue(1, options)
		esjob.StartENWorkers()
		err := esjob.loadCacheFromGob()
		if err == nil {
			gologger.Info().Msgf("使用缓存队列进行任务，如需重新开始请删除缓存文件，缓存大小: %d", len(esjob.dt))
		} else {
			esjob.AddTask(options.KeyWord)
		}
		// 等待任务完成
		esjob.wg.Wait()
		// 等待数据处理完成
		esjob.closeCH()
		esjob.esWg.Wait()
		_ = esjob.OutFileByEnInfo(options.KeyWord)
		_ = esjob.doneCacheGob()

	}
}

// RunJob 运行项目 添加新参数记得去Config添加
func (q *ESJob) processTask(task ENJobTask) {
	keyword := task.Keyword
	gologger.Info().Msgf("【%d/%d】正在获取 ⌈%s⌋ 信息 %s", q.processed+1, q.total, keyword, q.op.GetField)
	gologger.Debug().Msgf("关键词：⌈%s⌋ 数据源：%s 数据字段：%s\n", keyword, task.Typ, q.op.GetField)

	// 初始化任务模式
	enJob := q.NewEnJob(keyword + task.Typ)
	enJob.enWg.Add(1)
	enJobLen := len(*enJob.task)
	enJob.tpy = task.Typ
	if app, ok := q.enApps[task.Typ]; ok {
		enJob.app = app
	} else if job, isJob := q.enJobs[task.Typ]; isJob {
		enJob.job = job
	} else {
		gologger.Error().Msgf("未找到 %s 任务模式", task.Typ)
		// 如果既没有app也没有job，说明任务类型不支持，应该跳过
		enJob.enWg.Done()
		// 创建一个空的结果数据，避免程序卡住
		rdata := map[string][]map[string]string{
			"enterprise_info": {},
		}
		q.ch <- rdata
		q.wg.Done()
		return
	}
	enJob.startCH()
	// 如果是插件模式就不需要获取企业信息
	if enJob.app != nil {
		if err := enJob.getAppByKeyWord(keyword, q.op.GetField, task.Ref); err != nil {
			gologger.Error().Msgf("%s 查询信息失败\n%s", enJob.tpy, err.Error())
		}
	} else {
		if enJobLen > 0 {
			gologger.Info().Msgf("任务已存在，将直接继续进行任务查询")
			j := enJob
			j.newTaskQueue(enJobLen)
			j.reTaskQueue()
			j.StartWorkers()
			j.wg.Wait()
			j.closeCH()
		} else {
			pid := ""
			var err error
			if q.op.ISKeyPid {
				pid = keyword
			} else {
				pid, err = enJob.SearchByKeyWord(keyword)
				if err != nil {
					gologger.Error().Msgf("搜索关键词失败：%s，跳过该任务", err.Error())
					// 即使搜索失败，也要确保任务能够继续
					enJob.enWg.Done() // 先调用Done()，然后再Wait()
					enJob.enWg.Wait()
					// 创建一个空的结果数据，避免程序卡住
					rdata := map[string][]map[string]string{
						"enterprise_info": {},
					}
					q.ch <- rdata
					q.wg.Done()
					return
				}
			}
			// 获取企业信息，通过查询到的信息
			if err = enJob.getInfoById(pid, q.op.GetField); err != nil {
				gologger.Error().Msgf("获取企业信息失败：%s，跳过该任务", err.Error())
				// 即使获取企业信息失败，也要确保任务能够继续
				enJob.enWg.Done() // 先调用Done()，然后再Wait()
				enJob.enWg.Wait()
				// 创建一个空的结果数据，避免程序卡住
				rdata := map[string][]map[string]string{
					"enterprise_info": {},
				}
				q.ch <- rdata
				q.wg.Done()
				return
			}
		}
	}
	// 等待数据接受处理完成
	enJob.enWg.Wait()
	rdata := q.InfoToMap(enJob, fmt.Sprintf("%s⌈%s⌋", enJob.tpy, keyword))
	// TODO 在实时查询出来信息时自动调用
	if q.op.IsPlugins {
		for _, t := range rdata["enterprise_info"] {
			q.AddPluginsTask(t["name"])
		}
	}
	q.ch <- rdata
	// 处理完成数据进行导出
	if !q.op.IsMergeOut {
		err := common.OutFileByEnInfo(rdata, q.op.KeyWord, q.op.OutPutType, q.op.Output)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}
	}
	// 完成清理数据
	// 完成把任务清理干净
	delete(q.enJob, keyword+task.Typ)
	q.wg.Done()
}
func (q *ESJob) getENJob(tpy string) (enJob *EnJob) {
	enJob = q.NewEnJob(tpy)
	if app, ok := q.enApps[tpy]; ok {
		enJob.app = app
	} else if job, isJob := q.enJobs[tpy]; isJob {
		enJob.job = job
	} else {
		gologger.Error().Msgf("未找到 %s 任务模式", tpy)
	}
	return enJob
}

func NewENTaskQueue(capacity int, op *common.ENOptions) *ESJob {
	jobs := map[string]_interface.ENScan{
		"aqc":    &aiqicha.AQC{Options: op},
		"tyc":    &tianyancha.TYC{Options: op},
		"tycapi": &tycapi.TycAPI{Options: op},
		"kc":     &kuaicha.KC{Options: op},
		"rb":     &riskbird.RB{Options: op},
	}
	apps := map[string]_interface.App{
		"miit": &miit.Miit{Options: op},
	}
	return &ESJob{
		Jobs:    make(chan ENJobTask, capacity),
		Done:    make(chan struct{}),
		ch:      make(chan map[string][]map[string]string, capacity),
		dt:      make(map[string][]map[string]string),
		op:      op,
		gobPath: "enscan.gob",
		enJobs:  jobs,
		enApps:  apps,
		enJob:   make(map[string]*ENJob),
	}
}
func (q *ESJob) AddTask(keyword string) {
	for _, t := range q.op.GetType {
		// 如果是插件模式，而且不是指定添加插件跑
		if utils.IsInList(t, common.ENSApps) && q.op.IsPlugins {
			continue
		}
		q.AddENJobTask(ENJobTask{Keyword: keyword, Typ: t})
	}
}

func (q *ESJob) AddPluginsTask(keyword string) {
	for _, t := range q.op.GetType {
		// 如果是插件模式，而且不是指定添加插件跑
		if !utils.IsInList(t, common.ENSApps) {
			continue
		}
		q.AddENJobTask(ENJobTask{Keyword: keyword, Typ: t})
	}
}

func (q *ESJob) AddENJobTask(task ENJobTask) {
	q.Jobs <- task
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, task)
	q.wg.Add(1)
	q.total++

}

func (q *ESJob) StartENWorkers() {
	q.esWg.Add(1)
	go func() {
		for task := range q.Jobs {
			q.processTask(task)
			q.items = q.items[1:]
			q.mu.Lock()
			q.processed++
			q.mu.Unlock()
		}
	}()
	go func() {
		esJobTicker := time.NewTicker(1 * time.Second)
		var quitSig = make(chan os.Signal, 1)
		signal.Notify(quitSig, os.Interrupt, os.Kill)
		for {
			select {
			// 存储跑完的MAP数据
			case data, ok := <-q.ch:
				if !ok {
					q.esWg.Done()
					return
				}
				if data != nil {
					q.mu.Lock()
					for key, newResults := range data {
						q.dt[key] = append(q.dt[key], newResults...)
					}
					q.mu.Unlock()
				}
			case <-esJobTicker.C:
				if err := q.saveCacheToGob(); err != nil {
					gologger.Error().Msgf("保存缓存失败: %v", err)
				}
			case <-quitSig:
				if !q.op.IsApiMode && !q.op.IsMCPServer {
					gologger.Error().Msgf("任务未完成退出，将自动保存过程文件！")
					_ = q.OutFileByEnInfo("未完成任务结果")
				}
				log.Fatal("exit.by.signal")
			}
		}
	}()
}

func (q *ESJob) OutFileByEnInfo(name string) error {
	err := common.OutFileByEnInfo(q.dt, name, q.op.OutPutType, q.op.Output)
	if err != nil {
		gologger.Info().Msgf("尝试导出文件失败: %v", err)
		return err
	}
	return nil
}

func (q *ESJob) OutDataByEnInfo() map[string][]map[string]string {
	return q.dt
}

func (j *ESJob) saveCacheToGob() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// 创建或打开 Gob 文件
	file, err := os.Create(j.gobPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建可序列化的 enJob 映射
	cacheableEnJobs := make(map[string]CacheableENJob)
	for key, enJob := range j.enJob {
		if enJob == nil {
			continue
		}
		// 将 gjson.Result 转换为字符串以便存储
		cacheableEnJobs[key] = CacheableENJob{
			Task: enJob.Task,
			Data: enJob.Data,
		}
	}

	// 使用 Gob 编码器
	encoder := gob.NewEncoder(file)
	var c = &ENSCacheDTO{
		JobName: "缓存",
		Date:    time.Now().Format("2006-01-02 15:04:05"),
		Data:    j.dt,
		Tasks:   j.items,
		Jobs:    cacheableEnJobs,
	}
	return encoder.Encode(c)
}

func (j *ESJob) loadCacheFromGob() error {
	file, err := os.Open(j.gobPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，缓存为空
			j.dt = make(map[string][]map[string]string)
			return fmt.Errorf("缓存为空")
		}
		return err
	}
	defer file.Close()
	gologger.Info().Msgf("正在加载缓存数据...")
	decoder := gob.NewDecoder(file)
	c := &ENSCacheDTO{}
	if err = decoder.Decode(c); err != nil {
		return err
	}
	if c.Data == nil || c.Tasks == nil {
		return fmt.Errorf("缓存数据不完整")
	}
	j.dt = c.Data
	j.items = c.Tasks
	j.total = len(c.Tasks)

	gologger.Info().Msgf("ENSCAN[%s]加载缓存成功\n缓存时间：%s\n任务缓存数据大小: %d，剩余未完成的批量查询任务 %d\n", c.JobName, c.Date, len(c.Data), len(c.Tasks))
	mss := ""
	if len(c.Tasks) > 1 {
		for _, task := range c.Tasks {
			mss += fmt.Sprintf("- TASK[%s]关联原因：%s\n", task.Keyword, task.Ref)
		}
		gologger.Debug().Msgf(mss)
	}
	cacheableEnJobs := c.Jobs
	if cacheableEnJobs != nil {
		mss = ""
		gologger.Info().Msgf("加载JOB任务数量: %d\n", len(cacheableEnJobs))
		for key, cacheableEnJob := range cacheableEnJobs {
			enJob := &ENJob{
				Task:   cacheableEnJob.Task,
				Data:   cacheableEnJob.Data,
				DataCh: make(chan map[string][]gjson.Result, 100),
				TaskCh: make(chan DeepSearchTask, 100),
			}
			j.enJob[key] = enJob
			gologger.Info().Msgf(fmt.Sprintf("加载成功任务[%s] JOB:%d Data%d\n\n", key, len(cacheableEnJob.Task), len(cacheableEnJob.Data)))
			for _, task := range cacheableEnJob.Task {
				mss += fmt.Sprintf("- JOB[%s]关联原因：%s\n", task.Name, task.Ref)
			}
		}
		gologger.Debug().Msgf(mss)
	}
	gologger.Info().Msgf("============ 缓存加载完成，将在5S后开始任务！=============")
	time.Sleep(5 * time.Second)

	// 添加任务到队列
	for _, v := range c.Tasks {
		j.Jobs <- v
		j.wg.Add(1)
	}
	// TODO 这里需要改进，直接读取任务到队列里而不是添加
	//for _, t := range c.Tasks {
	//	j.AddENJobTask(t)
	//}

	return nil
}

// getRunningData 获取正在运行时的项目数据
func (q *ESJob) getRunData() map[string][]map[string]string {
	return q.dt
}

func (q *ESJob) getRunTaskData(tpy string) (data map[string][]gjson.Result, err error) {
	if _, ok := q.enJob[tpy]; ok {
		data = q.enJob[tpy].Data
	} else {
		return data, fmt.Errorf("获取任务数据不存在！")
	}
	return data, nil
}

func (j *ESJob) doneCacheGob() error {
	err := os.Remove(j.gobPath)
	if err != nil {
		if !os.IsNotExist(err) {
			gologger.Error().Msgf("删除文件失败: %v", err)
		}
	}
	return nil
}
