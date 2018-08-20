package main

import "github.com/skycatch/tasque-go/result"

// ExecutableInterface hello world
type ExecutableInterface interface {
	Execute(handler MessageHandler)
	Result() result.Result
}
