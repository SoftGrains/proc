package proc

import "time"

type timeoutAfter struct {
	After time.Duration
}

type timeoutAfterReply struct {
	After time.Duration
}

func timerProcess(self ProcessID, receive ReceiveDispatcher, args ...interface{}) {

	receive(func(sender ProcessID, message interface{}) {

		if sender == nil {
			return
		}

		switch msg := message.(type) {

		case timeoutAfter:
			time.Sleep(msg.After)
			sender.SendFrom(self, timeoutAfterReply{msg.After})
		}
	})
}
