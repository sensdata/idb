package sshman

import (
	"errors"
	"strings"

	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func (s *SSHMan) loadServiceName() (string, error) {
	if exist, _ := s.cmdHelper.IsExist("sshd"); exist {
		return "sshd", nil
	} else if exist, _ := s.cmdHelper.IsExist("ssh"); exist {
		return "ssh", nil
	}
	return "", errors.New("the ssh or sshd service is unavailable")
}

func (s *SSHMan) getSSHConfig() (*model.SSHInfo, error) {
	data := model.SSHInfo{
		AutoStart:              true,
		Status:                 constant.StatusEnable,
		Message:                "",
		Port:                   "22",
		ListenAddress:          "",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}

	serviceName, err := s.loadServiceName()
	if err != nil {
		data.Status = constant.StatusDisable
		data.Message = err.Error()
	} else {
		active, err := s.cmdHelper.IsActive(serviceName)
		if !active {
			data.Status = constant.StatusDisable
			data.Message = err.Error()
		} else {
			data.Status = constant.StatusEnable
		}
	}

	out, err := s.cmdHelper.RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		data.AutoStart = false
	} else {
		if out == "alias\n" {
			data.AutoStart, _ = s.cmdHelper.IsEnable("ssh")
		} else {
			data.AutoStart = out == "enabled\n"
		}
	}

	sshConf, err := s.cmdHelper.ReadFile(sshPath)
	if err != nil {
		data.Message = err.Error()
		data.Status = constant.StatusDisable
	}
	lines := strings.Split(string(sshConf.Content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Port ") {
			data.Port = strings.ReplaceAll(line, "Port ", "")
		}
		if strings.HasPrefix(line, "ListenAddress ") {
			itemAddr := strings.ReplaceAll(line, "ListenAddress ", "")
			if len(data.ListenAddress) != 0 {
				data.ListenAddress += ("," + itemAddr)
			} else {
				data.ListenAddress = itemAddr
			}
		}
		if strings.HasPrefix(line, "PasswordAuthentication ") {
			data.PasswordAuthentication = strings.ReplaceAll(line, "PasswordAuthentication ", "")
		}
		if strings.HasPrefix(line, "PubkeyAuthentication ") {
			data.PubkeyAuthentication = strings.ReplaceAll(line, "PubkeyAuthentication ", "")
		}
		if strings.HasPrefix(line, "PermitRootLogin ") {
			data.PermitRootLogin = strings.ReplaceAll(strings.ReplaceAll(line, "PermitRootLogin ", ""), "prohibit-password", "without-password")
		}
		if strings.HasPrefix(line, "UseDNS ") {
			data.UseDNS = strings.ReplaceAll(line, "UseDNS ", "")
		}
	}
	return &data, nil
}
