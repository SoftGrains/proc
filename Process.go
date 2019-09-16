package proc

import (
	"sync/atomic"
	"time"
)

const (
	idle int32 = iota
	running
	terminated
)

type startProcessMessage struct{}

// StopProcessMessage xxxx
type StopProcessMessage struct {
	Reason interface{}
}

// ProcessStoppedMessage xxx
type ProcessStoppedMessage struct {
	Reason interface{}
}

// TimeoutAfter xxxxx
type TimeoutAfter struct {
	After time.Duration
}

type receiveHandler struct {
	handler      Receive
	timeoutAfter time.Duration
	timerPid     ProcessID
}

// Process xxx
type process struct {
	pid             ProcessID
	receiveHandlers []*receiveHandler
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

func (proc *process) start() ProcessID {
	if proc.pid == nil {
		proc.pid = &localProcessID{
			process: proc,
		}
	}

	proc.Send(startProcessMessage{}, true)

	return proc.pid
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
func (proc *process) processMessages() {
processMessagesLabel:

	for {

		// check for next receive timeout
		if len(proc.receiveHandlers) > 0 {
			var nextHandler = proc.receiveHandlers[0]

			// start timer process
			if nextHandler.timeoutAfter >= 0 &&
				nextHandler.timerPid == nil {
				nextHandler.timerPid = Spawn(timerProcess)

				nextHandler.timerPid.SendFrom(proc.pid, timeoutAfter{
					After: nextHandler.timeoutAfter,
				})

			}
		}

		// fetch the next message
		var msg, ok = proc.pullMessage()
		if ok == false {
			break
		}

		var sender ProcessID
		if m, ok := msg.(messageWithSender); ok {
			sender = m.sender
			msg = m.message
		}

		switch m := msg.(type) {

		case startProcessMessage:

			proc.handler(proc.pid, func(receive Receive, after ...time.Duration) {
				if receive == nil {
					return
				}

				var handler = &receiveHandler{
					handler:      receive,
					timeoutAfter: -1,
				}

				if len(after) > 0 {
					handler.timeoutAfter = after[0]
				}

				proc.receiveHandlers = append(proc.receiveHandlers, handler)

			}, proc.args...)

			if len(proc.receiveHandlers) == 0 {
				atomic.StoreInt32(&proc.processStatus, terminated)

				return
			}

			continue

		case StopProcessMessage:
			atomic.StoreInt32(&proc.processStatus, terminated)

			proc.invokeReceive(sender, ProcessStoppedMessage{
				Reason: m.Reason,
			})

			return

		case timeoutAfterReply:
			// check that message fired for the right current receive handler
			if len(proc.receiveHandlers) > 0 {
				var handler = proc.receiveHandlers[0]
				if handler.timerPid != sender {
					continue
				}

				msg = TimeoutAfter{m.After}
			}
		}

		proc.invokeReceive(sender, msg)

		if len(proc.receiveHandlers) == 0 {
			atomic.StoreInt32(&proc.processStatus, terminated)

			return
		}
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
func (proc *process) invokeReceive(sender ProcessID, message interface{}) {

	if len(proc.receiveHandlers) == 0 {
		return
	}

	var handler = proc.receiveHandlers[0].handler
	proc.receiveHandlers = proc.receiveHandlers[1:]

	handler(sender, message)
}
