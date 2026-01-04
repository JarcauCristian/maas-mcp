package vault

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

var (
	vlt    *Vault
	vltErr error
	once   sync.Once
)

func getVault() (*Vault, error) {
	once.Do(func() {
		vlt, vltErr = NewVault()
	})
	return vlt, vltErr
}

func MustVault() *Vault {
	client, err := getVault()
	if err != nil {
		panic(fmt.Errorf("failed to initialize MAAS client: %w", err))
	}
	return client
}

type Vault struct {
	client   *vault.Client
	baseUrl  *url.URL
	roleName string
	roleId   string
	secretId string
}

func NewVault() (*Vault, error) {
	ctx := context.Background()
	roleId := os.Getenv("VAULT_ROLE_ID")
	if roleId == "" {
		return nil, errors.New("Environment VAULT_ROLE_ID required, but is empty.")
	}

	secretId := os.Getenv("VAULT_SECRET_ID")
	if secretId == "" {
		return nil, errors.New("Environment VAULT_SECRET_ID required, but is empty.")
	}

	baseUrl, err := url.Parse(os.Getenv("VAULT_BASE_URL"))
	if err != nil {
		return nil, err
	}

	role_name := os.Getenv("VAULT_ROLE_NAME")
	if role_name == "" {
		return nil, errors.New("Environment VAULT_ROLE_NAME required, but is empty.")
	}

	client, err := vault.New(
		vault.WithAddress(baseUrl.String()),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	response, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   roleId,
			SecretId: secretId,
		},
	)
	if err != nil {
		return nil, err
	}

	if err := client.SetToken(response.Auth.ClientToken); err != nil {
		return nil, err
	}

	vlt := Vault{
		client:   client,
		baseUrl:  &baseUrl,
		roleName: roleName,
		roleId:   roleId,
		secretId: secretId,
	}
	return &vlt, nil
}

func (v *Vault) GetSshKey(ctx context.Context) (string, error) {
	response, err := v.client.Read(ctx, fmt.Sprintf("/v1/secrets-%s/data/app/config/ssh/ztp_key", v.roleName))
	if err != nil {
		return "", err
	}

	data, ok := response.Data["data"]
	if !ok {
		return "", errors.New("There is no data field in response.")
	}

	publicKey, ok := data.(map[string]any)["private_key"]
	if !ok {
		return "", errors.New("There is no public_key field in data.")
	}
	return publicKey.(string), nil
}
