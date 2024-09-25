package file

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sensdata/idb/agent/db"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/spf13/afero"
)

type FileService struct {
}

type IFileService interface {
	GetFileList(op model.FileOption) (*model.FileInfo, error)
	SearchUploadWithPage(req model.SearchUploadWithPage) (int64, interface{}, error)
	GetFileTree(op model.FileOption) ([]model.FileTree, error)
	Create(op model.FileCreate) error
	Delete(op model.FileDelete) error
	BatchDelete(op model.FileBatchDelete) error
	Compress(c model.FileCompress) error
	DeCompress(c model.FileDeCompress) error
	GetContent(op model.FileContentReq) (*model.FileInfo, error)
	SaveContent(edit model.FileEdit) error
	FileDownload(d model.FileDownload) (string, error)
	DirSize(req model.DirSizeReq) (*model.DirSizeRes, error)
	ChangeName(req model.FileRename) error
	Wget(w model.FileWget) (string, error)
	MvFile(m model.FileMove) error
	ChangeOwner(req model.FileRoleUpdate) error
	ChangeMode(op model.FileCreate) error
	BatchChangeModeAndOwner(op model.FileRoleReq) error

	GetFavoriteList(req model.PageInfo) (*model.PageResult, error)
	CreateFavorite(req model.FavoriteCreate) (*model.Favorite, error)
	DeleteFavorite(req model.FavoriteDelete) error
}

func NewIFileService() IFileService {
	return &FileService{}
}

func (f *FileService) GetFileList(op model.FileOption) (*model.FileInfo, error) {
	var fileInfo model.FileInfo
	if _, err := os.Stat(op.Path); err != nil && os.IsNotExist(err) {
		return &fileInfo, nil
	}
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return &fileInfo, err
	}
	fileInfo.FileInfo = *info
	return &fileInfo, nil
}

func (f *FileService) SearchUploadWithPage(req model.SearchUploadWithPage) (int64, interface{}, error) {
	var (
		files    []model.UploadInfo
		backData []model.UploadInfo
	)
	_ = filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			files = append(files, model.UploadInfo{
				CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
				Size:      int(info.Size()),
				Name:      info.Name(),
			})
		}
		return nil
	})
	total, start, end := len(files), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backData = make([]model.UploadInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		backData = files[start:end]
	}
	return int64(total), backData, nil
}

func (f *FileService) GetFileTree(op model.FileOption) ([]model.FileTree, error) {
	var treeArray []model.FileTree
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return nil, err
	}
	node := model.FileTree{
		ID:   utils.GenerateUuid(),
		Name: info.Name,
		Path: info.Path,
	}
	for _, v := range info.Items {
		if v.IsDir {
			node.Children = append(node.Children, model.FileTree{
				ID:   utils.GenerateUuid(),
				Name: v.Name,
				Path: v.Path,
			})
		}
	}
	return append(treeArray, node), nil
}

func (f *FileService) Create(op model.FileCreate) error {
	if files.IsInvalidChar(op.Path) {
		return errors.New("ErrInvalidChar")
	}
	fo := files.NewFileOp()
	if fo.Stat(op.Path) {
		return errors.New(constant.ErrFileIsExit)
	}
	mode := op.Mode
	if mode == 0 {
		fileInfo, err := os.Stat(filepath.Dir(op.Path))
		if err == nil {
			mode = int64(fileInfo.Mode().Perm())
		} else {
			mode = 0755
		}
	}
	if op.IsDir {
		return fo.CreateDirWithMode(op.Path, fs.FileMode(mode))
	}
	if op.IsLink {
		if !fo.Stat(op.LinkPath) {
			return errors.New(constant.ErrLinkPathNotFound)
		}
		return fo.LinkFile(op.LinkPath, op.Path, op.IsSymlink)
	}
	return fo.CreateFileWithMode(op.Path, fs.FileMode(mode))
}

