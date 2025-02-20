package reader

// Reader 日志读取器接口
type Reader interface {
	// Read 从指定位置读取日志
	Read(offset int64) ([]byte, error)

	// Follow 持续读取新日志
	Follow() (<-chan []byte, error)

	// Close 关闭读取器
	Close() error
}
