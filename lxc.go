/*
Author: John Connor Sanders
License: Apache Version 2.0
Version: 0.0.2
Released: 01/18/2021
Copyright 2021 John Connor Sanders

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
------------GO-CONTAINERS----------------
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package containers

import (
	"encoding/json"
	"log"
	"strings"
)

// TODO: Build in system for generating cloud init config with proper ssh keys!
const SRCYAML = `cat <<EOF 
#cloud-config
users:
  - name: {{Username}}
    plain_text_passwd: {{Password}}
    ssh_pwauth: True
    lock_passwd: False
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    groups: sudo
    shell: /bin/bash
write_files:
  - path: /etc/ssh/sshd_config
    content: |
      Port {{Port}}
      Protocol 2
      UsePrivilegeSeparation yes
      KeyRegenerationInterval 3600
      ServerKeyBits 1024
      SyslogFacility AUTH
      LogLevel INFO
      LoginGraceTime 120
      PermitRootLogin no
      StrictModes yes
      RSAAuthentication yes
      PubkeyAuthentication yes
      PasswordAuthentication yes
      IgnoreRhosts yes
      RhostsRSAAuthentication no
      HostbasedAuthentication no
      PermitEmptyPasswords no
      ChallengeResponseAuthentication no
      X11Forwarding yes
      X11DisplayOffset 10
      PrintMotd no
      PrintLastLog yes
      TCPKeepAlive yes
      AcceptEnv LANG LC_*
      Subsystem sftp /usr/lib/openssh/sftp-server
      UsePAM yes
      AllowUsers {{Username}}
runcmd:
  - systemctl restart sshd
EOF
`

// generateCloudInit
func generateCloudInit(auth *Auth) []byte {
	newInitFile := SRCYAML
	newInitFile = strings.Replace(newInitFile, "{{Username}}", auth.User, -1)
	newInitFile = strings.Replace(newInitFile, "{{Password}}", auth.Credential, -1)
	newInitFile = strings.Replace(newInitFile, "{{Port}}", auth.Port, -1)
	return []byte(newInitFile)
}

// LXCConfig
type LXCConfig struct {
	ImageArchitecture string `json:"image.architecture,omitempty"`
	ImageDescription  string `json:"image.description,omitempty"`
	ImageOS           string `json:"image.os,omitempty"`
	ImageRelease      string `json:"image.release,omitempty"`
}

// ContainerNetwork
type ContainerNetwork struct {
	Usage     int `json:"usage,omitempty"`
	UsagePeak int `json:"usage_peak,omitempty"`
}

// ContainerMemory
type ContainerMemory struct {
	Usage     int `json:"usage,omitempty"`
	UsagePeak int `json:"usage_peak,omitempty"`
}

// ContainerCPU
type ContainerCPU struct {
	Usage int `json:"usage,omitempty"`
}

// ContainSnapshot
type ContainSnapshot struct {
	Name         string    `json:"name,omitempty"`
	Architecture string    `json:"architecture,omitempty"`
	Config       LXCConfig `json:"config,omitempty"`
	LastUsed     string    `json:"last_used_at,omitempty"`
}

// ContainerState
type ContainerState struct {
	Memory    ContainerMemory   `json:"memory,omitempty"`
	CPU       ContainerCPU      `json:"cpu,omitempty"`
	PID       int               `json:"pid,omitempty"`
	Processes int               `json:"processes,omitempty"`
	SnapShots []ContainSnapshot `json:"snapshots,omitempty"`
}

// ContainerOutput
type ContainerOutput struct {
	Architecture string         `json:"architecture,omitempty"`
	Config       LXCConfig      `json:"config,omitempty"`
	State        ContainerState `json:"state,omitempty"`
	Name         string         `json:"name,omitempty"`
	Status       string         `json:"status,omitempty"`
	StatusCode   int            `json:"status_code,omitempty"`
	LastUsed     string         `json:"last_used_at,omitempty"`
}

// ListOutput
type ListOutput struct {
	Outputs []ContainerOutput `json:"output,omitempty"`
}

// LoadNetworkOutput
func LoadListOut(jsonStr string) (*ListOutput, error) {
	var lOutput ListOutput
	var output []ContainerOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		log.Fatal(err.Error())
		return &lOutput, err
	}
	lOutput.Outputs = output
	return &lOutput, nil
}

// GetContainers
func (lo *ListOutput) GetContainers() []*GoContainer {
	var containers []*GoContainer
	for _, out := range lo.Outputs {
		var container GoContainer
		_ = container.loadListOutput(&out)
		_ = container.loadNetworkData("lxdbr0")
		containers = append(containers, &container)
	}
	return containers
}

// NetworkEntry
type NetworkEntry struct {
	Hostname string `json:"hostname,omitempty"`
	HWAddr   string `json:"hwaddr,omitempty"`
	Address  string `json:"address,omitempty"`
	Type     string `json:"type,omitempty"`
	Location string `json:"location,omitempty"`
}

// NetworkOutput
type NetworkOutput struct {
	NetworkEntries []NetworkEntry `json:"network_entries,omitempty"`
}

// LoadNetworkOutput
func LoadNetworkOutput(jsonStr string) (*NetworkOutput, error) {
	var lOutput NetworkOutput
	var nEntries []NetworkEntry
	if err := json.Unmarshal([]byte(jsonStr), &nEntries); err != nil {
		log.Fatal(err.Error())
		return &lOutput, err
	}
	lOutput.NetworkEntries = nEntries
	return &lOutput, nil
}

// GetContainerEntry
func (nwo *NetworkOutput) GetContainerEntry(containerName string) *Network {
	var reNetwork Network
	for _, nw := range nwo.NetworkEntries {
		if nw.Hostname == containerName {
			_ = reNetwork.LoadEntry(&nw)
			return &reNetwork
		}
	}
	return &reNetwork
}

// ImageProperties
type ImageProperties struct {
	Architecture string `json:"architecture,omitempty"`
	OSType       string `json:"os,omitempty"`
	OSRelease    string `json:"os_release,omitempty"`
}

// ImageOutput
type ImageOutput struct {
	Public      bool            `json:"public,omitempty"`
	Props       ImageProperties `json:"properties,omitempty"`
	Filename    string          `json:"filename,omitempty"`
	Fingerprint string          `json:"fingerprint,omitempty"`
	Size        int             `json:"size,omitempty"`
	Type        string          `json:"type,omitempty"`
	Created     string          `json:"created_at,omitempty"`
}

// ImagesOutput
type ImagesOutput struct {
	Outputs []ImageOutput `json:"output,omitempty"`
}

// LoadImagesOutput
func LoadImagesOutput(jsonStr string) *ImagesOutput {
	var imagesOutput ImagesOutput
	if err := json.Unmarshal([]byte(jsonStr), &imagesOutput); err != nil {
		log.Fatal(err.Error())
		return &imagesOutput
	}
	return &imagesOutput
}

// GetImages
func (imo *ImagesOutput) GetImages() []*GoImage {
	var reImgs []*GoImage
	for _, imgOut := range imo.Outputs {
		name := strings.Replace(imgOut.Filename, ".tar.xz", "", 1)
		newImg := NewGoImage(name, imgOut.Type, imgOut.Fingerprint, imgOut.Created)
		reImgs = append(reImgs, newImg)
	}
	return reImgs
}
