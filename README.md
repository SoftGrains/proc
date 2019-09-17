# proc
golang actor process library

	var pid = proc.Spawn(MyActor1)
	pid.Send("hello pid")

	pid = proc.Spawn(MyActor2, 1, 2, 3)
	pid.Send("hello pid")
	pid.Send(0xff)
	pid.Send(PrintState{})

TODO:
 supervisor