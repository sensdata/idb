package script

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type ScriptService struct {
	initialized  bool
	scriptConfig *model.Script
}

type IScriptService interface {
	Initialize() error
	GetScriptList(req model.QueryScript) (*model.PageResult, error)
	Create(req model.CreateScript) error
	Update(req model.UpdateScript) error
	Delete(req model.DeleteScript) error
}

func NewIScriptService() IScriptService {
	s := &ScriptService{
		initialized:  false,
		scriptConfig: nil,
	}
	if err := s.Initialize(); err != nil {
		global.LOG.Error("Failed to initialize ScriptService: %v", err)
	}
	return s
}

func (s *ScriptService) Initialize() error {

	confPath := filepath.Join(constant.CenterConfDir, "script", "script.toml")
	// 检查配置文件是否存在
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 创建配置文件并写入默认内容
		defaultConfig := "[script]\ndata_path = /var/lib/idb/data/script\nlog_path = /var/lib/idb/script.log\n"
		if err := os.WriteFile(confPath, []byte(defaultConfig), 0644); err != nil {
			global.LOG.Error("Failed to create script toml: %v", err)
			return err
		}
	}
	if _, err := toml.DecodeFile(confPath, &s.scriptConfig); err != nil {
		global.LOG.Error("Failed to load script toml: %v", err)
		return err
	}

	if err := initDirectories(s.scriptConfig.Script.DataPath); err != nil {
		global.LOG.Error("Failed to init script data dir: %v", err)
		return err
	}

	if err := initGitRepositories(s.scriptConfig.Script.DataPath); err != nil {
		global.LOG.Error("Failed to init script git repo: %v", err)
		return err
	}

	s.initialized = true
	return nil
}

func initDirectories(dataPath string) error {
	globalDir := filepath.Join(dataPath, "global")
	localDir := filepath.Join(dataPath, "local")

	// 检查并创建 global 和 local 目录
	if err := utils.EnsurePaths([]string{globalDir, localDir}); err != nil {
		return err
	}

	global.LOG.Info("Directories initialized: %s, %s", globalDir, localDir)
	return nil
}

func initGitRepositories(dataPath string) error {
	globalDir := filepath.Join(dataPath, "global")
	localDir := filepath.Join(dataPath, "local")

	// 初始化 global 仓库
	if err := initGitRepo(globalDir); err != nil {
		global.LOG.Error("Failed to init global repo %v", err)
		return err
	}

	// 初始化 local 仓库
	if err := initGitRepo(localDir); err != nil {
		global.LOG.Error("Failed to init local repo %v", err)
		return err
	}

	global.LOG.Info("script repos initialized")
	return nil
}

// 检查并初始化为 Git 仓库
func initGitRepo(path string) error {
	_, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		global.LOG.Info("Initializing Git repository at: %s", path)
		_, err := git.PlainInit(path, false) // 初始化非裸仓库
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (s *ScriptService) GetScriptList(req model.QueryScript) (*model.PageResult, error) {
	var pageResult model.PageResult

	var scripts []string

	// 打开仓库
	repoPath := filepath.Join(s.scriptConfig.Script.DataPath, req.Type)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return &pageResult, err
	}

	// 获取工作目录的路径
	worktree, err := repo.Worktree()
	if err != nil {
		return &pageResult, err
	}

	workDir := worktree.Filesystem.Root()
	if req.Category != "" {
		workDir += "/" + req.Category
	}

	// 遍历工作目录下的所有文件
	err = filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查文件扩展名，确认是 .sh 文件
		if !info.IsDir() && filepath.Ext(path) == ".sh" {
			scripts = append(scripts, path)
		}
		return nil
	})

	if err != nil {
		return &pageResult, err
	}

	// 处理分页 req.Page, req.PageSize
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start > len(scripts) {
		start = len(scripts)
	}
	if end > len(scripts) {
		end = len(scripts)
	}

	pageResult.Items = scripts[start:end]

	return &pageResult, nil
}

func (s *ScriptService) Create(req model.CreateScript) error {
	// 打开仓库
	repoPath := filepath.Join(s.scriptConfig.Script.DataPath, req.Type)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// 检查目录是否存在
	scriptPath := req.Name
	if req.Category != "" {
		scriptPath = filepath.Join(req.Category, req.Name)
	}
	fullScriptPath := filepath.Join(repoPath, scriptPath)
	dir := filepath.Dir(fullScriptPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// 创建并写入内容到脚本文件
	if err := os.WriteFile(fullScriptPath, []byte(req.Content), 0644); err != nil {
		return fmt.Errorf("failed to write script content: %w", err)
	}

	// 获取工作目录并添加文件到 Git
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	_, err = worktree.Add(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to add file to git: %w", err)
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Add script %s", scriptPath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB Script Manager",
			Email: "idb-scripts@sensdata.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	fmt.Println("Script created and committed:", fullScriptPath)
	return nil
}

func (s *ScriptService) Update(req model.UpdateScript) error {
	// 打开仓库
	repoPath := filepath.Join(s.scriptConfig.Script.DataPath, req.Type)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// 检查目录是否存在
	scriptPath := req.Name
	if req.Category != "" {
		scriptPath = filepath.Join(req.Category, req.Name)
	}
	fullScriptPath := filepath.Join(repoPath, scriptPath)
	dir := filepath.Dir(fullScriptPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullScriptPath)
	}

	// 创建并写入内容到脚本文件
	if err := os.WriteFile(fullScriptPath, []byte(req.Content), 0644); err != nil {
		return fmt.Errorf("failed to write script content: %w", err)
	}

	// 获取工作目录并添加文件到 Git
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	_, err = worktree.Add(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to add file to git: %w", err)
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Update script %s", scriptPath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB Script Manager",
			Email: "idb-scripts@sensdata.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	fmt.Println("Script updated and committed:", fullScriptPath)
	return nil
}

func (s *ScriptService) Delete(req model.DeleteScript) error {
	// 打开仓库
	repoPath := filepath.Join(s.scriptConfig.Script.DataPath, req.Type)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// 检查目录是否存在
	scriptPath := req.Name
	if req.Category != "" {
		scriptPath = filepath.Join(req.Category, req.Name)
	}
	fullScriptPath := filepath.Join(repoPath, scriptPath)

	// 检查文件是否存在
	if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullScriptPath)
	}

	// 删除脚本文件
	if err := os.Remove(fullScriptPath); err != nil {
		return fmt.Errorf("failed to delete script file: %w", err)
	}

	// 获取工作目录并添加删除操作到 Git
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	_, err = worktree.Remove(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to remove file from git: %w", err)
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Delete script %s", scriptPath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB Script Manager",
			Email: "idb-scripts@sensdata.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	fmt.Println("Script deleted and committed:", fullScriptPath)
	return nil
}
