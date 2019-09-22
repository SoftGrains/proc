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

type receiveQueueType struct {
	handler      ReceiveHandler
	timeoutAfter time.Duration
	timerPid     ProcessID
}

// Process xxx
type process struct {
	pid             ProcessID
	receiveQueue    []*receiveQueueType
	handler         ProcessHandler
	args            []interface{}
	internalMailbox *Mailbox
	mailbox         *Mailbox
	messagesCount   int32
	processStatus   int32
	followers       []ProcessID
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

	proc.Send(startProcessMessage{})

	return proc.pid
}

// Send xxxx
func (proc *process) Send(message interface{}) {

	if atomic.LoadInt32(&proc.processStatus) == terminated {
		return
	}

	// add message to mailbox
	atomic.AddInt32(&proc.messagesCount, 1)

	if isInternalMessage(message) {
		proc.internalMailbox.Enqueue(message)

	} else {
		proc.mailbox.Enqueue(message)
	}

	if ok := atomic.CompareAndSwapInt32(&proc.processStatus, idle, running); ok == false {
		return
	}

	go proc.processMessages()
}

func (proc *process) processMessages() {
processMessagesLabel:

	for {

		// check for next receive timeout
		if len(proc.receiveQueue) > 0 {
			var nextHandler = proc.receiveQueue[0]

			// start timer process
			if nextHandler.timeoutAfter >= 0 &&
				nextHandler.timerPid == nil {
				nextHandler.timerPid = Spawn(timerProcess,
					proc.pid,
					nextHandler.timeoutAfter)
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

			var context = newContext(proc.pid, func(receive ReceiveHandler, after ...time.Duration) {
				if receive == nil {
					return
				}

				var handler = &receiveQueueType{
					handler:      receive,
					timeoutAfter: -1,
				}

				if len(after) > 0 {
					handler.timeoutAfter = after[0]
				}

				proc.receiveQueue = append(proc.receiveQueue, handler)

			}, proc.args)

			proc.handler(context)

			if len(proc.receiveQueue) == 0 {
				proc.stopProcess(sender, "normal")

				return
			}

			continue

		case StopProcessMessage:
			proc.stopProcess(sender, m.Reason)

			return

		case timeoutMessage:
			// check that message fired for the right current receive handler
			if len(proc.receiveQueue) == 0 {
				continue

			} else {
				var handler = proc.receiveQueue[0]
				if handler.timerPid != sender {
					continue
				}
			}

			msg = TimeoutMessage{m.After}

		case FollowMessage:
			if sender != nil && sender != proc.pid {
				if proc.getFollowerIndex(sender) == -1 {
					proc.followers = append(proc.followers, sender)
				}
			}

			continue

		case UnfollowMessage:
			if sender != nil && sender != proc.pid {
				if idx := proc.getFollowerIndex(sender); idx > -1 {
					proc.followers = append(
						proc.followers[:idx],
						proc.followers[idx+1:]...)
				}
			}

			continue
		}

		var panicValue = proc.invokeReceive(sender, msg)

		if panicValue != nil {
			proc.stopProcess(sender, panicValue)

			return

		} else if len(proc.receiveQueue) == 0 {
			proc.stopProcess(sender, "normal")

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

func (proc *process) invokeReceive(sender ProcessID, message interface{}) (panicValue interface{}) {

	if len(proc.receiveQueue) == 0 {
		return
	}

	var handler = proc.receiveQueue[0].handler
	proc.receiveQueue = proc.receiveQueue[1:]

	defer func() {
		if r := recover(); r != nil {
			panicValue = r
		}
	}()

	handler(
		func() ProcessID { return sender },
		func() interface{} { return message })

	return
}

func (proc *process) stopProcess(sender ProcessID, reason interface{}) {

	atomic.StoreInt32(&proc.processStatus, terminated)

	// notify the follower processes
	for _, fpid := range proc.followers {
		fpid.SendFrom(proc.pid, ProcessStoppedMessage{
			Reason: reason,
		})
	}
}

func (proc *process) getFollowerIndex(follower ProcessID) int {

	for idx, pid := range proc.followers {
		if pid == follower {
			return idx
		}
	}

	return -1
}
