package config

// G 应用全局配置
var G GlobalConfig

// GlobalConfig 配置文件绑定结构体
type GlobalConfig struct {
	AppConf     AppConfig     `mapstructure:"app" yaml:"app"`
	StorageConf StorageConfig `mapstructure:"storage" yaml:"storage"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name                string       `mapstructure:"name" yaml:"name"`                                   // 应用名称
	Env                 string       `mapstructure:"env" yaml:"env"`                                     // 环境
	DataDir             string       `mapstructure:"data_dir" yaml:"data_dir"`                           // 应用数据目录
	TmdBinaryName       string       `mapstructure:"tmd_binary_name" yaml:"tmd_binary_name"`             // 二进制文件名称
	ScanIntervalMinutes int          `mapstructure:"scan_interval_minutes" yaml:"scan_interval_minutes"` // 扫描间隔
	Proxy               *ProxyConfig `mapstructure:"proxy" yaml:"proxy"`                                 // 代理配置
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	HttpProxy  string `mapstructure:"http_proxy" yaml:"http_proxy"`   // HTTP 代理
	HttpsProxy string `mapstructure:"https_proxy" yaml:"https_proxy"` // HTTPS 代理
	NoProxy    string `mapstructure:"no_proxy" yaml:"no_proxy"`       // 不走代理的地址
}

// StorageConfig 数据库配置
type StorageConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
	Database string `mapstructure:"database" yaml:"database"`
	Charset  string `mapstructure:"charset" yaml:"charset"`
}

// Init 初始化配置
func Init(conf *GlobalConfig) {
	G = *conf
}
