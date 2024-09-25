package model

import (
	"time"

	"github.com/sensdata/idb/core/files"
)

type FileOption struct {
	HostID uint `json:"host_id"`
	files.FileOption
}

type FileContentReq struct {
	HostID uint   `json:"host_id"`
	Path   string `json:"path" validate:"required"`
}

type SearchUploadWithPage struct {
	PageInfo
	Path string `json:"path" validate:"required"`
}

type FileCreate struct {
	HostID    uint   `json:"host_id"`
	Path      string `json:"path" validate:"required"`
	Content   string `json:"content"`
	IsDir     bool   `json:"is_dir"`
	Mode      int64  `json:"mode"`
	IsLink    bool   `json:"is_link"`
	IsSymlink bool   `json:"is_symlink"`
	LinkPath  string `json:"link_path"`
	Sub       bool   `json:"sub"`
}

type FileRoleReq struct {
	HostID uint     `json:"host_id"`
	Paths  []string `json:"paths" validate:"required"`
	Mode   int64    `json:"mode" validate:"required"`
	User   string   `json:"user" validate:"required"`
	Group  string   `json:"group" validate:"required"`
	Sub    bool     `json:"sub"`
}

type FileDelete struct {
	HostID      uint   `json:"host_id"`
	Path        string `json:"path" validate:"required"`
	ForceDelete bool   `json:"force_delete"`
	IsDir       bool   `json:"is_dir"`
}

type FileBatchDelete struct {
	HostID uint     `json:"host_id"`
	Paths  []string `json:"paths" validate:"required"`
	IsDir  bool     `json:"is_dir"`
}

type FileCompress struct {
	HostID  uint     `json:"host_id"`
	Files   []string `json:"files" validate:"required"`
	Dst     string   `json:"dst" validate:"required"`
	Type    string   `json:"type" validate:"required"`
	Name    string   `json:"name" validate:"required"`
	Replace bool     `json:"replace"`
}

type FileDeCompress struct {
	HostID uint   `json:"host_id"`
	Dst    string `json:"dst"  validate:"required"`
	Type   string `json:"type"  validate:"required"`
	Path   string `json:"path" validate:"required"`
}

type FileEdit struct {
	HostID  uint   `json:"host_id"`
	Path    string `json:"path"  validate:"required"`
	Content string `json:"content"`
}

type FileRename struct {
	HostID  uint   `json:"host_id"`
	Path    string `json:"path"  validate:"required"`
	NewName string `json:"new_name" validate:"required"`
}

type FilePathCheck struct {
	Path string `json:"path" validate:"required"`
}

type FileWget struct {
	Url               string `json:"url" validate:"required"`
	Path              string `json:"path" validate:"required"`
	Name              string `json:"name" validate:"required"`
	IgnoreCertificate bool   `json:"ignore_certificate"`
}

type FileMove struct {
	HostID   uint     `json:"host_id"`
	Type     string   `json:"type" validate:"required"`
	OldPaths []string `json:"old_paths" validate:"required"`
	NewPath  string   `json:"new_path" validate:"required"`
	Name     string   `json:"name"`
	Cover    bool     `json:"cover"`
}

type FileUpload struct {
	Path      string `json:"path" validate:"required"` // 文件路径
	TotalSize int64  `json:"total_size"`               // 文件总大小（可选，Agent端可校验完整性）
	Offset    int64  `json:"offset"`                   // 当前文件块的起始偏移量
	ChunkSize int    `json:"chunk_size"`               // 当前文件块的大小
	Chunk     []byte `json:"chunk"`
}

type FileDownload struct {
	Paths    []string `json:"paths" validate:"required"`
	Type     string   `json:"type" validate:"required"`
	Name     string   `json:"name" validate:"required"`
	Compress bool     `json:"compress"`
}

type FileChunkDownload struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type DirSizeReq struct {
	HostID uint   `json:"host_id"`
	Path   string `json:"path" validate:"required"`
}

type FileProcessReq struct {
	Key string `json:"key"`
}

type FileRoleUpdate struct {
	HostID uint   `json:"host_id"`
	Path   string `json:"path" validate:"required"`
	User   string `json:"user" validate:"required"`
	Group  string `json:"group" validate:"required"`
	Sub    bool   `json:"sub"`
}

type FileReadByLineReq struct {
	Page     int    `json:"page" validate:"required"`
	PageSize int    `json:"page_size" validate:"required"`
	Type     string `json:"type" validate:"required"`
	ID       uint   `json:"id"`
	Name     string `json:"name"`
}

type FileExistReq struct {
	Name string `json:"name" validate:"required"`
	Dir  string `json:"dir" validate:"required"`
}

type FileInfo struct {
	files.FileInfo
}

type UploadInfo struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	CreatedAt string `json:"created_at"`
}

type FileTree struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Children []FileTree `json:"children"`
}

type DirSizeRes struct {
	Size float64 `json:"size" validate:"required"`
}

type FileProcessKeys struct {
	Keys []string `json:"keys"`
}

type FileWgetRes struct {
	Key string `json:"key"`
}

type FileLineContent struct {
	Content string `json:"content"`
	End     bool   `json:"end"`
	Path    string `json:"path"`
}

type FileExist struct {
	Exist bool `json:"exist"`
}

type RecycleBinCreate struct {
	SourcePath string `json:"source_path" validate:"required"`
}

type RecycleBinReduce struct {
	From  string `json:"from" validate:"required"`
	RName string `json:"rname" validate:"required"`
	Name  string `json:"name"`
}

type RecycleBinDTO struct {
	Name       string    `json:"name"`
	Size       int       `json:"size"`
	Type       string    `json:"type"`
	DeleteTime time.Time `json:"delete_time"`
	RName      string    `json:"rName"`
	SourcePath string    `json:"source_path"`
	IsDir      bool      `json:"is_dir"`
	From       string    `json:"from"`
}
