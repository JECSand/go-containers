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
	"errors"
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

// Auth
type Auth struct {
	User           string
	Type           string
	Credential     string
	SecurityString string
	Port           string
}

// NewAuth create a pointer to a new Connection
func NewAuth(name string, cType string, credential string, SecurityString string, port string) *Auth {
	return &Auth{name, cType, credential, SecurityString, port}
}

// Connection
type Connection struct {
	Name string
	Type string
	Port string
}

// NewConnection create a pointer to a new Connection
func NewConnection(name string, cType string, port string) *Connection {
	return &Connection{name, cType, port}
}

// Network
type Network struct {
	PublicIP    string
	PrivateIP   string
	HWAddr      string
	Type        string
	HostName    string
	SSL         bool
	DNS         string
	Connections []*Connection
}

// AddConnection to a Network
func (n *Network) AddConnection(conn *Connection) {
	n.Connections = append(n.Connections, conn)
}

// LoadEntry
func (n *Network) LoadEntry(nw *NetworkEntry) error {
	n.PublicIP = ""
	n.HostName = nw.Hostname
	n.PrivateIP = nw.Address
	n.HWAddr = nw.HWAddr
	n.Type = nw.Type
	n.SSL = false
	return nil
}

// SSHClient
type SSHClient struct {
	SSHConn *ssh.Client
}

// newSSHClient
func newSSHClient(c *ssh.Client) *SSHClient {
	return &SSHClient{c}
}

