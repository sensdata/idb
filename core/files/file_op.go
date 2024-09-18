package files

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	cZip "github.com/klauspost/compress/zip"
	"github.com/mholt/archiver/v4"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/utils"
	"github.com/spf13/afero"
)

type FileOp struct {
	Fs afero.Fs
}

func NewFileOp() FileOp {
	return FileOp{
		Fs: afero.NewOsFs(),
	}
}

func (f FileOp) OpenFile(dst string) (fs.File, error) {
	return f.Fs.Open(dst)
}

func (f FileOp) GetContent(dst string) ([]byte, error) {
	afs := &afero.Afero{Fs: f.Fs}
	cByte, err := afs.ReadFile(dst)
	if err != nil {
		return nil, err
	}
	return cByte, nil
}

func (f FileOp) CreateDir(dst string, mode fs.FileMode) error {
	return f.Fs.MkdirAll(dst, mode)
}

func (f FileOp) CreateDirWithMode(dst string, mode fs.FileMode) error {
	if err := f.Fs.MkdirAll(dst, mode); err != nil {
		return err
	}
	return f.ChmodRWithMode(dst, mode, true)
}

func (f FileOp) CreateFile(dst string) error {
	if _, err := f.Fs.Create(dst); err != nil {
		return err
	}
	return nil
}

func (f FileOp) CreateFileWithMode(dst string, mode fs.FileMode) error {
	file, err := f.Fs.OpenFile(dst, os.O_CREATE, mode)
	if err != nil {
		return err
	}
	return file.Close()
}

func (f FileOp) LinkFile(source string, dst string, isSymlink bool) error {
	if isSymlink {
		osFs := afero.OsFs{}
		return osFs.SymlinkIfPossible(source, dst)
	} else {
		return os.Link(source, dst)
	}
}

func (f FileOp) DeleteDir(dst string) error {
	return f.Fs.RemoveAll(dst)
}

func (f FileOp) Stat(dst string) bool {
	info, _ := f.Fs.Stat(dst)
	return info != nil
}

func (f FileOp) DeleteFile(dst string) error {
	return f.Fs.Remove(dst)
}

func (f FileOp) CleanDir(dst string) error {
	return utils.ExecCmd(fmt.Sprintf("rm -rf %s/*", dst))
}

func (f FileOp) RmRf(dst string) error {
	return utils.ExecCmd(fmt.Sprintf("rm -rf %s", dst))
}

func (f FileOp) WriteChunkToFile(dst string, name string, offset int64, chunkSize int, chunk []byte) error {
	// 拼接完整路径
	fullPath := filepath.Join(dst, name)

	// 检查目录是否存在，不存在则创建
	if !f.Stat(dst) {
		if err := f.CreateDir(dst, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dst, err)
		}
	}

	// 检查文件是否存在，如果不存在则创建
	if !f.Stat(fullPath) {
		if err := f.CreateFileWithMode(fullPath, 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %v", fullPath, err)
		}
	}

	// 打开文件进行读写
	file, err := f.Fs.OpenFile(fullPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", fullPath, err)
	}
	defer file.Close()

	// 设置文件偏移
	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek to offset %d in file %s: %v", offset, fullPath, err)
	}

	// 检查 chunk 大小是否合适
	if len(chunk) > chunkSize {
		return fmt.Errorf("chunk size exceeds the specified limit of %d bytes", chunkSize)
	}

	// 写入数据
	written, err := file.Write(chunk)
	if err != nil {
		return fmt.Errorf("failed to write chunk to file %s: %v", fullPath, err)
	}

	// 检查写入的数据是否与 chunkSize 一致
	if written != chunkSize {
		return fmt.Errorf("written size %d does not match chunk size %d in file %s", written, chunkSize, fullPath)
	}

	return nil
}

func (f FileOp) WriteFile(dst string, in io.Reader, mode fs.FileMode) error {
	file, err := f.Fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = io.Copy(file, in); err != nil {
		return err
	}

	if _, err = file.Stat(); err != nil {
		return err
	}
	return nil
}

