package proc

import "sync/atomic"

const (
	idle int32 = iota
	running
	terminated
)

type startProcessMessage struct{}
type stopProcessMessage struct{}

// ProcessStartedMessage xxx
type ProcessStartedMessage struct{}

// ProcessStoppedMessage xxx
type ProcessStoppedMessage struct{}

// Process xxx
type process struct {
	pid             ProcessID
	receive         func(Context)
	actor           Actor
	args            []interface{}
	internalMailbox *Mailbox
	mailbox         *Mailbox
	messagesCount   int32
	processStatus   int32
}

// NewProcess xxxx
func newProcess(actor Actor, args []interface{}) *process {
	return &process{
		actor:           actor,
		args:            args,
		internalMailbox: NewMailbox(),
		mailbox:         NewMailbox(),
		processStatus:   idle,
	}
}

func (proc *process) ID() ProcessID {

	if proc.pid == nil {
		proc.pid = &localProcessID{
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
	atomic.AddInt32(&proc.messagesCount, 1)

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
func (proc *process) pullMessage() (interface{}, bool) {

	var msg, ok = proc.internalMailbox.Dequeue()

	if ok == false {
		msg, ok = proc.mailbox.Dequeue()
	}

	if ok {
		atomic.AddInt32(&proc.messagesCount, -1)
	}

	return msg, ok
}

// processMessages xxxx
func (proc *process) processMessages() {
processMessagesLabel:

	for {

		var msg, ok = proc.pullMessage()
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

			msg = ProcessStartedMessage{}

		case stopProcessMessage:
			atomic.StoreInt32(&proc.processStatus, terminated)

			proc.receive(newContext(
				proc.pid,
				ProcessStoppedMessage{}))

			return
		}

		proc.receive(newContext(
			proc.pid,
			msg))
	}

	if ok := atomic.CompareAndSwapInt32(&proc.processStatus, running, idle); ok == false {
		return
	}

	// check if there are still messages to process
	if atomic.LoadInt32(&proc.messagesCount) > 0 &&
		atomic.CompareAndSwapInt32(&proc.processStatus, idle, running) {

		goto processMessagesLabel
	}
}
