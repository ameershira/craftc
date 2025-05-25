package main

// Cmd is an interface that defines a command
type Cmd interface {
	// run runs the command and returns whether it succesfully executed its operation,
	// and whether it built an artifact or not.
	run() (built bool, err error)
}
