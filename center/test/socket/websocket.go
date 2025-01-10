package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensdata/idb/core/model"
)

func main() {
	// 连接到 WebSocket 服务器
	url := "ws://127.0.0.1:9918/api/v1/terminals/1/start"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// 构造 TerminalStart 消息
	msgObj := model.TerminalMessage{
		Type:      "attach",
		Session:   "",
		Data:      "",
		Timestamp: time.Now().Unix(),
	}

	// 将消息编码为 JSON
	wsData, err := json.Marshal(msgObj)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	// 发送消息
	err = conn.WriteMessage(websocket.TextMessage, wsData)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// 创建一个通道用于接收中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 持续读取消息直到收到中断信号
	for {
		select {
		case <-interrupt:
			log.Println("收到中断信号,准备退出...")
			// 关闭WebSocket连接
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("写入关闭消息失败: %v", err)
			}
			return
		default:
			// 读取消息
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("读取消息失败: %v", err)
				return
			}
			log.Printf("收到消息: %s", message)
		}
	}
}
