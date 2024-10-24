package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	ServerAddress string `env:"RUN_ADDRESS"`
	RootCertPath  string `env:"ROOT_CERT_PATH"`
}

func (c *config) initEnv() error {
	err := env.Parse(c)
	if err != nil {
		return fmt.Errorf("не удалось спарсить env: %w", err)
	}

	return nil
}

func (c *config) parseFlags() {
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "net address host:port")
	flag.StringVar(&c.RootCertPath, "ca", "./ca.pem", "root cert path")
	flag.Parse()
}

// NewConfig конструктор конфига, в котором идёт инициализация флагов и env переменных.
func NewConfig() *config {
	cfg := new(config)

	cfg.parseFlags()
	if err := cfg.initEnv(); err != nil {
		log.Fatalf("Ошибка при инициализации переменных окружения: %v", err)
	}

	return cfg
}

// GetServerAddress геттер для хоста.
func (c config) GetServerAddress() string {
	return c.ServerAddress
}

// GetRootCertPath геттер для пути к корневому сертификату.
func (c config) GetRootCertPath() string {
	return c.RootCertPath
}
