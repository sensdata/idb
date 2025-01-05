package main

import (
	"encoding/json"
	"log"
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
		Type:    "start",
		Session: "test",
		Data:    "",
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

	// 等待响应（可选）
	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	log.Printf("Received response: %s", response)

	// 等待一段时间以确保消息发送完成
	time.Sleep(1000 * time.Second)
}
