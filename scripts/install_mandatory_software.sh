#! /usr/bin/env bash

set -euo pipefail

declare -gA settings=(
  [OS_NAME]=${OS_NAME:-""}
  [OS_VERSION]=${OS_VERSION:-""}
  [WAZUH_MANAGER]=${WAZUH_MANAGER:-"{{ .WazuhManager }}"}
  [ELASTIC_SEARCH]=${ELASTIC_SEARCH:-"{{ .ElasticSearch }}"}
)

set_os_information() {
  if [[ -z "${settings[OS_NAME]}" ]]; then
    settings[OS_NAME]=$(lsb_release -i -s)
  fi

  if [[ -z "${settings[OS_VERSION]}" ]]; then
    settings[OS_VERSION]=$(lsb_release -r -s)
  fi

  echo "Running script on ${settings[OS_NAME]}:${settings[OS_VERSION]}"
}

install_wazuh_agent() {
  case "${settings[OS_NAME]}"+"${settings[OS_VERSION]}" in
    Ubuntu+14*|Debian+*[78]*)
      apt-get install gnupg apt-transport-https
      curl -s https://packages.wazuh.com/key/GPG-KEY-WAZUH | apt-key add -
      echo "deb https://packages.wazuh.com/4.x/apt/ stable main" | tee -a /etc/apt/sources.list.d/wazuh.list
      apt-get update
      ;;
    *)
      apt-get install gnupg apt-transport-https
      curl -s https://packages.wazuh.com/key/GPG-KEY-WAZUH | gpg --no-default-keyring --keyring gnupg-ring:/usr/share/keyrings/wazuh.gpg --import && chmod 644 /usr/share/keyrings/wazuh.gpg
      echo "deb [signed-by=/usr/share/keyrings/wazuh.gpg] https://packages.wazuh.com/4.x/apt/ stable main" | tee -a /etc/apt/sources.list.d/wazuh.list
      apt-get update
      ;;
  esac

  WAZUH_MANAGER="${settings[WAZUH_MANAGER]}" apt-get install wazuh-agent

  systemctl daemon-reload
  systemctl enable wazuh-agent
  systemctl start wazuh-agent
}

