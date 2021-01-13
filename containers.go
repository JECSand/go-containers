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
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/url"
)


// Auth
type Auth struct {
	User			string
	Type			string
	Credential		string
	SecurityString	string
	Port			int
}

// Connection
type Connection struct {
	Name			string
	Type			string
	Port			int
}

// NewConnection create a pointer to a new Connection
func NewConnection(name string, cType string, port int) *Connection {
	return &Connection{name, cType, port}
}

// Network
type Network struct {
	PublicIP		string
	PrivateIP		string
	HostName		string
	SSL				string
	DNS				string
	Connections		[]*Connection
}

// AddConnection to a Network
func (n *Network) AddConnection(conn *Connection) {
	n.Connections = append(n.Connections, conn)
}

// A GoContainer defines the structure of a lxc container
type GoContainer struct {
	Name			string
	Controller		bool
	Type			string
	Services		[]string
	InitFile      	[]byte
	Storage			string
	Network			*Network
	Auth			*Auth
}

// NewGoContainer creates a pointer to a new GoContainer
func NewGoContainer(name string, controller bool, cType string, services []string, initFile []byte, storage string, network *Network, auth *Auth) *GoContainer {
	return &GoContainer{name, controller, cType, services, initFile, storage, network, auth}
}


// Create a new GoContainer
func (co *GoContainer) Create() err {

}

// A GoCluster is a deployment of GoContainers
type GoCluster struct {
	Name			string
	Type			string
	ReverseProxy	string
	LoadBalancer	string
	Controller		string
	Containers		[]*GoContainer
	Network			*Network
}

// NewGoCluster creates a pointer to a new GoCluster
func NewGoCluster(name string, cType string, reverseProxy string, loadBalancer string, controller string) *GoCluster {
	return &GoCluster{name, cType, reverseProxy, loadBalancer, controller,[]*GoContainer{}, &Network{}}
}

// CreateContainer
func (cu *GoCluster) CreateContainer(name string, cType string, cRelease string, config []byte) error {
	// name string, sType string, env string, commands []string
	cmdStr := `lxc launch images:` + cType + `/` + cRelease + `/amd64 ` + name
	if len(config) != 0 {
		cmdStr = `lxc launch ` + cType + `: ` + name + ` --config=user.user-data="` + string(config) + `"`
	}
	commands := []string{cmdStr}
	newShell, err := NewShell(name, cType, cRelease, commands)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if err = newShell.Execute(); err != nil {
		log.Fatal(err)
		return err
	}
	if err = newShell.Resolve(); err != nil {
		log.Fatal(err)
		return err
	}
	if err = newShell.Error(); err != nil {
		log.Fatal(err)
		return err
	}
	// TODO - Load Networking info (like IP address, etc.)
	return nil
}