func (f *FileService) Delete(op model.FileDelete) error {
	fo := files.NewFileOp()
	if op.ForceDelete {
		if op.IsDir {
			return fo.DeleteDir(op.Path)
		} else {
			return fo.DeleteFile(op.Path)
		}
	}
	if err := NewIRecycleBinService().Create(model.RecycleBinCreate{SourcePath: op.Path}); err != nil {
		return err
	}
	return db.FavoriteRepo.Delete(db.FavoriteRepo.WithByPath(op.Path))
}

func (f *FileService) BatchDelete(op model.FileBatchDelete) error {
	fo := files.NewFileOp()
	if op.IsDir {
		for _, file := range op.Paths {
			if err := fo.DeleteDir(file); err != nil {
				return err
			}
		}
	} else {
		for _, file := range op.Paths {
			if err := fo.DeleteFile(file); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FileService) ChangeMode(op model.FileCreate) error {
	fo := files.NewFileOp()
	return fo.ChmodR(op.Path, op.Mode, op.Sub)
}

func (f *FileService) BatchChangeModeAndOwner(op model.FileRoleReq) error {
	fo := files.NewFileOp()
	for _, path := range op.Paths {
		if !fo.Stat(path) {
			return errors.New(constant.ErrPathNotFound)
		}
		if err := fo.ChownR(path, op.User, op.Group, op.Sub); err != nil {
			return err
		}
		if err := fo.ChmodR(path, op.Mode, op.Sub); err != nil {
			return err
		}
	}
	return nil

}

func (f *FileService) ChangeOwner(req model.FileRoleUpdate) error {
	fo := files.NewFileOp()
	return fo.ChownR(req.Path, req.User, req.Group, req.Sub)
}

func (f *FileService) Compress(c model.FileCompress) error {
	fo := files.NewFileOp()
	if !c.Replace && fo.Stat(filepath.Join(c.Dst, c.Name)) {
		return errors.New(constant.ErrFileIsExit)
	}
	return fo.Compress(c.Files, c.Dst, c.Name, files.CompressType(c.Type))
}

func (f *FileService) DeCompress(c model.FileDeCompress) error {
	fo := files.NewFileOp()
	return fo.Decompress(c.Path, c.Dst, files.CompressType(c.Type))
}

func (f *FileService) GetContent(op model.FileContentReq) (*model.FileInfo, error) {
	info, err := files.NewFileInfo(files.FileOption{
		Path:   op.Path,
		Expand: true,
	})
	if err != nil {
		return &model.FileInfo{}, err
	}
	return &model.FileInfo{FileInfo: *info}, nil
}

func (f *FileService) SaveContent(edit model.FileEdit) error {
	info, err := files.NewFileInfo(files.FileOption{
		Path:   edit.Path,
		Expand: false,
	})
	if err != nil {
		return err
	}

	fo := files.NewFileOp()
	return fo.WriteFile(edit.Path, strings.NewReader(edit.Content), info.FileMode)
}

func (f *FileService) ChangeName(req model.FileRename) error {
	fo := files.NewFileOp()
	if !fo.Stat(req.Path) {
		return errors.New(constant.ErrPathNotFound)
	}

	if files.IsInvalidChar(req.NewName) {
		return errors.New("ErrInvalidChar")
	}

	return fo.Rename(req.Path, req.NewName)
}

func (f *FileService) Wget(w model.FileWget) (string, error) {
	fo := files.NewFileOp()
	key := "file-wget-" + utils.GenerateUuid()
	return key, fo.DownloadFileWithProcess(w.Url, filepath.Join(w.Path, w.Name), key, w.IgnoreCertificate, func(process files.Process) {})
}

func (f *FileService) MvFile(m model.FileMove) error {
	fo := files.NewFileOp()
	if !fo.Stat(m.NewPath) {
		return errors.New(constant.ErrPathNotFound)
	}
	for _, oldPath := range m.OldPaths {
		if !fo.Stat(oldPath) {
			return errors.New(constant.ErrFileNotFound)
		}
		if oldPath == m.NewPath || strings.Contains(m.NewPath, filepath.Clean(oldPath)+"/") {
			return errors.New(constant.ErrMovePathFailed)
		}
	}
	if m.Type == "cut" {
		return fo.Cut(m.OldPaths, m.NewPath, m.Name, m.Cover)
	}
	var errs []error
	if m.Type == "copy" {
		for _, src := range m.OldPaths {
			if err := fo.CopyAndReName(src, m.NewPath, m.Name, m.Cover); err != nil {
				errs = append(errs, err)
				global.LOG.Error("copy file [%s] to [%s] failed, err: %s", src, m.NewPath, err.Error())
			}
		}
	}

	var errString string
	for _, err := range errs {
		errString += err.Error() + "\n"
	}
	if errString != "" {
		return errors.New(errString)
	}
	return nil
}

func (f *FileService) FileDownload(d model.FileDownload) (string, error) {
	filePath := d.Paths[0]
	if d.Compress {
		tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().UnixNano()))
		if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
			return "", err
		}
		fo := files.NewFileOp()
		if err := fo.Compress(d.Paths, tempPath, d.Name, files.CompressType(d.Type)); err != nil {
			return "", err
		}
		filePath = filepath.Join(tempPath, d.Name)
	}
	return filePath, nil
}

