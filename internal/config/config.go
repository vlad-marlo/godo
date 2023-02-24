package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/caarlos0/env/v7"
	"github.com/pelletier/go-toml/v2"
	"github.com/vlad-marlo/godo/internal/model"
	"go.uber.org/zap"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type (
	// Postgres ...
	Postgres struct {
		User     string `env:"POSTGRES_USER" toml:"user"`
		Password string `env:"POSTGRES_PASSWORD" toml:"password"`
		Port     uint   `env:"POSTGRES_PORT" toml:"port"`
		Addr     string `env:"POSTGRES_ADDR" toml:"host"`
		Name     string `env:"POSTGRES_NAME" toml:"name"`
		URI      string `env:"DB_URI"`
	}
	HTTPS struct {
		CertFile     string   `env:"CERT_FILE"`
		KeyFile      string   `env:"KEY_FILE"`
		AllowedHosts []string `env:"ALLOWED_HOSTS" envSeparator:":"`
	}
	Auth struct {
		AccessTokenLifeTime  time.Duration `env:"ACCESS_TOKEN_LIFETIME" envDefault:"720h" toml:"access_token_lifetime"`
		RefreshTokenLifeTime time.Duration `env:"REFRESH_TOKEN_LIFETIME" envDefault:"720h" toml:"refresh_token_lifetime"`
		PasswordDifficult    float64       `env:"MIN_PASSWORD_ENTROPY" toml:"password_difficult"`
		AuthTokenSize        int           `env:"AUTH_TOKEN_SIZE" toml:"auth_token_size"`
	}
	// Server is internal configuration of server.
	Server struct {
		Type       string `env:"SERVER_TYPE" toml:"type" valid:"in(grpc|http|https|GRPC|HTTP|HTTPS)" envDefault:"HTTP"`
		EnableHTTP bool   `toml:"-"`
		EnableGRPC bool   `toml:"-"`
		Salt       string `env:"ENCRYPT_SALT" valid:"required~use generated salt from the logs" toml:"salt"`
		SecretKey  string `env:"SECRET_KEY" valid:"required" toml:"secret_key"`
		IsDev      bool   `env:"IS_DEV" toml:"is_dev"`
		IsProd     bool   `env:"IS_PRODUCTION" toml:"is_prod"`
		Port       uint   `env:"BIND_PORT" toml:"port"`
		Addr       string `env:"BIND_ADDR" toml:"addr"`
		TimeFormat string `env:"TIME_LAYOUT" toml:"time_layout"`
	}
	// Test is a configuration that is using in tests
	Test struct {
		DatabaseURI string `env:"TEST_DB_URI"`
		Enable      bool   `env:"TEST"`
	}
	Roles struct {
		Default model.Role `toml:"default_user"`
		Admin   model.Role `toml:"admin_user"`
	}

	// Config ...
	Config struct {
		Postgres Postgres `toml:"postgres"`
		HTTPS    HTTPS    `toml:"https"`
		Server   Server   `toml:"server"`
		Test     Test     `toml:"-"`
		Auth     Auth     `toml:"auth"`
		//Roles    Roles    `toml:"roles"`
	}
)

var (
	// c ...
	c *Config
	// once ...
	once sync.Once
	// useFileConfig ...
	useFileConfig bool
)

const (
	generatedTokenSize = 30
	generatedSaltSize  = 12
	defaultPGHost      = "localhost"
	defaultPGPort      = 5432
	defaultAddr        = "localhost"
	defaultPort        = 8080
	defaultPassDif     = 40
	defaultConfigPath  = "configs/config.toml"
	defaultType        = "http"
	defaultTokenSize   = 20
	defaultTimeLayout  = time.RFC3339
)

func New() *Config {
	once.Do(initConfig)
	return c
}

