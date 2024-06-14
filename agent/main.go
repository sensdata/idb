package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/sensdata/idb/agent/channel"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/log"
	"github.com/sensdata/idb/agent/utils"
)

func main() {
	fmt.Println("Agent Starting")

	configPath := "config.json"

	manager, err := config.NewManager(configPath)
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v \n", err)
	}

	cfg := manager.GetConfig()
	fmt.Println("Get config:")
	fmt.Printf("%+v \n", *cfg)

	// 初始化日志模块
	if err := log.InitLogger(cfg.LogPath); err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
	}

	//启动服务
	done := make(chan struct{})
	defer close(done)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startServer(cfg, done)
	}()

	// 等待服务完全启动
	time.Sleep(time.Second * 5)

	// 等待服务和测试完成
	wg.Add(1)
	go func() {
		defer wg.Done()
		testSendMessage(cfg)
	}()

	// 等待信号
	waitForSignal()

	// 等待服务和测试完成
	wg.Wait()

	fmt.Println("Agent Exited")
}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func startServer(cfg *config.Config, done <-chan struct{}) {
	// 将端口号转换为字符串
	portStr := strconv.Itoa(cfg.Port)
	log.Info("Agent started, try listen on port %s", portStr)

	// 监听端口
	lis, err := net.Listen("tcp", ":"+portStr)
	if err != nil {
		log.Error("Failed to listen on port %s: %v", portStr, err)
		// TODO: 后续完善单一端口的问题
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return
	}

	log.Info("Starting TCP server on port %d", cfg.Port)

	channelService := channel.NewChannelService(cfg.SecretKey)

	go func() {
		for {
			select {
			case <-done:
				log.Info("Shutting down server...")
				return
			default:
				conn, err := lis.Accept()
				if err != nil {
					log.Error("Failed to accept connection: %v", err)
					continue
				}
				// 处理连接
				go handleConnection(conn, channelService)
			}
		}
	}()
}

func handleConnection(conn net.Conn, service *channel.ChannelService) {
	defer conn.Close()

	var buffer []byte
	tmp := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				log.Error("Read error: %v", err)
			}
			break
		}

		buffer = append(buffer, tmp[:n]...)

		// 尝试解析消息
		if err := service.AddMessage(buffer); err != nil {
			log.Error("Error processing message: %v", err)
		}

		// 清空缓冲区
		buffer = buffer[:0]
	}
}

func testSendMessage(cfg *config.Config) {
	// 构造消息
	nonce := utils.GenerateNonce(16)
	msg, err := channel.CreateMessage(
		"10000001",
		"Hello, this is a test message!",
		cfg.SecretKey,
		nonce,
	)
	if err != nil {
		fmt.Printf("Error creating message: %v\n", err)
		return
	}

	// 发送消息
	err = channel.SendMessage("127.0.0.1", cfg.Port, msg)
	if err != nil {
		fmt.Printf("Failed to send message: %v \n", err)
	} else {
		fmt.Printf("Message sent successfully! \n")
	}
}