func (f *FileService) DirSize(req model.DirSizeReq) (*model.DirSizeRes, error) {
	var (
		res model.DirSizeRes
	)
	if req.Path == "/proc" {
		return &res, nil
	}
	cmd := exec.Command("du", "-s", req.Path)
	output, err := cmd.Output()
	if err == nil {
		fields := strings.Fields(string(output))
		if len(fields) == 2 {
			var cmdSize int64
			_, err = fmt.Sscanf(fields[0], "%d", &cmdSize)
			if err == nil {
				res.Size = float64(cmdSize * 1024)
				return &res, nil
			}
		}
	}
	fo := files.NewFileOp()
	size, err := fo.GetDirSize(req.Path)
	if err != nil {
		return &res, err
	}
	res.Size = size
	return &res, nil
}

func (f *FileService) GetFavoriteList(req model.PageInfo) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}
	total, favorites, err := db.FavoriteRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return &pageResult, err
	}
	pageResult.Total = total
	pageResult.Items = favorites
	return &pageResult, nil
}

func (f *FileService) CreateFavorite(req model.FavoriteCreate) (*model.Favorite, error) {
	exist, _ := db.FavoriteRepo.GetFirst(db.FavoriteRepo.WithByPath(req.Path))
	if exist.ID > 0 {
		return nil, errors.New(constant.ErrFavoriteExist)
	}
	op := files.NewFileOp()
	if !op.Stat(req.Path) {
		return nil, errors.New(constant.ErrLinkPathNotFound)
	}
	openFile, err := op.OpenFile(req.Path)
	if err != nil {
		return nil, err
	}
	fileInfo, err := openFile.Stat()
	if err != nil {
		return nil, err
	}
	favorite := &model.Favorite{
		Name:  fileInfo.Name(),
		IsDir: fileInfo.IsDir(),
		Path:  req.Path,
	}
	if fileInfo.Size() <= 10*1024*1024 {
		afs := &afero.Afero{Fs: op.Fs}
		cByte, err := afs.ReadFile(req.Path)
		if err == nil {
			if len(cByte) > 0 && !files.DetectBinary(cByte) {
				favorite.IsTxt = true
			}
		}
	}
	if err := db.FavoriteRepo.Create(favorite); err != nil {
		return nil, err
	}
	return favorite, nil
}

func (f *FileService) DeleteFavorite(req model.FavoriteDelete) error {
	if err := db.FavoriteRepo.Delete(db.CommonRepo.WithByID(req.ID)); err != nil {
		return err
	}
	return nil
}
