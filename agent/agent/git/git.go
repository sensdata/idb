package git

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

//go:embed git_sync.sh
var gitSyncShell []byte

type GitService struct{}

type IGitService interface {
	InitRepo(repoPath string, isBare bool) error
	SyncRepo(remoteUrl string, repoPath string) error
	GetFileList(repoPath string, relativePath string, extension string, page int, pageSize int) (*model.PageResult, error)
	GetFile(repoPath string, relativePath string) (*model.GitFile, error)
	Create(repoPath string, relativePath string, dir bool, content string) error
	Update(repoPath string, relativePath string, newRelativePath string, dir bool, content string) error
	Delete(repoPath string, relativePath string, dir bool) error
	Restore(repoPath string, relativePath string, commitHash string) error
	Log(repoPath string, relativePath string, page int, pageSize int) (*model.PageResult, error)
	Diff(repoPath string, relativePath string, commitHash string) (string, error)
}

func NewIGitService() IGitService {
	return &GitService{}
}

func (s *GitService) InitRepo(repoPath string, isBare bool) error {
	global.LOG.Info("init repo %s", repoPath)

	// 检查目录是否存在
	if err := utils.EnsurePaths([]string{repoPath}); err != nil {
		global.LOG.Error("Failed to create dir %s, %v", repoPath, err)
		return err
	}

	// 检查目录是否已经是一个仓库了
	repo, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		global.LOG.Info("Initializing Git repository at: %s", repoPath)
		_, err := git.PlainInit(repoPath, isBare)
		if err != nil {
			global.LOG.Error("Failed to init repo %s, %v", repoPath, err)
			return err
		}
		// 获取工作区路径
		worktree, err := repo.Worktree()
		if err != nil {
			global.LOG.Error("Failed to get work tree in repo %s, %v", repoPath, err)
			return err
		}
		rootPath := worktree.Filesystem.Root()

		// 创建 default/.gitkeep 文件，并确保目录存在
		gitkeepPath := filepath.Join(rootPath, "default", ".gitkeep")
		if err := os.MkdirAll(filepath.Dir(gitkeepPath), os.ModePerm); err != nil {
			global.LOG.Error("Failed to create directory for .gitkeep %s, %v", gitkeepPath, err)
			return err
		}
		if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
			global.LOG.Error("Failed to create .gitkeep at %s, %v", gitkeepPath, err)
			return err
		}

		// 添加并提交
		relPath := filepath.ToSlash(filepath.Join("default", ".gitkeep"))
		if _, err := worktree.Add(relPath); err != nil {
			global.LOG.Error("Failed to add file to repo %s, %v", repoPath, err)
			return err
		}

		_, err = worktree.Commit("Initialize repo with default directory", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "IDB",
				Email: "idb@sensdata.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			global.LOG.Error("Failed to commit to repo %s, %v", repoPath, err)
			return err
		}
	} else if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return err
	}

	return nil
}

func (s *GitService) SyncRepo(remoteUrl string, repoPath string) error {
	global.LOG.Info("sync from %s in %s", remoteUrl, repoPath)

	// 尝试打开已存在的仓库
	repo, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		// 仓库不存在，执行 clone
		global.LOG.Info("Repo does not exist, cloning from %s", remoteUrl)
		_, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:             remoteUrl,
			Progress:        nil,
			InsecureSkipTLS: true, // 跳过 SSL 验证
		})
		if err != nil {
			global.LOG.Error("Failed to clone repo: %v", err)
			return err
		}
		global.LOG.Info("Repo cloned successfully")
	} else if err != nil {
		global.LOG.Error("Failed to open repo: %v", err)
		return err
	} else {
		// 仓库存在，执行 pull
		global.LOG.Info("Repo exists, pulling latest changes")
		w, err := repo.Worktree()
		if err != nil {
			global.LOG.Error("Failed to get worktree: %v", err)
			return err
		}

		// 先重置工作区
		global.LOG.Info("Resetting worktree to HEAD")
		err = w.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
		if err != nil {
			global.LOG.Error("Failed to reset worktree: %v", err)
			return err
		}

		// 执行 pull
		err = w.Pull(&git.PullOptions{
			Force:           true,
			InsecureSkipTLS: true, // 跳过 SSL 验证
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			global.LOG.Error("Failed to pull repo: %v", err)
			return err
		}
		global.LOG.Info("Repo updated successfully")
	}

	return nil
}

