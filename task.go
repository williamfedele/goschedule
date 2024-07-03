package main

type Task struct {
	ID       int
	Priority int
	Execute  func() error
}
