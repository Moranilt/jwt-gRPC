package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/Moranilt/jwt-http2/logger"
	"github.com/Moranilt/jwt-http2/utils"
	capi "github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v2"
)

type TokenTime interface {
	time.Duration | string
}
type Config struct {
	App    *AppConfig[time.Duration]
	base64 string
	mu     sync.RWMutex
}

type AppConfig[T TokenTime] struct {
	Issuer   string   `yaml:"issuer"`
	Subject  string   `yaml:"subject"`
	Audience []string `yaml:"audience"`
	TTL      *TTL[T]  `yaml:"ttl"`
}

type TTL[T TokenTime] struct {
	Access  T `yaml:"access"`
	Refresh T `yaml:"refresh"`
}

type WatchConsulBody struct {
	Key         string
	CreateIndex int
	Flags       int
	Value       string
}

func New(log *logger.Logger) *Config {
	return &Config{}
}

func (c *Config) ReadConsul(ctx context.Context, env *ConsulEnv, cc *capi.Client) error {
	kv := cc.KV()
	pair, _, err := kv.Get(env.Key(), nil)
	if err != nil {
		return err
	}

	if pair == nil {
		return fmt.Errorf("empty data in consul %q", env.Key())
	}

	err = c.setNewConfig(pair.Value)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) WatchConsul(ctx context.Context, env *ConsulEnv, newConfigs []WatchConsulBody) error {
	var consulConfig *WatchConsulBody
	for _, nc := range newConfigs {
		if nc.Key == env.Key() {
			consulConfig = &nc
			break
		}
	}
	if consulConfig == nil {
		return nil
	}

	c.mu.Lock()
	if consulConfig.Value == c.base64 {
		return nil
	} else {
		c.base64 = consulConfig.Value
	}

	base64Decoded, err := base64.StdEncoding.DecodeString(consulConfig.Value)
	if err != nil {
		return err
	}
	c.mu.Unlock()

	err = c.setNewConfig(base64Decoded)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) setNewConfig(newValue []byte) error {
	var newConfig *AppConfig[string]
	err := yaml.Unmarshal(newValue, &newConfig)
	if err != nil {
		return err
	}

	access, err := utils.MakeTimeFromString(newConfig.TTL.Access)
	if err != nil {
		return fmt.Errorf("access TTL: %w", err)
	}

	refresh, err := utils.MakeTimeFromString(newConfig.TTL.Refresh)
	if err != nil {
		return fmt.Errorf("refresh TTL: %w", err)
	}

	c.mu.Lock()
	c.App = &AppConfig[time.Duration]{
		Issuer:   newConfig.Issuer,
		Subject:  newConfig.Subject,
		Audience: newConfig.Audience,
		TTL: &TTL[time.Duration]{
			Access:  access,
			Refresh: refresh,
		},
	}
	c.mu.Unlock()

	return nil
}