func (s *GitService) GetFileList(repoPath string, relativePath string, extension string, page int, pageSize int) (*model.PageResult, error) {
	var pageResult model.PageResult
	var files []model.GitFile

	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return &pageResult, err
	}

	// 获取工作区的路径
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get work tree of repo %s, %v", repoPath, err)
		return &pageResult, err
	}
	rootPath := worktree.Filesystem.Root() // 获取工作区的根路径

	// 确定目标目录
	dirPath := rootPath
	if relativePath != "" {
		dirPath = filepath.Join(rootPath, relativePath)
	}

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		global.LOG.Error("Directory %s does not exist %v", dirPath, err)
		return &pageResult, fmt.Errorf("directory %s does not exist", dirPath)
	}

	// 遍历目录，获取文件信息
	global.LOG.Info("Scan file in directory %s", dirPath)
	extList := strings.Split(extension, ";") //支持多后缀筛选
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 排除 .git 目录及其内容
		if info.Name() == ".git" || strings.Contains(path, "/.git/") {
			return filepath.SkipDir
		}

		// 当 extension 为 "directory" 时，只收集目录
		if extension == "directory" {
			// 非目录
			if !info.IsDir() {
				return nil
			}
			// 排除根目录本身
			if path == dirPath {
				return nil
			}
		} else {
			// 排除非文件或不符合后缀条件的文件
			if !info.Mode().IsRegular() || (extension != "" && !isValidExtension(info.Name(), extList)) {
				return nil
			}
		}

		// 文件名和后缀
		ext := filepath.Ext(info.Name())
		name := strings.TrimSuffix(info.Name(), ext)

		// 读取文件内容
		var content string
		if !info.IsDir() {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				global.LOG.Error("Failed to read file %s, %v", path, err)
				return err
			}
			content = string(fileContent)
		}

		// 查看同目录下，是否存在.linked 文件(name.linked)
		var linked bool
		linkedFilePath := strings.TrimSuffix(path, ext) + ".linked"
		if _, err := os.Stat(linkedFilePath); err == nil {
			linked = true
		}

		// 填充 GitFile 信息
		file := model.GitFile{
			Source:    path,
			Name:      name,
			Extension: ext,
			Content:   content,
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Linked:    linked,
		}
		files = append(files, file)
		return nil
	})

	if err != nil {
		return &pageResult, err
	}

	// 分页处理
	totalFiles := int64(len(files))

	// 检查 page 和 pageSize 是否有效
	if page > 0 && pageSize > 0 {
		startIndex := (page - 1) * pageSize
		endIndex := startIndex + pageSize

		if startIndex >= int(totalFiles) {
			// 页数超出范围，返回空列表
			pageResult = model.PageResult{Total: totalFiles, Items: []model.GitFile{}}
			return &pageResult, nil
		}

		if endIndex > int(totalFiles) {
			endIndex = int(totalFiles)
		}

		pageResult = model.PageResult{
			Total: totalFiles,
			Items: files[startIndex:endIndex],
		}
	} else {
		// 如果 page 和 pageSize 无效，返回所有文件
		pageResult = model.PageResult{
			Total: totalFiles,
			Items: files,
		}
	}

	return &pageResult, nil
}

func isValidExtension(fileName string, extensions []string) bool {
	for _, ext := range extensions {
		if filepath.Ext(fileName) == ext {
			return true
		}
	}
	return false
}

