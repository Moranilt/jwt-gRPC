package certs

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"

	"github.com/Moranilt/jwt-http2/config"
	vault "github.com/hashicorp/vault/api"
)

type Certs struct {
	private  []byte
	public   []byte
	vault    *vault.Client
	vaultCfg *config.VaultEnv
}

func NewKeys(v *vault.Client, env *config.VaultEnv) *Certs {
	k := new(Certs)
	k.vault = v
	k.vaultCfg = env
	k.generateKeys()
	return k
}

func (k *Certs) Public() []byte {
	return k.public
}

func (k *Certs) Private() []byte {
	return k.private
}

func (k *Certs) StoreToVault() error {
	_, err := k.vault.KVv2(k.vaultCfg.MountPath).Put(
		context.Background(),
		k.vaultCfg.PublicCertPath,
		map[string]interface{}{
			"key": string(k.public),
		},
	)
	if err != nil {
		return err
	}

	_, err = k.vault.KVv2(k.vaultCfg.MountPath).Put(
		context.Background(),
		k.vaultCfg.PrivateCertPath,
		map[string]interface{}{
			"key": string(k.private),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (k *Certs) generateKeys() {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		log.Fatal(err)
	}

	k.public = k.makePublicPEMKey(&key.PublicKey)
	k.private = k.makePrivatePEMKey(key)
}

func (k *Certs) makePrivatePEMKey(privatekey *rsa.PrivateKey) []byte {
	pemkey := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privatekey),
	}

	return pem.EncodeToMemory(pemkey)
}

func (k *Certs) makePublicPEMKey(pubkey *rsa.PublicKey) []byte {
	key, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		log.Fatal(err)
	}
	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: key,
	}
	return pem.EncodeToMemory(pemkey)
}
