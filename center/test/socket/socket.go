package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensdata/idb/core/message"
)

var curSession string

func main() {
	// 连接到 WebSocket 服务器
	url := "ws://127.0.0.1:9918/api/v1/terminals/1/start"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// 构造 TerminalStart 消息
	msgObj := message.WsMessage{
		Type:      "attach",
		Session:   "",
		Data:      "",
		Timestamp: int(time.Now().Unix()),
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

	// 创建一个通道用于接收用户输入
	inputChan := make(chan string)

	// 启动一个协程来捕获用户输入
	go func() {
		reader := bufio.NewReader(os.Stdin) // 使用 bufio.Reader
		for {
			input, err := reader.ReadString('\n') // 读取整行输入
			if err != nil {
				log.Printf("读取输入失败: %v", err)
				continue
			}
			// log.Printf("输入: %s", input)
			inputChan <- input // 保留换行符
		}
	}()

	go func() {
		for {
			// 读取消息
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Printf("读取消息失败: %v", err)
				return
			}
			// 将消息解码为 message.WsMessage 类型
			var wsMessage message.WsMessage
			err = json.Unmarshal(p, &wsMessage)
			if err != nil {
				log.Printf("解码消息失败: %v", err)
				return
			}
			if wsMessage.Type == "attach" {
				curSession = wsMessage.Session
			}
			fmt.Printf("%s", wsMessage.Data)
		}
	}()

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

		case userInput := <-inputChan:
			// log.Printf("接收到用户输入: %s", userInput) // 确保这里能打印
			// 处理用户输入
			msgObj := message.WsMessage{
				Type:      "cmd",
				Session:   curSession,
				Data:      userInput, // 使用用户输入
				Timestamp: int(time.Now().Unix()),
			}

			// 将消息编码为 JSON
			wsData, err := json.Marshal(msgObj)
			if err != nil {
				log.Fatalf("Failed to marshal message: %v", err)
			}

			// 发送消息
			// log.Printf("Send message: %v\n", wsData)
			err = conn.WriteMessage(websocket.TextMessage, wsData)
			if err != nil {
				log.Fatalf("Failed to send message: %v", err)
			}
		}
	}
}
