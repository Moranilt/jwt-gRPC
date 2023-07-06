# jwt-gRPC

The JWT is implemented according to [RFC7519](https://datatracker.ietf.org/doc/html/rfc7519) using [gRPC](https://grpc.io/).

Core idea is to make fast and secure service to create and manage JWT-tokens and not store secure data into payload of tokens.

To create and validate tokens application using private and public RSA certificates. That is why you need **public** and **private** certificates. **Private** - needs to sign JWT, **public** needs to other services to validate it.

## Environments
| Name | Type | Description |
| ---- | ---- | ----------- |
| PORT_GRPC | integer | gRPC port for main server |
| PORT_REST | integer | Port for REST **/watch** endpoint |
| PRODUCTION | boolean | Turn on/off production mode |
| CONSUL_HOST | string | Consul host. Only hostname and port(localhost:8500) |
| CONSUL_TOKEN | string | Consul [ACL](https://developer.hashicorp.com/consul/tutorials/security/access-control-setup-production) token. It can be empty. |
| CONSUL_KEY_FOLDER | string | Core folder of all configuration files |
| CONSUL_KEY_VERSION | string | Name of folder with version |
| CONSUL_KEY_FILE | string | Name of configuration file |
| TRACER_URL | string | URL of jaeger with protocol(http://localhost:14268/api/traces) |
| TRACER_NAME | string | name of application in Jaeger UI |
| VAULT_MOUNT_PATH | string | Vault mount name, simply name of KV storage |
| VAULT_PUBLIC_CERT_PATH | string | Path to store public certificate |
| VAULT_PRIVATE_CERT_PATH | string | Path to store private certificate |
| VAULT_REDIS_CREDS_PATH | string | Path to store redis connection data |
| VAULT_TOKEN | string | Vault token to connect using client |
| VAULT_HOST | string | Vault host with protocol and port(http://localhost:8200) |

## Main tools

### Redis
Using to store secret data by token-uuid. Every token has his own UUID which is key in redis, and userId which is value for this key.

### Consul
Store you configuration to [Consul](https://www.consul.io/) by versioning your configs with app. App has endpoint **/watch** on which Consul will send data to if you would change configs.

If you have app of version **v1.0.1** so it will be waiting for changes of config with version **v1.0.1**.

Path is formed as:  
`/folder/version/file`

Using ENV variables:  
`/CONSUL_KEY_FOLDER/CONSUL_KEY_VERSION/CONSUL_KEY_FILE`

Example:  
`/jwt_auth/v1.0.0/config.yaml`

### Jaeger
[Jaeger](https://www.jaegertracing.io/): open source, end-to-end distributed tracing
Monitor and troubleshoot transactions in complex distributed systems. Fast and very comfortable to use.

### Vault
[Vault](https://www.vaultproject.io/) is a complex tool. In our situation we are using it to store Redis connection data and certificates.

The main reason to use it, that you can create tokens for others applications and apply policies to read only public certificate.

For exampple: 

We have **public** and **private** certificates in folder **crt**. Paths will be **/authentication/crt/public** and **/authentication/private**.

We have application **users** which want to validate JWT tokens created by our application.

We need to follow next steps to allow to read public certificate from Vault:
1. Create policy to only read from **/authentication/crt/public**(user-policy.hcl)
Example of **user-policy.hcl**:
```hcl
path "authentication/data/crt/public" {
  capabilities = ["read"]
}
```
2. Write this policy to Vault
```bash
vault policy write user-policy /policies/user-policy.hcl
```
3. Create vault-token with this policy
```bash
vault token create -policy=user-policy
```
4. Give this new token to **user**-app

Now you can read data only from **/authentication/crt/public**.

## Configuration
You can find default configuration in repository [config.yaml](https://github.com/Moranilt/jwt-gRPC/blob/main/config.yaml)

### Consul

| Name || Type | Description |
| ---- | - | ---- | ----------- |
| issuer | | string | [JWT iss](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1) |
| subject | | string | [JWT sub](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.2) |
| audience | | string[] | [JWT aud](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.3) |
| ttl | | object | TTL data for tokens |
| | access | string | TTL for access token |
| | refresh | string | TTL for refresh token |

TTL using his own measurement system. You can pass `s`, `m`, `h` and `d`.

`s` - seconds  
`m` - minutes  
`h` - hours  
`d` - days

### Vault
By default we have Redis data and certificates in Vault.

Redis data([redis_config.json](https://github.com/Moranilt/jwt-gRPC/blob/main/init/vault/redis_config.json)):
```json
{
  "host": "localhost:6379",
  "password": ""
}
```

Certificates:
```json
{
  "key": "certificate string"
}
```
