package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

func main() {
	// 启动一个新的 screen 会话，指定启动 bash shell
	screenCmd := exec.Command("screen", "-d", "-r", "1452888")

	// 设置环境变量
	homeDir, _ := os.UserHomeDir()
	fmt.Printf("homedir: %s\n", homeDir)
	screenCmd.Env = append(os.Environ(),
		"TERM=xterm",              // 设置为xterm以兼容xterm.js
		"SHELL=/bin/bash",         // 设置默认shell
		"HOME="+homeDir,           // 设置用户主目录
		"PATH="+os.Getenv("PATH"), // 确保PATH包含必要的命令
	)

	// 使用 pty.Open 启动进程并获取主设备
	// ptyMaster, err := pty.Start(screenCmd)
	ws := &pty.Winsize{Rows: 24, Cols: 80}
	ptyMaster, err := pty.StartWithSize(screenCmd, ws)
	if err != nil {
		log.Fatalf("Error opening pty: %v", err)
	}
	defer ptyMaster.Close()

	// 启动一个 goroutine 处理 screen 会话的输出
	go func() {
		buf := make([]byte, 1024)
		for {
			// 读取 screen 会话的输出
			n, err := ptyMaster.Read(buf)
			if err != nil {
				log.Fatalf("Error reading from pty: %v", err)
			}

			output := string(buf[:n])

			// 输出命令执行结果
			fmt.Print(output)

			// 如果输出末尾有提示符，表示命令执行完成
			// if strings.HasSuffix(output, "$ ") || strings.HasSuffix(output, "# ") {
			// 	// 看到提示符，表示当前命令执行完毕，准备接受新的输入
			// 	fmt.Print("\nCommand finished. Ready for new input:\n")
			// }
		}
	}()

	// 现在你可以开始输入命令并通过 ptyMaster 向 screen 会话发送输入
	for {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("读取输入失败: %v", err)
		}

		input = strings.TrimSpace(input)

		// 将命令发送到 screen 会话的 bash shell
		_, err = fmt.Fprintf(ptyMaster, "%s\n", input)
		if err != nil {
			log.Fatalf("Error writing to ptySlave: %v", err)
		}

		// 延时一小段时间，等待命令执行
		time.Sleep(500 * time.Millisecond) // Adjust delay if necessary
	}
}
