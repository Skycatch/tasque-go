package main

// ExecutableInterface hello world
type ExecutableInterface interface {
	Execute(handler MessageHandler)
	Result() Result
}
