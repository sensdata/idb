package model

type App struct {
	ID             uint         `yaml:"-" json:"id"`
	Type           string       `yaml:"type" json:"type"`
	Name           string       `yaml:"name" json:"name"`
	DisplayName    string       `yaml:"display_name" json:"display_name"`
	Category       string       `yaml:"category" json:"category"`
	Tags           []string     `yaml:"tags" json:"tags"`
	Title          string       `yaml:"title" json:"title"`
	Description    string       `yaml:"description" json:"description"`
	Vendor         NameUrl      `yaml:"vendor" json:"vendor"`
	Packager       NameUrl      `yaml:"packager" json:"packager"`
	HasUpgrade     bool         `yaml:"-" json:"has_upgrade"`
	Versions       []AppVersion `json:"versions"`
	CurrentVersion string       `json:"current_version"`
	Form           Form         `json:"form"`
	Status         string       `json:"status"`
}
type AppVersion struct {
	ID             uint   `json:"id"`
	Version        string `json:"version"`
	UpdateVersion  string `json:"update_version"`
	ComposeContent string `json:"compose_content"`
	EnvContent     string `json:"env_content"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	CanUpgrade     bool   `json:"can_upgrade"`
}
type NameUrl struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
}
type RemoveApp struct {
	ID uint `json:"id"`
}
type QueryApp struct {
	PageInfo
	Name     string `form:"name" json:"name"`
	Category string `form:"category" json:"category"`
}
type QueryInstalledApp struct {
	PageInfo
	Name string `form:"name" json:"name"`
}
type QueryAppDetail struct {
	ID uint `json:"id"`
}
type InstallApp struct {
	ID             uint       `json:"id"`
	VersionID      uint       `json:"version_id"`
	ComposeName    string     `json:"compose_name"`
	ComposeContent string     `json:"compose_content"`
	EnvContent     string     `json:"env_content"`
	FormParams     []KeyValue `json:"form_params"`
	ExtraParams    []KeyValue `json:"extra_params"`
}

type UninstallApp struct {
	ComposeName string `json:"compose_name"`
}

type UpgradeApp struct {
	ID               uint   `json:"id"`
	UpgradeVersionID uint   `json:"upgrade_version_id"`
	ComposeName      string `json:"compose_name"`
}
