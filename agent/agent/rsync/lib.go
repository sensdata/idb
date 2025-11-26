package rsync

import (
	"errors"
	"path/filepath"

	"github.com/sensdata/idb/agent/agent/rsync/pkg"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// Simple API wrapper

type RsyncLib struct {
	m *pkg.Manager
}

func NewRsyncLib() *RsyncLib {
	storagePath := filepath.Join(constant.AgentDataDir, "rsync_tasks.json")
	storage, err := pkg.NewFileJSONStorage(storagePath)
	if err != nil {
		panic(err)
	}
	m := pkg.NewManager(storage, 1, 100)
	return &RsyncLib{m: m}
}

func (api *RsyncLib) Create(req model.RsyncClientCreateTaskRequest) (string, error) {
	if req.Name == "" {
		return "", errors.New("name required")
	}
	t := &pkg.RsyncTask{
		Name:          req.Name,
		Direction:     pkg.SyncDirection(req.Direction),
		LocalPath:     req.LocalPath,
		RemoteType:    pkg.RemoteType(req.RemoteType),
		RemoteHost:    req.RemoteHost,
		RemotePort:    req.RemotePort,
		Username:      req.Username,
		Password:      req.Password,
		SSHPrivateKey: req.SSHPrivateKey,
		RemotePath:    req.RemotePath,
		Module:        req.Module,
	}
	id, err := api.m.CreateTask(t, req.Enqueue)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (api *RsyncLib) List(req model.RsyncListTaskRequest) ([]*pkg.RsyncTask, error) {
	return api.m.ListTasks(req.Page, req.PageSize)
}

func (api *RsyncLib) All() ([]*pkg.RsyncTask, error) {
	return api.m.AllTasks()
}

func (api *RsyncLib) Detail(id string) (*pkg.RsyncTask, error) {
	return api.m.GetTask(id)
}

func (api *RsyncLib) Stop(id string) error {
	return api.m.StopTask(id)
}

func (api *RsyncLib) Retry(id string) error {
	return api.m.RetryTask(id)
}

func (api *RsyncLib) Delete(id string) error {
	return api.m.DeleteTask(id)
}
