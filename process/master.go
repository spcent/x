package process

/*
Master is the main package that keeps everything running as it should be.
It's responsible for starting, stopping and deleting processes.
It also will keep an eye on the Watcher in case a process dies so it can restart it again.
RemoteMaster is responsible for exporting the main APM operations as HTTP requests.

If you want to start a Remote Server, run:
- remoteServer, err := master.StartRemoteMasterServer(dsn, configFile)

It will start a remote master and return the instance.
To make remote requests, use the Remote Client by instantiating using:
- remoteClient, err := master.StartRemoteClient(dsn, timeout)

It will start the remote client and return the instance so you can use to initiate requests, such as:
- remoteClient.StartGoBin(sourcePath, name, keepAlive, args)
*/

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spcent/x/lock"
)

// Master is the main module that keeps everything in place and execute
// the necessary actions to keep the process running as they should be.
type Master struct {
	sync.Mutex

	SysFolder string   // SysFolder is the main APM folder where the necessary config files will be stored.
	PidFile   string   // PidFille is the APM pid file path.
	OutFile   string   // OutFile is the APM output log file path.
	ErrFile   string   // ErrFile is the APM err log file path.
	Watcher   *Watcher // Watcher is a watcher instance.

	Procs map[string]ProcContainer // Procs is a map containing all procs started on APM.
}

// DecodableMaster is a struct that the config toml file will decode to.
// It is needed because toml decoder doesn't decode to interfaces, so the
// Procs map can't be decoded as long as we use the ProcContainer interface
type DecodableMaster struct {
	SysFolder string
	PidFile   string
	OutFile   string
	ErrFile   string

	Watcher *Watcher

	Procs map[string]*Proc
}

// SafeReadTomlFile will try to acquire a lock on the file and then read its content afterwards.
// Returns an error in case there's any.
func SafeReadTomlFile(filename string, v any) error {
	fileLock := lock.MakeFileMutex(filename)
	ctx := context.Background()
	_, err := fileLock.Lock(ctx)
	defer fileLock.Unlock(ctx)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := toml.NewDecoder(f)
	return decoder.Decode(&v)
}

// SafeWriteTomlFile will try to acquire a lock on the file and then write to it.
// Returns an error in case there's any.
func SafeWriteTomlFile(v any, filename string) error {
	fileLock := lock.MakeFileMutex(filename)
	ctx := context.Background()
	_, err := fileLock.Lock(ctx)
	defer fileLock.Unlock(ctx)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(v)
}

// NewMaster will start a master instance with configFile.
// It returns a Master instance.
func NewMaster(configFile string) *Master {
	watcher := NewWatcher()
	decodableMaster := &DecodableMaster{}
	decodableMaster.Procs = make(map[string]*Proc)

	err := SafeReadTomlFile(configFile, decodableMaster)
	if err != nil {
		panic(err)
	}

	procs := make(map[string]ProcContainer)
	for k, v := range decodableMaster.Procs {
		procs[k] = v
	}
	// We need this hack because toml decoder doesn't decode to interfaces
	master := &Master{
		SysFolder: decodableMaster.SysFolder,
		PidFile:   decodableMaster.PidFile,
		OutFile:   decodableMaster.OutFile,
		ErrFile:   decodableMaster.ErrFile,
		Watcher:   decodableMaster.Watcher,
		Procs:     procs,
	}

	if master.SysFolder == "" {
		os.MkdirAll(path.Dir(configFile), 0777)
		master.SysFolder = path.Dir(configFile) + "/"
	}
	master.Watcher = watcher
	master.Revive()
	log.Printf("All procs revived...")
	go master.WatchProcs()
	go master.SaveProcsLoop()
	go master.UpdateStatus()
	return master
}

// WatchProcs will keep the procs running forever.
func (master *Master) WatchProcs() {
	for proc := range master.Watcher.RestartProc() {
		if !proc.ShouldKeepAlive() {
			master.Lock()
			master.updateStatus(proc)
			master.Unlock()
			log.Printf("Proc %s does not have keep alive set. Will not be restarted.", proc.Identifier())
			continue
		}
		log.Printf("Restarting proc %s.", proc.Identifier())
		if proc.IsAlive() {
			log.Printf("Proc %s was supposed to be dead, but it is alive.", proc.Identifier())
		}
		master.Lock()
		proc.AddRestart()
		err := master.restart(proc)
		master.Unlock()
		if err != nil {
			log.Printf("Could not restart process %s due to %s.", proc.Identifier(), err)
		}
	}
}

// Prepare will compile the source code into a binary and return a preparable
// ready to be executed.
func (master *Master) Prepare(sourcePath string, name string, language string, keepAlive bool, args []string) (ProcPreparable, []byte, error) {
	procPreparable := &Preparable{
		Name:       name,
		SourcePath: sourcePath,
		SysFolder:  master.SysFolder,
		Language:   language,
		KeepAlive:  keepAlive,
		Args:       args,
	}
	output, err := procPreparable.PrepareBin()
	return procPreparable, output, err
}

