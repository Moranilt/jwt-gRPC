package clients

import (
	"context"

	"github.com/Moranilt/jwt-http2/config"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

type VaultClient struct {
	client *vault.Client
	cfg    *config.VaultEnv
}

type VaultCfg struct {
	MountPath       string
	PublicCertPath  string
	PrivateCertPath string
	RedisCredsPath  string
	Token           string
	Host            string
}

type RedisCreds struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
}

type CertificateValue struct {
	Key string `mapstructure:"key"`
}

// New Vault client
func Vault(cfg *config.VaultEnv) (*VaultClient, error) {
	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = cfg.Host
	client, err := vault.NewClient(vaultCfg)
	if err != nil {
		return nil, err
	}
	client.SetToken(cfg.Token)

	newClient := &VaultClient{
		client: client,
		cfg:    cfg,
	}

	return newClient, nil
}

func (v *VaultClient) GetClient() *vault.Client {
	return v.client
}

func (v *VaultClient) GetRedisCreds(ctx context.Context) (*RedisCreds, error) {
	kvSecret, err := v.client.KVv2(v.cfg.MountPath).Get(ctx, v.cfg.RedisCredsPath)
	if err != nil {
		return nil, err
	}
	var creds *RedisCreds
	err = mapstructure.Decode(kvSecret.Data, &creds)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (v *VaultClient) GetPublicCert(ctx context.Context) ([]byte, error) {
	kvSecret, err := v.client.KVv2(v.cfg.MountPath).Get(ctx, v.cfg.PublicCertPath)
	if err != nil {
		return nil, err
	}
	var cert *CertificateValue
	err = mapstructure.Decode(kvSecret.Data, &cert)
	if err != nil {
		return nil, err
	}

	return []byte(cert.Key), nil
}

func (v *VaultClient) GetPrivateCert(ctx context.Context) ([]byte, error) {
	kvSecret, err := v.client.KVv2(v.cfg.MountPath).Get(ctx, v.cfg.PrivateCertPath)
	if err != nil {
		return nil, err
	}
	var cert *CertificateValue
	err = mapstructure.Decode(kvSecret.Data, &cert)
	if err != nil {
		return nil, err
	}

	return []byte(cert.Key), nil
}

func Redis(ctx context.Context, creds *RedisCreds) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     creds.Host,
		Password: creds.Password,
	})

	if ping := redisClient.Ping(ctx); ping.Err() != nil {
		return nil, ping.Err()
	}

	return redisClient, nil

}
