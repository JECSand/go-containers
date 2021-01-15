# go-containers

A pre-made Golang module for easily spinning up and managing lxd containers and cluster.

[![Go Report Card](https://goreportcard.com/badge/github.com/JECSand/go-containers)](https://goreportcard.com/report/github.com/JECSand/go-containers)

* Author(s): John Connor Sanders
* License: Apache Version 2.0
* Version Release Date: 01/13/2021
* Current Version: 0.0.1
* Developed for Ubuntu Linux

## License
* Copyright 2021 John Connor Sanders

This source code of this package is released under the Apache Version 2.0 license. Please see
the [LICENSE](https://github.com/JECSand/go-containers/blob/main/LICENSE) for the full
content of the license.

## Installation
```bash
$ go get github.com/JECSand/go-containers
```

## Dependencies
####Ensure that LXD is installed and running on your environment
* Use the provided lxd installer script if needed:
```bash
$ . ./build_lxd.sh
```

## Usage Examples
```go
package main

import (
	"fmt"
	"github.com/JECSand/go-containers"
	"log"
)

func main() {
	// #1: Declare a Cluster of LXC GoContainers
	goCluster := containers.NewGoCluster("test", "ubuntu", "HA-Proxy", "nginx", "")
	// #2: Scan your cluster for existing containers at any time to reload the GoContainer Map
	clusterContains, err := goCluster.Scan()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(clusterContains)
	// #3: Create a New GoContainer
	//       -Params: Username, AuthType, Pass/Key, SecretKey, SSH Port
	cAuth := containers.NewAuth("envUserName", "password", "envPW", "SecurityString*", "22") // Auth for the GoContainer
	// *Note: SecurityString is for SSH Key Auth, which is coming soon
	isCluterControllerNode := true
	//  B. Create the new GoContainer
	//       -Params: Auth, isController, ContainerName, ContainerOS, osRelease, CloudInitFile
	err = goCluster.CreateContainer(cAuth, isCluterControllerNode, "testContainer", "ubuntu", "xenial", []byte{})
	if err != nil {
		log.Fatal(err.Error())
	}
	// #4: Get a GoContainer from GoCluster
	goCon, err := goCluster.GetContainer("testContainer")
	if err != nil {
		log.Fatal(err.Error())
	}
	// #5: Open a SSHClient on your GoContainer
	err = goCon.OpenSSH()
	if err != nil {
		log.Fatal(err.Error())
	}
	// #6: Execute Shell Commands on your GoContainer
	cmd := "echo Hello GoContainer ; sudo apt-get -y update"
	sshSession, _ := goCon.SSHClient.SSHConn.NewSession()
	defer sshSession.Close()
	sshSession.Run(cmd) // Execute the SSH Command
	// #7: Close the SSHClient Session
	goCon.SSHClient.close()
	// #8: Delete the GoContainer from GoCluster
	err = goCluster.DeleteContainer("testContainer")
	if err != nil {
		log.Fatal(err.Error())
	}
	// #9: Get all GoContainers in GoCluster
	goCons, err := goCluster.GetContainers()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(goCons)
}
```