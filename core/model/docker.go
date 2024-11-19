package model

// container
type Container struct{}

type ContainerLog struct{}

type ContainerStatus struct{}

type CreateContainer struct{}

type UpdateContainer struct{}

// image
type Image struct{}

type PruneImage struct{}

type PullImage struct{}

type ImportImage struct{}

type BuildImage struct{}

type PushImage struct{}

type ExportImage struct{}

type SetImageTag struct{}

// volume
type Volume struct{}

type CreateVolume struct{}

// network
type Network struct{}

type CreateNetwork struct{}
