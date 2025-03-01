package config

import (
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/yaml.v3"
)

var (
	ConfigInstance *Config
	once           sync.Once
)

// Config 总配置结构
type Config struct {
	Bot *BotConfig `yaml:"bot"`
	AI  *AIConfig  `yaml:"ai"`
}

// BotConfig 机器人配置
type BotConfig struct {
	Feishu   *FeishuConfig   `yaml:"feishu"`
	Weixin   *WeixinConfig   `yaml:"weixin"`
	Dingtalk *DingtalkConfig `yaml:"dingtalk"`
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	Enable               bool   `yaml:"enable"`
	AppID                string `yaml:"app_id"`
	AppSecret            string `yaml:"app_secret"`
	AppEncryptKey        string `yaml:"app_encrypt_key"`
	AppVerificationToken string `yaml:"app_verification_token"`
	BotName              string `yaml:"bot_name"`
	UnionID              string `yaml:"union_id"`
}

// WeixinConfig 微信配置
type WeixinConfig struct {
	Enable bool `yaml:"enable"`
}

// DingtalkConfig 钉钉配置
type DingtalkConfig struct {
	Enable bool `yaml:"enable"`
}

// AIConfig AI配置
type AIConfig struct {
	OpenAI *OpenAIConfig `yaml:"openai"`
	Volc   *VolcConfig   `yaml:"volc"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	Enable bool   `yaml:"enable"`
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
	APIURL string `yaml:"api_url"`
}

// VolcConfig 火山引擎配置
type VolcConfig struct {
	Enable bool   `yaml:"enable"`
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
	APIURL string `yaml:"api_url"`
}

// LoadConfig 从文件加载配置
func LoadConfig() error {
	var err error
	once.Do(func() {
		filename := getConfigPath()
		data, readErr := os.ReadFile(filename)
		if readErr != nil {
			err = readErr
			return
		}

		config := &Config{}
		if unmarshalErr := yaml.Unmarshal(data, config); unmarshalErr != nil {
			err = unmarshalErr
			return
		}

		// 验证配置
		if validateErr := validateConfig(config); validateErr != nil {
			err = validateErr
			return
		}

		ConfigInstance = config
	})
	return err
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	if ConfigInstance == nil {
		hlog.Fatal("配置未初始化")
	}
	return ConfigInstance
}

// GetFeishuConfig 获取飞书配置
func GetFeishuConfig() *FeishuConfig {
	cfg := GetConfig()
	if cfg.Bot == nil || cfg.Bot.Feishu == nil {
		return nil
	}
	return cfg.Bot.Feishu
}

// GetWeixinConfig 获取微信配置
func GetWeixinConfig() *WeixinConfig {
	cfg := GetConfig()
	if cfg.Bot == nil || cfg.Bot.Weixin == nil {
		return nil
	}
	return cfg.Bot.Weixin
}

// GetDingtalkConfig 获取钉钉配置
func GetDingtalkConfig() *DingtalkConfig {
	cfg := GetConfig()
	if cfg.Bot == nil || cfg.Bot.Dingtalk == nil {
		return nil
	}
	return cfg.Bot.Dingtalk
}

// GetOpenAIConfig 获取 OpenAI 配置
func GetOpenAIConfig() *OpenAIConfig {
	cfg := GetConfig()
	if cfg.AI == nil || cfg.AI.OpenAI == nil {
		return nil
	}
	return cfg.AI.OpenAI
}

// GetVolcConfig 获取火山引擎配置
func GetVolcConfig() *VolcConfig {
	cfg := GetConfig()
	if cfg.AI == nil || cfg.AI.Volc == nil {
		return nil
	}
	return cfg.AI.Volc
}

// IsFeishuEnabled 检查飞书是否启用
func IsFeishuEnabled() bool {
	cfg := GetFeishuConfig()
	return cfg != nil && cfg.Enable
}

// IsWeixinEnabled 检查微信是否启用
func IsWeixinEnabled() bool {
	cfg := GetWeixinConfig()
	return cfg != nil && cfg.Enable
}

// IsDingtalkEnabled 检查钉钉是否启用
func IsDingtalkEnabled() bool {
	cfg := GetDingtalkConfig()
	return cfg != nil && cfg.Enable
}

// IsOpenAIEnabled 检查 OpenAI 是否启用
func IsOpenAIEnabled() bool {
	cfg := GetOpenAIConfig()
	return cfg != nil && cfg.Enable
}

// IsVolcEnabled 检查火山引擎是否启用
func IsVolcEnabled() bool {
	cfg := GetVolcConfig()
	return cfg != nil && cfg.Enable
}

// 获取配置文件路径
func getConfigPath() string {
	// 获取环境变量，默认为 dev
	env := os.Getenv("ENV")
	if env == "" {
		hlog.Warnf("配置文件环境变量 ENV 不存在")
		os.Exit(1)
	}

	// 根据环境变量构造配置文件路径
	configPath := fmt.Sprintf("config_%s.yaml", env)

	// 如果文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		hlog.Warnf("配置文件 %s 不存在，使用默认配置 config_dev.yaml", configPath)
		return "config_dev.yaml"
	}

	return configPath
}

// validateConfig 验证配置是否有效
func validateConfig(cfg *Config) error {
	v := reflect.ValueOf(cfg).Elem()

	// 检查机器人服务
	hasBotEnabled := hasEnabledService(v.FieldByName("Bot"))
	// 检查 AI 服务
	hasAIEnabled := hasEnabledService(v.FieldByName("AI"))

	if !hasBotEnabled || !hasAIEnabled {
		return fmt.Errorf("配置无效: 必须至少启用一个机器人服务和一个 AI 服务")
	}

	return nil
}

// hasEnabledService 检查结构体中是否有启用的服务
func hasEnabledService(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return false
	}

	for i := range make([]struct{}, v.NumField()) {
		field := v.Field(i)

		// 如果字段是指针类型且不为空
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			// 获取结构体字段
			structField := field.Elem()
			// 查找 Enable 字段
			if enableField := structField.FieldByName("Enable"); enableField.IsValid() {
				if enableField.Bool() {
					return true
				}
			}
		}
	}
	return false
}
