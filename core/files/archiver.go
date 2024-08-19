package files

import (
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/utils"
)

type ShellArchiver interface {
	Extract(filePath, dstDir string) error
	Compress(sourcePaths []string, dstFile string) error
}

func NewShellArchiver(compressType CompressType) (ShellArchiver, error) {
	switch compressType {
	case Tar:
		if err := checkCmdAvailability("tar"); err != nil {
			return nil, err
		}
		return NewTarArchiver(compressType), nil
	case Zip:
		if err := checkCmdAvailability("zip"); err != nil {
			return nil, err
		}
		return NewZipArchiver(), nil
	default:
		return nil, errors.New("unsupported compress type")
	}
}

func checkCmdAvailability(cmdStr string) error {
	if utils.Which(cmdStr) {
		return nil
	}
	return errors.New(constant.ErrCmdNotFound)
}
