package model

type App struct {
	ID          uint         `yaml:"-" json:"id"`
	Type        string       `yaml:"type" json:"type"`
	Name        string       `yaml:"name" json:"name"`
	DisplayName string       `yaml:"display_name" json:"display_name"`
	Category    string       `yaml:"category" json:"category"`
	Tags        []string     `yaml:"tags" json:"tags"`
	Title       string       `yaml:"title" json:"title"`
	Description string       `yaml:"description" json:"description"`
	Vendor      NameUrl      `yaml:"vendor" json:"vendor"`
	Packager    NameUrl      `yaml:"packager" json:"packager"`
	HasUpdate   bool         `yaml:"-" json:"has_update"`
	Versions    []AppVersion `json:"versions"`
	Form        Form         `json:"form"`
	Status      string       `json:"status"`
}
type AppVersion struct {
	ID             uint   `json:"id"`
	Version        string `json:"version"`
	UpdateVersion  string `json:"update_version"`
	ComposeContent string `json:"compose_content"`
}
type NameUrl struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
}
type QueryApp struct {
	PageInfo
	Name     string `json:"name"`
	Category string `json:"category"`
}
type QueryInstalledApp struct {
	PageInfo
	Name string `json:"name"`
}
type QueryAppDetail struct {
	ID uint `json:"id"`
}
type InstallApp struct {
	ID             uint       `json:"id"`
	VersionID      uint       `json:"version_id"`
	ComposeContent string     `json:"compose_content"`
	FormParams     []KeyValue `json:"form_params"`
	ExtraParams    []KeyValue `json:"extra_params"`
}

type UninstallApp struct {
	ID uint `json:"id"`
}
