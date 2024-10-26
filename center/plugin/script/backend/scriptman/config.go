package scriptman

// 配置结构
type Config struct {
	Script ScriptConfig `toml:"script"`
}

type ScriptConfig struct {
	DataPath string `toml:"data_path"`
	LogPath  string `toml:"log_path"`
}