func (s *GitService) GetFile(repoPath string, relativePath string) (*model.GitFile, error) {
	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return nil, err
	}

	// 获取工作区的路径
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get work tree in repo %s, %v", repoPath, err)
		return nil, err
	}
	rootPath := worktree.Filesystem.Root()

	// 确定目标文件的完整路径
	realPath := filepath.Join(rootPath, relativePath)

	// 检查文件是否存在
	global.LOG.Info("Try get file  %s", realPath)
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		global.LOG.Error("File %s does not exist, %v", realPath, err)
		return nil, fmt.Errorf("file %s does not exist", realPath)
	}

	// 读取文件内容
	content, err := os.ReadFile(realPath)
	if err != nil {
		global.LOG.Error("Failed to read file %s, %v", realPath, err)
		return nil, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		global.LOG.Error("Failed to get stat of file %s, %v", realPath, err)
		return nil, err
	}

	ext := filepath.Ext(fileInfo.Name())

	// 查看同目录下，是否存在.linked 文件(name.linked)
	var linked bool
	linkedFilePath := strings.TrimSuffix(realPath, ext) + ".linked"
	if _, err := os.Stat(linkedFilePath); err == nil {
		linked = true
	}

	// 填充到结果
	gitFile := &model.GitFile{
		Source:    realPath,
		Name:      strings.TrimSuffix(fileInfo.Name(), ext),
		Extension: ext,
		Content:   string(content),
		Size:      fileInfo.Size(),
		ModTime:   fileInfo.ModTime(),
		Linked:    linked,
	}

	return gitFile, nil
}

func (s *GitService) Create(repoPath string, relativePath string, dir bool, content string) error {
	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return err
	}

	// 获取工作区路径
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get work tree in repo %s, %v", repoPath, err)
		return err
	}
	rootPath := worktree.Filesystem.Root()

	// 确定目标文件的完整路径
	realPath := filepath.Join(rootPath, relativePath)

	// 检查文件是否已存在
	if _, err := os.Stat(realPath); err == nil {
		global.LOG.Error("File %s already exists, %v", realPath, err)
		return fmt.Errorf("file %s already exists", realPath)
	}

	if dir {
		// 目录操作只需要在文件系统层面处理
		if err := os.MkdirAll(realPath, os.ModePerm); err != nil {
			global.LOG.Error("Failed to create directory %s, %v", realPath, err)
			return err
		}
		global.LOG.Info("Created directory %s", realPath)
		return nil
	}

	// 确保父目录存在
	dirPath := filepath.Dir(realPath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		global.LOG.Error("Failed to create parent directory for %s, %v", realPath, err)
		return err
	}

	// 创建并写入文件
	if err := os.WriteFile(realPath, []byte(content), 0644); err != nil {
		global.LOG.Error("Failed to write file %s, %v", realPath, err)
		return err
	}
	global.LOG.Info("Created file %s", realPath)

	// 只有文件才需要添加到 Git 索引
	if _, err = worktree.Add(relativePath); err != nil {
		global.LOG.Error("Failed to add %s to repo %s, %v", relativePath, repoPath, err)
		return err
	}
	global.LOG.Info("Added %s to repo %s", relativePath, repoPath)

	// 提交更改
	commitMsg := fmt.Sprintf("Add %s", relativePath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB",
			Email: "idb@sensdata.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: true,
	})
	if err != nil {
		global.LOG.Error("Failed to commit %s%s, %v", repoPath, relativePath, err)
		return err
	}

	return nil
}

