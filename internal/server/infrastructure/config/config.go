package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	RunAddress    string `env:"RUN_ADDRESS"`
	DatabaseURI   string `env:"DATABASE_URI"`
	SecretKey     string `env:"SECRET_KEY"`
	CryptoKey     string `env:"CRYPTO_KEY"`
	ServerKeyPath string `env:"SERVER_KEY_PATH"`
	ServerCrtPath string `env:"SERVER_CRT_PATH"`
}

func (c *config) initEnv() error {
	err := env.Parse(c)
	if err != nil {
		return fmt.Errorf("не удалось спарсить env: %w", err)
	}

	return nil
}

func (c *config) parseFlags() {
	flag.StringVar(&c.RunAddress, "a", "localhost:8080", "net address host:port")
	flag.StringVar(&c.DatabaseURI, "d",
		"user=nikolos "+
			"password=abc123 "+
			"dbname=gophkeeper "+
			"sslmode=disable",
		"data source name for connection")
	flag.StringVar(&c.SecretKey, "k", "abc", "secret key for hash")
	flag.StringVar(&c.CryptoKey, "crypto-key", "01234567890123456789012345678901", "crypto key")
	flag.StringVar(&c.ServerKeyPath, "server-key", "./server.key", "path to server key")
	flag.StringVar(&c.ServerCrtPath, "server-crt", "./server.crt", "path to server crt")
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

// GetAddress геттер для хоста.
func (c config) GetRunAddress() string {
	return c.RunAddress
}

// GetDatabaseURI геттер для подключения к бд.
func (c config) GetDatabaseURI() string {
	return c.DatabaseURI
}

// GetSecretKey геттер для секретного ключа для хеширования.
func (c config) GetSecretKey() string {
	return c.SecretKey
}

// GetCryptoKeyPath геттер для ключа шифрования.
func (c config) GetCryptoKeyPath() string {
	return c.CryptoKey
}

// GetServerKeyPath геттер для пути к приватному ключу сервера.
func (c config) GetServerKeyPath() string {
	return c.ServerKeyPath
}

// GetServerCrtPath геттер для пути к сертификату сервера, подписанному CA.
func (c config) GetServerCrtPath() string {
	return c.ServerCrtPath
}
