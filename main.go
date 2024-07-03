package main

import (
	"container/heap"
	"fmt"
)

func main() {
	scheduler := NewScheduler()
	heap.Init(&scheduler.tasks)

	scheduler.AddTask(&Task{
		ID:       1,
		Priority: 3,
		Execute: func() error {
			fmt.Println("Executing task with priority 3")
			return nil
		}})
	scheduler.AddTask(&Task{
		ID:       2,
		Priority: 1,
		Execute: func() error {
			fmt.Println("Executing task with priority 1")
			return nil
		}})
	scheduler.AddTask(&Task{
		ID:       3,
		Priority: 2,
		Execute: func() error {
			fmt.Println("Executing task with priority 2")
			return nil
		}})

	scheduler.Run()
}
