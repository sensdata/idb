package main

import (
	"io"

	"github.com/gin-gonic/gin"
	logstream "github.com/sensdata/idb/core/logstream"
	"github.com/sensdata/idb/core/logstream/internal/config"
	"github.com/sensdata/idb/core/logstream/pkg/types"
)

var ls *logstream.LogStream

func main() {
	cfg := config.DefaultConfig()
	var err error
	ls, err = logstream.New(cfg)
	if err != nil {
		panic(err)
	}
	defer ls.Close()

	r := gin.Default()
	r.Static("/static", "./static")
	r.POST("/tasks", createTask)
	r.GET("/tasks/:taskID/logs", streamLogs)
	r.Run(":8080")
}

func createTask(c *gin.Context) {
	taskID, err := ls.CreateTask("demo", nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 模拟写入一些初始日志
	writer, _ := ls.GetWriter(taskID)
	writer.Write(types.LogLevelInfo, "任务已创建", nil)

	c.JSON(200, gin.H{"task_id": taskID})
}

func streamLogs(c *gin.Context) {
	taskID := c.Param("taskID")
	reader, err := ls.GetReader(taskID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	logCh, _ := reader.Follow()
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-logCh; ok {
			c.SSEvent("log", string(msg))
			return true
		}
		return false
	})
}
