package proc

import "time"

type messageWithSender struct {
	sender  ProcessID
	message interface{}
}

type startProcessMessage struct{}

// StopProcessMessage xxxx
type StopProcessMessage struct {
	Reason interface{}
}

// ProcessStoppedMessage xxx
type ProcessStoppedMessage struct {
	Reason interface{}
}

type timeoutMessage struct {
	After time.Duration
}

// TimeoutMessage xxxxx
type TimeoutMessage struct {
	After time.Duration
}

// FollowMessage xxxx
type FollowMessage struct{}

// UnfollowMessage xxxx
type UnfollowMessage struct{}

// FollowerStoppedMessage xxx
type FollowerStoppedMessage struct {
	Reason interface{}
}

func isInternalMessage(message interface{}) bool {

	switch msg := message.(type) {

	case messageWithSender:
		return isInternalMessage(msg.message)

	case startProcessMessage,
		StopProcessMessage,
		FollowMessage,
		UnfollowMessage,
		FollowerStoppedMessage:

		return true
	}

	return false
}
