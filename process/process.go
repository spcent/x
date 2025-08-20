package process

import (
	"os/exec"
	"strings"
)

type ProcPreparable interface {
	PrepareBin() ([]byte, error)
	Start() (ProcContainer, error)
	Identifier() string
}

// ProcPreparable is a preparable with all the necessary informations to run
// a process. To actually run a process, call the Start() method.
type Preparable struct {
	Name       string   // The name of the process
	SourcePath string   // The path to the source code of the process
	Cmd        string   // The command to be executed
	SysFolder  string   // The folder where all the process files will be stored
	Language   string   // The language of the process, ie: go, python, nodejs, etc.
	KeepAlive  bool     // If true, the process will be kept alive after it exits
	Args       []string // The arguments to be passed to the process
	Env        []string // The environment variables to be passed to the process, ie: PORT=8080, DEBUG=true
}

// PrepareBin will compile the Golang project from SourcePath and populate Cmd with the proper
// command for the process to be executed.
// Returns the compile command output.
func (preparable *Preparable) PrepareBin() ([]byte, error) {
	// Remove the last character '/' if present
	if preparable.SourcePath[len(preparable.SourcePath)-1] == '/' {
		preparable.SourcePath = strings.TrimSuffix(preparable.SourcePath, "/")
	}
	cmd := ""
	cmdArgs := []string{}
	binPath := preparable.getBinPath()
	if preparable.Language == "go" {
		cmd = "go"
		cmdArgs = []string{"build", "-o", binPath, preparable.SourcePath + "/."}
	}

	preparable.Cmd = preparable.getBinPath()
	return exec.Command(cmd, cmdArgs...).Output()
}

// Start will execute the process based on the information presented on the preparable.
// This function should be called from inside the master to make sure
// all the watchers and process handling are done correctly.
// Returns a tuple with the process and an error in case there's any.
func (preparable *Preparable) Start() (ProcContainer, error) {
	proc := &Proc{
		Name:      preparable.Name,
		Cmd:       preparable.Cmd,
		Args:      preparable.Args,
		Env:       preparable.Env,
		Path:      preparable.getPath(),
		Pidfile:   preparable.getPidPath(),
		Outfile:   preparable.getOutPath(),
		Errfile:   preparable.getErrPath(),
		KeepAlive: preparable.KeepAlive,
		Status:    &ProcStatus{},
	}

	err := proc.Start()
	return proc, err
}

func (preparable *Preparable) Identifier() string {
	return preparable.Name
}

func (preparable *Preparable) getPath() string {
	if preparable.SysFolder[len(preparable.SysFolder)-1] == '/' {
		preparable.SysFolder = strings.TrimSuffix(preparable.SysFolder, "/")
	}
	return preparable.SysFolder + "/" + preparable.Name
}

func (preparable *Preparable) getBinPath() string {
	return preparable.getPath() + "/" + preparable.Name
}

func (preparable *Preparable) getPidPath() string {
	return preparable.getBinPath() + ".pid"
}

func (preparable *Preparable) getOutPath() string {
	return preparable.getBinPath() + ".out"
}

func (preparable *Preparable) getErrPath() string {
	return preparable.getBinPath() + ".err"
}
