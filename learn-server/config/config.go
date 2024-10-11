package config

import (
	"fmt"
	"learn/pkg/utils"
	"time"

	"github.com/spf13/viper"
)

// ServerConfig 包含服务器相关配置，包括是否启用 Swagger 和 CORS
type ServerConfig struct {
	Address        string        `mapstructure:"address"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	EnableSwagger  bool          `mapstructure:"enable_swagger"`
	AllowedOrigins []string      `mapstructure:"allowed_origins"` // 新增字段，用于配置允许的跨域源
}

// DatabaseConfig 包含数据库相关配置
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// JWTConfig 包含JWT相关配置
type JWTConfig struct {
	Secret               string        `mapstructure:"secret"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
}

// Config 是包含所有配置的主结构体
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	DefaultAdmin DefaultAdminConfig `mapstructure:"default_admin"`
}

// DefaultAdminConfig 包含默认 admin 用户配置
type DefaultAdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LoadConfig 使用 Viper 从配置文件和环境变量加载配置
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv() // 允许从环境变量读取配置

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 检查 JWT Secret 是否配置，如果未配置则询问是否自动生成
	if config.JWT.Secret == "" {
		fmt.Println("JWT Secret is not configured.")
		fmt.Println("Would you like to generate a new secret? (y/n)")

		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			return nil, err
		}

		if response == "y" || response == "Y" {
			config.JWT.Secret, err = utils.GenerateSecureToken(32)
			if err != nil {
				return nil, err
			}
			fmt.Printf("Generated JWT Secret: %s\n", config.JWT.Secret)
			// 可以选择将新的 Secret 保存到配置文件中
			err = saveConfigWithSecret(path, config.JWT.Secret)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("JWT Secret must be configured")
		}
	}

	return &config, nil
}

// 保存生成的 JWT Secret 到配置文件
func saveConfigWithSecret(path string, secret string) error {
	viper.Set("jwt.secret", secret)
	return viper.WriteConfigAs(path)
}
