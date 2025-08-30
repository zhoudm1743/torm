package db

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncProcessor 异步处理器
type AsyncProcessor struct {
	workers    []*AsyncWorker
	workerPool chan *AsyncWorker
	taskQueue  chan *AsyncTask
	resultPool sync.Pool

	// 统计信息
	totalTasks     int64
	completedTasks int64
	failedTasks    int64

	// 配置
	maxWorkers   int
	maxQueueSize int
	timeout      time.Duration

	// 状态
	running bool
	mutex   sync.RWMutex
}

// AsyncWorker 异步工作器
type AsyncWorker struct {
	id        int
	processor *AsyncProcessor
	taskChan  chan *AsyncTask
	quit      chan bool
	wg        *sync.WaitGroup
}

// AsyncTask 异步任务
type AsyncTask struct {
	ID          int64
	Type        TaskType
	Query       *QueryBuilder
	Data        interface{}
	Callback    func(*AsyncResult)
	ErrCallback func(error)
	Context     context.Context
	CreatedAt   time.Time

	// 批量操作字段
	BatchData []interface{}
	BatchSize int

	// 优先级
	Priority TaskPriority
}

// AsyncResult 异步结果
type AsyncResult struct {
	TaskID    int64
	Type      TaskType
	Data      interface{}
	Error     error
	Duration  time.Duration
	Timestamp time.Time

	// 批量结果
	BatchResults []interface{}
	BatchErrors  []error
}

// TaskType 任务类型
type TaskType int

const (
	TaskTypeSelect TaskType = iota
	TaskTypeInsert
	TaskTypeUpdate
	TaskTypeDelete
	TaskTypeBatchInsert
	TaskTypeBatchUpdate
	TaskTypeBatchDelete
	TaskTypeTransaction
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow TaskPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// BatchOperation 批量操作接口
type BatchOperation interface {
	Execute(ctx context.Context) (interface{}, error)
	GetBatchSize() int
	GetPriority() TaskPriority
}

// BatchInsertOperation 批量插入操作
type BatchInsertOperation struct {
	query     *QueryBuilder
	data      []map[string]interface{}
	batchSize int
	priority  TaskPriority
}

// BatchUpdateOperation 批量更新操作
type BatchUpdateOperation struct {
	query      *QueryBuilder
	conditions []map[string]interface{}
	updates    []map[string]interface{}
	batchSize  int
	priority   TaskPriority
}

// NewAsyncProcessor 创建异步处理器
func NewAsyncProcessor(maxWorkers, maxQueueSize int) *AsyncProcessor {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU() * 2
	}
	if maxQueueSize <= 0 {
		maxQueueSize = 10000
	}

	ap := &AsyncProcessor{
		workers:      make([]*AsyncWorker, 0, maxWorkers),
		workerPool:   make(chan *AsyncWorker, maxWorkers),
		taskQueue:    make(chan *AsyncTask, maxQueueSize),
		maxWorkers:   maxWorkers,
		maxQueueSize: maxQueueSize,
		timeout:      30 * time.Second,
		running:      false,
	}

	ap.resultPool.New = func() interface{} {
		return &AsyncResult{}
	}

	return ap
}

// Start 启动异步处理器
func (ap *AsyncProcessor) Start() error {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.running {
		return ErrAsyncProcessorRunning
	}

	ap.running = true

	// 创建工作器
	for i := 0; i < ap.maxWorkers; i++ {
		worker := &AsyncWorker{
			id:        i,
			processor: ap,
			taskChan:  make(chan *AsyncTask, 10),
			quit:      make(chan bool),
			wg:        &sync.WaitGroup{},
		}

		ap.workers = append(ap.workers, worker)
		ap.workerPool <- worker

		// 启动工作器
		go worker.start()
	}

	// 启动任务分发器
	go ap.dispatcher()

	return nil
}

