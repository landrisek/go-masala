package masala

type Config struct {
		Host string
		Name string
		Password string
		Port int
		User string
}

type IConfig interface {
	GetHost() string
	GetName() string
	GetPassword() string
	GetPort() int
	GetUser() string
}

func (config *Config) GetHost() string {
	return config.Host
}

func (config *Config) GetName() string {
	return config.Name
}

func (config *Config) GetPassword() string {
	return config.Password
}

func (config *Config) GetPort() int {
	return config.Port
}

func (config *Config) GetUser() string {
	return config.User
}