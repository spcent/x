package helper

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"log"
)

// RunWithRetry
func RunWithRetry(runFunc func(idx int) (retry bool, err error), maxCnt int) error {
	var finalErr error
	for idxCnt := 0; idxCnt < maxCnt; idxCnt++ {
		retry, err := runFunc(idxCnt)
		if maxCnt == 1 {
			return err
		}

		finalErr = err
		if !retry {
			break
		}

		log.Printf("exec failed: %s, retry[%d]\n", err, idxCnt+1)
	}

	return finalErr
}

// GoRecover take a function as input, if the function panics, the panic will be recovered and the error will be returned
func GoRecover(f func(), name string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\x1b[31m", "panic occurred at: \n", name, "\npanic: \n", r, "\x1b[0m")
		}
	}()

	f()
}

// SafeGo go with recover
func SafeGo(fn func()) {
	go func() {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			var stack [4096]byte
			n := runtime.Stack(stack[:], false)
			fmt.Printf("panic err: %v\n\n%s\n", err, stack[:n])
		}()

		fn()
	}()
}

func WaitForSignal() {
	// Setting up signal capturing
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV,
		syscall.SIGINT, syscall.SIGQUIT)

	// Waiting for SIGINT (pkill -2)
	<-quit
}