// Stop 停止异步处理器
func (ap *AsyncProcessor) Stop() error {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if !ap.running {
		return ErrAsyncProcessorNotRunning
	}

	ap.running = false

	// 停止所有工作器
	for _, worker := range ap.workers {
		worker.stop()
	}

	// 等待所有任务完成
	for _, worker := range ap.workers {
		worker.wg.Wait()
	}

	return nil
}

// SubmitAsync 提交异步任务
func (ap *AsyncProcessor) SubmitAsync(task *AsyncTask) error {
	if !ap.running {
		return ErrAsyncProcessorNotRunning
	}

	task.ID = atomic.AddInt64(&ap.totalTasks, 1)
	task.CreatedAt = time.Now()

	if task.Context == nil {
		task.Context = context.Background()
	}

	select {
	case ap.taskQueue <- task:
		return nil
	default:
		return ErrTaskQueueFull
	}
}

// SelectAsync 异步查询
func (ap *AsyncProcessor) SelectAsync(query *QueryBuilder, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:     TaskTypeSelect,
		Query:    query,
		Callback: callback,
		Priority: PriorityNormal,
	}

	return ap.SubmitAsync(task)
}

// InsertAsync 异步插入
func (ap *AsyncProcessor) InsertAsync(query *QueryBuilder, data interface{}, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:     TaskTypeInsert,
		Query:    query,
		Data:     data,
		Callback: callback,
		Priority: PriorityNormal,
	}

	return ap.SubmitAsync(task)
}

// UpdateAsync 异步更新
func (ap *AsyncProcessor) UpdateAsync(query *QueryBuilder, data interface{}, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:     TaskTypeUpdate,
		Query:    query,
		Data:     data,
		Callback: callback,
		Priority: PriorityNormal,
	}

	return ap.SubmitAsync(task)
}

// DeleteAsync 异步删除
func (ap *AsyncProcessor) DeleteAsync(query *QueryBuilder, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:     TaskTypeDelete,
		Query:    query,
		Callback: callback,
		Priority: PriorityNormal,
	}

	return ap.SubmitAsync(task)
}

// BatchInsertAsync 异步批量插入
func (ap *AsyncProcessor) BatchInsertAsync(query *QueryBuilder, data []interface{}, batchSize int, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:      TaskTypeBatchInsert,
		Query:     query,
		BatchData: data,
		BatchSize: batchSize,
		Callback:  callback,
		Priority:  PriorityHigh,
	}

	return ap.SubmitAsync(task)
}

// BatchUpdateAsync 异步批量更新
func (ap *AsyncProcessor) BatchUpdateAsync(query *QueryBuilder, data []interface{}, batchSize int, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:      TaskTypeBatchUpdate,
		Query:     query,
		BatchData: data,
		BatchSize: batchSize,
		Callback:  callback,
		Priority:  PriorityHigh,
	}

	return ap.SubmitAsync(task)
}

// TransactionAsync 异步事务
func (ap *AsyncProcessor) TransactionAsync(operations []func(*QueryBuilder) error, callback func(*AsyncResult)) error {
	task := &AsyncTask{
		Type:     TaskTypeTransaction,
		Data:     operations,
		Callback: callback,
		Priority: PriorityCritical,
	}

	return ap.SubmitAsync(task)
}

// GetStats 获取统计信息
func (ap *AsyncProcessor) GetStats() map[string]int64 {
	return map[string]int64{
		"total_tasks":     atomic.LoadInt64(&ap.totalTasks),
		"completed_tasks": atomic.LoadInt64(&ap.completedTasks),
		"failed_tasks":    atomic.LoadInt64(&ap.failedTasks),
		"pending_tasks":   int64(len(ap.taskQueue)),
		"active_workers":  int64(len(ap.workers)),
	}
}