func (f FileOp) SaveFile(dst string, content string, mode fs.FileMode) error {
	if !f.Stat(path.Dir(dst)) {
		_ = f.CreateDir(path.Dir(dst), mode.Perm())
	}
	file, err := f.Fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(content)
	write.Flush()
	return nil
}

func (f FileOp) SaveFileWithByte(dst string, content []byte, mode fs.FileMode) error {
	if !f.Stat(path.Dir(dst)) {
		_ = f.CreateDir(path.Dir(dst), mode.Perm())
	}
	file, err := f.Fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.Write(content)
	write.Flush()
	return nil
}

func (f FileOp) ChownR(dst string, uid string, gid string, sub bool) error {
	cmdStr := fmt.Sprintf(`chown %s:%s "%s"`, uid, gid, dst)
	if sub {
		cmdStr = fmt.Sprintf(`chown -R %s:%s "%s"`, uid, gid, dst)
	}
	if utils.HasNoPasswordSudo() {
		cmdStr = fmt.Sprintf("sudo %s", cmdStr)
	}
	if msg, err := utils.ExecWithTimeOut(cmdStr, 10*time.Second); err != nil {
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

func (f FileOp) ChmodR(dst string, mode int64, sub bool) error {
	cmdStr := fmt.Sprintf(`chmod %v "%s"`, fmt.Sprintf("%04o", mode), dst)
	if sub {
		cmdStr = fmt.Sprintf(`chmod -R %v "%s"`, fmt.Sprintf("%04o", mode), dst)
	}
	if utils.HasNoPasswordSudo() {
		cmdStr = fmt.Sprintf("sudo %s", cmdStr)
	}
	if msg, err := utils.ExecWithTimeOut(cmdStr, 10*time.Second); err != nil {
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

func (f FileOp) ChmodRWithMode(dst string, mode fs.FileMode, sub bool) error {
	cmdStr := fmt.Sprintf(`chmod %v "%s"`, fmt.Sprintf("%o", mode.Perm()), dst)
	if sub {
		cmdStr = fmt.Sprintf(`chmod -R %v "%s"`, fmt.Sprintf("%o", mode.Perm()), dst)
	}
	if utils.HasNoPasswordSudo() {
		cmdStr = fmt.Sprintf("sudo %s", cmdStr)
	}
	if msg, err := utils.ExecWithTimeOut(cmdStr, 10*time.Second); err != nil {
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

func (f FileOp) Rename(oldName string, newName string) error {
	return f.Fs.Rename(oldName, newName)
}

type WriteCounter struct {
	Total    uint64
	Written  uint64
	Key      string
	Name     string
	Callback func(process Process)
}

type Process struct {
	Total   uint64  `json:"total"`
	Written uint64  `json:"written"`
	Percent float64 `json:"percent"`
	Name    string  `json:"name"`
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Written += uint64(n)

	// 计算当前下载百分比
	var percent float64
	if wc.Total > 0 {
		percent = (float64(wc.Written) / float64(wc.Total)) * 100
	}

	// 每次写入数据后，调用回调函数，传递当前的下载进度
	if wc.Callback != nil {
		wc.Callback(Process{
			Total:   wc.Total,
			Written: wc.Written,
			Percent: percent,
			Name:    wc.Name,
		})
	}

	return n, nil
}

func (w *WriteCounter) SaveProcess() Process {
	percentValue := 0.0
	if w.Total > 0 {
		percent := float64(w.Written) / float64(w.Total) * 100
		percentValue, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", percent), 64)
	}
	process := Process{
		Total:   w.Total,
		Written: w.Written,
		Percent: percentValue,
		Name:    w.Name,
	}
	return process
}

func (f FileOp) DownloadFileWithProcess(url, dst, key string, ignoreCertificate bool, callback func(process Process)) error {
	client := &http.Client{}
	if ignoreCertificate {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept-Encoding", "identity")

	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("获取下载文件 [%s] 出错, 错误: %s", dst, err.Error())
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		fmt.Printf("创建下载文件 [%s] 出错, 错误: %s", dst, err.Error())
		return err
	}
	defer out.Close()

	// 初始化 WriteCounter 并设置回调函数
	counter := &WriteCounter{
		Total:    uint64(resp.ContentLength),
		Key:      key,
		Name:     filepath.Base(dst),
		Callback: callback, // 设置回调函数
	}

	// 使用 io.TeeReader 来更新进度
	if _, err := io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		fmt.Printf("保存下载文件 [%s] 出错, 错误: %s", dst, err.Error())
		return err
	}

	// 下载完成后调用最终回调更新状态
	if callback != nil {
		callback(Process{
			Total:   counter.Total,
			Written: counter.Written,
			Percent: 100,
			Name:    counter.Name,
		})
	}

	return nil
}

func (f FileOp) DownloadFile(url, dst string) error {
	resp, err := utils.GetHttpRes(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("save download file [%s] error, err %s", dst, err.Error())
	}
	return nil
}

func (f FileOp) Cut(oldPaths []string, dst, name string, cover bool) error {
	for _, p := range oldPaths {
		var dstPath string
		if name != "" {
			dstPath = filepath.Join(dst, name)
			if f.Stat(dstPath) {
				dstPath = dst
			}
		} else {
			base := filepath.Base(p)
			dstPath = filepath.Join(dst, base)
		}
		coverFlag := ""
		if cover {
			coverFlag = "-f"
		}

		cmdStr := fmt.Sprintf(`mv %s '%s' '%s'`, coverFlag, p, dstPath)
		if err := utils.ExecCmd(cmdStr); err != nil {
			return err
		}
	}
	return nil
}

func (f FileOp) Mv(oldPath, dstPath string) error {
	cmdStr := fmt.Sprintf(`mv '%s' '%s'`, oldPath, dstPath)
	if err := utils.ExecCmd(cmdStr); err != nil {
		return err
	}
	return nil
}

func (f FileOp) Copy(src, dst string) error {
	if src = path.Clean("/" + src); src == "" {
		return os.ErrNotExist
	}
	if dst = path.Clean("/" + dst); dst == "" {
		return os.ErrNotExist
	}
	if src == "/" || dst == "/" {
		return os.ErrInvalid
	}
	if dst == src {
		return os.ErrInvalid
	}
	info, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return f.CopyDir(src, dst)
	}
	return f.CopyFile(src, dst)
}

func (f FileOp) CopyAndReName(src, dst, name string, cover bool) error {
	if src = path.Clean("/" + src); src == "" {
		return os.ErrNotExist
	}
	if dst = path.Clean("/" + dst); dst == "" {
		return os.ErrNotExist
	}
	if src == "/" || dst == "/" {
		return os.ErrInvalid
	}
	if dst == src {
		return os.ErrInvalid
	}

	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		dstPath := dst
		if name != "" && !cover {
			dstPath = filepath.Join(dst, name)
		}
		return utils.ExecCmd(fmt.Sprintf(`cp -rf '%s' '%s'`, src, dstPath))
	} else {
		dstPath := filepath.Join(dst, name)
		if cover {
			dstPath = dst
		}
		return utils.ExecCmd(fmt.Sprintf(`cp -f '%s' '%s'`, src, dstPath))
	}
}

func (f FileOp) CopyDir(src, dst string) error {
	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	dstDir := filepath.Join(dst, srcInfo.Name())
	if err = f.Fs.MkdirAll(dstDir, srcInfo.Mode()); err != nil {
		return err
	}
	return utils.ExecCmd(fmt.Sprintf(`cp -rf '%s' '%s'`, src, dst+"/"))
}

func (f FileOp) CopyFile(src, dst string) error {
	dst = filepath.Clean(dst) + string(filepath.Separator)
	return utils.ExecCmd(fmt.Sprintf(`cp -f '%s' '%s'`, src, dst+"/"))
}

func (f FileOp) GetDirSize(path string) (float64, error) {
	var size int64
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return float64(size), nil
}

func getFormat(cType CompressType) archiver.CompressedArchive {
	format := archiver.CompressedArchive{}
	switch cType {
	case Tar:
		format.Archival = archiver.Tar{}
	case TarGz, Gz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case SdkTarGz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case SdkZip, Zip:
		format.Archival = archiver.Zip{
			Compression: zip.Deflate,
		}
	case Bz2:
		format.Compression = archiver.Bz2{}
		format.Archival = archiver.Tar{}
	case Xz:
		format.Compression = archiver.Xz{}
		format.Archival = archiver.Tar{}
	}
	return format
}

func (f FileOp) Compress(srcRiles []string, dst string, name string, cType CompressType) error {
	format := getFormat(cType)

	fileMaps := make(map[string]string, len(srcRiles))
	for _, s := range srcRiles {
		base := filepath.Base(s)
		fileMaps[s] = base
	}

	if !f.Stat(dst) {
		_ = f.CreateDir(dst, 0755)
	}

	files, err := archiver.FilesFromDisk(nil, fileMaps)
	if err != nil {
		return err
	}
	dstFile := filepath.Join(dst, name)
	out, err := f.Fs.Create(dstFile)
	if err != nil {
		return err
	}

	switch cType {
	case Zip:
		if err := ZipFile(files, out); err == nil {
			return nil
		}
		_ = f.DeleteFile(dstFile)
		return NewZipArchiver().Compress(srcRiles, dstFile)
	default:
		err = format.Archive(context.Background(), out, files)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	}
	return nil
}

func isIgnoreFile(name string) bool {
	return strings.HasPrefix(name, "__MACOSX") || strings.HasSuffix(name, ".DS_Store") || strings.HasPrefix(name, "._")
}

func decodeGBK(input string) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded, _, err := transform.String(decoder, input)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

func (f FileOp) decompressWithSDK(srcFile string, dst string, cType CompressType) error {
	format := getFormat(cType)
	handler := func(ctx context.Context, archFile archiver.File) error {
		info := archFile.FileInfo
		if isIgnoreFile(archFile.Name()) {
			return nil
		}
		fileName := archFile.NameInArchive
		var err error
		if header, ok := archFile.Header.(cZip.FileHeader); ok {
			if header.NonUTF8 && header.Flags == 0 {
				fileName, err = decodeGBK(fileName)
				if err != nil {
					return err
				}
			}
		}
		filePath := filepath.Join(dst, fileName)
		if archFile.FileInfo.IsDir() {
			if err := f.Fs.MkdirAll(filePath, info.Mode()); err != nil {
				return err
			}
			return nil
		} else {
			parentDir := path.Dir(filePath)
			if !f.Stat(parentDir) {
				if err := f.Fs.MkdirAll(parentDir, info.Mode()); err != nil {
					return err
				}
			}
		}
		fr, err := archFile.Open()
		if err != nil {
			return err
		}
		defer fr.Close()
		fw, err := f.Fs.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer fw.Close()
		if _, err := io.Copy(fw, fr); err != nil {
			return err
		}

		return nil
	}
	input, err := f.Fs.Open(srcFile)
	if err != nil {
		return err
	}
	return format.Extract(context.Background(), input, nil, handler)
}

func (f FileOp) Decompress(srcFile string, dst string, cType CompressType) error {
	if err := f.decompressWithSDK(srcFile, dst, cType); err != nil {
		if cType == Tar || cType == Zip {
			shellArchiver, err := NewShellArchiver(cType)
			if err != nil {
				return err
			}
			return shellArchiver.Extract(srcFile, dst)
		}
		return err
	}
	return nil
}

func ZipFile(files []archiver.File, dst afero.File) error {
	zw := zip.NewWriter(dst)
	defer zw.Close()

	for _, file := range files {
		hdr, err := zip.FileInfoHeader(file)
		if err != nil {
			return err
		}
		hdr.Method = zip.Deflate
		hdr.Name = file.NameInArchive
		if file.IsDir() {
			if !strings.HasSuffix(hdr.Name, "/") {
				hdr.Name += "/"
			}
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if file.IsDir() {
			continue
		}

		if file.LinkTarget != "" {
			_, err = w.Write([]byte(filepath.ToSlash(file.LinkTarget)))
			if err != nil {
				return err
			}
		} else {
			fileReader, err := file.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(w, fileReader)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
