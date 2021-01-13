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
	"testing"
)


// TestContainers
func TestContainers(t *testing.T) {
	t.Run("Create", testCreate)
}

// testCreate
func testCreate(t *testing.T) {
	goCluster := NewGoCluster("test", "ubuntu", "xenial", "", "")
	err := goCluster.CreateContainer("CreateTest", "ubuntu", "xenial", []byte{})
	if err != nil {
		t.Errorf("Error Creating New Container: %v", err)
	}
}