// dispatcher 任务分发器
func (ap *AsyncProcessor) dispatcher() {
	for {
		if !ap.running {
			break
		}

		select {
		case task := <-ap.taskQueue:
			// 获取空闲工作器
			select {
			case worker := <-ap.workerPool:
				worker.taskChan <- task
			default:
				// 没有空闲工作器，任务回到队列
				select {
				case ap.taskQueue <- task:
				default:
					// 队列满了，丢弃任务
					atomic.AddInt64(&ap.failedTasks, 1)
					if task.ErrCallback != nil {
						task.ErrCallback(ErrTaskQueueFull)
					}
				}
			}
		}
	}
}

// AsyncWorker 方法

// start 启动工作器
func (aw *AsyncWorker) start() {
	aw.wg.Add(1)
	defer aw.wg.Done()

	for {
		select {
		case task := <-aw.taskChan:
			aw.processTask(task)
			// 工作器回到池中
			aw.processor.workerPool <- aw

		case <-aw.quit:
			return
		}
	}
}

// stop 停止工作器
func (aw *AsyncWorker) stop() {
	close(aw.quit)
}

// processTask 处理任务
func (aw *AsyncWorker) processTask(task *AsyncTask) {
	startTime := time.Now()
	result := aw.processor.resultPool.Get().(*AsyncResult)
	defer aw.processor.resultPool.Put(result)

	// 重置结果
	result.TaskID = task.ID
	result.Type = task.Type
	result.Error = nil
	result.Data = nil
	result.BatchResults = nil
	result.BatchErrors = nil
	result.Timestamp = time.Now()

	// 检查超时
	ctx := task.Context
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), aw.processor.timeout)
		defer cancel()
	}

	// 执行任务
	switch task.Type {
	case TaskTypeSelect:
		result.Data, result.Error = aw.executeSelect(ctx, task)
	case TaskTypeInsert:
		result.Data, result.Error = aw.executeInsert(ctx, task)
	case TaskTypeUpdate:
		result.Data, result.Error = aw.executeUpdate(ctx, task)
	case TaskTypeDelete:
		result.Data, result.Error = aw.executeDelete(ctx, task)
	case TaskTypeBatchInsert:
		result.BatchResults, result.BatchErrors = aw.executeBatchInsert(ctx, task)
	case TaskTypeBatchUpdate:
		result.BatchResults, result.BatchErrors = aw.executeBatchUpdate(ctx, task)
	case TaskTypeBatchDelete:
		result.BatchResults, result.BatchErrors = aw.executeBatchDelete(ctx, task)
	case TaskTypeTransaction:
		result.Data, result.Error = aw.executeTransaction(ctx, task)
	default:
		result.Error = ErrUnsupportedTaskType
	}

	result.Duration = time.Since(startTime)

	// 更新统计
	if result.Error != nil {
		atomic.AddInt64(&aw.processor.failedTasks, 1)
		if task.ErrCallback != nil {
			task.ErrCallback(result.Error)
		}
	} else {
		atomic.AddInt64(&aw.processor.completedTasks, 1)
	}

	// 调用回调
	if task.Callback != nil {
		task.Callback(result)
	}
}

// executeSelect 执行查询
func (aw *AsyncWorker) executeSelect(ctx context.Context, task *AsyncTask) (interface{}, error) {
	query := task.Query.WithContext(ctx)
	return query.Get()
}

// executeInsert 执行插入
func (aw *AsyncWorker) executeInsert(ctx context.Context, task *AsyncTask) (interface{}, error) {
	query := task.Query.WithContext(ctx)
	if data, ok := task.Data.(map[string]interface{}); ok {
		return query.Insert(data)
	}
	return nil, ErrUnsupportedTaskType
}

// executeUpdate 执行更新
func (aw *AsyncWorker) executeUpdate(ctx context.Context, task *AsyncTask) (interface{}, error) {
	query := task.Query.WithContext(ctx)
	if data, ok := task.Data.(map[string]interface{}); ok {
		return query.Update(data)
	}
	return nil, ErrUnsupportedTaskType
}

