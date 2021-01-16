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
	"github.com/gofrs/uuid"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// GenerateUuid
func GenerateUuid() (string, error) {
	uuId, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}
	return uuId.String(), nil
}

// createYMLDirs
func createYMLDirs() error {
	_, err := os.Stat("inits")
	if os.IsNotExist(err) {
		errDir := os.MkdirAll("inits", 0755)
		if errDir != nil {
			log.Fatal(errDir.Error())
			return errDir
		}
	}
	return nil
}

// cleanBashFile
func cleanBashFile(fName string) error {
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
	fName, err := GenerateUuid()
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

// buildInitStr
func (cm *CMD) buildInitStr() string {
	initStr := ""
	for _, c := range cm.Args {
		initStr = initStr + " " + c
	}
	return initStr
}

// buildInitCmd
func (cm *CMD) buildInitCmd() error {
	contentStr := cm.buildInitStr()
	fName, err := createBashFile("INIT", contentStr)
	cm.ScriptName = fName
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	cm.Cmd = buildBashCommand(fName)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// CMD defines a async os command
type CMD struct {
	Type        string
	ScriptName  string
	Raw         string
	Args        []string
	Cmd         *exec.Cmd
	Status      int // 0 - Error, 1 - Initialized, 2 - Executed, 3 - Finished
	OutputBytes []byte
}

// newCMD returns a pointer to a new CMD
func newCMD(raw string) (*CMD, error) {
	var cmd CMD
	cmd.Status = 0
	cmd.Raw = raw
	err := cmd.loadArgs()
	if err != nil {
		log.Fatal(err.Error())
		return &cmd, err
	}
	if strings.Contains(raw, "#cloud-config") {
		cmd.Type = "script"
		err = cmd.buildInitCmd()
	} else {
		cmd.Type = "command"
		err = cmd.buildCmd()
	}
	if err != nil {
		log.Fatal(err.Error())
		return &cmd, err
	}
	cmd.Status = 1
	return &cmd, nil
}

// loadArgs from cmd.Raw
func (cm *CMD) loadArgs() error {
	var subs []string
	reString := `\$\(\"((?:.*?\r?\n?)*)\"\)\$`
	regx := regexp.MustCompile(reString)
	doubleQuotes := regx.FindAll([]byte(cm.Raw), -1)
	for _, dq := range doubleQuotes {
		key, err := GenerateUuid()
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		subEnt := key + "|||" + string(dq)
		if strings.Contains(cm.Raw, string(dq)) {
			cm.Raw = strings.Replace(cm.Raw, string(dq), key, 1)
		}
		subs = append(subs, subEnt)
	}
	cm.Args = strings.Split(cm.Raw, " ")
	for ind, arg := range cm.Args {
		for _, sub := range subs {
			sSub := strings.Split(sub, "|||")
			if strings.Contains(arg, sSub[0]) {
				nArg := strings.Replace(arg, sSub[0], sSub[1], 1)
				nArg = strings.Replace(strings.Replace(nArg, `$("`, `"$(`, -1), `")$`, `)"`, -1)
				cm.Args[ind] = nArg
			}
		}
	}
	return nil
}

// buildCmd a CMD
func (cm *CMD) buildCmd() error {
	cmd := exec.Command(cm.Args[0], cm.Args[1:]...)
	cm.Cmd = cmd
	return nil
}

// cleanScript
func (cm *CMD) cleanScript() error {
	return cleanBashFile(cm.ScriptName)
}

// Execute a CMD
func (cm *CMD) Execute() error {
	stdout, err := cm.Cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cm.Cmd.Start()
	dat, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cm.Cmd.Wait()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	chkRes := string(dat)
	chkRes = strings.Replace(chkRes, " ", "", -1)
	chkRes = strings.Replace(chkRes, "\n", "", -1)
	chkRes = strings.Replace(chkRes, "\t", "", -1)
	if cm.Type == "script" {
		time.Sleep(20 * time.Second)
		err = cm.cleanScript()
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
	} else if chkRes == "[]" {
		dat = []byte("[RETRY]")
		cm.OutputBytes = dat
		return nil
	}
	cm.OutputBytes = dat
	cm.Status = 2
	return nil
}

// Shell wraps around a slice of sh CMD
type Shell struct {
	Name     string
	Type     string
	Commands []*CMD
	Status   int // 0 - Error, 1 - Initialized, 2 - Executed, 3 - Finished
}

// NewShell initializes a new sh Shell
func NewShell(name string, sType string, commands []string) (*Shell, error) {
	var sCMDs []*CMD
	for _, command := range commands {
		sCMD, err := newCMD(command)
		if err != nil {
			log.Fatal(err.Error())
			return &Shell{Status: 0}, err
		}
		sCMDs = append(sCMDs, sCMD)
	}
	return &Shell{name, sType, sCMDs, 1}, nil
}

// Execute a Shell
func (sh *Shell) Execute() error {
	var err error
	for ind, cMD := range sh.Commands {
		err = cMD.Execute()
		if err != nil {
			sh.Status = 0
			log.Fatal(err.Error())
			return err
		}
		sh.Commands[ind] = cMD
	}
	sh.Status = 2
	return nil
}

// OutputBytes from Shell Execution
func (sh *Shell) OutputBytes() [][]byte {
	var reContents [][]byte
	for _, cmd := range sh.Commands {
		reContents = append(reContents, cmd.OutputBytes)
	}
	return reContents
}

// Run a Shell
func (sh *Shell) Run() error {
	if err := sh.Execute(); err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}
