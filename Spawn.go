package proc

import "time"

// Receive xxx
type Receive = func(ProcessID, interface{})

// ReceiveDispatcher xxxx
type ReceiveDispatcher = func(Receive, ...time.Duration)

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(ProcessID, ReceiveDispatcher, ...interface{})

// Spawn xxxx
func Spawn(handler ProcessHandler, args ...interface{}) ProcessID {

	return newProcess(handler, args).
		start()
}
