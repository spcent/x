package lock

import (
	"context"
	"os"
	"sync"
	"syscall"
)

// FileMutex is a wrapper used to create lock on files.
type FileMutex struct {
	mutex *sync.Mutex
	file  *os.File
}

// MakeFileMutex will create a FileMutex intance.
// Returns a FileMutex instance.
func MakeFileMutex(filename string) *FileMutex {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return &FileMutex{file: nil}
	}
	mutex := &sync.Mutex{}
	return &FileMutex{file: file, mutex: mutex}
}

// Lock will try to acquire a lock on the file.
func (fMutex *FileMutex) Lock(ctx context.Context) (bool, error) {
	fMutex.mutex.Lock()
	if fMutex.file != nil {
		if err := syscall.Flock(int(fMutex.file.Fd()), syscall.LOCK_EX); err != nil {
			return false, err
		}
	}

	return true, nil
}

// Unlock will try to release a lock on a file.
func (fMutex *FileMutex) Unlock(ctx context.Context) (bool, error) {
	fMutex.mutex.Unlock()
	if fMutex.file != nil {
		if err := syscall.Flock(int(fMutex.file.Fd()), syscall.LOCK_UN); err != nil {
			return false, err
		}
	}

	return true, nil
}