// RunPreparable will run procPreparable and add it to the watch list in case everything goes well.
func (master *Master) RunPreparable(procPreparable ProcPreparable) error {
	master.Lock()
	defer master.Unlock()

	if _, ok := master.Procs[procPreparable.Identifier()]; ok {
		log.Printf("Proc %s already exist.", procPreparable.Identifier())
		return errors.New("trying to start a process that already exist")
	}

	proc, err := procPreparable.Start()
	if err != nil {
		return err
	}

	master.Procs[proc.Identifier()] = proc
	master.saveProcsWrapper()
	master.Watcher.AddProcWatcher(proc)
	proc.SetStatus("running")
	return nil
}

// ListProcs will return a list of all procs.
func (master *Master) ListProcs() []ProcContainer {
	procsList := []ProcContainer{}
	for _, v := range master.Procs {
		procsList = append(procsList, v)
	}
	return procsList
}

// RestartProcess will restart a process.
func (master *Master) RestartProcess(name string) error {
	err := master.StopProcess(name)
	if err != nil {
		return err
	}
	return master.StartProcess(name)
}

// StartProcess will a start a process.
func (master *Master) StartProcess(name string) error {
	master.Lock()
	defer master.Unlock()
	if proc, ok := master.Procs[name]; ok {
		return master.start(proc)
	}
	return errors.New("unknown process")
}

// StopProcess will stop a process with the given name.
func (master *Master) StopProcess(name string) error {
	master.Lock()
	defer master.Unlock()
	if proc, ok := master.Procs[name]; ok {
		return master.stop(proc)
	}

	return errors.New("unknown process")
}

// DeleteProcess will delete a process and all its files and childs forever.
func (master *Master) DeleteProcess(name string) error {
	master.Lock()
	defer master.Unlock()

	log.Printf("Trying to delete proc %s", name)
	if proc, ok := master.Procs[name]; ok {
		err := master.stop(proc)
		if err != nil {
			return err
		}
		delete(master.Procs, name)
		err = master.delete(proc)
		if err != nil {
			return err
		}
		log.Printf("Successfully deleted proc %s", name)
	}
	return nil
}

// Revive will revive all procs listed on ListProcs. This should ONLY be called
// during Master startup.
func (master *Master) Revive() error {
	master.Lock()
	defer master.Unlock()

	procs := master.ListProcs()
	log.Printf("Reviving all processes")
	for id := range procs {
		proc := procs[id]
		if !proc.ShouldKeepAlive() {
			log.Printf("Proc %s does not have KeepAlive set. Will not revive it.", proc.Identifier())
			continue
		}
		log.Printf("Reviving proc %s", proc.Identifier())
		err := master.start(proc)
		if err != nil {
			return fmt.Errorf("failed to revive proc %s due to %s", proc.Identifier(), err)
		}
	}

	return nil
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *Master) start(proc ProcContainer) error {
	if !proc.IsAlive() {
		err := proc.Start()
		if err != nil {
			return err
		}

		master.Watcher.AddProcWatcher(proc)
		proc.SetStatus("running")
	}
	return nil
}

func (master *Master) delete(proc ProcContainer) error {
	return proc.Delete()
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *Master) stop(proc ProcContainer) error {
	if proc.IsAlive() {
		waitStop := master.Watcher.StopWatcher(proc.Identifier())
		err := proc.GracefullyStop()
		if err != nil {
			return err
		}
		if waitStop != nil {
			<-waitStop
			proc.NotifyStopped()
			proc.SetStatus("stopped")
		}
		log.Printf("Proc %s successfully stopped.", proc.Identifier())
	}

	return nil
}

// UpdateStatus will update a process status every 30s.
func (master *Master) UpdateStatus() {
	for {
		master.Lock()
		procs := master.ListProcs()
		for id := range procs {
			proc := procs[id]
			master.updateStatus(proc)
		}
		master.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (master *Master) updateStatus(proc ProcContainer) {
	if proc.IsAlive() {
		proc.SetStatus("running")
	} else {
		proc.NotifyStopped()
		proc.SetStatus("stopped")
	}
}

// NOT thread safe method. Lock should be acquire before calling it.
func (master *Master) restart(proc ProcContainer) error {
	err := master.stop(proc)
	if err != nil {
		return err
	}

	return master.start(proc)
}

// SaveProcsLoop will loop forever to save the list of procs onto the proc file.
func (master *Master) SaveProcsLoop() {
	for {
		log.Printf("Saving list of procs.")
		master.Lock()
		master.saveProcsWrapper()
		master.Unlock()
		time.Sleep(5 * time.Minute)
	}
}

// Stop will stop APM and all of its running procs.
func (master *Master) Stop() error {
	log.Printf("Stopping APM...")
	procs := master.ListProcs()
	for id := range procs {
		proc := procs[id]
		log.Printf("Stopping proc %s", proc.Identifier())
		master.stop(proc)
	}
	log.Printf("Saving and returning list of procs.")
	return master.saveProcsWrapper()
}

// SaveProcs will save a list of procs onto a file inside configPath.
// Returns an error in case there's any.
func (master *Master) SaveProcs() error {
	master.Lock()
	defer master.Unlock()

	return master.saveProcsWrapper()
}

// NOT Thread Safe. Lock should be acquired before calling it.
func (master *Master) saveProcsWrapper() error {
	configPath := master.getConfigPath()
	return SafeWriteTomlFile(master, configPath)
}

func (master *Master) getConfigPath() string {
	return path.Join(master.SysFolder, "config.toml")
}
