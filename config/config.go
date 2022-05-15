package config

import (
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

// ApplicationVersion represents the version of current application.
var ApplicationVersion string

// Config is a structure for values of the environment variables.
type Config struct {
	MigrationMode         bool          `default:"false"`
	GracefulShutdownDelay time.Duration `default:"15s"`

	App      ApplicationConfig
	Server   ServerConfig
	Bot      BotConfig
	Postgres PostgresConfig
	Log      *LoggerConfig
}

type ApplicationConfig struct {
	Version          string `default:"v0.0.1"`
	Name             string `default:"tg-bot"`
	DisableDBRestore bool   `default:"false"`
}

type ServerConfig struct {
	Addr string

	ViewsFolder     string `default:"./static/templates"`
	ViewsExt        string `default:".html"`
	StaticFolder    string `default:"./static"`
	StaticPrefix    string `default:"/"`
	TemplatesPrefix string `default:"templates"`

	Limiter ServerLimiterConfig
	Cache   ServerCacheConfig
}

type ServerLimiterConfig struct {
	MaxRequests int           `default:"100"`
	Expiration  time.Duration `default:"1m"`
}

type ServerCacheConfig struct {
	Expiration time.Duration `default:"1h"`
}

type BotConfig struct {
	Token string
}

type PostgresConfig struct {
	DSN           string
	LogLevel      string `default:"info"`
	MigrationsDir string `default:"./database/sql/migrations"`
}

type LoggerConfig struct {
	Level      string `default:"debug"`
	WithCaller int    `default:"1"`
}

// NewConfig loads values from environment variables and returns loaded configuration.
func NewConfig(file string) (*Config, error) {
	config := &Config{}
	loader := aconfig.LoaderFor(config, aconfig.Config{
		SkipFlags:        true,
		EnvPrefix:        "",
		AllowUnknownEnvs: true,
		AllFieldRequired: true,
		Files:            []string{file},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yml": aconfigyaml.New(),
		},
	})
	if err := loader.Load(); err != nil {
		return nil, err
	}
	if config.App.Version == "v0.0.1" || config.App.Version == "" {
		config.App.Version = ApplicationVersion
	}

	return config, nil
}
