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
	"github.com/gofrs/uuid"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// getTimeStamp
func getTimeStamp() string {
	currentTime := time.Now().UTC()
	t := strings.Split(currentTime.String(), " +")[0]
	t = strings.Replace(t, " ", "T", 1)
	t = strings.Replace(t, " ", "", -1)
	t = strings.Replace(t, ":", "HH", 1)
	t = strings.Replace(t, ":", "MM", 1)
	t = strings.Split(t, ".")[0]
	t = t + "SS-UTC"
	return t
}

// generateUuid
func generateUuid() (string, error) {
	uuId, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}
	return uuId.String(), nil
}

// createDir
func createDir(dirName string) error {
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			fmt.Println("Error in utils.go on line 53: Creating the directory for ", dirName, " ", errDir.Error())
			log.Fatal(errDir.Error())
			return errDir
		}
	}
	return nil
}

// createYMLDirs
func createYMLDirs() error {
	return createDir("inits")
}

// createJobDirectory
func createJobDirectory(jType string) (string, error) {
	jobId, err := generateUuid()
	if err != nil {
		fmt.Println("Error in utils.go on line 68: create the Job Uuid for ", jType, " ", err.Error())
		log.Fatal(err.Error())
		return jobId, err
	}
	err = createDir(jType)
	if err != nil {
		fmt.Println("Error in utils.go on line 76: create the directory for ", jType, ":", jobId, " ", err.Error())
		log.Fatal(err.Error())
		return jobId, err
	}
	jobDir := jType + "/" + jobId
	return jobId, createDir(jobDir)
}

// deleteJobDirectory
func deleteJobDirectory(jType string, jobId string) error {
	jobDir := jType + "/" + jobId
	return os.RemoveAll(jobDir)
}

// scanFile
func scanFile(fileName string) ([]byte, string, error) {
	var contents []byte
	var fName string
	sFName := strings.Split(fileName, "/")
	fName = sFName[len(sFName)-1]
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error in utils.go on line 96: Opening the tar.gz file, ", fileName, " ", err.Error())
		log.Fatal(err.Error())
		return contents, fName, err
	}
	defer file.Close()
	contents, err = ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error in utils.go on line 103: Reading the tar.gz file, ", fileName, " ", err.Error())
		log.Fatal(err.Error())
		return contents, fName, err
	}
	return contents, fName, nil
}

// scanJobDirectory
func scanJobDirectory(jType string, jobId string) ([][]byte, []string, error) {
	var dirContents [][]byte
	var dirNames []string
	var files []string
	root := jType + "/" + jobId
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".gz" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		fmt.Println("Error in utils.go on line 118: Scan the Job Directory, ", jType, " ", jobId, " ", err.Error())
		log.Fatal(err.Error())
		return dirContents, dirNames, err
	}
	for _, file := range files {
		fContents, fName, fErr := scanFile(file)
		if fErr != nil {
			fmt.Println("Error in utils.go on line 135: Scanning the tar.gz file,  ", fName, " ", fErr.Error())
			log.Fatal(fErr.Error())
			return dirContents, dirNames, fErr
		}
		dirContents = append(dirContents, fContents)
		dirNames = append(dirNames, fName)
	}
	return dirContents, dirNames, nil
}

// createFile
func createFile(fName string, fContent []byte) error {
	f, err := os.Create(fName)
	if err != nil {
		fmt.Println("Error in utils.go on line 149: Creating the new tar.gz file,  ", fName, " ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	defer f.Close()
	_, err = f.Write(fContent)
	if err != nil {
		fmt.Println("Error in utils.go on line 135: Writing the new tar.gz file,  ", fName, " ", err.Error())
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// deleteFile
func deleteFile(fName string) error {
	err := os.Remove(fName)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// createBashFile
func createBashFile(fType string, contents string) (string, error) {
	err := createYMLDirs()
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}
	fName, err := generateUuid()
	if err != nil {
		log.Fatal(err.Error())
		return fName, err
	}
	if fType == "INIT" {
		fName = "inits/" + fName + ".sh"
	}
	f, err := os.Create(fName)
	if err != nil {
		log.Fatal(err.Error())
		return fName, nil
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		log.Fatal(err.Error())
		return fName, nil
	}
	return fName, nil
}

// buildBashCommand
func buildBashCommand(fName string) *exec.Cmd {
	return exec.Command("/bin/sh", fName)
}
