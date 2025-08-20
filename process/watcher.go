package process

import (
	"log"
	"os"
	"sync"
)

// ProcState is a wrapper with the process state and an error in case there's any.
type ProcState struct {
	state *os.ProcessState
	err   error
}

// ProcWatcher is a wrapper that act as a object that watches a process.
type ProcWatcher struct {
	procStatus  chan *ProcState
	proc        ProcContainer
	stopWatcher chan bool
}

// Watcher is responsible for watching a list of processes and report to Master in
// case the process dies at some point.
type Watcher struct {
	sync.Mutex

	restartProc chan ProcContainer
	watchProcs  map[string]*ProcWatcher
}

// NewWatcher will create a Watcher instance.
// Returns a Watcher instance.
func NewWatcher() *Watcher {
	watcher := &Watcher{
		restartProc: make(chan ProcContainer),
		watchProcs:  make(map[string]*ProcWatcher),
	}
	return watcher
}

// RestartProc is a wrapper to export the channel restartProc. It basically keeps track of
// all the processes that died and need to be restarted.
// Returns a channel with the dead processes that need to be restarted.
func (watcher *Watcher) RestartProc() chan ProcContainer {
	return watcher.restartProc
}

// AddProcWatcher will add a watcher on proc.
func (watcher *Watcher) AddProcWatcher(proc ProcContainer) {
	watcher.Lock()
	defer watcher.Unlock()
	if _, ok := watcher.watchProcs[proc.Identifier()]; ok {
		log.Printf("A watcher for this process already exists.")
		return
	}
	procWatcher := &ProcWatcher{
		procStatus:  make(chan *ProcState, 1),
		proc:        proc,
		stopWatcher: make(chan bool, 1),
	}
	watcher.watchProcs[proc.Identifier()] = procWatcher
	go func() {
		log.Printf("Starting watcher on proc %s", proc.Identifier())
		state, err := proc.Watch()
		procWatcher.procStatus <- &ProcState{
			state: state,
			err:   err,
		}
	}()
	go func() {
		defer delete(watcher.watchProcs, procWatcher.proc.Identifier())
		select {
		case procStatus := <-procWatcher.procStatus:
			log.Printf("Proc %s is dead, advising master...", procWatcher.proc.Identifier())
			log.Printf("State is %s", procStatus.state.String())
			watcher.restartProc <- procWatcher.proc
			break
		case <-procWatcher.stopWatcher:
			break
		}
	}()
}

// StopWatcher will stop a running watcher on a process with identifier 'identifier'
// Returns a channel that will be populated when the watcher is finally done.
func (watcher *Watcher) StopWatcher(identifier string) chan bool {
	if watcher, ok := watcher.watchProcs[identifier]; ok {
		log.Printf("Stopping watcher on proc %s", identifier)
		watcher.stopWatcher <- true
		waitStop := make(chan bool, 1)
		go func() {
			<-watcher.procStatus
			waitStop <- true
		}()
		return waitStop
	}

	return nil
}
