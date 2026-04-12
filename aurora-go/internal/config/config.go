package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用全局配置结构体
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	RabbitMQ  RabbitMQConfig  `mapstructure:"rabbitmq"`
	ES        ESConfig        `mapstructure:"elasticsearch"`
	MinIO     MinIOConfig     `mapstructure:"minio"`       // MinIO对象存储配置
	OSS       *OSSConfig      `mapstructure:"oss,omitempty"` // 阿里云OSS配置(可选)
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	Email     EmailConfig     `mapstructure:"email"`      // 邮件SMTP配置
	BaiduSEO  BaiduSEOSConfig  `mapstructure:"baidu_seo"`   // 百度SEO推送配置
	Search    SearchConfig    `mapstructure:"search"`        // 搜索模式配置 (对标Java search.mode)
	Upload    UploadConfig    `mapstructure:"upload"`        // 上传模式配置 (对标Java upload.mode)
	QQ        QQConfig        `mapstructure:"qq"`         // QQ OAuth登录配置
	Agent     AgentConfig     `mapstructure:"agent"`       // 可选Agent配置
}

// Load 从YAML/ENV加载配置
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("AURORA")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件 %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置: %w", err)
	}

	return &cfg, nil
}

// ServerConfig HTTP服务配置
type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"` // debug/release
	SiteURL string `mapstructure:"site_url"` // 网站访问URL (用于邮件中的链接)
}

// GetSiteURL 获取网站URL (带默认值兜底)
func (c *ServerConfig) GetSiteURL() string {
	if c.SiteURL != "" {
		return c.SiteURL
	}
	return "https://www.aurora.blog"
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"` // 秒
}

// DSN 返回MySQL连接字符串
func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// RedisConfig Redis缓存配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回Redis连接地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RabbitMQConfig 消息队列配置
type RabbitMQConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	VHost        string `mapstructure:"vhost"`
	PrefetchCount int   `mapstructure:"prefetch_count"` // QoS预取计数 (默认10)
}

// URL 返回RabbitMQ AMQP连接URL
func (c *RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.User, c.Password, c.Host, c.Port, c.VHost)
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	URLs         []string `mapstructure:"urls"`           // ES节点URL列表
	Host         string   `mapstructure:"-"`               // 主节点地址 (从URLs[0]提取)
	IndexName    string   `mapstructure:"index_name"`      // 索引名 (如 article)
	IndexPrefix  string   `mapstructure:"index_prefix"`     // 索引前缀 (可选)
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	SniffEnabled bool     `mapstructure:"sniff_enabled"`
	Timeout      int      `mapstructure:"timeout"` // 请求超时(秒), 默认10s
}

// GetPrimaryURL 获取主ES节点URL
func (c *ESConfig) GetPrimaryURL() string {
	if len(c.URLs) > 0 {
		return c.URLs[0]
	}
	return ""
}

// MinIOConfig 对象存储配置 (对标 Java MinioProperties)
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	URL       string `mapstructure:"url"` // 外部访问URL (如 https://ws.aqi125.cn/)
}

// JWTConfig JWT认证配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int64  `mapstructure:"expire_time"` // 小时
	Issuer     string `mapstructure:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`       // debug/info/warn/error
	Format     string `mapstructure:"format"`      // json/console
	OutputPath string `mapstructure:"output_path"` // 输出路径，空值=stdout
}

// EmailConfig 邮件SMTP配置 (对标Java spring.mail.*)
type EmailConfig struct {
	Enabled          bool   `mapstructure:"enabled"`            // 是否启用邮件服务
	Host             string `mapstructure:"host"`               // SMTP服务器地址
	Port             int    `mapstructure:"port"`               // SMTP端口 (465/587/25)
	Username         string `mapstructure:"username"`           // 发件人邮箱账号
	Password         string `mapstructure:"password"`           // SMTP密码或授权码
	FromName         string `mapstructure:"from_name"`          // 发件人显示名称
	IsSSL            bool   `mapstructure:"is_ssl"`             // 是否使用SSL
	VerifyCodeExpire int64  `mapstructure:"verify_code_expire"`  // 验证码过期时间(分钟)
}

// BaiduSEOSConfig 百度SEO推送配置 (对标Java baiduSeo任务)
type BaiduSEOSConfig struct {
	Enabled bool   `mapstructure:"enabled"` // 是否启用百度SEO推送
	Token   string `mapstructure:"token"`   // 百度推送Token (从百度站长平台获取)
}

// OSSConfig 阿里云OSS配置 (对标Java OssConfigProperties)
type OSSConfig struct {
	URL            string `mapstructure:"url"`             // 访问URL前缀
	Endpoint       string `mapstructure:"endpoint"`        // OSS endpoint
	AccessKeyId     string `mapstructure:"access_key_id"`      // AccessKey ID
	AccessKeySecret string `mapstructure:"access_key_secret"`   // AccessKey Secret
	BucketName     string `mapstructure:"bucket_name"`      // Bucket名称
}

// SearchConfig 搜索模式配置 (对标Java search.mode)
type SearchConfig struct {
	Mode string `mapstructure:"mode"` // "mysql" 或 "elasticsearch"
}

// UploadConfig 上传模式配置 (对标Java upload.mode)  
type UploadConfig struct {
	Mode string `mapstructure:"mode"` // "minio" 或 "oss"
}

// AgentConfig AI Agent模块配置（可选）
type AgentConfig struct {
	Enabled bool          `mapstructure:"enabled"`           // 总开关
	LLM     AgentLLMConfig `mapstructure:"llm"`               // LLM模型配置
	Memory  AgentMemoryConfig `mapstructure:"memory"`         // 记忆配置
}

// AgentLLMConfig LLM提供商配置
type AgentLLMConfig struct {
	DefaultProvider string            `mapstructure:"default_provider"` // openai/deepseek/qwen/claude
	Providers       map[string]LLMProvider `mapstructure:"providers"`
}

// LLMProvider 单个LLM提供商配置
type LLMProvider struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"` // 可选，用于自定义端点
	Model   string `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens    int     `mapstructure:"max_tokens"`
}

// AgentMemoryConfig 记忆系统配置
type AgentMemoryConfig struct {
	Type     string `mapstructure:"type"`      // inmemory/redis
	MaxTurns int    `mapstructure:"max_turns"`  // 最大对话轮数
}

// QQConfig QQ OAuth登录配置 (对标Java QQConfigProperties)
type QQConfig struct {
	AppID         string `mapstructure:"app_id"`          // QQ应用AppID
	CheckTokenURL string `mapstructure:"check_token_url"`  // Token校验URL
	UserInfoURL   string `mapstructure:"user_info_url"`    // 用户信息获取URL
}