func (s *GitService) Update(repoPath string, relativePath string, newRelativePath string, dir bool, content string) error {
	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return err
	}

	// 获取工作区路径
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get work tree in repo %s, %v", repoPath, err)
		return err
	}
	rootPath := worktree.Filesystem.Root()

	if newRelativePath != "" && newRelativePath != relativePath {
		oldRealPath := filepath.Join(rootPath, relativePath)
		newRealPath := filepath.Join(rootPath, newRelativePath)

		global.LOG.Info("Moving from %s to %s", oldRealPath, newRealPath)

		// 检查源路径是否存在
		_, err := os.Stat(oldRealPath)
		if os.IsNotExist(err) {
			global.LOG.Error("Source path %s does not exist", oldRealPath)
			return fmt.Errorf("source path %s does not exist", oldRealPath)
		}

		// 检查目标路径是否已存在
		if _, err := os.Stat(newRealPath); err == nil {
			global.LOG.Error("Target path %s already exists", newRealPath)
			return fmt.Errorf("target path %s already exists", newRealPath)
		}

		// 检查目标路径是否合法
		if oldRealPath == newRealPath || strings.Contains(newRealPath, filepath.Clean(oldRealPath)+string(os.PathSeparator)) {
			global.LOG.Error("Invalid move operation from %s to %s", oldRealPath, newRealPath)
			return fmt.Errorf("invalid move operation")
		}

		// 确保目标目录的父目录存在
		if err := os.MkdirAll(filepath.Dir(newRealPath), os.ModePerm); err != nil {
			global.LOG.Error("Failed to create parent directory for %s: %v", newRealPath, err)
			return err
		}

		if dir {
			// 检查目录是否为空
			files, err := os.ReadDir(oldRealPath)
			if err != nil {
				return fmt.Errorf("failed to read directory: %w", err)
			}

			if len(files) == 0 {
				// 空目录只需要文件系统操作
				if err := os.Rename(oldRealPath, newRealPath); err != nil {
					return fmt.Errorf("failed to move directory: %w", err)
				}
				return nil
			}

			// 非空目录
			// 目录移动
			if err := os.Rename(oldRealPath, newRealPath); err != nil {
				return fmt.Errorf("failed to move directory: %w", err)
			}

			// 处理目录下的所有文件的 Git 索引
			status, err := worktree.Status()
			if err != nil {
				return fmt.Errorf("failed to get worktree status: %w", err)
			}

			for filePath := range status {
				if strings.HasPrefix(filePath, relativePath+"/") {
					newPath := strings.Replace(filePath, relativePath, newRelativePath, 1)
					if _, err = worktree.Remove(filePath); err != nil {
						global.LOG.Warn("Failed to remove old path %s from index: %v", filePath, err)
					}
					if _, err = worktree.Add(newPath); err != nil {
						return fmt.Errorf("failed to add new path %s to index: %w", newPath, err)
					}
				}
			}
		} else {
			// 如果是单个文件，先修改内容（如果需要）再移动
			if content != "" {
				if err := os.WriteFile(oldRealPath, []byte(content), 0644); err != nil {
					global.LOG.Error("Failed to write to file %s, %v", oldRealPath, err)
					return err
				}
			}

			// 执行文件系统移动
			if err := os.Rename(oldRealPath, newRealPath); err != nil {
				global.LOG.Error("Failed to move from %s to %s: %v", oldRealPath, newRealPath, err)
				return err
			}

			// 更新 Git 索引
			if _, err = worktree.Remove(relativePath); err != nil {
				global.LOG.Warn("Failed to remove old path %s from index: %v", relativePath, err)
			}
			if _, err = worktree.Add(newRelativePath); err != nil {
				return fmt.Errorf("failed to add new path to index: %w", err)
			}
		}
	} else if !dir && content != "" {
		// 仅更新文件内容
		targetPath := filepath.Join(rootPath, relativePath)
		if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to update file content: %w", err)
		}

		if _, err = worktree.Add(relativePath); err != nil {
			return fmt.Errorf("failed to add updated content to index: %w", err)
		}
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Update %s", relativePath)
	if newRelativePath != relativePath {
		commitMsg = fmt.Sprintf("Move %s to %s", relativePath, newRelativePath)
	}
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB",
			Email: "idb@sensdata.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: true,
	})
	if err != nil {
		global.LOG.Error("Failed to commit changes: %v", err)
		return err
	}

	return nil
}

