package action

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sensdata/idb/core/model"
	"github.com/shirou/gopsutil/v4/process"
)

var (
	cacheMu      sync.Mutex
	lastProcList []*model.ProcessSummary
	lastUpdate   time.Time
	cacheTTL     = 2 * time.Second
)

func ListProcesses(req model.ProcessListRequest) (*model.ProcessListResponse, error) {
	// 短时间缓存
	cacheMu.Lock()
	if time.Since(lastUpdate) < cacheTTL && len(lastProcList) > 0 {
		defer cacheMu.Unlock()
		pagedList := filterAndPage(lastProcList, req.Name, req.Pid, req.User, req.Page, req.PageSize)
		return &model.ProcessListResponse{
			Total: int64(len(pagedList)),
			Items: pagedList,
		}, nil
	}
	cacheMu.Unlock()

	var result model.ProcessListResponse

	// 获取所有进程
	procs, err := process.Processes()
	if err != nil {
		return &result, err
	}

	ctx := context.Background()
	var filteredProcs []*process.Process
	// Step1: 根据过滤条件筛选进程
	for _, p := range procs {
		name, _ := p.NameWithContext(ctx)
		user, _ := p.UsernameWithContext(ctx)
		if req.Name != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(req.Name)) {
			continue
		}
		if req.Pid != 0 && p.Pid != req.Pid {
			continue
		}
		if req.User != "" && user != req.User {
			continue
		}
		filteredProcs = append(filteredProcs, p)
	}

	// Step2: 顺序采集完整指标
	var fullList []*model.ProcessSummary
	for _, p := range filteredProcs {
		ps, _ := collectProcessSummary(p, ctx)
		fullList = append(fullList, ps)
	}

	// Step3: 排序
	switch strings.ToLower(req.Sort) {
	case "cpu":
		sort.Slice(fullList, func(i, j int) bool {
			return fullList[i].CPUPercent > fullList[j].CPUPercent
		})
	case "mem":
		sort.Slice(fullList, func(i, j int) bool {
			return fullList[i].MemPercent > fullList[j].MemPercent
		})
	}

	// Step4: 分页
	pagedList := filterAndPage(fullList, "", 0, "", req.Page, req.PageSize)

	// Step5: 更新缓存
	cacheMu.Lock()
	lastProcList = fullList
	lastUpdate = time.Now()
	cacheMu.Unlock()

	result.Total = int64(len(fullList))
	result.Items = pagedList

	return &result, nil
}

// collectProcessSummary 获取单个进程完整指标（顺序）
func collectProcessSummary(p *process.Process, ctx context.Context) (*model.ProcessSummary, error) {
	name, _ := p.NameWithContext(ctx)
	ppid, _ := p.PpidWithContext(ctx)
	cpuPercent, _ := p.CPUPercentWithContext(ctx)
	memPercent, _ := p.MemoryPercentWithContext(ctx)
	memInfo, _ := p.MemoryInfoWithContext(ctx)
	io, _ := p.IOCountersWithContext(ctx)
	user, _ := p.UsernameWithContext(ctx)
	conns, _ := p.ConnectionsWithContext(ctx)
	threads, _ := p.NumThreadsWithContext(ctx)
	create, _ := p.CreateTimeWithContext(ctx)

	return &model.ProcessSummary{
		PID:         p.Pid,
		PPID:        ppid,
		Name:        name,
		CPUPercent:  cpuPercent,
		MemPercent:  float64(memPercent),
		MemRSS:      memInfo.RSS,
		Swap:        memInfo.Swap,
		DiskRead:    io.ReadBytes,
		DiskWrite:   io.WriteBytes,
		Connections: len(conns),
		User:        user,
		Threads:     threads,
		CreateTime:  create,
	}, nil
}

// filterAndPage 过滤 + 分页逻辑
func filterAndPage(list []*model.ProcessSummary, name string, pid int32, user string, page, pageSize int) []*model.ProcessSummary {
	var filtered []*model.ProcessSummary
	for _, p := range list {
		if name != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
			continue
		}
		if pid != 0 && p.PID != pid {
			continue
		}
		if user != "" && p.User != user {
			continue
		}
		filtered = append(filtered, p)
	}

	// 分页逻辑
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []*model.ProcessSummary{}
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end]
}

// ----------------------
// 进程详情
// ----------------------

func GetProcessDetail(pid int32) (*model.ProcessDetail, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	name, _ := p.NameWithContext(ctx)
	status, _ := p.StatusWithContext(ctx)
	ppid, _ := p.PpidWithContext(ctx)
	threads, _ := p.NumThreadsWithContext(ctx)
	io, _ := p.IOCountersWithContext(ctx)
	user, _ := p.UsernameWithContext(ctx)
	create, _ := p.CreateTimeWithContext(ctx)
	cmdline, _ := p.CmdlineWithContext(ctx)
	exe, _ := p.ExeWithContext(ctx)
	cwd, _ := p.CwdWithContext(ctx)
	conns, _ := p.ConnectionsWithContext(ctx)

	basic := model.ProcessBasicInfo{
		Name:        name,
		Status:      strings.Join(status, ","),
		PID:         pid,
		PPID:        ppid,
		Threads:     threads,
		Connections: len(conns),
		DiskRead:    io.ReadBytes,
		DiskWrite:   io.WriteBytes,
		User:        user,
		CreateTime:  create,
		Cmdline:     cmdline,
		Exe:         exe,
		Cwd:         cwd,
	}

	memInfo, _ := p.MemoryInfoWithContext(ctx)
	memory := model.ProcessMemoryInfo{
		RSS:    memInfo.RSS,
		Swap:   memInfo.Swap,
		VMS:    memInfo.VMS,
		HWM:    memInfo.HWM,
		Data:   memInfo.Data,
		Stack:  memInfo.Stack,
		Locked: memInfo.Locked,
	}

	envs, _ := p.EnvironWithContext(ctx)

	var netConns []model.ProcessNetConn
	for _, c := range conns {
		proto := "tcp"
		if c.Type == 2 {
			proto = "udp"
		}
		netConns = append(netConns, model.ProcessNetConn{
			Protocol:   proto,
			LocalAddr:  c.Laddr.IP,
			LocalPort:  c.Laddr.Port,
			RemoteAddr: c.Raddr.IP,
			RemotePort: c.Raddr.Port,
			Status:     c.Status,
		})
	}

	files, _ := p.OpenFilesWithContext(ctx)
	var openFiles []model.ProcessOpenFile
	for _, f := range files {
		openFiles = append(openFiles, model.ProcessOpenFile{Path: f.Path})
	}

	return &model.ProcessDetail{
		Basic:     basic,
		Memory:    memory,
		Envs:      envs,
		NetConns:  netConns,
		OpenFiles: openFiles,
	}, nil
}

// ----------------------
// 结束进程
// ----------------------

func KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
