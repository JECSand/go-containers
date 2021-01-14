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
	"fmt"
	"io/ioutil"
	"testing"
)

// TestContainers
func TestContainers(t *testing.T) {
	t.Run("BasicCreate", testBasicCreate)
	t.Run("InitCreate", testInitCreate)
	t.Run("ClusterScan", testClusterScan)
	t.Run("InitAuth", testInitAuth)
}

// testBasicCreate
func testBasicCreate(t *testing.T) {
	fmt.Println("<-----------BEGINNING 1: testBasicCreate...")
	goCluster := NewGoCluster("test", "ubuntu", "", "", "")
	cAuth := &Auth{}
	err := goCluster.CreateContainer(cAuth, true, "CreateTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		t.Errorf("Error Creating Test Container: %v", err)
	}
	err = goCluster.DeleteContainer("CreateTest")
	if err != nil {
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("<-----------testBasicCreate COMPLETE")
}

// testInitCreate
func testInitCreate(t *testing.T) {
	fmt.Println("<-----------BEGINNING 2: testInitCreate...")
	username := "tester"
	password := "l0lThis1sAWeak1"
	aType := "password"
	port := "22"
	dat, err := ioutil.ReadFile("test-init-conf.yml")
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := NewAuth(username, aType, password, "", port)
	err = goCluster.CreateContainer(cAuth, true, "CreateInitTest", "ubuntu", "xenial", dat)
	if err != nil {
		goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Creating Test Container: %v", err)
	}
	err = goCluster.DeleteContainer("CreateInitTest")
	if err != nil {
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("<-----------testInitCreate COMPLETE")
}

// testClusterScan
func testClusterScan(t *testing.T) {
	fmt.Println("<-----------BEGINNING 3: testClusterScan...")
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := &Auth{}
	err := goCluster.CreateContainer(cAuth, true, "ClusterTest1", "ubuntu", "xenial", []byte{})
	if err != nil {
		t.Errorf("Error Creating Test Container: %v", err)
	}
	clusterContains, err := goCluster.Scan()
	if err != nil {
		t.Errorf("Error Scanning Test Cluster's Containers: %v", err)
	} else if len(clusterContains) == 0 {
		goCluster.DeleteContainer("ClusterTest1")
		t.Errorf("Failed to Scan all Test Containers in Cluster!")
	}
	err = goCluster.DeleteContainer("ClusterTest1")
	if err != nil {
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("<-----------testClusterScan COMPLETE")
}

// testInitAuth
func testInitAuth(t *testing.T) {
	fmt.Println("<-----------BEGINNING 4: testInitAuth...")
	username := "tester"
	password := "l0lThis1sAWeak1"
	aType := "password"
	port := "22"
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	cAuth := NewAuth(username, aType, password, "", port)
	err := goCluster.CreateContainer(cAuth, true, "CreateInitTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Creating Test Container: %v", err)
	}
	goCon, err := goCluster.GetContainer("CreateInitTest")
	goCon.Auth = cAuth
	if err != nil {
		goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	err = goCon.OpenSSH()
	if err != nil {
		goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Opening SSHClient for Test Container: %v", err)
	}
	err = goCon.SSHClient.Close()
	if err != nil {
		goCluster.DeleteContainer("CreateInitTest")
		t.Errorf("Error Closing Test Container SSH Client: %v", err)
	}
	err = goCluster.DeleteContainer("CreateInitTest")
	if err != nil {
		t.Errorf("Error Deleting Test Container: %v", err)
	}
	fmt.Println("<-----------testInitAuth COMPLETE")
}