// initConfig creates new config instance. Should be called just once. Always use sync.Once to access to this function.
func initConfig() {
	c = &Config{}
	if err := env.Parse(c); err != nil {
		log.Fatalf("env: parse: %v", err)
	}
	var configPath string
	// flag vars
	{
		flag.StringVar(&configPath, "config-path", defaultConfigPath, "specify config path")
		flag.StringVar(&configPath, "c", configPath, "specify config path")
		flag.StringVar(&c.Postgres.URI, "d", c.Postgres.URI, "db uri like postgresql://(username):(password)@(addr):(port)/(db_name)?(params)")
		flag.StringVar(&c.Postgres.URI, "db-uri", c.Postgres.URI, "db uri like postgresql://(username):(password)@(addr):(port)/(db_name)?(params)")
		flag.StringVar(&c.Server.Addr, "a", c.Server.Addr, "bind address without port for example 127.0.0.1")
		flag.StringVar(&c.Server.Addr, "bind-addr", c.Server.Addr, "bind address without port for example 127.0.0.1")
		flag.UintVar(&c.Server.Port, "p", c.Server.Port, "bind port as uint from 0 to 65536")
		flag.UintVar(&c.Server.Port, "bind-port", c.Server.Port, "bind port as uint from 0 to 65536")

		flag.BoolVar(&c.Server.IsDev, "is-dev", c.Server.IsDev, "if true server will be running with development mode")
		flag.BoolVar(&c.Server.IsProd, "is-prod", c.Server.IsProd, "if true server will be running with production mode")
		flag.BoolVar(&useFileConfig, "file-cfg", false, "if true than will parse config from file")
	}

	flag.Parse()

	(&c.Postgres).SetDatabaseURI()
	switch c.Server.Type {
	case "HTTP", "HTTPS", "https", "http":
		c.Server.EnableHTTP = true
	case "GRPC", "grpc":
		c.Server.EnableGRPC = true
	}

	if useFileConfig {
		zap.L().Info("parsing cfg from file")
		ex, err := os.Executable()
		if err == nil {
			configPath = path.Join(path.Dir(ex), configPath)

			if err = c.ParseFromFile(configPath); err != nil {
				zap.L().Info("config: parse from file", zap.Error(err))
			}
			c.setDefaultVars()
			if err = c.WriteToFile(configPath); err != nil {
				zap.L().Info("config: write to file", zap.Error(err))
			}
		} else {
			zap.L().Info("get os executable", zap.Error(err))
		}
	}
	c.setDefaultVars()

}

// SetDatabaseURI ...
func (p *Postgres) SetDatabaseURI() {
	if p.URI == "" {
		p.URI = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", p.User, p.Password, p.Addr, p.Port, p.Name)
	}
}

// Valid ...
func (c *Config) Valid() (bool, error) {
	if c == nil {
		return false, nil
	}
	ok, err := govalidator.ValidateStruct(c)
	// validate that enabled
	ok = ok && (c.Server.EnableGRPC != c.Server.EnableHTTP) && (c.Server.EnableHTTP || c.Server.EnableGRPC)
	return ok, err
}

// setDefaultVars add default values or generate it for config.
func (c *Config) setDefaultVars() {

	// configure secret key.
	if c.Server.SecretKey == "" {
		data, err := generateRandom(generatedTokenSize)
		if err != nil {
			zap.L().Warn("set default config vars", zap.Error(err))
			return
		}
		c.Server.SecretKey = byteToString(data)
	}

	// configure server salt.
	if c.Server.Salt == "" {
		data, err := generateRandom(generatedSaltSize)
		if err != nil {
			zap.L().Warn("set default config vars", zap.Error(err))
			return
		}
		c.Server.Salt = byteToString(data)
	}

	// configure other default values
	if c.Server.Addr == "" {
		c.Server.Addr = defaultAddr
	}
	if c.Server.Port == 0 {
		c.Server.Port = defaultPort
	}
	if c.Postgres.Port == 0 {
		c.Postgres.Port = defaultPGPort
	}
	if c.Auth.PasswordDifficult == 0.0 {
		c.Auth.PasswordDifficult = defaultPassDif
	}
	if c.Postgres.Addr == "" {
		c.Postgres.Addr = defaultPGHost
	}
	if c.Server.Type == "" {
		c.Server.Type = defaultType
	}
	if c.Server.TimeFormat == "" {
		c.Server.TimeFormat = defaultTimeLayout
	}
	if c.Auth.AuthTokenSize == 0 {
		c.Auth.AuthTokenSize = defaultTokenSize
	}
}

// byteToString is helper func that calls base64.StdEncoding.EncodeToString.
func byteToString(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// generateRandom returns random byte slice with len(b) == size.
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("rand: read: %w", err)
	}
	return b, nil
}

// ParseFromFile ...
func (c *Config) ParseFromFile(file string) error {
	f, err := os.Open(file)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return nil
	} else if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			zap.L().Error("os: file: close", zap.Error(err))
		}

	}()

	if err = toml.NewDecoder(f).Decode(c); err != nil {
		return fmt.Errorf("toml: decode: %w", err)
	}

	return nil
}

func (c *Config) WriteToFile(file string) error {
	_ = os.Remove(file)

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("os: create: %w", err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			zap.L().Error("config: write to file: close file", zap.Error(err))
		}
	}()

	if err = toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("config: write to file: toml: new encoder: encode: %w", err)
	}
	return nil
}
