package files

import (
	"fmt"

	"github.com/sensdata/idb/core/utils"
)

type TarArchiver struct {
	Cmd          string
	CompressType CompressType
}

func NewTarArchiver(compressType CompressType) ShellArchiver {
	return &TarArchiver{
		Cmd:          "tar",
		CompressType: compressType,
	}
}

func (t TarArchiver) Extract(FilePath string, dstDir string) error {
	return utils.ExecCmd(fmt.Sprintf("%s %s %s -C %s", t.Cmd, t.getOptionStr("extract"), FilePath, dstDir))
}

func (t TarArchiver) Compress(sourcePaths []string, dstFile string) error {
	return nil
}

func (t TarArchiver) getOptionStr(Option string) string {
	switch t.CompressType {
	case Tar:
		if Option == "compress" {
			return "cvf"
		} else {
			return "xf"
		}
	}
	return ""
}
