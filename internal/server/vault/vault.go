package vault

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type Vault struct {
	client   *vault.Client
	baseUrl  *url.URL
	roleName string
	roleId   string
	secretId string
}

func NewVault(baseUrl url.URL, roleName string) (*Vault, error) {
	ctx := context.Background()
	roleId := os.Getenv("ROLE_ID")
	if roleId == "" {
		return nil, errors.New("Environment ROLE_ID required, but is empty.")
	}

	secretId := os.Getenv("SECRET_ID")
	if secretId == "" {
		return nil, errors.New("Environment SECRET_ID required, but is empty.")
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

	publicKey, ok := data.(map[string]any)["public_key"]
	if !ok {
		return "", errors.New("There is no public_key field in data.")
	}
	return publicKey.(string), nil
}
