package files

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sensdata/idb/core/utils"
)

type ZipArchiver struct {
}

func NewZipArchiver() ShellArchiver {
	return &ZipArchiver{}
}

func (z ZipArchiver) Extract(filePath, dstDir string) error {
	if err := checkCmdAvailability("unzip"); err != nil {
		return err
	}
	return utils.ExecCmd(fmt.Sprintf("unzip -qo %s -d %s", filePath, dstDir))
}

func (z ZipArchiver) Compress(sourcePaths []string, dstFile string) error {
	var err error
	tmpFile := path.Join(os.TempDir(), fmt.Sprintf("%s%s.zip", utils.GenerateNonce(8), time.Now().Format("20060102150405")))
	op := NewFileOp()
	defer func() {
		_ = op.DeleteFile(tmpFile)
		if err != nil {
			_ = op.DeleteFile(dstFile)
		}
	}()
	baseDir := path.Dir(sourcePaths[0])
	relativePaths := make([]string, len(sourcePaths))
	for i, sp := range sourcePaths {
		relativePaths[i] = path.Base(sp)
	}
	cmdStr := fmt.Sprintf("zip -qr %s  %s", tmpFile, strings.Join(relativePaths, " "))
	if err = utils.ExecCmdWithDir(cmdStr, baseDir); err != nil {
		return err
	}
	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}
