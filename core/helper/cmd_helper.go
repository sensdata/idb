package helper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/core/model"
)

type CmdHelper struct {
	Addr        string
	Port        string
	RestyClient *resty.Client
}

func NewCmdHelper(addr string, port string, client *resty.Client) *CmdHelper {
	var restyClient *resty.Client
	if client != nil {
		restyClient = client
	} else {
		restyClient = resty.New().
			SetBaseURL(fmt.Sprintf("http://%s:%s", addr, port)).
			SetHeader("Content-Type", "application/json")
	}

	return &CmdHelper{
		Addr:        addr,
		Port:        port,
		RestyClient: restyClient,
	}
}

func (c *CmdHelper) RunSystemCtl(hostId uint, args ...string) (string, error) {
	// 在args前面添加"systemctl"
	fullArgs := append([]string{"systemctl"}, args...)

	// 将所有参数拼接成一个字符串
	cmdStr := strings.Join(fullArgs, " ")
	rsp, err := c.SendCommand(hostId, cmdStr)
	if err != nil {
		return rsp.Result, fmt.Errorf("failed to run command: %w", err)
	}
	return rsp.Result, nil
}

func (c *CmdHelper) StartService(hostId uint, serviceName string) error {
	_, err := c.RunSystemCtl(hostId, "start", serviceName)
	if err != nil {
		return err
	}
	return nil
}

func (c *CmdHelper) StopService(hostId uint, serviceName string) error {
	_, err := c.RunSystemCtl(hostId, "stop", serviceName)
	if err != nil {
		return err
	}
	return nil
}

func (c *CmdHelper) RestartService(hostId uint, serviceName string) error {
	_, err := c.RunSystemCtl(hostId, "restart", serviceName)
	if err != nil {
		return err
	}
	return nil
}

func (c *CmdHelper) IsExist(hostId uint, serviceName string) (bool, error) {
	out, err := c.RunSystemCtl(hostId, "is-enabled", serviceName)
	if err != nil {
		if strings.Contains(out, "disabled") {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func (c *CmdHelper) IsActive(hostId uint, serviceName string) (bool, error) {
	out, err := c.RunSystemCtl(hostId, "is-active", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}

func (c *CmdHelper) IsEnable(hostId uint, serviceName string) (bool, error) {
	out, err := c.RunSystemCtl(hostId, "is-enabled", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}

func (c *CmdHelper) SendCommand(hostId uint, command string) (*model.CommandResult, error) {
	var commandResult model.CommandResult

	commandRequest := model.Command{
		HostID:  hostId,
		Command: command,
	}

	var commandResponse model.CommandResponse

	resp, err := c.RestyClient.R().
		SetBody(commandRequest).
		SetResult(&commandResponse).
		Post("/idb/api/commands")

	if err != nil {
		fmt.Printf("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		fmt.Printf("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("received error response: %s", resp.Status())
	}

	fmt.Printf("cmd response: %v", commandResponse)

	commandResult = commandResponse.Data

	return &commandResult, nil
}

func (c *CmdHelper) ReadFile(hostId uint, path string) (*model.FileInfo, error) {
	var fileInfo model.FileInfo

	req := model.FileContentReq{
		HostID: hostId,
		Path:   path,
	}

	var response model.Response

	resp, err := c.RestyClient.R().
		SetBody(req).
		SetResult(&response).
		Post("/idb/files/content")

	if err != nil {
		fmt.Printf("failed to send request: %v", err)
		return &fileInfo, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		fmt.Printf("failed to send request: %v", err)
		return &fileInfo, fmt.Errorf("received error response: %s", resp.Status())
	}

	// fmt.Printf("file response: %v", response)

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		fmt.Println("Error marshaling Data:", err)
		return &fileInfo, fmt.Errorf("failed to marshaling response data: %v", err)
	}
	if err := json.Unmarshal(dataBytes, &fileInfo); err != nil {
		fmt.Println("Error unmarshaling Data to fileinfo:", err)
		return &fileInfo, fmt.Errorf("failed to unmarshaling response data: %v", err)
	}

	return &fileInfo, nil
}

func (c *CmdHelper) WriteFile(hostId uint, path string, content string) error {

	req := model.FileEdit{
		HostID:  hostId,
		Path:    path,
		Content: content,
	}

	var response model.Response

	resp, err := c.RestyClient.R().
		SetBody(req).
		SetResult(&response).
		Post("/idb/files/content/save")

	if err != nil {
		fmt.Printf("failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		fmt.Printf("failed to send request: %v", err)
		return fmt.Errorf("received error response: %s", resp.Status())
	}

	fmt.Printf("file response: %v", response)

	return nil
}
