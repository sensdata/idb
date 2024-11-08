package git

import (
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

type GitService struct{}

type IGitService interface {
	InitRepo(repoPath string, isBare bool) error
	GetFileList(repoPath string, relativePath string, extension string, page int, pageSize int) (*model.PageResult, error)
	GetFile(repoPath string, relativePath string) (*model.GitFile, error)
	Create(repoPath string, relativePath string, content string) error
	Update(repoPath string, relativePath string, content string) error
	Delete(repoPath string, relativePath string) error
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
	_, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		global.LOG.Info("Initializing Git repository at: %s", repoPath)
		_, err := git.PlainInit(repoPath, isBare)
		if err != nil {
			global.LOG.Error("Failed to init repo %s, %v", repoPath, err)
			return err
		}
	} else if err != nil {
		global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
		return err
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

		// 排除非文件或不符合后缀条件的文件
		if !info.Mode().IsRegular() || (extension != "" && !isValidExtension(info.Name(), extList)) {
			return nil
		}

		// 填充 GitFile 信息
		file := model.GitFile{
			Source:    path,
			Name:      info.Name(),
			Extension: filepath.Ext(info.Name()),
			Content:   "",
			Size:      info.Size(),
			ModTime:   info.ModTime(),
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
	filePath := filepath.Join(rootPath, relativePath)

	// 检查文件是否存在
	global.LOG.Info("Try get file  %s", filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		global.LOG.Error("File %s does not exist, %v", filePath, err)
		return nil, fmt.Errorf("file %s does not exist", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		global.LOG.Error("Failed to read file %s, %v", filePath, err)
		return nil, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		global.LOG.Error("Failed to get stat of file %s, %v", filePath, err)
		return nil, err
	}

	// 填充到结果
	gitFile := &model.GitFile{
		Source:    filePath,
		Name:      filepath.Base(relativePath),
		Extension: filepath.Ext(relativePath),
		Content:   string(content), // 将内容转换为字符串
		Size:      fileInfo.Size(),
		ModTime:   fileInfo.ModTime(),
	}

	return gitFile, nil
}

func (s *GitService) Create(repoPath string, relativePath string, content string) error {
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
	filePath := filepath.Join(rootPath, relativePath)

	global.LOG.Info("Try create file %s", filePath)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		global.LOG.Error("File %s already exists, %v", filePath, err)
		return fmt.Errorf("file %s already exists", filePath)
	}

	// 确保相对目录的创建（若存在目录部分）
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		global.LOG.Error("Failed to make dir for file %s, %v", filePath, err)
		return err
	}

	// 创建文件并写入内容
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		global.LOG.Error("Failed to write to file %s, %v", filePath, err)
		return err
	}

	// 将新创建的改动添加到 Git 索引
	_, err = worktree.Add(relativePath)
	if err != nil {
		global.LOG.Error("Failed to add %s to repo %s, %v", relativePath, repoPath, err)
		return err
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Add %s", relativePath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB",
			Email: "idb@sensdata.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: false,
	})
	if err != nil {
		global.LOG.Error("Failed to commit %s%s, %v", repoPath, relativePath, err)
		return err
	}

	return nil
}

func (s *GitService) Update(repoPath string, relativePath string, content string) error {
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
	filePath := filepath.Join(rootPath, relativePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		global.LOG.Error("File %s does not exists, %v", filePath, err)
		return fmt.Errorf("file %s does not exist", filePath)
	}

	// 更新文件内容
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		global.LOG.Error("Failed to write to file %s, %v", filePath, err)
		return err
	}

	// 将更新的文件添加到 Git 索引
	_, err = worktree.Add(relativePath)
	if err != nil {
		global.LOG.Error("Failed to add %s to repo %s, %v", relativePath, repoPath, err)
		return err
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Update %s", relativePath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB",
			Email: "idb@sensdata.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: false,
	})
	if err != nil {
		global.LOG.Error("Failed to commit %s%s, %v", repoPath, relativePath, err)
		return err
	}

	return nil
}

func (s *GitService) Delete(repoPath string, relativePath string) error {
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
	filePath := filepath.Join(rootPath, relativePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		global.LOG.Error("File does not exist %s: %v", filePath, err)
		return fmt.Errorf("file %s does not exist", filePath)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		global.LOG.Error("Failed to remove file %s: %v", filePath, err)
		return fmt.Errorf("failed to remove file %s: %w", filePath, err)
	}

	// 将删除的文件从 Git 索引中删除
	if _, err = worktree.Remove(relativePath); err != nil {
		global.LOG.Error("Failed to remove %s from index: %v", relativePath, err)
		return fmt.Errorf("failed to remove file from index: %w", err)
	}

	// 检查工作区状态，确保删除操作生效
	status, err := worktree.Status()
	if err != nil {
		global.LOG.Error("Failed to get worktree status: %v", err)
		return fmt.Errorf("failed to get worktree status: %w", err)
	}
	if status.IsClean() {
		global.LOG.Error("No changes to commit for  %s", relativePath)
		return fmt.Errorf("no changes to commit for %s", relativePath)
	}

	// 提交更改
	commitMsg := fmt.Sprintf("Delete %s", relativePath)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "IDB",
			Email: "idb@sensdata.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: false,
	})
	if err != nil {
		global.LOG.Error("Failed to commit %s%s, %v", repoPath, relativePath, err)
		return fmt.Errorf("failed to commit changes: %w", err)
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
	filePath := filepath.Join(repoPath, relativePath)

	// 将内容写入目标文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		global.LOG.Error("Failed to write to file %s: %v", filePath, err)
		return err
	}

	// 提交恢复的文件
	w, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Failed to get worktree %s: %v", repoPath, err)
		return err
	}

	// 添加文件到工作树
	if _, err := w.Add(relativePath); err != nil {
		global.LOG.Error("Failed to add %s to repo %s, %v", relativePath, repoPath, err)
		return err
	}

	// 提交恢复操作
	_, err = w.Commit(fmt.Sprintf("Restore %s to %s", relativePath, commitHash), &git.CommitOptions{
		All: true,
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
