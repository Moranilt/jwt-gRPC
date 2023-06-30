package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const (
	PORT_GRPC  = "PORT_GRPC"
	PORT_REST  = "PORT_REST"
	PRODUCTION = "PRODUCTION"

	CONSUL_HOST        = "CONSUL_HOST"
	CONSUL_TOKEN       = "CONSUL_TOKEN"
	CONSUL_KEY_FOLDER  = "CONSUL_KEY_FOLDER"
	CONSUL_KEY_VERSION = "CONSUL_KEY_VERSION"
	CONSUL_KEY_FILE    = "CONSUL_KEY_FILE"

	TRACER_URL  = "TRACER_URL"
	TRACER_NAME = "TRACER_NAME"

	VAULT_MOUNT_PATH        = "VAULT_MOUNT_PATH"
	VAULT_PUBLIC_CERT_PATH  = "VAULT_PUBLIC_CERT_PATH"
	VAULT_PRIVATE_CERT_PATH = "VAULT_PRIVATE_CERT_PATH"
	VAULT_REDIS_CREDS_PATH  = "VAULT_REDIS_CREDS_PATH"
	VAULT_TOKEN             = "VAULT_TOKEN"
	VAULT_HOST              = "VAULT_HOST"
)

type VaultEnv struct {
	MountPath       string `mapstructure:"VAULT_MOUNT_PATH"`
	PublicCertPath  string `mapstructure:"VAULT_PUBLIC_CERT_PATH"`
	PrivateCertPath string `mapstructure:"VAULT_PRIVATE_CERT_PATH"`
	RedisCredsPath  string `mapstructure:"VAULT_REDIS_CREDS_PATH"`
	Token           string `mapstructure:"VAULT_TOKEN"`
	Host            string `mapstructure:"VAULT_HOST"`
}

type ConsulEnv struct {
	Host       string `mapstructure:"CONSUL_HOST"`
	Token      string `mapstructure:"CONSUL_TOKEN"`
	KeyFolder  string `mapstructure:"CONSUL_KEY_FOLDER"`
	KeyVersion string `mapstructure:"CONSUL_KEY_VERSION"`
	KeyFile    string `mapstructure:"CONSUL_KEY_FILE"`
}

func (c *ConsulEnv) Key() string {
	return strings.Join([]string{c.KeyFolder, c.KeyVersion, c.KeyFile}, "/")
}

type JaegerEnv struct {
	URL  string `mapstructure:"TRACER_URL"`
	Name string `mapstructure:"TRACER_NAME"`
}

type Env struct {
	Vault      *VaultEnv
	Consul     *ConsulEnv
	Jaeger     *JaegerEnv
	PortGRPC   string
	PortREST   string
	Production bool
}

func ReadEnv() (*Env, error) {
	keys := []string{
		PORT_GRPC,
		PORT_REST,
		PRODUCTION,
		CONSUL_HOST,
		CONSUL_TOKEN,
		CONSUL_KEY_FOLDER,
		CONSUL_KEY_VERSION,
		CONSUL_KEY_FILE,
		TRACER_URL,
		TRACER_NAME,
		VAULT_MOUNT_PATH,
		VAULT_PUBLIC_CERT_PATH,
		VAULT_PRIVATE_CERT_PATH,
		VAULT_REDIS_CREDS_PATH,
		VAULT_TOKEN,
		VAULT_HOST,
	}

	result := make(map[string]string, len(keys))

	for _, key := range keys {
		if val, err := os.LookupEnv(key); !err {
			return nil, fmt.Errorf("env %q is not provided", key)
		} else {
			result[key] = val
		}
	}

	var consul *ConsulEnv
	err := mapstructure.Decode(result, &consul)
	if err != nil {
		return nil, err
	}

	var vault *VaultEnv
	err = mapstructure.Decode(result, &vault)
	if err != nil {
		return nil, err
	}

	var jaeger *JaegerEnv
	err = mapstructure.Decode(result, &jaeger)
	if err != nil {
		return nil, err
	}

	var production bool
	if result[PRODUCTION] == "true" {
		production = true
	}

	return &Env{
		Vault:      vault,
		Consul:     consul,
		Jaeger:     jaeger,
		PortGRPC:   result[PORT_GRPC],
		PortREST:   result[PORT_REST],
		Production: production,
	}, nil
}