// executeDelete 执行删除
func (aw *AsyncWorker) executeDelete(ctx context.Context, task *AsyncTask) (interface{}, error) {
	query := task.Query.WithContext(ctx)
	return query.Delete()
}

// executeBatchInsert 执行批量插入
func (aw *AsyncWorker) executeBatchInsert(ctx context.Context, task *AsyncTask) ([]interface{}, []error) {
	batchSize := task.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	results := make([]interface{}, 0, len(task.BatchData))
	errors := make([]error, 0)

	for i := 0; i < len(task.BatchData); i += batchSize {
		end := i + batchSize
		if end > len(task.BatchData) {
			end = len(task.BatchData)
		}

		batch := task.BatchData[i:end]

		// 逐个插入（批量操作）
		for _, item := range batch {
			if data, ok := item.(map[string]interface{}); ok {
				query := task.Query.WithContext(ctx)
				result, err := query.Insert(data)
				results = append(results, result)
				if err != nil {
					errors = append(errors, err)
				} else {
					errors = append(errors, nil)
				}
			} else {
				results = append(results, nil)
				errors = append(errors, ErrUnsupportedTaskType)
			}
		}
	}

	return results, errors
}

// executeBatchUpdate 执行批量更新
func (aw *AsyncWorker) executeBatchUpdate(ctx context.Context, task *AsyncTask) ([]interface{}, []error) {
	batchSize := task.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	results := make([]interface{}, 0, len(task.BatchData))
	errors := make([]error, 0)

	for i := 0; i < len(task.BatchData); i += batchSize {
		end := i + batchSize
		if end > len(task.BatchData) {
			end = len(task.BatchData)
		}

		batch := task.BatchData[i:end]

		for _, data := range batch {
			query := task.Query.WithContext(ctx)
			if updateData, ok := data.(map[string]interface{}); ok {
				result, err := query.Update(updateData)
				results = append(results, result)
				if err != nil {
					errors = append(errors, err)
				} else {
					errors = append(errors, nil)
				}
			} else {
				results = append(results, nil)
				errors = append(errors, ErrUnsupportedTaskType)
			}
		}
	}

	return results, errors
}

// executeBatchDelete 执行批量删除
func (aw *AsyncWorker) executeBatchDelete(ctx context.Context, task *AsyncTask) ([]interface{}, []error) {
	batchSize := task.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	results := make([]interface{}, 0, len(task.BatchData))
	errors := make([]error, 0)

	for i := 0; i < len(task.BatchData); i += batchSize {
		end := i + batchSize
		if end > len(task.BatchData) {
			end = len(task.BatchData)
		}

		for j := i; j < end; j++ {
			query := task.Query.WithContext(ctx)
			result, err := query.Delete()

			results = append(results, result)
			if err != nil {
				errors = append(errors, err)
			} else {
				errors = append(errors, nil)
			}
		}
	}

	return results, errors
}

// executeTransaction 执行事务
func (aw *AsyncWorker) executeTransaction(ctx context.Context, task *AsyncTask) (interface{}, error) {
	operations, ok := task.Data.([]func(*QueryBuilder) error)
	if !ok {
		return nil, ErrUnsupportedTaskType
	}

	tx, err := task.Query.connection.BeginTx(nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 执行所有操作
	for _, operation := range operations {
		query := task.Query.WithContext(ctx)
		query.transaction = tx

		if err := operation(query); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return "transaction completed", nil
}

// 错误定义
var (
	ErrAsyncProcessorRunning    = NewError(ErrCodeInvalidModelState, "异步处理器已运行，不能重复启动")
	ErrAsyncProcessorNotRunning = NewError(ErrCodeInvalidModelState, "异步处理器未运行，请先启动")
	ErrTaskQueueFull            = NewError(ErrCodeTimeout, "任务队列已满，请稍后重试")
	ErrUnsupportedTaskType      = NewError(ErrCodeNotImplemented, "不支持的任务类型")
)
