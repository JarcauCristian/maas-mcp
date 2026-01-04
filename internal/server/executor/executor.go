package executor

import (
	"context"
	"strings"

	"github.com/JarcauCristian/ztp-mcp/internal/server/vault"
	"golang.org/x/crypto/ssh"
)

type Executor struct {
	vault    *vault.Vault
	hostname string
}

func NewExecutor(vault *vault.Vault, hostname string) Executor {
	return Executor{
		vault:    vault,
		hostname: hostname,
	}
}

func (e Executor) Execute(ctx context.Context, commands []string) (string, error) {
	keyStr, err := e.vault.GetSshKey(ctx)
	if err != nil {
		return "", err
	}

	key, err := ssh.ParsePrivateKey([]byte(keyStr))
	if err != nil {
		return "", err
	}

	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", e.hostname+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	script := strings.Join(commands, "\n")
	output, err := session.CombinedOutput(script)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
