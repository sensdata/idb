package factory

import (
	"fmt"

	"github.com/sensdata/idb/core/logstream/internal/config"
	"github.com/sensdata/idb/core/logstream/pkg/reader"
	readerAdater "github.com/sensdata/idb/core/logstream/pkg/reader/adapters"
	"github.com/sensdata/idb/core/logstream/pkg/task"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/logstream/pkg/writer"
	writerAdapter "github.com/sensdata/idb/core/logstream/pkg/writer/adapters"
)

// NewReader 根据任务类型创建对应的读取器
func NewReader(task *types.Task, taskMgr task.Manager, cfg *config.Config) (reader.Reader, error) {
	switch task.Type {
	case types.TaskTypeBuffer:
		buffer, err := taskMgr.GetBuffer(task.ID)
		if err != nil {
			return nil, fmt.Errorf("get buffer failed: %w", err)
		}
		if buffer == nil {
			return nil, fmt.Errorf("buffer not initialized for task: %s", task.ID)
		}
		return readerAdater.NewBufferReader(buffer), nil
	case types.TaskTypeFile:
		return readerAdater.NewTailReader(task.LogPath, cfg)
	case types.TaskTypeRemote:
		return readerAdater.NewRemoteReader(task.LogPath, cfg), nil
	default:
		return nil, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}

// NewWriter 根据任务类型创建对应的写入器
func NewWriter(task *types.Task, taskMgr task.Manager, cfg *config.Config) (writer.Writer, error) {
	switch task.Type {
	case types.TaskTypeBuffer:
		buffer, err := taskMgr.GetBuffer(task.ID)
		if err != nil {
			return nil, err
		}
		return writerAdapter.NewBufferWriter(buffer), nil
	case types.TaskTypeFile:
		return writerAdapter.NewFileWriter(task.LogPath, cfg.LogBufferSize)
	default:
		return nil, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}
