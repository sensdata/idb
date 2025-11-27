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

func (api *RsyncLib) Create(req model.RsyncClientCreateTaskRequest) (*model.RsyncCreateTaskResponse, error) {
	var rsp model.RsyncCreateTaskResponse
	if req.Name == "" {
		return &rsp, errors.New("name required")
	}
	t := &pkg.RsyncTask{
		Name:          req.Name,
		Direction:     pkg.SyncDirection(req.Direction),
		LocalPath:     req.LocalPath,
		RemoteType:    pkg.RemoteType(req.RemoteType),
		RemoteHost:    req.RemoteHost,
		RemotePort:    req.RemotePort,
		Username:      req.Username,
		AuthMode:      pkg.AuthMode(req.AuthMode),
		Password:      req.Password,
		SSHPrivateKey: req.SSHPrivateKey,
		RemotePath:    req.RemotePath,
		Module:        req.Module,
	}
	id, err := api.m.CreateTask(t, req.Enqueue)
	if err != nil {
		return &rsp, err
	}
	rsp.ID = id
	return &rsp, nil
}

func (api *RsyncLib) List(req model.RsyncListTaskRequest) (*model.RsyncClientListTaskResponse, error) {
	var resp model.RsyncClientListTaskResponse
	tasks, err := api.m.ListTasks(req.Page, req.PageSize)
	if err != nil {
		return &resp, err
	}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, &model.RsyncClientTask{
			ID:            t.ID,
			Name:          t.Name,
			Direction:     string(t.Direction),
			LocalPath:     t.LocalPath,
			RemoteType:    string(t.RemoteType),
			RemoteHost:    t.RemoteHost,
			RemotePort:    t.RemotePort,
			Username:      t.Username,
			Password:      t.Password,
			SSHPrivateKey: t.SSHPrivateKey,
			RemotePath:    t.RemotePath,
			Module:        t.Module,
			State:         string(t.State),
			Attempt:       t.Attempt,
		})
	}

	resp.Total = len(tasks)
	return &resp, nil
}

func (api *RsyncLib) All() (*model.RsyncClientListTaskResponse, error) {
	var resp model.RsyncClientListTaskResponse
	tasks, err := api.m.AllTasks()
	if err != nil {
		return &resp, err
	}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, &model.RsyncClientTask{
			ID:            t.ID,
			Name:          t.Name,
			Direction:     string(t.Direction),
			LocalPath:     t.LocalPath,
			RemoteType:    string(t.RemoteType),
			RemoteHost:    t.RemoteHost,
			RemotePort:    t.RemotePort,
			Username:      t.Username,
			Password:      t.Password,
			SSHPrivateKey: t.SSHPrivateKey,
			RemotePath:    t.RemotePath,
			Module:        t.Module,
			State:         string(t.State),
			Attempt:       t.Attempt,
		})
	}
	resp.Total = len(tasks)
	return &resp, nil
}

func (api *RsyncLib) Detail(id string) (*model.RsyncClientTask, error) {
	var resp model.RsyncClientTask
	t, err := api.m.GetTask(id)
	if err != nil {
		return &resp, err
	}
	return &model.RsyncClientTask{
		ID:            t.ID,
		Name:          t.Name,
		Direction:     string(t.Direction),
		LocalPath:     t.LocalPath,
		RemoteType:    string(t.RemoteType),
		RemoteHost:    t.RemoteHost,
		RemotePort:    t.RemotePort,
		Username:      t.Username,
		Password:      t.Password,
		SSHPrivateKey: t.SSHPrivateKey,
		RemotePath:    t.RemotePath,
		Module:        t.Module,
		State:         string(t.State),
		Attempt:       t.Attempt,
	}, nil
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
