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

type receiveNodeType struct {
	parent       *receiveNodeType
	childs       []*receiveNodeType
	handler      ReceiveHandler
	timeoutAfter time.Duration
	timerPid     ProcessID
}

// Process xxx
type process struct {
	pid             ProcessID
	receiveNode     *receiveNodeType
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
		receiveNode:     &receiveNodeType{},
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

		var nextHandler = getNextHandler(proc.receiveNode)

		// check for next receive timeout
		if nextHandler != nil {
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

				var node = &receiveNodeType{
					parent:       proc.receiveNode,
					handler:      receive,
					timeoutAfter: -1,
				}

				if len(after) > 0 {
					node.timeoutAfter = after[0]
				}

				proc.receiveNode.childs = append(proc.receiveNode.childs, node)

			}, proc.args)

			proc.handler(context)

			continue

		case StopProcessMessage:

			proc.invokeReceive(nextHandler, sender, ProcessStoppedMessage{
				Reason: m.Reason,
			})

			proc.stopProcess(sender, m.Reason)

			return

		case timeoutMessage:
			// check that message fired for the right current receive handler
			if nextHandler == nil {
				continue

			} else {
				if nextHandler.timerPid != sender {
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

		if nextHandler == nil {
			proc.stopProcess(sender, "normal")

			return
		}

		var panicValue = proc.invokeReceive(nextHandler, sender, msg)
		if panicValue != nil {
			proc.stopProcess(sender, panicValue)

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

func (proc *process) invokeReceive(node *receiveNodeType, sender ProcessID, message interface{}) (panicValue interface{}) {

	proc.receiveNode = node

	var handler = node.handler

	defer func() {
		proc.receiveNode = getNextNode(node)

		if r := recover(); r != nil {
			panicValue = r
		}
	}()

	handler(newReceiveContext(
		sender,
		message,
	))

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

func getNextNode(node *receiveNodeType) *receiveNodeType {

	if len(node.childs) == 0 {

		if node.parent != nil {
			node.parent.childs = node.parent.childs[1:]

			return getNextNode(node.parent)
		}
	}

	return node
}

func getNextHandler(curNode *receiveNodeType) *receiveNodeType {

	if len(curNode.childs) > 0 {
		return curNode.childs[0]
	}

	if curNode.parent == nil {
		return nil
	}

	return getNextHandler(curNode.parent)
}
