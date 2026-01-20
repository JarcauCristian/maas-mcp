package tools

type Machine struct {
	SystemID string `json:"system_id"`
	Hostname string `json:"hostname"`
	FQDN     string `json:"fqdn"`

	// Status
	StatusName string `json:"status_name"`
	PowerState string `json:"power_state"`
	Locked     bool   `json:"locked"`

	// Hardware
	Architecture string  `json:"architecture"`
	CPUCount     int     `json:"cpu_count"`
	Memory       int     `json:"memory"`  // in MB
	Storage      float64 `json:"storage"` // in MB
	CPUModel     string  `json:"cpu_model,omitempty"`

	// OS
	OSSystem     string `json:"osystem,omitempty"`
	DistroSeries string `json:"distro_series,omitempty"`

	// Network
	IPAddresses   []string    `json:"ip_addresses,omitempty"`
	Gateway       string      `json:"gateway,omitempty"`
	BootInterface *Interface  `json:"boot_interface,omitempty"`
	Interfaces    []Interface `json:"interfaces,omitempty"`

	// Storage
	BootDisk *BlockDevice  `json:"boot_disk,omitempty"`
	Disks    []BlockDevice `json:"disks,omitempty"`

	// Organization
	Zone     string   `json:"zone"`
	Pool     string   `json:"pool"`
	TagNames []string `json:"tags,omitempty"`
}

type Interface struct {
	Name       string `json:"name"`
	MACAddress string `json:"mac_address"`
	IPAddress  string `json:"ip_address,omitempty"`
	CIDR       string `json:"cidr,omitempty"`
	VLAN       int    `json:"vlan,omitempty"`
}

type BlockDevice struct {
	Name       string      `json:"name"`
	Size       int64       `json:"size"`
	Model      string      `json:"model,omitempty"`
	Serial     string      `json:"serial,omitempty"`
	Partitions []Partition `json:"partitions,omitempty"`
}

type Partition struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	FSType     string `json:"fstype,omitempty"`
	MountPoint string `json:"mount_point,omitempty"`
}
