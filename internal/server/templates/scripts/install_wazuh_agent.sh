#! /usr/bin/env bash

set -euo pipefail

install_wazuh_agent() {
  local wazuh_manager={{ .WazuhManger }}
  local wazuh_agent_group={{ .WazuhAgentGroup }}
  local wazuh_agent_name={{ .WazuhAgentName }}

  wget -O /tmp/wazuh.deb https://packages.wazuh.com/4.x/apt/pool/main/w/wazuh-agent/wazuh-agent_4.14.4-1_amd64.deb 
  WAZUH_MANAGER="${wazuh_manager}" WAZUH_AGENT_GROUP="${wazuh_agent_group}" WAZUH_AGENT_NAME="${wazuh_agent_name}" dpkg -i /tmp/wazuh.deb
  rm -rf /tmp/wazuh.deb
}

start_wazuh_agent() {
  systemctl daemon-reload
  systemctl enable wazuh-agent
  systemctl start wazuh-agent
}

main() {
  install_wazuh_agent || exit $?
  start_wazuh_agent || exit $?
}

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

main || exit $?
