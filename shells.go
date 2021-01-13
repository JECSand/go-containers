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
	"log"
	"os"
	"os/exec"
	"strings"
)

// CMD defines a async os command
type CMD struct {
	Raw			string
	Path		string
	Args		[]string
	Cmd			*exec.Cmd
	Status		int // 0 - Error, 1 - Initialized, 2 - Executed, 3 - Finished
	Output		os.Stdout
	Error		os.Stdout
}

// newCMD returns a pointer to a new CMD
func newCMD(raw string) (*CMD, error) {
	var cmd CMD
	cmd.Status = 0
	goExe, err := exec.LookPath("go")
	if err != nil {
		log.Fatal(err)
		return &cmd, err
	}
	cmd.Raw = raw
	cmd.Path = goExe
	cmd.loadArgs()
	err = cmd.buildCmd()
	if err != nil {
		return &cmd, err
	}
	cmd.Status = 1
	return &cmd, nil
}

// loadArgs from cmd.Raw
func (cm *CMD) loadArgs() {
	// cm.Raw = "lxc launch ubuntu:CONTAINER --config=user.user-data="$(cat cloud-init-config.yml)"
	// cm.Raw = "lxc launch images:ubuntu/xenial/amd64 demo"
	cm.Args = strings.Split(cm.Raw, " ")
}

// buildCmd a CMD
func (cm *CMD) buildCmd() error {
	cmd := exec.Command(cm.Args)
	cm.Cmd = cmd
	stderr, err := cm.Cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
		return err
	}
	cm.Error = stderr
	stdout, err := cm.Cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return err
	}
	cm.Output = stdout
	return nil
}

// Execute a CMD
func (cm *CMD) Execute() error {
	if err := cm.Cmd.Start(); err != nil {
		cm.Status = 0
		log.Fatal(err)
		return err
	}
	cm.Status = 2
	return nil
}

// Resolve a CMD
func (cm *CMD) Resolve() error {
	if err := cm.Cmd.Wait(); err != nil {
		cm.Status = 0
		log.Fatal(err)
		return err
	}
	cm.Status = 3
	return nil
}

// Shell wraps around a slice of sh CMD
type Shell struct {
	Name		string
	Type		string
	Env			string
	Commands	[]*CMD
	Outputs		[]os.Stdout
	Errors		[]os.Stdout
	Status		int // 0 - Error, 1 - Initialized, 2 - Executed, 3 - Finished
}

// NewShell initializes a new sh Shell
func NewShell(name string, sType string, env string, commands []string) (*Shell, error) {
	var sCMDs []*CMD
	for _, command := range commands {
		sCMD, err := newCMD(command)
		if err != nil {
			log.Fatal(err)
			return &Shell{Status: 0}, err
		}
		sCMDs = append(sCMDs, sCMD)
	}
	return &Shell{name, sType, env, sCMDs, []os.Stdout{}, []os.Stdout{}, 1}, nil
}

// Execute a Shell
func (sh *Shell) Execute() error {
	var err error
	for ind, cMD :=  range sh.Commands {
		err = cMD.Execute()
		if err != nil {
			sh.Status = 0
			log.Fatal(err)
			return err
		}
		sh.Commands[ind] = cMD
		sh.Outputs = append(sh.Outputs, cMD.Output)
		sh.Errors = append(sh.Errors, cMD.Error)
	}
	sh.Status = 2
	return nil
}

// Resolve a Shell
func (sh *Shell) Resolve() error {
	var err error
	for ind, cMD :=  range sh.Commands {
		err = cMD.Resolve()
		if err != nil {
			sh.Status = 0
			log.Fatal(err)
			return err
		}
		sh.Commands[ind] = cMD
	}
	sh.Status = 3
	return nil
}

// Error checks a resolved Shell for an errored CMD
func (sh *Shell) Error() error {
	for _, err := range sh.Errors {
		if err != nil {
			sh.Status = 0
			log.Fatal(err)
			return err
		}
	}
	return nil
}