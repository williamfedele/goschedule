package main

import (
	"container/heap"
	"fmt"
)

type Scheduler struct {
	tasks PriorityQueue
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(PriorityQueue, 0),
	}
}

func (s *Scheduler) AddTask(t *Task) {
	heap.Push(&s.tasks, t)
}

func (s *Scheduler) Run() {
	for s.tasks.Len() > 0 {
		task := heap.Pop(&s.tasks).(*Task)
		err := task.Execute()
		if err != nil {
			fmt.Println("Error executing task %d: %v\n", task.ID, err)
		}
	}
}
