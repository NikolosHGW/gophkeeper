package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_initEnv_Success(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "127.0.0.1:9090")

	cfg := new(config)
	err := cfg.initEnv()

	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddress)
}

func TestConfig_parseFlags_Success(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	osArgs := []string{
		"cmd",
		"-a=127.0.0.1:9090",
	}

	cfg := NewConfig()

	flagSet.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "net address host:port")
	err := flagSet.Parse(osArgs[1:])
	if err != nil {
		t.Errorf("failed flagSet.Parse: %v", err)
	}

	assert.Equal(t, "127.0.0.1:9090", cfg.GetServerAddress())
}

func TestConfig_GetServerAddress(t *testing.T) {
	cfg := &config{
		ServerAddress: "127.0.0.1:9090",
	}

	assert.Equal(t, "127.0.0.1:9090", cfg.GetServerAddress())
}

func TestConfig_initEnv_RootCertPath_Success(t *testing.T) {
	t.Setenv("ROOT_CERT_PATH", "/path/to/custom/ca.pem")

	cfg := new(config)
	err := cfg.initEnv()

	assert.NoError(t, err)
	assert.Equal(t, "/path/to/custom/ca.pem", cfg.RootCertPath)
}

func TestConfig_parseFlags_RootCertPath_Success(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	osArgs := []string{
		"cmd",
		"-ca=/path/to/custom/ca.pem",
	}

	cfg := new(config)

	flagSet.StringVar(&cfg.RootCertPath, "ca", "./ca.pem", "root cert path")
	err := flagSet.Parse(osArgs[1:])
	if err != nil {
		t.Errorf("failed flagSet.Parse: %v", err)
	}

	assert.Equal(t, "/path/to/custom/ca.pem", cfg.GetRootCertPath())
}

func TestConfig_GetRootCertPath(t *testing.T) {
	cfg := &config{
		RootCertPath: "/path/to/custom/ca.pem",
	}

	assert.Equal(t, "/path/to/custom/ca.pem", cfg.GetRootCertPath())
}
