package plugin

type IdbPlugin interface {
	Initialize()
	Start()
	Release()
}
