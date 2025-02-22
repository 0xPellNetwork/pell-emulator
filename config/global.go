package config

var globalConfig = DefaultConfig()

func SetGlobalConfig(cfg *Config) {
	globalConfig = cfg
}

func GetGlobalConfig() *Config {
	return globalConfig
}
