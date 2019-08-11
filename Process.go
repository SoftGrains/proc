package proc

import (
	"sync/atomic"
)

const (
	idle int32 = iota
	running
	terminated
)

type startProcessMessage struct{}
type stopMessage struct{}

type ProcessStartedMessage struct{}
type ProcessStoppedMessage struct{}

// Process xxx
type process struct {
	pid             ProcessId
	internalMailbox *Mailbox
	mailbox         *Mailbox
	actor           Actor
	receive         func(Context)
	args            []interface{}
	processStatus   int32
}

// NewProcess xxxx
func newProcess(actor Actor, args []interface{}) *process {
	return &process{
		internalMailbox: NewMailbox(),
		mailbox:         NewMailbox(),
		actor:           actor,
		args:            args,
		processStatus:   idle,
	}
}

func (proc *process) ID() ProcessId {

	if proc.pid == nil {
		proc.pid = &localProcessId{
			process: proc,
		}
	}

	return proc.pid
}

func (proc *process) start() {
	proc.Send(startProcessMessage{}, true)
}

// Send xxxx
func (proc *process) Send(message interface{}, internal bool) {

	if atomic.LoadInt32(&proc.processStatus) == terminated {
		return
	}

	// add message to mailbox
	if internal {
		proc.internalMailbox.Enqueue(message)

	} else {
		proc.mailbox.Enqueue(message)
	}

	if ok := atomic.CompareAndSwapInt32(&proc.processStatus, idle, running); ok == false {
		return
	}

	go proc.processMessages()
}

// processMessages xxxx
func (proc *process) processMessages() {

	// push the messages into the actor
	for {

		var msg, ok = proc.internalMailbox.Dequeue()
		if ok == false {
			msg, ok = proc.mailbox.Dequeue()
		}

		if ok == false {
			break
		}

		switch msg.(type) {

		case startProcessMessage:
			proc.receive = proc.actor(proc.args...)

			if proc.receive == nil {
				atomic.StoreInt32(&proc.processStatus, terminated)

				return
			}

			proc.internalMailbox.Enqueue(ProcessStartedMessage{})

			continue

		case stopMessage:
			atomic.StoreInt32(&proc.processStatus, terminated)

			proc.internalMailbox = NewMailbox()
			proc.mailbox = NewMailbox()

			proc.internalMailbox.Enqueue(ProcessStoppedMessage{})

			continue
		}

		var context = newContext(
			proc.pid,
			msg)

		proc.receive(context)
	}

	atomic.StoreInt32(&proc.processStatus, idle)
}
