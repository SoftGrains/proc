package proc

import "time"

// PIDFunction xxx
type PIDFunction = func() ProcessID

// MessageFunction xxx
type MessageFunction = func() interface{}

// ReceiveHandler xxx
type ReceiveHandler = func(PIDFunction, MessageFunction)

// ReceiveFunction xxxx
type ReceiveFunction = func(ReceiveHandler, ...time.Duration)

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(*Context)

// Spawn xxxx
func Spawn(handler ProcessHandler, args ...interface{}) ProcessID {

	return newProcess(handler, args).
		start()
}