// Close wil close an open Container SSHClient
func (ssh *SSHClient) Close() error {
	err := ssh.SSHConn.Close()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// A GoContainer defines the structure of a lxc container
type GoContainer struct {
	Name       string
	Controller bool
	SSHClient  *SSHClient
	Type       string
	Release    string
	Services   []string
	InitFile   []byte
	Storage    string
	Network    *Network
	Auth       *Auth
	Status     string
}

// NewGoContainer creates a pointer to a new GoContainer
func NewGoContainer(name string, controller bool, cType string, release string, services []string, initFile []byte, storage string, network *Network, auth *Auth) *GoContainer {
	if len(initFile) == 0 && auth.Type == "password" {
		initFile = generateCloudInit(auth)
	}
	return &GoContainer{name, controller, &SSHClient{}, cType, release, services, initFile, storage, network, auth, "Initializing"}
}

// OpenSSH begins an SSHClient session
func (co *GoContainer) OpenSSH() error {
	var conn *ssh.Client
	var err error
	if co.Auth.Type == "" {
		err = errors.New("Error containers.go: No GoContainer Auth Profile has been set for SSH!")
		log.Fatal(err.Error())
		return err
	} else if co.Auth.Credential == "" {
		err = errors.New("Error containers.go: No GoContainer Auth Credential has been set for SSH!")
		log.Fatal(err.Error())
		return err
	} else if co.Auth.User == "" {
		err = errors.New("Error containers.go: No GoContainer Auth User has been set for SSH!")
		log.Fatal(err.Error())
		return err
	} else if co.Auth.Port == "" {
		err = errors.New("Error containers.go: No GoContainer Auth Port has been set for SSH!")
		log.Fatal(err.Error())
		return err
	}
	//TODO - Add Functionality to do Private or Public based on IP Type
	addr := co.Network.PrivateIP + ":" + co.Auth.Port
	config := &ssh.ClientConfig{
		User: co.Auth.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(co.Auth.Credential),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err = ssh.Dial("tcp", addr, config)
	if err != nil {
		nErr := errors.New("failed to ssh into container: " + err.Error())
		log.Fatal(nErr.Error())
		return nErr
	}
	co.SSHClient = newSSHClient(conn)
	return nil
}

// shellCMD
func (co *GoContainer) shellCMD(cmdStr string) ([][]byte, error) {
	var outBytes [][]byte
	commands := []string{cmdStr}
	newShell, err := NewShell(co.Name, co.Type, commands)
	if err != nil {
		log.Fatal(err.Error())
		return outBytes, err
	}
	if err = newShell.Run(); err != nil {
		log.Fatal(err.Error())
		return outBytes, err
	}
	return newShell.OutputBytes(), nil
}

// Create a new GoContainer
func (co *GoContainer) Create() error {
	cmdStr := `lxc launch images:` + co.Type + `/` + co.Release + `/amd64 ` + co.Name
	if len(co.InitFile) != 0 {
		cmdStr = `lxc launch ` + co.Type + `: ` + co.Name + ` --config=user.user-data=$("` + string(co.InitFile) + `")$`
	}
	_, err := co.shellCMD(cmdStr)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Delete
func (co *GoContainer) Delete() error {
	stopCmdStr := `lxc stop ` + co.Name
	delCmdStr := `lxc delete ` + co.Name
	_, err := co.shellCMD(stopCmdStr)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	_, err = co.shellCMD(delCmdStr)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// loadNetworkData
func (co *GoContainer) loadNetworkData(networkInt string) error {
	cmdStr := `lxc network list-leases ` + networkInt + ` --format json`
	oBytes, err := co.shellCMD(cmdStr)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	if string(oBytes[0]) == "[RETRY]" {
		time.Sleep(5 * time.Second)
		oBytes, err = co.shellCMD(cmdStr)
		if err != nil {
			log.Fatal(err.Error())
			return err
		} else {
			if string(oBytes[0]) == "[RETRY]" {
				time.Sleep(5 * time.Second)
				oBytes, err = co.shellCMD(cmdStr)
				if err != nil {
					log.Fatal(err.Error())
					return err
				}
			}
		}
	}
	nwOut, err := LoadNetworkOutput(string(oBytes[0]))
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	network := nwOut.GetContainerEntry(co.Name)
	co.Network = network
	return nil
}

// loadListOutput
func (co *GoContainer) loadListOutput(output *ContainerOutput) error {
	co.Name = output.Name
	co.Type = output.Config.ImageOS
	co.Release = output.Config.ImageRelease
	co.Status = output.Status
	return nil
}

// A GoCluster is a deployment of GoContainers
type GoCluster struct {
	Name         string
	Type         string
	ReverseProxy string
	LoadBalancer string
	Controller   string
	Containers   []*GoContainer
	Network      *Network
}

// NewGoCluster creates a pointer to a new GoCluster
func NewGoCluster(name string, cType string, reverseProxy string, loadBalancer string, controller string) *GoCluster {
	return &GoCluster{name, cType, reverseProxy, loadBalancer, controller, []*GoContainer{}, &Network{}}
}

// GetContainer gets a single container back from the Cluster with GoContainer name as the filter
func (cu *GoCluster) GetContainer(cName string) (*GoContainer, error) {
	var goContainer *GoContainer
	_, err := cu.Scan()
	if err != nil {
		log.Fatal(err.Error())
		return goContainer, err
	}
	for _, con := range cu.Containers {
		if con.Name == cName {
			con.loadNetworkData("lxdbr0")
			return con, nil
		}
	}
	return goContainer, nil
}

// GetContainers gets all containers for a given GoCluster
func (cu *GoCluster) GetContainers() ([]*GoContainer, error) {
	for ind, _ := range cu.Containers {
		cu.Containers[ind].loadNetworkData("lxdbr0")
	}
	return cu.Containers, nil
}

// shellCMD executes a unix shell Cmd
func (cu *GoCluster) shellCMD(cmdStr string) ([][]byte, error) {
	var outBytes [][]byte
	commands := []string{cmdStr}
	newShell, err := NewShell(cu.Name, cu.Type, commands)
	if err != nil {
		log.Fatal(err.Error())
		return outBytes, err
	}
	if err = newShell.Run(); err != nil {
		log.Fatal(err.Error())
		return outBytes, err
	}
	return newShell.OutputBytes(), nil
}

// Scan gets all containers for a given cluster
func (cu *GoCluster) Scan() ([]*GoContainer, error) {
	var reContains []*GoContainer
	cmdStr := `lxc ls --format json`
	oBytes, err := cu.shellCMD(cmdStr)
	if err != nil {
		log.Fatal(err.Error())
		return reContains, err
	}
	reOuts, err := LoadListOut(string(oBytes[0]))
	if err != nil {
		log.Fatal(err.Error())
		return reContains, err
	}
	cu.Containers = reOuts.GetContainers()
	return cu.Containers, nil
}

// DeleteContainer deleted the container whose name is inputted
func (cu *GoCluster) DeleteContainer(cName string) error {
	var newContainers []*GoContainer
	for _, con := range cu.Containers {
		if con.Name == cName {
			con.Delete()
		} else {
			newContainers = append(newContainers, con)
		}
	}
	cu.Containers = newContainers
	return nil
}

// CreateContainer create a new container
func (cu *GoCluster) CreateContainer(auth *Auth, controller bool, name string, cType string, cRelease string, config []byte) error {
	newContainer := NewGoContainer(name, controller, cType, cRelease, []string{}, config, "default", &Network{}, auth)
	err := newContainer.Create()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	newContainer.loadNetworkData("lxdbr0")
	if newContainer.Auth.Type != "" {
		newConn := NewConnection("auth", "ssh", newContainer.Auth.Port)
		newContainer.Network.AddConnection(newConn)
	}
	cu.Containers = append(cu.Containers, newContainer)
	return nil
}