install_osquery() {
  local osquery_version
  osquery_version=$(curl -fsSL https://api.github.com/repos/osquery/osquery/releases/latest | grep -oP '"tag_name":\s*"\K[^"]+')

  if [[ -z "${osquery_version}" ]]; then
    echo "Failed to fetch latest osquery version"
    return 1
  fi

  local deb_url="https://github.com/osquery/osquery/releases/download/${osquery_version}/osquery_${osquery_version}-1.linux_amd64.deb"
  local tmp_deb="/tmp/osquery_${osquery_version}.deb"

  curl -fsSL -o "${tmp_deb}" "${deb_url}"

  dpkg -i "${tmp_deb}"

  rm -f "${tmp_deb}"
}

setup_osquery() {
  mkdir -p /etc/osquery
  mkdir -p /var/log/osquery
  mkdir -p /var/osquery

  cat > /etc/osquery/osquery.conf << 'EOF'
{
  "options": {
    "config_plugin": "filesystem",
    "logger_plugin": "filesystem",
    "logger_path": "/var/log/osquery",
    "disable_logging": "false",
    "log_result_events": "true",
    "schedule_splay_percent": "10",
    "pidfile": "/var/osquery/osquery.pidfile",
    "events_expiry": "3600",
    "database_path": "/var/osquery/osquery.db",
    "verbose": "false",
    "worker_threads": "2",
    "enable_monitor": "true",
    "disable_events": "false",
    "disable_audit": "false",
    "audit_allow_config": "true",
    "audit_persist": "true",
    "enable_file_events": "true",
    "enable_syslog": "true"
  },

  "schedule": {
    "system_info": {
      "query": "SELECT hostname, cpu_brand, cpu_type, physical_memory, hardware_vendor, hardware_model FROM system_info;",
      "interval": 3600,
      "description": "System information"
    },
    "os_version": {
      "query": "SELECT name, version, major, minor, patch, build, platform FROM os_version;",
      "interval": 3600,
      "description": "Operating system version"
    },
    "uptime": {
      "query": "SELECT days, hours, minutes, total_seconds FROM uptime;",
      "interval": 1800,
      "description": "System uptime"
    },
    "mounts": {
      "query": "SELECT device, path, type, flags FROM mounts;",
      "interval": 3600,
      "description": "Mounted filesystems"
    },
    "disk_usage": {
      "query": "SELECT path, type, blocks_size, blocks_available, blocks_free, inodes, inodes_free FROM mounts WHERE type NOT IN ('devtmpfs', 'tmpfs', 'proc', 'sysfs', 'cgroup', 'cgroup2');",
      "interval": 900,
      "description": "Disk usage information"
    },
    "file_events": {
      "query": "SELECT target_path, category, action, md5, sha256, time FROM file_events;",
      "interval": 300,
      "description": "File integrity monitoring events"
    },
    "running_processes": {
      "query": "SELECT pid, name, path, cmdline, state, uid, gid, parent, start_time FROM processes;",
      "interval": 300,
      "description": "Currently running processes"
    },
    "process_open_sockets": {
      "query": "SELECT p.pid, p.name, p.path, pos.local_address, pos.local_port, pos.remote_address, pos.remote_port, pos.protocol FROM processes p JOIN process_open_sockets pos ON p.pid = pos.pid WHERE pos.remote_port != 0;",
      "interval": 300,
      "description": "Processes with open network sockets"
    },
    "listening_ports": {
      "query": "SELECT pid, port, protocol, family, address, path FROM listening_ports;",
      "interval": 300,
      "description": "Listening network ports"
    },
    "open_sockets": {
      "query": "SELECT pid, fd, socket, family, protocol, local_address, local_port, remote_address, remote_port, path FROM process_open_sockets WHERE remote_port != 0;",
      "interval": 300,
      "description": "Open network sockets"
    },
    "interface_addresses": {
      "query": "SELECT interface, address, mask, broadcast, type FROM interface_addresses;",
      "interval": 1800,
      "description": "Network interface addresses"
    },
    "arp_cache": {
      "query": "SELECT address, mac, interface, permanent FROM arp_cache;",
      "interval": 600,
      "description": "ARP cache entries"
    },
    "routes": {
      "query": "SELECT destination, netmask, gateway, source, interface, type FROM routes;",
      "interval": 1800,
      "description": "Network routing table"
    },
    "iptables": {
      "query": "SELECT filter_name, chain, policy, target, protocol, src_ip, dst_ip, src_port, dst_port FROM iptables;",
      "interval": 1800,
      "description": "IPTables firewall rules"
    },
    "users": {
      "query": "SELECT uid, gid, username, description, directory, shell FROM users;",
      "interval": 3600,
      "description": "Local user accounts"
    },
    "groups": {
      "query": "SELECT gid, groupname FROM groups;",
      "interval": 3600,
      "description": "Local groups"
    },
    "logged_in_users": {
      "query": "SELECT type, user, host, time, pid, tty FROM logged_in_users;",
      "interval": 300,
      "description": "Currently logged in users"
    },
    "last_logins": {
      "query": "SELECT username, time, pid, type, tty, host FROM last;",
      "interval": 900,
      "description": "Recent login history"
    },
    "sudoers": {
      "query": "SELECT source, header, rule_details FROM sudoers WHERE header NOT IN ('Defaults', 'Cmnd_Alias');",
      "interval": 3600,
      "description": "Sudoers configuration"
    },
    "authorized_keys": {
      "query": "SELECT uid, username, key_file, key, algorithm FROM users CROSS JOIN authorized_keys USING (uid);",
      "interval": 3600,
      "description": "SSH authorized keys"
    },
    "sshd_config": {
      "query": "SELECT * FROM sshd_config;",
      "interval": 3600,
      "description": "SSHD configuration"
    },
    "crontab": {
      "query": "SELECT event, minute, hour, day_of_month, month, day_of_week, command, path FROM crontab;",
      "interval": 900,
      "description": "Scheduled cron jobs"
    },
    "systemd_units": {
      "query": "SELECT id, description, load_state, active_state, sub_state, unit_file_state FROM systemd_units WHERE active_state = 'active';",
      "interval": 900,
      "description": "Active systemd services"
    },
    "deb_packages": {
      "query": "SELECT name, version, source, arch, status FROM deb_packages;",
      "interval": 3600,
      "description": "Installed packages"
    },
    "kernel_modules": {
      "query": "SELECT name, size, used_by, status FROM kernel_modules;",
      "interval": 1800,
      "description": "Loaded kernel modules"
    },
    "sysctl": {
      "query": "SELECT name, current_value FROM system_controls WHERE name IN ('kernel.randomize_va_space', 'net.ipv4.ip_forward', 'net.ipv4.conf.all.accept_redirects', 'net.ipv4.conf.all.send_redirects', 'net.ipv4.tcp_syncookies');",
      "interval": 3600,
      "description": "Security-relevant kernel parameters"
    },
    "docker_containers": {
      "query": "SELECT id, name, image, state, status, pid FROM docker_containers;",
      "interval": 300,
      "description": "Docker containers"
    }
  },

  "file_paths": {
    "configuration": [
      "/etc/%%",
      "/etc/ssh/%%"
    ],
    "binaries": [
      "/usr/bin/%%",
      "/usr/sbin/%%",
      "/bin/%%",
      "/sbin/%%",
      "/usr/local/bin/%%",
      "/usr/local/sbin/%%"
    ],
    "sensitive": [
      "/etc/passwd",
      "/etc/shadow",
      "/etc/sudoers",
      "/etc/sudoers.d/%%",
      "/root/.ssh/%%",
      "/home/%/.ssh/%%"
    ]
  },

  "exclude_paths": {
    "configuration": [
      "/etc/mtab",
      "/etc/resolv.conf"
    ]
  },

  "packs": {
    "osquery-monitoring": "/opt/osquery/share/osquery/packs/osquery-monitoring.conf",
    "incident-response": "/opt/osquery/share/osquery/packs/incident-response.conf",
    "vuln-management": "/opt/osquery/share/osquery/packs/vuln-management.conf"
  }
}
EOF

  chmod 644 /etc/osquery/osquery.conf
  chmod 755 /var/log/osquery
  chmod 755 /var/osquery

  systemctl enable osqueryd
  systemctl start osqueryd
}

install_logstash() {
  apt-get install -y apt-transport-https default-jre

  wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | sudo gpg --dearmor -o /usr/share/keyrings/elastic-keyring.gpg
  echo "deb [signed-by=/usr/share/keyrings/elastic-keyring.gpg] https://artifacts.elastic.co/packages/9.x/apt stable main" | sudo tee -a /etc/apt/sources.list.d/elastic-9.x.list
  apt-get update
  apt-get install -y logstash

  # Create osquery pipeline configuration
  cat > /etc/logstash/conf.d/osquery.conf << EOF
input {
  file {
    path => "/var/log/osquery/osqueryd.results.log"
    type => "osquery_json"
    codec => "json"
    start_position => "beginning"
    sincedb_path => "/var/lib/logstash/sincedb_osquery"
  }
  file {
    path => "/var/log/osquery/osqueryd.snapshots.log"
    type => "osquery_snapshot"
    codec => "json"
    start_position => "beginning"
    sincedb_path => "/var/lib/logstash/sincedb_osquery_snapshots"
  }
}

filter {
  if [type] == "osquery_json" or [type] == "osquery_snapshot" {
    date {
      match => ["unixTime", "UNIX"]
      target => "@timestamp"
    }

    mutate {
      add_field => {
        "host_identifier" => "%{hostIdentifier}"
        "query_name" => "%{name}"
      }
      remove_field => ["unixTime"]
    }

    if [decorations] {
      ruby {
        code => "
          if event.get('decorations').is_a?(Hash)
            event.get('decorations').each { |k, v| event.set(k, v) }
          end
        "
      }
    }
  }
}

output {
  if [type] == "osquery_json" {
    elasticsearch {
      hosts => ["${settings[ELASTIC_SEARCH]}"]
      index => "osquery-results-%{+YYYY.MM.dd}"
    }
  }

  if [type] == "osquery_snapshot" {
    elasticsearch {
      hosts => ["${settings[ELASTIC_SEARCH]}"]
      index => "osquery-snapshots-%{+YYYY.MM.dd}"
    }
  }
}
EOF

  usermod -a -G adm logstash
  chmod 750 /var/log/osquery
  chown root:adm /var/log/osquery

  mkdir -p /var/lib/logstash
  chown logstash:logstash /var/lib/logstash

  systemctl enable logstash
  systemctl start logstash
}

main() {
  set_os_information || exit $?
  install_wazuh_agent || exit $?
  install_osquery || exit $?
  setup_osquery || exit $?
  install_logstash || exit $?
}

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

main || exit $?
