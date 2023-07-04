# go-containers

A pre-made Golang module for easily spinning up and managing lxd containers.

[![Go Report Card](https://goreportcard.com/badge/github.com/JECSand/go-containers)](https://goreportcard.com/report/github.com/JECSand/go-containers)

* Author: John Connor Sanders
* License: Apache Version 2.0
* Version Release Date: 04/18/2021
* Current Version: 0.0.3
* Developed for Ubuntu 20.x

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

### Ensure that LXD is installed and running on your environment

* Use the provided lxd installer script if needed:
```bash
$ . ./build_lxd.sh
```
________
## Module Main Data Structs
###1. GoCluster
```go
type GoCluster struct {
    Name         string
    Type         string
    ReverseProxy string
    LoadBalancer string
    Controller   string
    Containers   []*GoContainer
    Images       []*GoImage
    Network      *Network
}
```

- ###GoCluster Methods

    I.  *Scan()*
    ```go
    func (gc *GoCluster) Scan() ([]*GoContainers, error)
    ```

    II.  *CreateContainer()*
    ```go
    func (gc *GoCluster) CreateContainer(auth *Auth, controller bool, name string, cType string, cRelease string, config []byte) error
    ```
  
    III. *GetContainers()*
    ```go
    func (gc *GoCluster) GetContainers() ([]*GoContainer, error)
    ```
  
    IV.  *GetContainer()*
    ```go
    func (gc *GoCluster) GetContainer(containerName string) (*GoContainer, error)
    ```

    V.  *DeleteContainer()*
    ```go
    func (gc *GoCluster) DeleteContainer(containerName string) error
    ```
  
    VI. *ExportContainer()*
    ```go
    func (gc *GoCluster) ExportContainer(containerName string) (*GoImage, error)
    ```

    VII. *ImportContainer()*
    ```go
    func (gc *GoCluster) ImportContainer(containerName string, image *GoImage) (*GoContainer, error)
    ```
  
    VIII. *ScanImages()*
    ```go
    func (gc *GoCluster) ScanImages() ([]*GoImage, error)
    ```
  
    IX. *CreateImage()*
    ```go
    func (gc *GoCluster) CreateImage(containerName string, sName string) error 
    ```
  
    X. *DeleteImage()*
    ```go
    func (gc *GoCluster) DeleteImage(fingerprint string) error
    ```

###2. Network
```go
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
```
###3. GoContainer
```go
type GoContainer struct {
    Name        string
    Controller  bool
    SSHClient   *SSHClient
    Type        string
    Release     string
    Services    []string
    InitFile    []byte
    Storage     string
    Network     *Network
    Auth        *Auth
    GoSnapshots []*GoSnapshot
    Status      string
}
```

- ###GoContainer Methods

  I.  *Create()*
    ```go
    func (c *GoContainer) Create() error
    ```

  II.  *Stop()*
    ```go
    func (c *GoContainer) Stop() error
    ```

  II.  *Boot()*
    ```go
    func (c *GoContainer) Boot() error
    ```

  III.  *Reboot()*
    ```go
    func (c *GoContainer) Reboot() error
    ```

  IV.  *Delete()*
    ```go
    func (c *GoContainer) Delete() error
    ```

  V.  *CreateSnapshot()*
    ```go
    func (c *GoContainer) CreateSnapshot() (string, error)
    ```

  VI.  *GetSnapshots()*
    ```go
    func (c *GoContainer) GetSnapshots() ([]*GoSnapshot, error) 
    ```

  VII.  *Restore()*
    ```go
    func (c *GoContainer) Restore(snapshotName string) error
    ```

  VIII.  *Image()*
    ```go
    func (c *GoContainer) Image(snapshotName string) (*GoImage, error) 
    ```

  IX.  *Export()*
    ```go
    func (c *GoContainer) Export() (*GoImage, error) 
    ```
  
  X.  *Import()*
    ```go
    func (c *GoContainer) Import(image *GoImage) error
    ```

  XI.  *CMD()*
    ```go
    func (c *GoContainer) CMD(cmd string, userName string, reErr bool) ([]byte, error)
    ```
  
  XII.  *OpenSSH()*
    ```go
    func (c *GoContainer) OpenSSH() error 
    ```


###4. GoContainer.Auth
```go
type Auth struct {
    User           string
    Type           string
    Credential     string
    SecurityString string
    Port           string
}
```

###5. GoContainer.GoSnapshot
```go
type GoSnapshot struct {
    Name       string
    DateTime   string
}
```
###6. GoContainer.SSHClient
```go
type SSHClient struct {
    SSHConn   *ssh.Client
}
```

- ###SSHClient Methods

  I.  *Close()*
    ```go
    func (ssh *SSHClient) Close() error
    ```
  
###7. GoImage
```go
type GoImage struct {
    Name        string
    Type        string
    Fingerprint string
    TarMeta     []string
    Contents    [][]byte
    DateTime    string
}
```

- ###GoImage Methods

  I.  *Import()*
    ```go
    func (im *GoImage) Import() error
    ```

  II.  *Export()*
    ```go
    func (im *GoImage) Export() error
    ```
  
__________
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
	//  Create the new GoContainer
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
	goCon.Auth = cAuth
	
	// #5: Open a SSHClient on your GoContainer
	err = goCon.OpenSSH()
	if err != nil {
		log.Fatal(err.Error())
	}
	
	// #6: Execute Shell Commands on your GoContainer via SSH
	cmd := "echo Hello GoContainer ; sudo apt-get -y update"
	sshSession, _ := goCon.SSHClient.SSHConn.NewSession()
	defer sshSession.Close()
	sshSession.Run(cmd) // Execute the SSH Command
	// #7: Close the SSHClient Session
	goCon.SSHClient.close()
	
	// #8: Execute a Shell Command on your GoContainer via lxc exec
	out, _ := goCon.CMD("pwd", "ubuntu", false)
	fmt.Println("Remote Command Output: ", string(out))
	
	// #9: Export GoContainer to GoImage
	reImg, err := goCluster.ExportContainer("testContainer")
	if err != nil {
		log.Fatal("Error Exporting the GoContainer: ", err.Error())
	}
	// Exported GoContainer's contents stored in tar.gz format
	fmt.Println(reImg.TarMeta[0])
	fmt.Println(len(reImg.Contents[0]))

	// #10: Import GoContainer from a GoImage
	imCon, err := goCluster.ImportContainer("ImportedContainer", reImg)
	if err != nil {
		log.Fatal("Error Importing the GoContainer: ", err.Error())
	}
	fmt.Println(imCon.Name)
	
	// #11: Delete GoContainer from GoCluster
	err = goCluster.DeleteContainer("testContainer")
	if err != nil {
		log.Fatal(err.Error())
	}
	
	// #12: Get all GoContainers in GoCluster
	goCons, err := goCluster.GetContainers()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(goCons)
}
```