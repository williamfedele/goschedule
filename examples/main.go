package main

import (
	"fmt"

	"github.com/williamfedele/goschedule"
)

func task1() error {
	fmt.Println("This task should complete")
	return nil
}

func task2() error {
	fmt.Println("This task should complete")
	return nil
}

func taskFail() error {
	return fmt.Errorf("this task should fail")
}

func main() {
	scheduler := goschedule.NewScheduler()

	// 1. taskErrorDependency runs first with highest priority and fails
	// 2. taskDependent task has next priority and should not run since it has a failing dependency
	// 3. taskSuccess runs and completes
	taskSuccess := scheduler.CreateTask(1, task1)
	taskDependent := scheduler.CreateTask(2, task2)
	taskErrorDependency := scheduler.CreateTask(3, taskFail)

	scheduler.AddDependency(taskDependent, taskErrorDependency)

	scheduler.AddTask(taskSuccess)
	scheduler.AddTask(taskDependent)
	scheduler.AddTask(taskErrorDependency)
	scheduler.Run()
}
