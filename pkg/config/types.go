package config

type ServerConfig struct {
	Port string `mapstructure:"port" yaml:"port"`
}

type PostgreConfig struct {
	DSN string `mapstructure:"dsn" yaml:"dsn"`
}

type ApplicationConfig struct {
	Server  ServerConfig  `mapstructure:"server" yaml:"server"`
	Postgre PostgreConfig `mapstructure:"postgre" yaml:"postgre"`
}