func (s *GitService) Delete(repoPath string, relativePath string, dir bool) error {
	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s: %v", repoPath, err)
		return err
	}

	// 获取工作区路径
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get worktree %s: %v", repoPath, err)
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	rootPath := worktree.Filesystem.Root()

	// 确定目标文件的完整路径
	realPath := filepath.Join(rootPath, relativePath)

	// 检查文件是否存在
	_, err = os.Stat(realPath)
	if os.IsNotExist(err) {
		global.LOG.Error("File does not exist %s: %v", realPath, err)
		return fmt.Errorf("file %s does not exist", realPath)
	}

	// 处理目录的情况
	if dir {
		// 先检查目录内容
		files, err := os.ReadDir(realPath)
		if err != nil {
			global.LOG.Error("Failed to read directory %s: %v", realPath, err)
			return fmt.Errorf("failed to read directory: %w", err)
		}

		if len(files) == 0 {
			// 空目录直接删除
			if err := os.RemoveAll(realPath); err != nil {
				global.LOG.Error("Failed to remove empty directory %s: %v", realPath, err)
				return fmt.Errorf("failed to remove empty directory: %w", err)
			}
			global.LOG.Info("Empty directory %s removed", relativePath)
			return nil
		}

		// 非空目录
		// 删除文件系统中的目录
		if err := os.RemoveAll(realPath); err != nil {
			global.LOG.Error("Failed to remove directory %s: %v", realPath, err)
			return fmt.Errorf("failed to remove directory: %w", err)
		}

		// 从 Git 索引中删除文件
		status, err := worktree.Status()
		if err != nil {
			return fmt.Errorf("failed to get worktree status: %w", err)
		}
		for filePath := range status {
			if strings.HasPrefix(filePath, relativePath+"/") {
				if _, err = worktree.Remove(filePath); err != nil {
					global.LOG.Warn("Failed to remove %s from index: %v", filePath, err)
				}
			}
		}

		// 提交更改
		commitMsg := fmt.Sprintf("Delete directory %s and its contents", relativePath)
		_, err = worktree.Commit(commitMsg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "IDB",
				Email: "idb@sensdata.com",
				When:  time.Now(),
			},
			AllowEmptyCommits: true,
		})
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	} else {
		// 删除单个文件
		// 从文件系统中删除
		if err := os.Remove(realPath); err != nil && !os.IsNotExist(err) {
			global.LOG.Error("Failed to remove file %s: %v", realPath, err)
			return fmt.Errorf("failed to remove file: %w", err)
		}
		// 再从 Git 索引中删除
		if _, err = worktree.Remove(relativePath); err != nil {
			global.LOG.Error("Failed to remove %s from index: %v", relativePath, err)
			return fmt.Errorf("failed to remove file from index: %w", err)
		}

		// 提交更改
		commitMsg := fmt.Sprintf("Delete %s", relativePath)
		_, err = worktree.Commit(commitMsg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "IDB",
				Email: "idb@sensdata.com",
				When:  time.Now(),
			},
			AllowEmptyCommits: true,
		})
		if err != nil {
			global.LOG.Error("Failed to commit %s/%s, %v", repoPath, relativePath, err)
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	return nil
}

func (s *GitService) Restore(repoPath string, relativePath string, commitHash string) error {
	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return err
	}

	// 获取目标提交对象
	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		global.LOG.Error("Commit %s does not exist: %v", commitHash, err)
		return fmt.Errorf("commit %s does not exist", commitHash)
	}

	// 获取提交的树对象
	tree, err := commit.Tree()
	if err != nil {
		global.LOG.Error("Failed to get commit tree %s: %v", commitHash, err)
		return err
	}

	// 获取指定路径的文件对象
	file, err := tree.File(relativePath)
	if err != nil {
		global.LOG.Error("Failed to get file %s in commit %s: %v", relativePath, commitHash, err)
		return fmt.Errorf("file %s does not exist in commit %s", relativePath, commitHash)
	}

	// 读取文件内容
	content, err := file.Contents()
	if err != nil {
		global.LOG.Error("Failed to get file %s content in commit %s: %v", relativePath, commitHash, err)
		return err
	}

	// 确定目标文件的完整路径
	realPath := filepath.Join(repoPath, relativePath)

	// 将内容写入目标文件
	if err := os.WriteFile(realPath, []byte(content), 0644); err != nil {
		global.LOG.Error("Failed to write to file %s: %v", realPath, err)
		return err
	}

	// 提交恢复的文件
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get work tree in repo %s, %v", repoPath, err)
		return err
	}

	// 将更新的文件添加到 Git 索引
	_, err = worktree.Add(relativePath)
	if err != nil {
		global.LOG.Error("Failed to add %s to repo %s, %v", relativePath, repoPath, err)
		return err
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Restore %s to %s", relativePath, commitHash)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB",
			Email: "idb@sensdata.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: true,
	})
	if err != nil {
		global.LOG.Error("Failed to commit %s%s, %v", repoPath, relativePath, err)
		return err
	}

	return nil
}

