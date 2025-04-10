package service

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/sensdata/idb/center/core/conn"
	db "github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type SettingsService struct{}

type ISettingsService interface {
	About() (*model.About, error)
	IPs() (*model.AvailableIps, error)
	Timezones(req model.SearchPageInfo) (*model.PageResult, error)
	Settings() (*model.SettingInfo, error)
	Update(req model.UpdateSettingRequest) (*model.UpdateSettingResponse, error)
	Upgrade() error
}

func NewISettingsService() ISettingsService {
	return &SettingsService{}
}

func (s *SettingsService) About() (*model.About, error) {
	var about model.About

	about.Version = global.Version

	// 获取新版本信息
	about.NewVersion = getLatestVersion()

	return &about, nil
}

func getLatestVersion() string {
	cmd := fmt.Sprintf("curl -sSL %s", conn.CONFMAN.GetConfig().Latest)
	global.LOG.Info("Getting latest version: %s", cmd)
	latest, err := utils.Exec(cmd)
	if err != nil {
		global.LOG.Error("Failed to get latest version: %v", err)
		return ""
	}
	global.LOG.Info("Got latest version: %s", latest)
	return strings.TrimSpace(latest)
}

func (s *SettingsService) IPs() (*model.AvailableIps, error) {
	var availableIps model.AvailableIps
	availableIps.IPs = make([]model.BindIp, 0)

	// 添加几项ip：
	// 所有IP - 0.0.0.0
	availableIps.IPs = append(availableIps.IPs, model.BindIp{IP: "0.0.0.0", Name: "All IP"})
	// 127.0.0.1 - 127.0.0.1
	availableIps.IPs = append(availableIps.IPs, model.BindIp{IP: "127.0.0.1", Name: "127.0.0.1"})
	// ::1 - ::1
	availableIps.IPs = append(availableIps.IPs, model.BindIp{IP: "::1", Name: "::1"})
	// Link-Local Address
	interfaces, err := net.Interfaces()
	if err != nil {
		return &availableIps, nil
	}
	for _, iface := range interfaces {
		// 只要 eth0
		if iface.Name != "eth0" {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			// 获取 Link-Local 地址
			if ipNet.IP.IsLinkLocalUnicast() {
				availableIps.IPs = append(
					availableIps.IPs,
					model.BindIp{IP: ipNet.IP.String(), Name: ipNet.IP.String()},
				)
			}
		}
	}

	return &availableIps, nil
}

func (s *SettingsService) Timezones(req model.SearchPageInfo) (*model.PageResult, error) {
	var (
		result    model.PageResult
		backDatas []db.Timezone
	)

	// 获取所有时区
	timezones, err := TimezoneRepo.GetList()
	if err != nil {
		return &result, err
	}
	// 分页
	total, start, end := len(timezones), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]db.Timezone, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = timezones[start:end]
	}

	result.Total = int64(total)
	result.Items = backDatas

	return &result, nil
}
func (s *SettingsService) Settings() (*model.SettingInfo, error) {
	bindIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return nil, err
	}
	bindPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
	if err != nil {
		return nil, err
	}
	bindPortValue, err := strconv.Atoi(bindPort.Value)
	if err != nil {
		return nil, err
	}
	https, err := SettingsRepo.Get(SettingsRepo.WithByKey("Https"))
	if err != nil {
		return nil, err
	}
	httpsCertType, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return nil, err
	}
	httpsCertPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return nil, err
	}
	httpsKeyPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return nil, err
	}

	return &model.SettingInfo{
		BindIP:        bindIP.Value,
		BindPort:      bindPortValue,
		Https:         https.Value,
		HttpsCertType: httpsCertType.Value,
		HttpsCertPath: httpsCertPath.Value,
		HttpsKeyPath:  httpsKeyPath.Value,
	}, nil
}

