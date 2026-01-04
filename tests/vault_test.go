package tests

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/JarcauCristian/ztp-mcp/internal/server/vault"
)

func TestNewVault_Integration(t *testing.T) {
	_, err := vault.NewVault()
	if err != nil {
		t.Errorf("Failed to create vault: %v", err)
	}
}

func TestGetSshKey_Integration(t *testing.T) {
	v, err := vault.NewVault()
	if err != nil {
		t.Errorf("Failed to create vault: %v", err)
	}

	ctx := context.Background()
	response, err := v.GetSshKey(ctx)
	if err != nil {
		t.Errorf("Failed to retrieve the public key: %v", err)
	}

	if !strings.Contains(response, "ssh") {
		t.Error("Response is not a public key.")
	}
}

func TestNewVault_MissingRoleID(t *testing.T) {
	os.Unsetenv("ROLE_ID")
	os.Unsetenv("SECRET_ID")

	_, err := vault.NewVault()
	if err == nil {
		t.Error("Expected error for missing ROLE_ID")
	}
	if err.Error() != "Environment ROLE_ID required, but is empty." {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestNewVault_MissingSecretID(t *testing.T) {
	os.Setenv("ROLE_ID", "test")
	os.Unsetenv("SECRET_ID")

	_, err := vault.NewVault()
	if err == nil {
		t.Error("Expected error for missing SECRET_ID")
	}
	if err.Error() != "Environment SECRET_ID required, but is empty." {
		t.Errorf("Unexpected error message: %v", err)
	}
}
