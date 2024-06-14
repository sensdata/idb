package main

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/byteseek/idb/agent/channel"
	"github.com/byteseek/idb/agent/config"
	"github.com/byteseek/idb/agent/log"
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

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Error("Failed to accept connection: %v", err)
			continue
		}

		// 处理连接
		go handleConnection(conn, channelService)
	}
}

func handleConnection(conn net.Conn, service *channel.ChannelService) {
	defer conn.Close()

	var buffer []byte
	tmp := make([]byte, 256)
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
