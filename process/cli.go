package process

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/spcent/x/logging"
)

// Cli is the command line client.
type Cli struct {
	remoteClient *RemoteClient  // remote client instance
	log          logging.Logger // logger instance
}

// NewCli initiates a remote client connecting to dsn.
// Returns a Cli instance.
func NewCli(dsn string, timeout time.Duration, l logging.Logger) (*Cli, error) {
	if l == nil {
		l = logging.NewLogAdapter(os.Stdout, logging.InfoLevel)
	}

	client, err := StartRemoteClient(dsn, timeout)
	if err != nil {
		l.Errorf("Failed to start remote client due to: %+v, dsn: %s", err, dsn)
		return nil, err
	}

	return &Cli{
		remoteClient: client,
		log:          l,
	}, nil
}

// Save will save all previously saved processes onto a list.
// Display an error in case there's any.
func (cli *Cli) Save() error {
	err := cli.remoteClient.Save()
	if err != nil {
		cli.log.Errorf("Failed to save list of processes due to: %+v", err)
		return err
	}
	return nil
}

// Resurrect will restore all previously save processes.
// Display an error in case there's any.
func (cli *Cli) Resurrect() error {
	err := cli.remoteClient.Resurrect()
	if err != nil {
		cli.log.Errorf("Failed to resurrect all previously save processes due to: %+v", err)
		return err
	}
	return nil
}

// StartGoBin will try to start a go binary process.
// Returns a fatal error in case there's any.
func (cli *Cli) StartGoBin(sourcePath string, name string, keepAlive bool, args []string) error {
	err := cli.remoteClient.StartGoBin(sourcePath, name, keepAlive, args)
	if err != nil {
		cli.log.Errorf("Failed to start go bin due to: %+v", err)
		return err
	}
	return nil
}

// RestartProcess will try to restart a process with procName. Note that this process
// must have been already started through StartGoBin.
func (cli *Cli) RestartProcess(procName string) error {
	err := cli.remoteClient.RestartProcess(procName)
	if err != nil {
		cli.log.Errorf("Failed to restart process due to: %+v, name: %s", err, procName)
		return err
	}
	return nil
}

// StartProcess will try to start a process with procName. Note that this process
// must have been already started through StartGoBin.
func (cli *Cli) StartProcess(procName string) error {
	err := cli.remoteClient.StartProcess(procName)
	if err != nil {
		cli.log.Errorf("Failed to start process due to: %+v, name: %s", err, procName)
		return err
	}
	return nil
}

// StopProcess will try to stop a process named procName.
func (cli *Cli) StopProcess(procName string) error {
	err := cli.remoteClient.StopProcess(procName)
	if err != nil {
		cli.log.Errorf("Failed to stop process due to: %+v, name: %s", err, procName)
		return err
	}
	return nil
}

// DeleteProcess will stop and delete all dependencies from process procName forever.
func (cli *Cli) DeleteProcess(procName string) error {
	err := cli.remoteClient.DeleteProcess(procName)
	if err != nil {
		cli.log.Errorf("Failed to delete process due to: %+v, name: %s", err, procName)
		return err
	}
	return nil
}

// Status will display the status of all procs started through StartGoBin.
func (cli *Cli) Status() error {
	procResponse, err := cli.remoteClient.MonitStatus()
	if err != nil {
		cli.log.Errorf("Failed to get status due to: %+v", err)
		return err
	}

	maxName := 0
	for id := range procResponse.Procs {
		proc := procResponse.Procs[id]
		maxName = int(math.Max(float64(maxName), float64(len(proc.Name))))
	}
	totalSize := maxName + 51
	topBar := ""
	for i := 1; i <= totalSize; i += 1 {
		topBar += "-"
	}
	infoBar := fmt.Sprintf("|%s|%s|%s|%s|",
		PadString("pid", 13),
		PadString("name", maxName+2),
		PadString("status", 16),
		PadString("keep-alive", 15))
	fmt.Println(topBar)
	fmt.Println(infoBar)
	for id := range procResponse.Procs {
		proc := procResponse.Procs[id]
		kp := "True"
		if !proc.KeepAlive {
			kp = "False"
		}
		fmt.Printf("|%s|%s|%s|%s|\n",
			PadString(fmt.Sprintf("%d", proc.Pid), 13),
			PadString(proc.Name, maxName+2),
			PadString(proc.Status.Status, 16),
			PadString(kp, 15))
	}
	fmt.Println(topBar)
	return nil
}

// PadString will add totalSize spaces evenly to the right and left side of str.
// Returns str after applying the pad.
func PadString(str string, totalSize int) string {
	turn := 0
	for {
		if len(str) >= totalSize {
			break
		}
		if turn == 0 {
			str = " " + str
			turn ^= 1
		} else {
			str = str + " "
			turn ^= 1
		}
	}
	return str
}
