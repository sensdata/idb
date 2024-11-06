package model

type FormField struct {
	Name       string      `yaml:"name"`
	Label      string      `yaml:"label"`
	Key        string      `yaml:"key"`
	Type       string      `yaml:"type"`
	Default    string      `yaml:"default"`
	Required   bool        `yaml:"required"`
	Hint       string      `yaml:"hint"`
	Options    []string    `yaml:"options,omitempty"`
	Validation *Validation `yaml:"validation,omitempty"`
}

type Validation struct {
	MinLength int    `yaml:"min_length,omitempty"`
	MaxLength int    `yaml:"max_length,omitempty"`
	Pattern   string `yaml:"pattern,omitempty"`
}

type Form struct {
	Fields []FormField `yaml:"form"`
}
