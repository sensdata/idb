package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	sessionName := "idb-1"
	sessionID, err := getSessionID(sessionName)
	if err != nil {
		fmt.Printf("获取会话ID失败: %v\n", err)
	} else {
		fmt.Printf("获取会话ID: %s, 会话名称: %s\n", sessionID, sessionName)
	}

	// sessionID := "1182713"
	// sessionName, err := getSessionName(sessionID)
	// if err != nil {
	// 	fmt.Printf("获取会话Name失败: %v\n", err)
	// } else {
	// 	fmt.Printf("获取会话名称: %s, 会话ID: %s\n", sessionName, sessionID)
	// }
}

func getSessionID(sessionName string) (string, error) {
	// 执行 screen -ls 命令获取所有会话列表
	output, err := exec.Command("screen", "-ls").Output()
	if strings.Contains(string(output), "No Sockets found") {
		fmt.Println("no session found")
		return "", fmt.Errorf("no session found")
	}
	if err != nil {
		fmt.Printf("failed to list sessions: %v", err)
		return "", fmt.Errorf("failed to list sessions: %v", err)
	}
	fmt.Printf("output: %s", output)

	// 处理返回的结果字符串
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 查找包含 .sessionName 的行
		if !strings.Contains(line, "."+sessionName) {
			continue
		}

		// 使用正则表达式提取会话ID,使用\s+匹配前后的空白字符
		re := regexp.MustCompile(fmt.Sprintf(`(\d+)\.%s\s+`, sessionName))
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("session %s not found", sessionName)
}

// func getSessionName(sessionID string) (string, error) {
// 	// 执行 screen -ls 命令获取所有会话列表
// 	output, err := exec.Command("screen", "-ls").Output()
// 	if strings.Contains(string(output), "No Sockets found") {
// 		fmt.Println("no session found")
// 		return "", fmt.Errorf("no session found")
// 	}
// 	if err != nil {
// 		fmt.Printf("failed to list sessions: %v", err)
// 		return "", fmt.Errorf("failed to list sessions: %v", err)
// 	}
// 	fmt.Printf("output: %s", output)

// 	// 处理返回的结果字符串
// 	lines := strings.Split(string(output), "\n")
// 	for _, line := range lines {
// 		// 查找包含 sessionID. 的行
// 		if !strings.Contains(line, sessionID+".") {
// 			continue
// 		}

// 		// 使用正则表达式提取会话信息，匹配 id.name 格式
// 		re := regexp.MustCompile(`\d+\.(\S+)\s+`)
// 		matches := re.FindStringSubmatch(line)
// 		if len(matches) >= 2 {
// 			return matches[1], nil
// 		}

// 		fmt.Printf("invalid session format: %s", line)
// 		return "", fmt.Errorf("invalid session format")
// 	}

// 	return "", fmt.Errorf("session %s not found", sessionID)
// }
