// config/config.go
package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config 结构体，用于存储读取的配置信息
type Config struct {
	ServerPort    string
	EnableSwagger bool
}

// LoadConfig 加载配置文件和环境变量
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // 配置文件名 (不带扩展名)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 配置文件路径

	// 将环境变量中的下划线转换为配置文件中的点
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 从环境变量中读取配置
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
	}

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.enableSwagger", false)

	config := &Config{
		ServerPort:    viper.GetString("server.port"),
		EnableSwagger: viper.GetBool("server.enableSwagger"),
	}

	return config, nil
}
