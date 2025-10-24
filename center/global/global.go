package global

import (
	_ "embed"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/logstream"
	"github.com/sensdata/idb/core/model"

	"gorm.io/gorm"
)

var (
	Version    string = "0.0.1"
	Host       string = "127.0.0.1"
	DefaultKey string = ""

	LOG       *log.Log
	LogStream *logstream.LogStream
	DB        *gorm.DB
	VALID     *validator.Validate

	//go:embed certs/cert.pem
	CaCertPem []byte
	//go:embed certs/key.pem
	CaKeyPem []byte
	// self-sign cert & key
	CertPem []byte
	KeyPem  []byte

	// HostStatusCache 用于缓存host的状态信息
	HostStatusCache sync.Map
	// AgentStatusCache 用于缓存agent的状态信息
	AgentStatusCache sync.Map
)

// GetHostStatus 获取指定host的状态
func GetHostStatus(hostID uint) *model.HostStatus {
	if val, ok := HostStatusCache.Load(hostID); ok {
		return val.(*model.HostStatus)
	}
	return nil
}

// SetHostStatus 设置指定host的状态
func SetHostStatus(hostID uint, status *model.HostStatus) {
	HostStatusCache.Store(hostID, status)
}

// DeleteHostStatus 删除指定host的状态
func DeleteHostStatus(hostID uint) {
	HostStatusCache.Delete(hostID)
}

// GetAllHostStatus 获取所有host的状态
func GetAllHostStatus() map[uint]*model.HostStatus {
	result := make(map[uint]*model.HostStatus)
	HostStatusCache.Range(func(key, value interface{}) bool {
		result[key.(uint)] = value.(*model.HostStatus)
		return true
	})
	return result
}

func GetAgentStatus(hostID uint) *model.AgentStatus {
	if val, ok := AgentStatusCache.Load(hostID); ok {
		return val.(*model.AgentStatus)
	}
	return nil
}

func SetAgentStatus(hostID uint, status *model.AgentStatus) {
	AgentStatusCache.Store(hostID, status)
}

func DeleteAgentStatus(hostID uint) {
	AgentStatusCache.Delete(hostID)
}

func GetAllAgentStatus() map[uint]*model.AgentStatus {
	result := make(map[uint]*model.AgentStatus)
	AgentStatusCache.Range(func(key, value interface{}) bool {
		result[key.(uint)] = value.(*model.AgentStatus)
		return true
	})
	return result
}
