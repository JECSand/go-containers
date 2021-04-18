/*
Author: John Connor Sanders
License: Apache Version 2.0
Version: 0.0.3
Released: 04/18/2021
Copyright 2021 John Connor Sanders

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
------------GO-CONTAINERS----------------
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package containers

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"strings"
	"time"
)

// Auth stores auth information for a GoContainer's user
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

// Connection stores info relating to externally connecting to a GoContainer
type Connection struct {
	Name string
	Type string
	Port string
}

// NewConnection create a pointer to a new Connection
func NewConnection(name string, cType string, port string) *Connection {
	return &Connection{name, cType, port}
}

// Network stores the network info for a GoContainer
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

// LoadEntry loads a lxc data into a pointer to NetworkEntry struct
func (n *Network) LoadEntry(nw *NetworkEntry) error {
	n.PublicIP = ""
	n.HostName = nw.Hostname
	n.PrivateIP = nw.Address
	n.HWAddr = nw.HWAddr
	n.Type = nw.Type
	n.SSL = false
	return nil
}

// SSHClient stores a pointer to a GoContainer ssh.Client
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
		fmt.Println("ERROR: containers.go, line 92: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// GoSnapshot represents a GoContainer's snapshot
type GoSnapshot struct {
	Name     string
	DateTime string
}

// NewGoSnapshot loads a GoSnapshot
func NewGoSnapshot(name string, dt string) *GoSnapshot {
	return &GoSnapshot{
		Name:     name,
		DateTime: dt,
	}
}

// GoImage represents a GoImage of a GoContainer stored in the cluster
type GoImage struct {
	Name        string
	Type        string
	Fingerprint string
	TarMeta     []string
	Contents    [][]byte
	DateTime    string
}

// NewGoImage loads a new GoImage
func NewGoImage(name string, iType string, fingerprint string, dt string) *GoImage {
	return &GoImage{
		Name:        name,
		Type:        iType,
		Fingerprint: fingerprint,
		TarMeta:     []string{},
		Contents:    [][]byte{},
		DateTime:    dt,
	}
}

// shellCMD executes a unix shell GoImage Cmd
func (im *GoImage) shellCMD(cmdStr string) ([][]byte, error) {
	return SHELL(im.Name, im.Type, cmdStr)
}

// loadNewOutput
func (im *GoImage) loadNewOutput(out []byte) {
	im.Fingerprint = "unknown"
	if bytes.Contains(out, []byte(" fingerprint: ")) {
		sOut := bytes.Split(out, []byte(" fingerprint: "))[1]
		sOut = bytes.Trim(sOut, "\n")
		im.Fingerprint = string(sOut)
	}
}

// Import a GoImage
func (im *GoImage) Import() error {
	cmdStr := `lxc image import `
	var importFiles []string
	jobId, err := createJobDirectory("imports")
	if err != nil {
		fmt.Println("ERROR: containers.go, line 156: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 162: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	importDir := pwd + `/imports/` + jobId
	for ind, fName := range im.TarMeta {
		iName := importDir + "/" + fName
		fContents := im.Contents[ind]
		err = createFile(iName, fContents)
		if err != nil {
			fmt.Println("ERROR: containers.go, line 172: ", err.Error())
			log.Fatal(err.Error())
			return err
		}
		importFiles = append(importFiles, iName)
	}
	cmdStr = cmdStr + importFiles[0] + ` --alias ` + im.Name
	_, err = im.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 281: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	err = deleteJobDirectory("imports", jobId)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 187: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Export a GoImage
func (im *GoImage) Export() error {
	cmdStr := `lxc image export ` + im.Name
	jobId, err := createJobDirectory("exports")
	if err != nil {
		fmt.Println("ERROR: containers.go, line 193: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 199: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	exportDir := pwd + `/exports/` + jobId
	cmdStr = cmdStr + ` ` + exportDir
	_, err = im.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 207: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	jobContents, jobMeta, err := scanJobDirectory("exports", jobId)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 213: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	im.TarMeta = jobMeta
	im.Contents = jobContents
	err = deleteJobDirectory("exports", jobId)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 221: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// A GoContainer defines the structure of a lxc container
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

// NewGoContainer creates a pointer to a new GoContainer
func NewGoContainer(name string, controller bool, cType string, release string, services []string, initFile []byte, storage string, network *Network, auth *Auth) *GoContainer {
	var goSnaps []*GoSnapshot
	if len(initFile) == 0 && auth.Type == "password" {
		initFile = generateCloudInit(auth)
	}
	return &GoContainer{
		name,
		controller,
		&SSHClient{},
		cType,
		release,
		services,
		initFile,
		storage,
		network,
		auth,
		goSnaps,
		"Initializing",
	}
}

// OpenSSH begins an SSHClient session
func (co *GoContainer) OpenSSH() error {
	var conn *ssh.Client
	var err error
	if co.Auth.Type == "" {
		err = errors.New("error containers.go: No GoContainer Auth Profile has been set for SSH")
		fmt.Println("ERROR: containers.go, line 272: ", err.Error())
		log.Fatal(err.Error())
		return err
	} else if co.Auth.Credential == "" {
		err = errors.New("error containers.go: No GoContainer Auth Credential has been set for SSH")
		fmt.Println("ERROR: containers.go, line 277: ", err.Error())
		log.Fatal(err.Error())
		return err
	} else if co.Auth.User == "" {
		err = errors.New("error containers.go: No GoContainer Auth User has been set for SSH")
		fmt.Println("ERROR: containers.go, line 282: ", err.Error())
		log.Fatal(err.Error())
		return err
	} else if co.Auth.Port == "" {
		err = errors.New("error containers.go: No GoContainer Auth Port has been set for SSH")
		fmt.Println("ERROR: containers.go, line 287: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	sshKeygenCMD := `ssh-keygen -R `
	if co.Auth.Port != "22" {
		sshKeygenCMD = sshKeygenCMD + `[` + co.Network.PrivateIP + `]:` + co.Auth.Port
	} else {
		sshKeygenCMD = sshKeygenCMD + co.Network.PrivateIP
	}
	_, _, _ = BASH(sshKeygenCMD)
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
		fmt.Println("ERROR: containers.go, line 302: ", nErr.Error())
		log.Fatal(nErr.Error())
		return nErr
	}
	co.SSHClient = newSSHClient(conn)
	return nil
}

// shellCMD executes a unix shell Cmd
func (co *GoContainer) shellCMD(cmdStr string) ([][]byte, error) {
	return SHELL(co.Name, co.Type, cmdStr)
}

// ensure that a container is done booting before continuing
func (co *GoContainer) ensure() error {
	chkCmd := `systemctl is-system-running`
	for i := 0; i < 36; i++ {
		out, err := co.CMD(chkCmd, "", false)
		if err != nil {
			fmt.Println("ERROR: containers.go, line 322: ", err.Error())
			log.Fatal(err.Error())
			return err
		}
		if strings.Contains(string(out), "running") {
			return nil
		} else if strings.Contains(string(out), "degraded") {
			resetCmd := `systemctl reset-failed`
			_, _ = co.CMD(resetCmd, "", false)
		}
		time.Sleep(5 * time.Second)
	}
	errStr := "Error: in containers.go, line 321: " + co.Name + " is not starting in a timely manner!"
	fmt.Println(errStr)
	return errors.New(errStr)
}

// CMD executes a command on a GoContainer
func (co *GoContainer) CMD(cmd string, userName string, reErr bool) ([]byte, error) {
	cmdStr := `lxc exec ` + co.Name + ` `
	xFlag := 0
	if userName != "" {
		cmdStr = cmdStr + ` -- sudo --login --user ` + userName + ` bash -ilc "`
		xFlag = 1
	} else if strings.Contains(cmd, " && ") || strings.Contains(cmd, " ; ") {
		cmdStr = cmdStr + ` -- sudo bash -ilc "`
		xFlag = 1
	}
	cmdStr = cmdStr + cmd
	if xFlag == 1 {
		cmdStr = cmdStr + `"`
	}
	out, errOut, err := BASH(cmdStr)
	if err != nil {
		if reErr {
			fmt.Println("ERROR: containers.go, line 356: ", err.Error())
			log.Fatal(err.Error())
			return []byte{}, err
		}
	}
	if string(errOut) != "" {
		if reErr {
			return out, errors.New(string(errOut))
		}
	}
	return out, nil
}

// Create a new GoContainer
func (co *GoContainer) Create() error {
	cmdStr := `lxc launch images:` + co.Type + `/` + co.Release + `/amd64 ` + co.Name
	if len(co.InitFile) != 0 {
		cmdStr = `lxc launch ` + co.Type + `: ` + co.Name + ` --config=user.user-data=$("` + string(co.InitFile) + `")$`
	}
	_, err := co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 378: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	err = co.ensure()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 384: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Stop shutdowns a GoContainer
func (co *GoContainer) Stop() error {
	stopCmdStr := `lxc stop ` + co.Name
	_, err := co.shellCMD(stopCmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 397: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Boot boots an offline GoContainer
func (co *GoContainer) Boot() error {
	stopCmdStr := `lxc start ` + co.Name
	_, err := co.shellCMD(stopCmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 397: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Reboot Stops then Boots an online GoContainer
func (co *GoContainer) Reboot() error {
	stopCmdStr := `lxc restart ` + co.Name
	_, err := co.shellCMD(stopCmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 397: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Delete an existing GoContainer from the GoCluster
func (co *GoContainer) Delete() error {
	stopCmdStr := `lxc stop ` + co.Name
	delCmdStr := `lxc delete ` + co.Name
	_, err := co.shellCMD(stopCmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 397: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	_, err = co.shellCMD(delCmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 403: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// checkSnapshots
func (co *GoContainer) checkSnapshots(newSnap *GoSnapshot) bool {
	for _, conSnap := range co.GoSnapshots {
		if conSnap.Name == newSnap.Name && conSnap.DateTime == newSnap.DateTime {
			return false
		}
	}
	return true
}

// loadSnapshots
func (co *GoContainer) loadSnapshots() error {
	cmdStr := `lxc info ` + co.Name
	oBytes, err := co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 425: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	if bytes.Contains(oBytes[0], []byte("Snapshots:\n")) {
		outSnaps := bytes.Split(bytes.Split(oBytes[0], []byte("Snapshots:\n"))[1], []byte("\n"))
		for _, outSnap := range outSnaps {
			if bytes.Contains(outSnap, []byte("-snap-")) {
				outSnap = bytes.Split(outSnap, []byte(" ("))[0]
				outSnap = bytes.Replace(outSnap, []byte(" "), []byte(""), -1)
				outSnap = bytes.Replace(outSnap, []byte("\t"), []byte(""), -1)
				sSnap := bytes.Split(outSnap, []byte("-snap-"))
				ss := NewGoSnapshot(string(outSnap), string(sSnap[1]))
				if co.checkSnapshots(ss) {
					co.GoSnapshots = append(co.GoSnapshots, ss)
				}
			}
		}
	}
	return nil
}

// CreateSnapshot creates a new GoSnapshot off an existing GoContainer
func (co *GoContainer) CreateSnapshot() (string, error) {
	ts := getTimeStamp()
	snapName := co.Name + "-snap-" + ts
	newSnap := NewGoSnapshot(snapName, ts)
	cmdStr := `lxc snapshot ` + co.Name + ` ` + snapName
	_, err := co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 455: ", err.Error())
		log.Fatal(err.Error())
		return snapName, err
	}
	co.GoSnapshots = append(co.GoSnapshots, newSnap)
	return snapName, nil
}

// DeleteSnapshot deletes a GoSnapshot
func (co *GoContainer) DeleteSnapshot(snapName string) error {
	cmdStr := `lxc delete ` + co.Name + `/` + snapName
	_, err := co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 468: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// GetSnapshots for all containers
func (co *GoContainer) GetSnapshots() ([]*GoSnapshot, error) {
	err := co.loadSnapshots()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 479: ", err.Error())
		log.Fatal(err.Error())
		return co.GoSnapshots, err
	}
	return co.GoSnapshots, nil
}

// Restore a GoContainer from a snapshot
func (co *GoContainer) Restore(snapName string) error {
	cmdStr := `lxc restore ` + co.Name + ` ` + snapName
	_, err := co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 491: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Image of the GoContainer
func (co *GoContainer) Image(snapShot string) (*GoImage, error) {
	var reImg GoImage
	ts := getTimeStamp()
	imgName := co.Name + "-image-"
	cmdStr := `lxc publish ` + co.Name
	reImg.Type = "Container"
	if snapShot != "" {
		reImg.Type = "Snapshot"
		cmdStr = cmdStr + `/` + snapShot
		imgName = imgName + "-snap-"
	}
	imgName = imgName + ts
	cmdStr = cmdStr + ` --alias ` + imgName
	reImg.Name = imgName
	out, err := co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 516: ", err.Error())
		log.Fatal(err.Error())
		return &reImg, err
	}
	reImg.loadNewOutput(out[0])
	return &reImg, nil
}

// Export a GoContainer from the GoCluster
func (co *GoContainer) Export() (*GoImage, error) {
	var exImage *GoImage
	imageSnap, err := co.CreateSnapshot()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 529: ", err.Error())
		log.Fatal(err.Error())
		return exImage, err
	}
	exImage, err = co.Image(imageSnap)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 535: ", err.Error())
		log.Fatal(err.Error())
		return exImage, err
	}
	err = exImage.Export()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 541: ", err.Error())
		log.Fatal(err.Error())
		return exImage, err
	}
	return exImage, nil
}

// Import a GoContainer into the GoCluster
func (co *GoContainer) Import(image *GoImage) error {
	cmdStr := `lxc launch ` + image.Name + ` ` + co.Name
	err := image.Import()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 553: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	_, err = co.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 559: ", err.Error())
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
		fmt.Println("ERROR: containers.go, line 571: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	if string(oBytes[0]) == "[RETRY]" {
		time.Sleep(5 * time.Second)
		oBytes, err = co.shellCMD(cmdStr)
		if err != nil {
			fmt.Println("ERROR: containers.go, line 579: ", err.Error())
			log.Fatal(err.Error())
			return err
		} else {
			if string(oBytes[0]) == "[RETRY]" {
				time.Sleep(5 * time.Second)
				oBytes, err = co.shellCMD(cmdStr)
				if err != nil {
					fmt.Println("ERROR: containers.go, line 587: ", err.Error())
					log.Fatal(err.Error())
					return err
				}
			}
		}
	}
	nwOut, err := LoadNetworkOutput(string(oBytes[0]))
	if err != nil {
		fmt.Println("ERROR: containers.go, line 596: ", err.Error())
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
	Images       []*GoImage
	Network      *Network
}

// NewGoCluster creates a pointer to a new GoCluster
func NewGoCluster(name string, cType string, reverseProxy string, loadBalancer string, controller string) *GoCluster {
	var imgs []*GoImage
	var containers []*GoContainer
	return &GoCluster{
		name,
		cType,
		reverseProxy,
		loadBalancer,
		controller,
		containers,
		imgs,
		&Network{},
	}
}

// shellCMD executes a unix shell Cmd
func (cu *GoCluster) shellCMD(cmdStr string) ([][]byte, error) {
	return SHELL(cu.Name, cu.Type, cmdStr)
}

// ScanImages from a GoCluster
func (cu *GoCluster) ScanImages() ([]*GoImage, error) {
	cmdStr := `lxc image list --format json`
	oBytes, err := cu.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 652: ", err.Error())
		log.Fatal(err.Error())
		return cu.Images, err
	}
	outImgs := LoadImagesOutput(string(oBytes[0]))
	cu.Images = outImgs.GetImages()
	return cu.Images, nil
}

// CreateImage an image of a GoCluster's GoContainer
func (cu *GoCluster) CreateImage(cName string, sName string) error {
	container, err := cu.GetContainer(cName)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 665: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	exImage, err := container.Image(sName)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 671: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	cu.Images = append(cu.Images, exImage)
	return nil
}

// DeleteImage an Image from the GoCluster
func (cu *GoCluster) DeleteImage(fingerprint string) error {
	cmdStr := `lxc image delete ` + fingerprint
	_, err := cu.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 684: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// ExportContainer from the GoCluster
func (cu *GoCluster) ExportContainer(cName string) (*GoImage, error) {
	var exImage *GoImage
	container, err := cu.GetContainer(cName)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 696: ", err.Error())
		log.Fatal(err.Error())
		return exImage, err
	}
	exImage, err = container.Export()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 702: ", err.Error())
		log.Fatal(err.Error())
		return exImage, err
	}
	err = cu.DeleteImage(exImage.Fingerprint)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 708: ", err.Error())
		log.Fatal(err.Error())
		return exImage, err
	}
	return exImage, nil
}

// ImportContainer into the GoCluster
func (cu *GoCluster) ImportContainer(containerName string, image *GoImage) (*GoContainer, error) {
	var newCon GoContainer
	newCon.Name = containerName
	err := newCon.Import(image)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 721: ", err.Error())
		log.Fatal(err.Error())
		return &newCon, err
	}
	err = cu.DeleteImage(image.Name)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 727: ", err.Error())
		log.Fatal(err.Error())
		return &newCon, err
	}
	return cu.GetContainer(newCon.Name)
}

// GetContainer gets a single container back from the GoCluster with GoContainer name as the filter
func (cu *GoCluster) GetContainer(cName string) (*GoContainer, error) {
	var goContainer *GoContainer
	_, err := cu.Scan()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 739: ", err.Error())
		log.Fatal(err.Error())
		return goContainer, err
	}
	for _, con := range cu.Containers {
		if con.Name == cName {
			err = con.loadNetworkData("lxdbr0")
			if err != nil {
				fmt.Println("ERROR: containers.go, line 747: ", err.Error())
				log.Fatal(err.Error())
				return con, err
			}
			return con, nil
		}
	}
	return goContainer, nil
}

// GetContainers gets all containers for a given GoCluster
func (cu *GoCluster) GetContainers() ([]*GoContainer, error) {
	for ind, _ := range cu.Containers {
		err := cu.Containers[ind].loadNetworkData("lxdbr0")
		if err != nil {
			fmt.Println("ERROR: containers.go, line 762: ", err.Error())
			log.Fatal(err.Error())
			return cu.Containers, err
		}
	}
	return cu.Containers, nil
}

// Scan gets each GoContainer in a given GoCluster
func (cu *GoCluster) Scan() ([]*GoContainer, error) {
	var reContains []*GoContainer
	cmdStr := `lxc ls --format json`
	oBytes, err := cu.shellCMD(cmdStr)
	if err != nil {
		fmt.Println("ERROR: containers.go, line 776: ", err.Error())
		log.Fatal(err.Error())
		return reContains, err
	}
	reOuts, err := LoadListOut(string(oBytes[0]))
	if err != nil {
		fmt.Println("ERROR: containers.go, line 782: ", err.Error())
		log.Fatal(err.Error())
		return reContains, err
	}
	cu.Containers = reOuts.GetContainers()
	return cu.Containers, nil
}

// DeleteContainer deletes the GoContainer whose name is inputted
func (cu *GoCluster) DeleteContainer(cName string) error {
	var newContainers []*GoContainer
	for _, con := range cu.Containers {
		if con.Name == cName {
			err := con.Delete()
			if err != nil {
				fmt.Println("ERROR: containers.go, line 797: ", err.Error())
				log.Fatal(err.Error())
				return err
			}
		} else {
			newContainers = append(newContainers, con)
		}
	}
	cu.Containers = newContainers
	return nil
}

// CreateContainer create a new GoContainer in the GoCluster
func (cu *GoCluster) CreateContainer(auth *Auth, controller bool, name string, cType string, cRelease string, config []byte) error {
	newContainer := NewGoContainer(name, controller, cType, cRelease, []string{}, config, "default", &Network{}, auth)
	err := newContainer.Create()
	if err != nil {
		fmt.Println("ERROR: containers.go, line 815: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	err = newContainer.loadNetworkData("lxdbr0")
	if err != nil {
		fmt.Println("ERROR: containers.go, line 820: ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	if newContainer.Auth.Type != "" {
		newConn := NewConnection("auth", "ssh", newContainer.Auth.Port)
		newContainer.Network.AddConnection(newConn)
	}
	cu.Containers = append(cu.Containers, newContainer)
	return nil
}
