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

func (c *CmdHelper) RunSystemCtl(args ...string) (string, error) {
	cmdStr := strings.Join(args, " ")
	rsp, err := c.SendCommand(cmdStr)
	if err != nil {
		return rsp.Result, fmt.Errorf("failed to run command: %w", err)
	}
	return rsp.Result, nil
}

func (c *CmdHelper) IsExist(serviceName string) (bool, error) {
	out, err := c.RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		if strings.Contains(out, "disabled") {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func (c *CmdHelper) IsActive(serviceName string) (bool, error) {
	out, err := c.RunSystemCtl("is-active", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}

func (c *CmdHelper) IsEnable(serviceName string) (bool, error) {
	out, err := c.RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}

func (c *CmdHelper) SendCommand(command string) (*model.CommandResult, error) {
	var commandResult model.CommandResult

	commandRequest := model.Command{
		HostID:  1,
		Command: command,
	}

	var commandRespons model.CommandResponse

	resp, err := c.RestyClient.R().
		SetBody(commandRequest).
		SetResult(&commandRespons).
		Post("/idb/api/cmd/send")

	if err != nil {
		fmt.Printf("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		fmt.Printf("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("received error response: %s", resp.Status())
	}

	fmt.Printf("cmd response: %v", commandRespons)

	commandResult = commandRespons.Data

	return &commandResult, nil
}

func (c *CmdHelper) ReadFile(path string) (*model.FileInfo, error) {
	var fileInfo model.FileInfo

	req := model.FileContentReq{
		Path: path,
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

func (c *CmdHelper) WriteFile(path string, content string) error {

	req := model.FileEdit{
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
