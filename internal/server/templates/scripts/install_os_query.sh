#!/usr/bin/env bash

set -euo pipefail

FINDINGS_API_HOST="{{ .FindingsApiHost }}"
ENROLL_SECRET="{{ .OsqueryEnrollSecret }}"
MACHINE_ID="{{ .MachineId }}"

install_osquery() {
  mkdir -p /etc/apt/keyrings
  curl -L https://pkg.osquery.io/deb/pubkey.gpg | tee /etc/apt/keyrings/osquery.asc >/dev/null
  add-apt-repository -y 'deb [arch=amd64 signed-by=/etc/apt/keyrings/osquery.asc] https://pkg.osquery.io/deb deb main'
  apt-get update -qq
  apt-get install -y osquery
}

configure_osquery() {
  cat > /etc/osquery/osquery.flags <<FLAGSEOF
--tls_hostname=${FINDINGS_API_HOST}
--config_plugin=tls
--logger_plugin=tls
--enroll_tls_endpoint=/api/v1/osquery/enroll
--config_tls_endpoint=/api/v1/osquery/config
--logger_tls_endpoint=/api/v1/osquery/log
--distributed_tls_read_endpoint=/api/v1/osquery/distributed/read
--distributed_tls_write_endpoint=/api/v1/osquery/distributed/write
--enroll_secret_env=OSQUERY_ENROLL_SECRET
--host_identifier=specified
--specified_identifier=${MACHINE_ID}
--tls_enroll_max_attempts=5
--logger_min_status=1
--disable_events=false
--enable_file_events=true
FLAGSEOF

  cat > /etc/osquery/osquery.conf <<'CONFEOF'
{
  "options": {
    "config_plugin": "tls",
    "logger_plugin": "tls"
  }
}
CONFEOF

  mkdir -p /etc/default
  cat > /etc/default/osquery <<ENVEOF
OSQUERY_ENROLL_SECRET=${ENROLL_SECRET}
ENVEOF
  chmod 600 /etc/default/osquery
}

start_osquery() {
  systemctl enable --now osqueryd
}

main() {
  install_osquery || exit $?
  configure_osquery || exit $?
  start_osquery || exit $?
}

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

main || exit $?
