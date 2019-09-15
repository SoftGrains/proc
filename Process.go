package proc

import "sync/atomic"

const (
	idle int32 = iota
	running
	terminated
)

type startProcessMessage struct{}

// StopProcessMessage xxxx
type StopProcessMessage struct{}

// ProcessStoppedMessage xxx
type ProcessStoppedMessage struct{}

// Process xxx
type process struct {
	pid             ProcessID
	receiveHandler  func(ProcessID, interface{})
	handler         ProcessHandler
	args            []interface{}
	internalMailbox *Mailbox
	mailbox         *Mailbox
	messagesCount   int32
	processStatus   int32
}

// NewProcess xxxx
func newProcess(handler ProcessHandler, args []interface{}) *process {
	return &process{
		handler:         handler,
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

		var sender ProcessID
		if m, ok := msg.(messageWithSender); ok {
			sender = m.sender
			msg = m.message
		}

		switch msg.(type) {

		case startProcessMessage:
			proc.receiveHandler = proc.handler(proc.pid, proc.args...)

			if proc.receiveHandler == nil {
				atomic.StoreInt32(&proc.processStatus, terminated)

				return
			}

			continue

		case StopProcessMessage:
			atomic.StoreInt32(&proc.processStatus, terminated)

			proc.receiveHandler(sender, ProcessStoppedMessage{})

			return
		}

		proc.receiveHandler(sender, msg)
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
