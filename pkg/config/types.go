package config

type ServerConfig struct {
	Port string `mapstructure:"port" yaml:"port"`
}

type PostgreConfig struct {
	DSN string `mapstructure:"dsn" yaml:"dsn"`
}

type SecurityConfig struct {
	JwtSecretKey        string `mapstructure:"jwtSecretKey" yaml:"jwtSecretKey"`
	MinutesOfExpiration int    `mapstructure:"minutesOfExpiration" yaml:"minutesOfExpiration"`
}

type ApplicationConfig struct {
	Server   ServerConfig   `mapstructure:"server" yaml:"server"`
	Postgre  PostgreConfig  `mapstructure:"postgre" yaml:"postgre"`
	Security SecurityConfig `mapstructure:"security" yaml:"security"`
}
