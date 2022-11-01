package configs

type ApplicationConfig interface {
	GetKey() string
	GetDefaultValue() string
}

type OptionalConfig struct {
	Key     string
	Default string
}

func NewOptionalConfig(key string, defaultValue string) *OptionalConfig {
	return &OptionalConfig{key, defaultValue}
}

func (config OptionalConfig) GetKey() string {
	return config.Key
}

func (config OptionalConfig) GetDefaultValue() string {
	return config.Default
}

type MandatoryConfig struct {
	Key string
}

func NewMandatoryConfig(key string) *MandatoryConfig {
	return &MandatoryConfig{key}
}

func (config MandatoryConfig) GetKey() string {
	return config.Key
}

func (config MandatoryConfig) GetDefaultValue() string {
	panic("mandatory configuration must be specified manually")
}
