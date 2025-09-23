package files

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/spf13/afero"
)

type FileInfo struct {
	Fs         afero.Fs    `json:"-"`
	Path       string      `json:"path"`
	Name       string      `json:"name"`
	User       string      `json:"user"`
	Group      string      `json:"group"`
	Uid        string      `json:"uid"`
	Gid        string      `json:"gid"`
	Extension  string      `json:"extension"`
	Content    string      `json:"content"`
	Size       int64       `json:"size"`
	IsDir      bool        `json:"is_dir"`
	IsSymlink  bool        `json:"is_symlink"`
	IsHidden   bool        `json:"is_hidden"`
	LinkPath   string      `json:"link_path"`
	Type       string      `json:"type"`
	Mode       string      `json:"mode"`
	MimeType   string      `json:"mime_type"`
	UpdateTime time.Time   `json:"update_time"`
	ModTime    time.Time   `json:"mod_time"`
	FileMode   os.FileMode `json:"-"`
	Items      []*FileInfo `json:"items"`
	ItemTotal  int         `json:"item_total"`
	FavoriteID uint        `json:"favorite_id"`
}

type FileOption struct {
	Path       string `json:"path"`
	Search     string `json:"search"`
	ContainSub bool   `json:"contain_sub"`
	Expand     bool   `json:"expand"`
	Dir        bool   `json:"dir"`
	ShowHidden bool   `json:"show_hidden"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	SortBy     string `json:"sort_by"`
	SortOrder  string `json:"sort_order"`
}

type FileSearchInfo struct {
	Path string `json:"path"`
	fs.FileInfo
}

func NewFileInfo(op FileOption) (*FileInfo, error) {
	var appFs = afero.NewOsFs()

	info, err := appFs.Stat(op.Path)
	if err != nil {
		return nil, err
	}

	file := &FileInfo{
		Fs:        appFs,
		Path:      op.Path,
		Name:      info.Name(),
		IsDir:     info.IsDir(),
		FileMode:  info.Mode(),
		ModTime:   info.ModTime(),
		Size:      info.Size(),
		IsSymlink: IsSymlink(info.Mode()),
		Extension: filepath.Ext(info.Name()),
		IsHidden:  IsHidden(op.Path),
		Mode:      fmt.Sprintf("%04o", info.Mode().Perm()),
		User:      GetUsername(info.Sys().(*syscall.Stat_t).Uid),
		Uid:       strconv.FormatUint(uint64(info.Sys().(*syscall.Stat_t).Uid), 10),
		Gid:       strconv.FormatUint(uint64(info.Sys().(*syscall.Stat_t).Gid), 10),
		Group:     GetGroup(info.Sys().(*syscall.Stat_t).Gid),
		MimeType:  GetMimeType(op.Path),
	}

	if file.IsSymlink {
		file.LinkPath = GetSymlink(op.Path)
	}
	if op.Expand {
		if file.IsDir {
			if err := file.listChildren(op); err != nil {
				return nil, err
			}
			return file, nil
		} else {
			if err := file.getContent(); err != nil {
				return nil, err
			}
		}
	}
	return file, nil
}

func (f *FileInfo) search(search string, count int) (files []FileSearchInfo, total int, err error) {
	cmd := exec.Command("find", f.Path, "-name", fmt.Sprintf("*%s*", search))
	output, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	defer func() {
		_ = cmd.Wait()
		_ = cmd.Process.Kill()
	}()

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		info, err := os.Stat(line)
		if err != nil {
			continue
		}
		total++
		if total > count {
			continue
		}
		files = append(files, FileSearchInfo{
			Path:     line,
			FileInfo: info,
		})
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}

func sortFileList(list []FileSearchInfo, sortBy, sortOrder string) {
	switch sortBy {
	case "name":
		if sortOrder == "ascending" {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Name() < list[j].Name()
			})
		} else {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Name() > list[j].Name()
			})
		}
	case "size":
		if sortOrder == "ascending" {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Size() < list[j].Size()
			})
		} else {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Size() > list[j].Size()
			})
		}
	case "modTime":
		if sortOrder == "ascending" {
			sort.Slice(list, func(i, j int) bool {
				return list[i].ModTime().Before(list[j].ModTime())
			})
		} else {
			sort.Slice(list, func(i, j int) bool {
				return list[i].ModTime().After(list[j].ModTime())
			})
		}
	}
}

func (f *FileInfo) listChildren(option FileOption) error {
	afs := &afero.Afero{Fs: f.Fs}
	var (
		files []FileSearchInfo
		err   error
		total int
	)

	if option.Search != "" && option.ContainSub {
		files, total, err = f.search(option.Search, option.Page*option.PageSize)
		if err != nil {
			return err
		}
	} else {
		dirFiles, err := afs.ReadDir(f.Path)
		if err != nil {
			return err
		}
		var (
			dirs     []FileSearchInfo
			fileList []FileSearchInfo
		)
		for _, file := range dirFiles {
			info := FileSearchInfo{
				Path:     f.Path,
				FileInfo: file,
			}
			if file.IsDir() {
				dirs = append(dirs, info)
			} else {
				fileList = append(fileList, info)
			}
		}
		sortFileList(dirs, option.SortBy, option.SortOrder)
		sortFileList(fileList, option.SortBy, option.SortOrder)
		files = append(dirs, fileList...)
	}

	var items []*FileInfo
	for _, df := range files {
		if option.Dir && !df.IsDir() {
			continue
		}
		name := df.Name()
		fPath := path.Join(df.Path, df.Name())
		if option.Search != "" {
			if option.ContainSub {
				fPath = df.Path
				name = strings.TrimPrefix(strings.TrimPrefix(fPath, f.Path), "/")
			} else {
				lowerName := strings.ToLower(name)
				lowerSearch := strings.ToLower(option.Search)
				if !strings.Contains(lowerName, lowerSearch) {
					continue
				}
			}
		}
		if !option.ShowHidden && IsHidden(name) {
			continue
		}
		f.ItemTotal++
		isSymlink, isInvalidLink := false, false
		if IsSymlink(df.Mode()) {
			isSymlink = true
			info, err := f.Fs.Stat(fPath)
			if err == nil {
				df.FileInfo = info
			} else {
				isInvalidLink = true
			}
		}

		file := &FileInfo{
			Fs:        f.Fs,
			Name:      name,
			Size:      df.Size(),
			ModTime:   df.ModTime(),
			FileMode:  df.Mode(),
			IsDir:     df.IsDir(),
			IsSymlink: isSymlink,
			IsHidden:  IsHidden(fPath),
			Extension: filepath.Ext(name),
			Path:      fPath,
			Mode:      fmt.Sprintf("%04o", df.Mode().Perm()),
			User:      GetUsername(df.Sys().(*syscall.Stat_t).Uid),
			Group:     GetGroup(df.Sys().(*syscall.Stat_t).Gid),
			Uid:       strconv.FormatUint(uint64(df.Sys().(*syscall.Stat_t).Uid), 10),
			Gid:       strconv.FormatUint(uint64(df.Sys().(*syscall.Stat_t).Gid), 10),
		}
		if isSymlink {
			file.LinkPath = GetSymlink(fPath)
		}
		if df.Size() > 0 {
			file.MimeType = GetMimeType(fPath)
		}
		if isInvalidLink {
			file.Type = "invalid_link"
		}
		items = append(items, file)
	}
	if option.ContainSub {
		f.ItemTotal = total
	}
	start := (option.Page - 1) * option.PageSize
	end := option.PageSize + start
	var result []*FileInfo
	if start < 0 || start > f.ItemTotal || end < 0 || start > end {
		result = items
	} else {
		if end > f.ItemTotal {
			result = items[start:]
		} else {
			result = items[start:end]
		}
	}

	f.Items = result
	return nil
}

func (f *FileInfo) getContent() error {
	if IsBlockDevice(f.FileMode) {
		return errors.New(constant.ErrFileCanNotRead)
	}
	if f.Size > 10*1024*1024 {
		return errors.New(constant.ErrFileToLarge)
	}
	afs := &afero.Afero{Fs: f.Fs}
	cByte, err := afs.ReadFile(f.Path)
	if err != nil {
		return nil
	}
	if len(cByte) > 0 && DetectBinary(cByte) {
		return errors.New(constant.ErrReadBinFile)
	}
	f.Content = string(cByte)
	return nil
}

func (f *FileInfo) Part(lines int64, whence int) (string, error) {
	return f.getContentPart(lines, whence)
}

func (f *FileInfo) getContentPart(lines int64, whence int) (string, error) {
	if IsBlockDevice(f.FileMode) {
		return "", errors.New(constant.ErrFileCanNotRead)
	}

	file, err := f.Fs.Open(f.Path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	switch whence {
	case io.SeekStart:
		if lines <= 0 {
			return "", errors.New("offsetLines must be positive for SeekStart")
		}
		// 从头读取指定行数
		scanner := bufio.NewScanner(file)
		var result []string
		for lineCount := int64(0); lineCount < lines && scanner.Scan(); lineCount++ {
			result = append(result, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		content := strings.Join(result, "\n")
		if DetectBinary([]byte(content)) {
			return "", errors.New(constant.ErrReadBinFile)
		}
		return content, nil

	case io.SeekEnd:
		if lines >= 0 {
			return "", errors.New("offsetLines must be negative for SeekEnd")
		}
		absLines := -lines

		// 使用4KB的缓冲区
		const bufSize = 4 * 1024
		buf := make([]byte, bufSize)
		lines := make([]string, 0, absLines)
		lineCount := int64(0)
		pos := f.Size

		// 从文件末尾开始，向前读取
		for pos > 0 && lineCount < absLines {
			readSize := bufSize
			if pos < int64(bufSize) {
				readSize = int(pos)
			}
			pos -= int64(readSize)

			// 设置读取位置
			if _, err := file.(io.ReadSeeker).Seek(pos, io.SeekStart); err != nil {
				return "", err
			}

			// 读取数据块
			n, err := file.Read(buf[:readSize])
			if err != nil && err != io.EOF {
				return "", err
			}

			// 处理当前数据块中的行
			chunk := buf[:n]
			for i := len(chunk) - 1; i >= 0 && lineCount < absLines; i-- {
				if chunk[i] == '\n' || i == 0 {
					start := i
					if chunk[i] == '\n' {
						start++
					}
					if start < len(chunk) {
						line := string(chunk[start:])
						if len(line) > 0 {
							lines = append([]string{line}, lines...)
							lineCount++
						}
					}
					chunk = chunk[:i]
				}
			}

			// 处理跨缓冲区的行
			if len(chunk) > 0 && lineCount < absLines {
				lines = append([]string{string(chunk)}, lines...)
				lineCount++
			}
		}

		content := strings.Join(lines, "\n")
		if DetectBinary([]byte(content)) {
			return "", errors.New(constant.ErrReadBinFile)
		}
		return content, nil

	default:
		return "", errors.New("whence must be either io.SeekStart or io.SeekEnd")
	}
}

// DetectBinary 判断数据片段是否为二进制文件内容。
func DetectBinary(buf []byte) bool {
	if len(buf) == 0 {
		return false // 空文件视为文本
	}

	n := min(1024, len(buf)) // 最多检查前 1024 字节
	printable := 0
	nullCount := 0
	controlCount := 0

	for i := 0; i < n; i++ {
		b := buf[i]
		switch {
		case b == 0:
			nullCount++
		case (b >= 0x20 && b <= 0x7E) || b == 9 || b == 10 || b == 13:
			printable++
		case b < 0x09 || (b >= 0x0E && b <= 0x1F):
			controlCount++
		}
	}

	// 规则：
	// 1. null 字节比例大于 20% → 二进制
	if float64(nullCount)/float64(n) > 0.2 {
		return true
	}
	// 2. 控制字符比例大于 5% → 二进制
	if float64(controlCount)/float64(n) > 0.05 {
		return true
	}
	// 3. 可打印字符比例过低 (<30%) → 二进制
	if float64(printable)/float64(n) < 0.3 {
		return true
	}

	// 默认认为是文本
	return false
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type CompressType string

const (
	Zip      CompressType = "zip"
	Gz       CompressType = "gz"
	Bz2      CompressType = "bz2"
	Tar      CompressType = "tar"
	TarGz    CompressType = "tar.gz"
	Xz       CompressType = "xz"
	SdkZip   CompressType = "sdkZip"
	SdkTarGz CompressType = "sdkTarGz"
)
