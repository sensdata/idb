package types

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	taskCount      int64
	activeTasks    int64
	completedTasks int64
	failedTasks    int64
	totalLogSize   int64
	readOps        int64
	writeOps       int64
	cleanupCount   int64
	errorCount     int64

	// 错误记录
	lastError struct {
		sync.RWMutex
		msg  string
		time time.Time
	}
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (mc *MetricsCollector) IncrTaskCount()        { atomic.AddInt64(&mc.taskCount, 1) }
func (mc *MetricsCollector) DecrTaskCount()        { atomic.AddInt64(&mc.taskCount, -1) }
func (mc *MetricsCollector) IncrActiveTasks()      { atomic.AddInt64(&mc.activeTasks, 1) }
func (mc *MetricsCollector) DecrActiveTasks()      { atomic.AddInt64(&mc.activeTasks, -1) }
func (mc *MetricsCollector) IncrCompletedTasks()   { atomic.AddInt64(&mc.completedTasks, 1) }
func (mc *MetricsCollector) IncrFailedTasks()      { atomic.AddInt64(&mc.failedTasks, 1) }
func (mc *MetricsCollector) AddLogSize(size int64) { atomic.AddInt64(&mc.totalLogSize, size) }
func (mc *MetricsCollector) IncrReadOps()          { atomic.AddInt64(&mc.readOps, 1) }
func (mc *MetricsCollector) IncrWriteOps()         { atomic.AddInt64(&mc.writeOps, 1) }
func (mc *MetricsCollector) IncrCleanupCount()     { atomic.AddInt64(&mc.cleanupCount, 1) }
func (mc *MetricsCollector) IncrErrorCount()       { atomic.AddInt64(&mc.errorCount, 1) }

// SetLogSize 设置日志总大小
func (mc *MetricsCollector) SetLogSize(size int64) {
	atomic.StoreInt64(&mc.totalLogSize, size)
}

// SetTaskCount 设置任务总数
func (mc *MetricsCollector) SetTaskCount(count int64) {
	atomic.StoreInt64(&mc.taskCount, count)
}

// RecordLastError 记录最后一次错误
func (mc *MetricsCollector) RecordLastError(err error) {
	if err == nil {
		return
	}
	mc.lastError.Lock()
	mc.lastError.msg = err.Error()
	mc.lastError.time = time.Now()
	mc.lastError.Unlock()
}

// GetLastError 获取最后一次错误信息
func (mc *MetricsCollector) GetLastError() (string, time.Time) {
	mc.lastError.RLock()
	defer mc.lastError.RUnlock()
	return mc.lastError.msg, mc.lastError.time
}

// 更新 GetMetrics 方法以包含错误信息
func (mc *MetricsCollector) GetMetrics() *Metrics {
	errMsg, errTime := mc.GetLastError()
	return &Metrics{
		TaskCount:      atomic.LoadInt64(&mc.taskCount),
		ActiveTasks:    atomic.LoadInt64(&mc.activeTasks),
		CompletedTasks: atomic.LoadInt64(&mc.completedTasks),
		FailedTasks:    atomic.LoadInt64(&mc.failedTasks),
		TotalLogSize:   atomic.LoadInt64(&mc.totalLogSize),
		ReadOps:        atomic.LoadInt64(&mc.readOps),
		WriteOps:       atomic.LoadInt64(&mc.writeOps),
		CleanupCount:   atomic.LoadInt64(&mc.cleanupCount),
		ErrorCount:     atomic.LoadInt64(&mc.errorCount),
		LastError:      errMsg,
		LastErrorTime:  errTime,
	}
}
