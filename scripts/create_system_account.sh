#! /usr/bin/env bash

set -euo pipefail

declare -gA settings=(
  [ACCOUNT_NAME]=${ACCOUNT_NAME:-"sysops"}
)

create_system_account() {
  local account_dir = "/var/lib/${settings[ACCOUNT_NAME]}"

  useradd -r -m -d "$account_dir" -s /bin/bash "${settings[ACCOUNT_NAME]}"
  usermod -aG sudo "${settings[ACCOUNT_NAME]}"

  mkdir -p "$account_dir/.ssh"
  chmod 700 "$account_dir/.ssh"
}

create_ssh_key() {
  local ssh_dir="/var/lib/${settings[ACCOUNT_NAME]}/.ssh"
  local key_file="$ssh_dir/id_ed25519"
  
  ssh-keygen -t ed25519 -f "$key_file" -N "" -C "${settings[ACCOUNT_NAME]}@$(hostname)"
  
  cat "$key_file.pub" >> "$ssh_dir/authorized_keys"
  chmod 600 "$ssh_dir/authorized_keys"
  chown -R "${settings[ACCOUNT_NAME]}:${settings[ACCOUNT_NAME]}" "$ssh_dir"
}

ensure_vault_cli() {
  wget -O - https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(grep -oP '(?<=UBUNTU_CODENAME=).*' /etc/os-release || lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
  sudo apt update && sudo apt install vault
}

push_to_vault() {
  local vault_address={{ .VaultAddress }}
  local vault_token={{ .VaultToken }}
  local vault_role={{ .VaultRole }}
  local machine_id={{ .MachineId }}
  local vault_path="secrets-$vault_role/app/config/ssh/$machine_id"

  local ssh_dir="/var/lib/${settings[ACCOUNT_NAME]}/.ssh"
  local private_key=$(cat "$ssh_dir/id_ed25519")
  local public_key=$(cat "$ssh_dir/id_ed25519.pub")

  VAULT_ADDR="$vault_address" VAULT_TOKEN="$vault_token" vault kv put "$vault_path" \
    username="${settings[ACCOUNT_NAME]}" \
    private_key="$private_key" \
    public_key="$public_key"
}

main() {
  create_system_account || exit $?
  create_ssh_key || exit $?
  ensure_vault || exit $?
  push_to_vault || exit $?
}

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

main || exit $?
