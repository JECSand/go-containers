/*
Author: John Connor Sanders
License: Apache Version 2.0
Version: 0.0.1
Released: 01/13/2021
Copyright 2021 John Connor Sanders

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
------------GO-CONTAINERS----------------
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package containers

import (
	"encoding/json"
	"fmt"
	//"gopkg.in/yaml.v3"
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
/*
// CloudInit
type CloudInit struct {
	Users []struct {
		Name            string   `yaml:"name"`
		PlainTextPasswd string   `yaml:"plain_text_passwd"`
		SSHPwauth       bool     `yaml:"ssh_pwauth"`
		LockPasswd      bool     `yaml:"lock_passwd"`
		Sudo            []string `yaml:"sudo"`
		Groups          string   `yaml:"groups"`
		Shell           string   `yaml:"shell"`
	} `yaml:"users"`
	WriteFiles []struct {
		Path    string `yaml:"path"`
		Content string `yaml:"content"`
	} `yaml:"write_files"`
	Runcmd []string `yaml:"runcmd"`
}
 */

// generateCloudInit
func generateCloudInit(auth *Auth) []byte {
	//var newCloudInit CloudInit
	newInitFile := SRCYAML
	newInitFile = strings.Replace(newInitFile, "{{Username}}", auth.User, -1)
	newInitFile = strings.Replace(newInitFile, "{{Password}}", auth.Credential, -1)
	newInitFile = strings.Replace(newInitFile, "{{Port}}", auth.Port, -1)
	/*
	err := yaml.Unmarshal([]byte(newInitFile), &newCloudInit)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	b, err := yaml.Marshal(&newCloudInit)
	if err != nil {
		log.Fatal(err.Error())
	}
	appendStr := "#cloud-config\n"
	b = append([]byte(appendStr), b...)
	 */
	return []byte(newInitFile)
}

// LXCConfig
type LXCConfig struct {
	ImageArchitecture	string					`json:"image.architecture,omitempty"`
	ImageDescription	string					`json:"image.description,omitempty"`
	ImageOS				string					`json:"image.os,omitempty"`
	ImageRelease		string					`json:"image.release,omitempty"`
}

// ContainerNetwork
type ContainerNetwork struct {
	Usage				int						`json:"usage,omitempty"`
	UsagePeak			int						`json:"usage_peak,omitempty"`
}

// ContainerMemory
type ContainerMemory struct {
	Usage				int						`json:"usage,omitempty"`
	UsagePeak			int						`json:"usage_peak,omitempty"`
}

// ContainerCPU
type ContainerCPU struct {
	Usage				int						`json:"usage,omitempty"`
}

// ContainSnapshot
type ContainSnapshot struct {
	Name				string					`json:"name,omitempty"`
	Architecture		string					`json:"architecture,omitempty"`
	Config				LXCConfig				`json:"config,omitempty"`
	LastUsed			string					`json:"last_used_at,omitempty"`
}

// ContainerState
type ContainerState struct {
	Memory				ContainerMemory			`json:"memory,omitempty"`
	CPU					ContainerCPU			`json:"cpu,omitempty"`
	PID					int						`json:"pid,omitempty"`
	Processes			int						`json:"processes,omitempty"`
	SnapShots			[]ContainSnapshot		`json:"snapshots,omitempty"`
}

// ContainerOutput
type ContainerOutput struct {
	Architecture		string					`json:"architecture,omitempty"`
	Config				LXCConfig				`json:"config,omitempty"`
	State				ContainerState			`json:"state,omitempty"`
	Name				string					`json:"name,omitempty"`
	Status				string					`json:"status,omitempty"`
	StatusCode			int						`json:"status_code,omitempty"`
	LastUsed			string					`json:"last_used_at,omitempty"`
}

// ListOutput
type ListOutput struct {
	Outputs				[]ContainerOutput		`json:"output,omitempty"`
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
		container.loadListOutput(&out)
		container.loadNetworkData("lxdbr0")
		containers = append(containers, &container)
	}
	return containers
}

// NetworkEntry
type NetworkEntry struct {
	Hostname			string					`json:"hostname,omitempty"`
	HWAddr				string					`json:"hwaddr,omitempty"`
	Address				string					`json:"address,omitempty"`
	Type				string					`json:"type,omitempty"`
	Location			string					`json:"location,omitempty"`
}

// NetworkOutput
type NetworkOutput struct {
	NetworkEntries		[]NetworkEntry			`json:"network_entries,omitempty"`
}

// LoadNetworkOutput
func LoadNetworkOutput(jsonStr string) (*NetworkOutput, error) {
	var lOutput NetworkOutput
	var nEntries []NetworkEntry
	fmt.Println("CHECK OUTLINE JSON")
	fmt.Println(jsonStr)
	fmt.Println("END OUTLINE JSON")
	if err := json.Unmarshal([]byte(jsonStr), &nEntries); err != nil {
		log.Fatal(err.Error())
		return &lOutput, err
	}
	fmt.Println("CHECK OUTLINE ENTRIES")
	for _, en := range nEntries {
		fmt.Println(en)
	}
	fmt.Println("END OUTLINE ENTRIES")
	lOutput.NetworkEntries = nEntries
	return &lOutput, nil
}

// GetContainerEntry
func (nwo *NetworkOutput) GetContainerEntry(containerName string) *Network {
	var reNetwork Network
	for _, nw := range nwo.NetworkEntries {
		if nw.Hostname == containerName {
			reNetwork.LoadEntry(&nw)
			return &reNetwork
		}
	}
	return &reNetwork
}