func (s *GitService) Log(repoPath string, relativePath string, page int, pageSize int) (*model.PageResult, error) {
	var pageResult model.PageResult
	var commits []model.GitCommit

	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return &pageResult, err
	}

	// 获取引用
	ref, err := repo.Reference(plumbing.HEAD, true)
	if err != nil {
		global.LOG.Error("Failed to get reference of repo %s: %v", repoPath, err)
		return &pageResult, err
	}

	// 获取历史记录
	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		global.LOG.Error("Failed to get log iter of repo %s: %v", repoPath, err)
		return &pageResult, err
	}

	// 遍历历史记录
	err = iter.ForEach(func(c *object.Commit) error {
		// 获取当前提交的文件列表
		files, err := c.Files()
		if err != nil {
			return err
		}

		// 遍历文件，检查是否包含指定文件的更改
		err = files.ForEach(func(file *object.File) error {
			if file.Name == relativePath {
				commits = append(
					commits,
					model.GitCommit{
						CommitHash: c.Hash.String(), // 提交Hash
						Author:     c.Author.Name,   // 添加作者
						Email:      c.Author.Email,  // 添加作者邮箱
						Time:       c.Author.When,   // 添加时间
						Message:    c.Message,       // 添加提交信息
					},
				)
			}
			return nil
		})

		return err
	})

	if err != nil {
		global.LOG.Error("Failed to scan log for %s: %v", relativePath, err)
		return &pageResult, err
	}

	// 分页处理
	totalFiles := int64(len(commits))

	// 检查 page 和 pageSize 是否有效
	if page > 0 && pageSize > 0 {
		startIndex := (page - 1) * pageSize
		endIndex := startIndex + pageSize

		if startIndex >= int(totalFiles) {
			// 页数超出范围，返回空列表
			pageResult = model.PageResult{Total: totalFiles, Items: []model.GitCommit{}}
			return &pageResult, nil
		}

		if endIndex > int(totalFiles) {
			endIndex = int(totalFiles)
		}

		pageResult = model.PageResult{
			Total: totalFiles,
			Items: commits[startIndex:endIndex],
		}
	} else {
		// 如果 page 和 pageSize 无效，返回所有文件
		pageResult = model.PageResult{
			Total: totalFiles,
			Items: commits,
		}
	}

	return &pageResult, nil
}

func (s *GitService) Diff(repoPath string, relativePath string, commitHash string) (string, error) {
	// 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return "", err
	}

	// 获取当前文件版本
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get worktree %s: %v", repoPath, err)
		return "", err
	}
	rootPath := worktree.Filesystem.Root()

	currentFilePath := filepath.Join(rootPath, relativePath)
	currentContent, err := os.ReadFile(currentFilePath)
	if err != nil {
		global.LOG.Error("Failed to read file %s: %v", currentFilePath, err)
		return "", err
	}

	// 获取历史版本的提交对象
	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		global.LOG.Error("Commit %s of file %s does not exist: %v", commitHash, repoPath, err)
		return "", fmt.Errorf("commit %s does not exist", commitHash)
	}

	// 获取指定提交中的文件内容
	file, err := commit.File(relativePath)
	if err != nil {
		global.LOG.Error("File %s does not exist in commit %s: %v", relativePath, commitHash, err)
		return "", fmt.Errorf("file %s does not exist in commit %s", relativePath, commitHash)
	}

	historicalContent, err := file.Contents()
	if err != nil {
		global.LOG.Error("Failed to get file %s content in commit %s: %v", relativePath, commitHash, err)
		return "", err
	}

	// 比较内容并生成差异
	diff := diffText(string(currentContent), historicalContent)

	return diff, nil
}

// diffText 简单实现比较两个文本并返回差异，使用html格式
func diffText(currentContent string, historicalContent string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(currentContent, historicalContent, false)
	dmp.DiffCleanupSemantic(diffs)

	// 将差异转换为 HTML 格式
	html := dmp.DiffPrettyHtml(diffs)
	return html
}
