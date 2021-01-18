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
	"fmt"
	"strings"
	"testing"
)

// TestContainers
func TestContainers(t *testing.T) {
	t.Run("ClusterScan", testClusterScan)
	t.Run("InitAuth", testInitAuth)
	t.Run("testSnapshot", testSnapshot)
	t.Run("testContainerCMD", testContainerCMD)
	t.Run("testImportExport", testImportExport)
}

// testClusterScan
func testClusterScan(t *testing.T) {
	fmt.Println("\n<-----------BEGINNING 1: testClusterScan...")
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := &Auth{}
	fmt.Println("----------->BEGINNING 1.A: Create a Container...")
	err := goCluster.CreateContainer(cAuth, true, "ClusterTest1", "ubuntu", "xenial", []byte{})
	if err != nil {
		fmt.Println("Error Creating Test Scan Container: ", err.Error())
		t.Errorf("Error Creating Test Container: %v", err)
	}
	fmt.Println("----------->PASSED 1.A: Create Snapshot...")
	fmt.Println("----------->BEGINNING 1.B: Run a Cluster Scan...")
	clusterContains, err := goCluster.Scan()
	if err != nil {
		fmt.Println("Error Scanning Test Cluster's Containers: ", err.Error())
		t.Errorf("Error Scanning Test Cluster's Containers: %v", err)
	} else if len(clusterContains) == 0 {
		fmt.Println("Error: Failed to Scan all Test Containers in Cluster!")
		_ = goCluster.DeleteContainer("ClusterTest1")
		t.Errorf("Failed to Scan all Test Containers in Cluster!")
	}
	fmt.Println("----------->PASSED 1.B: Run a Cluster Scan...")
	fmt.Println("----------->BEGINNING 1.C: Delete a Container...")
	err = goCluster.DeleteContainer("ClusterTest1")
	if err != nil {
		fmt.Println("Error Deleting Test Container: ", err.Error())
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("----------->PASSED 1.C: Delete a Container...")
	fmt.Println("<-----------testClusterScan COMPLETE")
}

// testInitAuth
func testInitAuth(t *testing.T) {
	fmt.Println("\n<-----------BEGINNING 2: testInitAuth...")
	username := "tester"
	password := "l0lThis1sAWeak1"
	aType := "password"
	port := "2222"
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := NewAuth(username, aType, password, "", port)
	fmt.Println("----------->BEGINNING 2.A: Create an Auth Container...")
	err := goCluster.CreateContainer(cAuth, true, "CreateInitTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		fmt.Println("Error Creating Test Auth Container: ", err.Error())
		_ = goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Creating Test Container: %v", err)
	}
	fmt.Println("----------->PASSED 2.A: Create an Auth Container...")
	fmt.Println("----------->BEGINNING 2.B: Get a Container...")
	goCon, err := goCluster.GetContainer("CreateInitTest")
	if err != nil {
		fmt.Println("Error Getting Test Auth Container: ", err.Error())
		_ = goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	goCon.Auth = cAuth
	fmt.Println("----------->PASSED 2.B: Get a Container...")
	fmt.Println("----------->BEGINNING 2.C: Open a Container SSH Client...")
	err = goCon.OpenSSH()
	if err != nil {
		fmt.Println("Error Opening Test Container SSHClient: ", err.Error())
		_ = goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Opening SSHClient for Test Container: %v", err)
	}
	fmt.Println("----------->PASSED 2.C: Open a Container SSH Client...")
	fmt.Println("----------->BEGINNING 2.D: Close Container SSH Client...")
	err = goCon.SSHClient.Close()
	if err != nil {
		fmt.Println("Error Closing Container SSHClient: ", err.Error())
		_ = goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Closing Test Container SSH Client: %v", err)
	}
	fmt.Println("----------->PASSED 2.D: Close Container SSH Client...")
	err = goCluster.DeleteContainer("CreateInitTest")
	if err != nil {
		fmt.Println("Error Deleting Test Auth Container: ", err.Error())
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("<-----------testInitAuth COMPLETE")
}

// testSnapshot
func testSnapshot(t *testing.T) {
	fmt.Println("\n<-----------BEGINNING 3: testSnapshot...")
	username := "tester"
	password := "l0lThis1sAWeak1"
	aType := "password"
	port := "22"
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := NewAuth(username, aType, password, "", port)
	err := goCluster.CreateContainer(cAuth, true, "SnapshotTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		fmt.Println("Error Creating Test Snapshot Container: ", err.Error())
		t.Errorf("Error Creating Test Container: %v", err)
	}
	goCon, err := goCluster.GetContainer("SnapshotTest")
	if err != nil {
		_ = goCluster.DeleteContainer("SnapshotTest")
		fmt.Println("Error Getting Test Snapshot Container: ", err.Error())
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("----------->BEGINNING 3.A: Create Snapshot...")
	_, err = goCon.CreateSnapshot()
	if err != nil {
		_ = goCluster.DeleteContainer("SnapshotTest")
		fmt.Println("Error Creating Test Snapshot: ", err.Error())
		t.Errorf("Error Creating Test Snapshot: %v", err)
	}
	fmt.Println("----------->PASSED 3.A: Create Snapshot...")
	fmt.Println("----------->BEGINNING 3.B: Get Snapshot...")
	conSnaps, err := goCon.GetSnapshots()
	if err != nil {
		_ = goCluster.DeleteContainer("SnapshotTest")
		fmt.Println("Error Getting Test Snapshot: ", err.Error())
		t.Errorf("Error Getting Test Snapshots: %v", err)
	}
	fmt.Println("----------->PASSED 3.B: Create Snapshot...")
	fmt.Println("----------->BEGINNING 3.C: Delete Snapshot...")
	err = goCon.DeleteSnapshot(conSnaps[0].Name)
	if err != nil {
		_ = goCluster.DeleteContainer("SnapshotTest")
		fmt.Println("Error Deleting Test Snapshot: ", err.Error())
		t.Errorf("Error Deleting Test Snapshot: %v", err)
	}
	fmt.Println("----------->PASSED 3.C: Delete Snapshot...")
	err = goCluster.DeleteContainer("SnapshotTest")
	if err != nil {
		fmt.Println("Error Deleting Test Snapshot Container: ", err.Error())
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("<-----------testSnapshot COMPLETE")
}

// testContainerCMD
func testContainerCMD(t *testing.T) {
	fmt.Println("\n<-----------BEGINNING 4: testContainerCMD...")
	expectOuts := []string{"/root\n", "/home\n", "/home/ubuntu\n"}
	goCluster := NewGoCluster("test", "ubuntu", "", "", "")
	cAuth := &Auth{}
	err := goCluster.CreateContainer(cAuth, true, "CMDTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		fmt.Println("Error Creating Test CMD Container: ", err.Error())
		t.Errorf("Error Creating Test CMD Container: %v", err)
	}
	goCon, err := goCluster.GetContainer("CMDTest")
	if err != nil {
		fmt.Println("Error Getting Test CMD Container: ", err.Error())
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error Deleting CMD Container: %v", err)
	}
	/*=============================TEST-1===============================*/
	fmt.Println("----------->BEGINNING 4.A: Basic Command...")
	output, err := goCon.CMD("pwd", "", true)
	if err != nil {
		fmt.Println("Error Executing Test CMD 1: ", err.Error())
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error Executing Test CMD 1: %v", err)
	}
	if string(output) != expectOuts[0] {
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error COMMAND TEST 1")
	}
	fmt.Println("----------->PASSED 4.A: Basic Command...")
	/*=============================TEST-2===============================*/
	fmt.Println("----------->BEGINNING 4.B: Double Command...")
	output, err = goCon.CMD("cd /home/ && pwd", "", false)
	if err != nil {
		fmt.Println("Error Executing Test CMD: ", err.Error())
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error Executing Test CMD: %v", err)
	}
	if string(output) != expectOuts[1] {
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error COMMAND TEST 2")
	}
	fmt.Println("----------->PASSED 4.B: Double Command...")
	/*=============================TEST-3===============================*/
	fmt.Println("----------->BEGINNING 4.C: Login User Command...")
	output, err = goCon.CMD("cd ~ && pwd", "ubuntu", false)
	if err != nil {
		fmt.Println("Error Executing Test CMD: ", err.Error())
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error Executing Test CMD: %v", err)
	}
	if !strings.Contains(string(output), expectOuts[2]) {
		_ = goCluster.DeleteContainer("CMDTest")
		t.Errorf("Error COMMAND TEST 3")
	}
	fmt.Println("----------->PASSED 4.C:  Login User Command...")
	err = goCluster.DeleteContainer("CMDTest")
	if err != nil {
		fmt.Println("Error Deleting Test CMD Container: ", err.Error())
		t.Errorf("Error Deleting Test CMD Container: %v", err)
	}
	fmt.Println("<-----------testContainerCMD COMPLETE")
}

// testImportExport
func testImportExport(t *testing.T) {
	fmt.Println("\n<-----------BEGINNING 5: testImportExport...")
	username := "tester"
	password := "l0lThis1sAWeak1"
	aType := "password"
	port := "22"
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := NewAuth(username, aType, password, "", port)
	err := goCluster.CreateContainer(cAuth, true, "ImportTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		fmt.Println("Error Creating Test Export Container: ", err.Error())
		_ = goCluster.DeleteContainer("ImportTest")
		t.Errorf("Error Creating Test Export Container: %v", err)
	}
	goCon, err := goCluster.GetContainer("ImportTest")
	if err != nil {
		fmt.Println("Error Getting Test Export Container: ", err.Error())
		_ = goCluster.DeleteContainer("ImportTest")
		t.Errorf("Error Getting Test Export Container: %v", err)
	}
	fmt.Println("----------->BEGINNING 5.A: Test Export Container...")
	reImg, err := goCluster.ExportContainer(goCon.Name)
	if err != nil {
		fmt.Println("Error Exporting the Test Container: ", err.Error())
		t.Errorf("Error Exporting the Test Container: %v", err)
	}
	if len(reImg.TarMeta) != 1 || len(reImg.Contents) != 1 {
		fmt.Println("Error scanning the Import Test Exported .tar.gz file!")
		t.Errorf("Error scanning the Import Test Exported .tar.gz file!")
	}
	fmt.Println("----------->PASSED 5.A: Test Export Container...")
	err = goCluster.DeleteContainer("ImportTest")
	if err != nil {
		fmt.Println("Error Deleting Test Import Container: ", err.Error())
		t.Errorf("Error Deleting Test Import Container: %v", err)
	}
	fmt.Println("----------->BEGINNING 5.B: Test Import Container...")
	imCon, err := goCluster.ImportContainer("ImportTest2", reImg)
	if err != nil {
		fmt.Println("Error Importing the Test Container: ", err.Error())
		t.Errorf("Error Importing the Test Container: %v", err)
	}
	err = goCluster.DeleteContainer(imCon.Name)
	if err != nil {
		fmt.Println("Error Deleting Imported Test Container: ", err.Error())
		t.Errorf("Error Deleting Imported Test Container: %v", err)
	}
	fmt.Println("----------->PASSED 5.B: Test Import Container...")
	fmt.Println("<-----------testImportExport COMPLETE")
}