func (s *SettingsService) Update(req model.UpdateSettingRequest) (*model.UpdateSettingResponse, error) {
	var response model.UpdateSettingResponse
	var scheme string
	switch req.Https {
	case "no":
		scheme = "http"
	case "yes":
		scheme = "https"
		// 检查type和path
		switch req.HttpsCertType {
		case "default":
		case "path":
			if len(req.HttpsCertPath) == 0 || len(req.HttpsKeyPath) == 0 {
				return &response, errors.New("invalid cert path or key path")
			}
		default:
			return &response, errors.New("invalid cert type")
		}
	default:
		return &response, errors.New("invalid https value")
	}

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return &response, err
	}

	// 开始事务
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			global.LOG.Error("Transaction failed: %v  - rollback", r)
		} else if err != nil {
			tx.Rollback() // 如果发生错误，回滚事务
			global.LOG.Error("Error Happend - rollback")
		}
	}()

	if err = s.updateBindIP(req.BindIP); err != nil {
		global.LOG.Error("Failed to save BindIP to %s: %v", req.BindIP, err)
		return &response, err
	}

	if err = s.updateBindPort(req.BindPort); err != nil {
		global.LOG.Error("Failed to save BindPort to %d: %v", req.BindPort, err)
		return &response, err
	}

	if err = s.updateBindDomain(req.BindDomain); err != nil {
		global.LOG.Error("Failed to save BindDomain to %s: %v", req.BindDomain, err)
		return &response, err
	}

	if err = s.updateHttps(req); err != nil {
		global.LOG.Error("Failed to save Https settings: %v", err)
		return &response, err
	}

	// 提交事务
	tx.Commit()

	var url string
	if len(req.BindDomain) == 0 {
		url = fmt.Sprintf("%s://%s:%d/manage/settings", scheme, host.Addr, req.BindPort)
	} else {
		url = fmt.Sprintf("%s://%s/manage/settings", scheme, req.BindDomain)
	}
	response.RedirectUrl = url

	go func() {
		// 发送 SIGTERM 信号给主进程，触发容器重启
		if err := syscall.Kill(1, syscall.SIGTERM); err != nil {
			global.LOG.Error("Failed to send termination signal: %v", err)
		}
	}()

	return &response, nil
}

func (s *SettingsService) updateBindIP(newIP string) error {
	if len(newIP) == 0 {
		return errors.New("invalid bind ip")
	}

	oldIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return err
	}
	if newIP == oldIP.Value {
		return nil
	}

	return SettingsRepo.Update("BindIP", newIP)
}

func (s *SettingsService) updateBindPort(newPort int) error {
	if newPort <= 0 || newPort > 65535 {
		return errors.New("server port must between 1 - 65535")
	}
	oldPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
	if err != nil {
		return err
	}
	newPortStr := strconv.Itoa(newPort)
	if newPortStr == oldPort.Value {
		return nil
	}

	if common.ScanPort(newPort) {
		return errors.New(constant.ErrPortInUsed)
	}

	// TODO: 处理port的更换（调用nftables）

	return SettingsRepo.Update("BindPort", newPortStr)
}

func (s *SettingsService) updateBindDomain(newDomain string) error {
	if len(newDomain) == 0 {
		return nil
	}
	oldDomain, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindDomain"))
	if err != nil {
		return err
	}
	if newDomain == oldDomain.Value {
		return nil
	}
	return SettingsRepo.Update("BindDomain", newDomain)
}

func (s *SettingsService) updateHttps(req model.UpdateSettingRequest) error {
	if err := SettingsRepo.Update("Https", req.Https); err != nil {
		return err
	}

	if req.Https == "yes" {
		if err := SettingsRepo.Update("HttpsCertType", req.HttpsCertType); err != nil {
			return err
		}

		if req.HttpsCertType == "path" {
			if err := SettingsRepo.Update("HttpsCertPath", req.HttpsCertPath); err != nil {
				return err
			}

			if err := SettingsRepo.Update("HttpsKeyPath", req.HttpsKeyPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SettingsService) Upgrade() error {
	newVersion := getLatestVersion()
	if len(newVersion) == 0 {
		return errors.New("failed to get latest version")
	}

	if global.Version == newVersion {
		return errors.New("already latest version")
	}

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return err
	}

	agentConn, err := conn.CENTER.GetAgentConn(&host)
	if err != nil {
		global.LOG.Error("Failed to get agent connection")
		return err
	}

	// 创建消息
	cmd := fmt.Sprintf("curl -sSL https://static.sensdata.com/idb/release/upgrade.sh -o /tmp/upgrade.sh && bash /tmp/upgrade.sh %s", newVersion)

	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		cmd,
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
		message.CmdMessage,
	)
	err = message.SendMessage(*agentConn, msg)
	if err != nil {
		global.LOG.Error("Failed to send command message: %v", err)
		return err
	}

	return nil
}
