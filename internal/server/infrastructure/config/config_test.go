package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_initEnv_Success(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "127.0.0.1:9090")
	t.Setenv("DATABASE_URI", "user=test password=test dbname=testdb sslmode=disable")
	t.Setenv("SECRET_KEY", "supersecret")
	t.Setenv("CRYPTO_KEY", "/path/to/crypto.key")

	cfg := new(config)
	err := cfg.initEnv()

	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", cfg.RunAddress)
	assert.Equal(t, "user=test password=test dbname=testdb sslmode=disable", cfg.DatabaseURI)
	assert.Equal(t, "supersecret", cfg.SecretKey)
	assert.Equal(t, "/path/to/crypto.key", cfg.CryptoKey)
}

func TestConfig_parseFlags_Success(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	osArgs := []string{
		"cmd",
		"-a=127.0.0.1:9090",
		"-d=user=test password=test dbname=testdb sslmode=disable",
		"-k=supersecret",
		"--crypto-key=/path/to/crypto.key",
	}

	cfg := NewConfig()

	flagSet.StringVar(&cfg.RunAddress, "a", "localhost:8080", "net address host:port")
	flagSet.StringVar(
		&cfg.DatabaseURI,
		"d",
		"user=nikolos password=abc123 dbname=gophkeeper sslmode=disable",
		"data source name for connection",
	)
	flagSet.StringVar(&cfg.SecretKey, "k", "abc", "secret key for hash")
	flagSet.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to private crypto key")
	err := flagSet.Parse(osArgs[1:])
	if err != nil {
		t.Errorf("failed flagSet.Parse: %v", err)
	}

	assert.Equal(t, "127.0.0.1:9090", cfg.RunAddress)
	assert.Equal(t, "user=test password=test dbname=testdb sslmode=disable", cfg.DatabaseURI)
	assert.Equal(t, "supersecret", cfg.SecretKey)
	assert.Equal(t, "/path/to/crypto.key", cfg.CryptoKey)
}

func TestConfig_Getters(t *testing.T) {
	cfg := &config{
		RunAddress:  "127.0.0.1:9090",
		DatabaseURI: "user=test password=test dbname=testdb sslmode=disable",
		SecretKey:   "supersecret",
		CryptoKey:   "/path/to/crypto.key",
	}

	assert.Equal(t, "127.0.0.1:9090", cfg.GetRunAddress())
	assert.Equal(t, "user=test password=test dbname=testdb sslmode=disable", cfg.GetDatabaseURI())
	assert.Equal(t, "supersecret", cfg.GetSecretKey())
	assert.Equal(t, "/path/to/crypto.key", cfg.GetCryptoKeyPath())
}
