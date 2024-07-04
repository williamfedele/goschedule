package goschedule

import (
	"container/heap"
	"log"
	"sync"
)

type Scheduler struct {
	tasks     PriorityQueue
	nextId    int
	completed map[int]struct{}
	mu        sync.Mutex
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		tasks:     make(PriorityQueue, 0),
		nextId:    0,
		completed: make(map[int]struct{}),
	}
	heap.Init(&s.tasks)
	return s
}

func (s *Scheduler) getNextId() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextId++
	return s.nextId
}

func (s *Scheduler) AddTask(t *Task) {
	heap.Push(&s.tasks, t)
}

func (s *Scheduler) CreateTask(priority int, execute func() error) *Task {
	return &Task{
		ID:          s.getNextId(),
		Priority:    priority,
		ExecuteFunc: execute,
	}
}

func (s *Scheduler) RunNext() {
	if s.tasks.Len() > 0 {
		task := heap.Pop(&s.tasks).(*Task)

		for _, dep := range task.dependencies {
			log.Printf("Executing dependency %d for task %d", dep.ID, task.ID)
			if err := dep.Execute(); err != nil {
				log.Printf("Error executing dependency %d for task %d: %v", dep.ID, task.ID, err)
				log.Printf("Task %d will not be executed", task.ID)
				// TODO add maxretries to task and retry if it fails
				// currently will not add the task back to the queue
				return
			} else {
				s.mu.Lock()
				s.completed[dep.ID] = struct{}{}
				s.mu.Unlock()
			}
		}

		err := task.Execute()

		if err != nil {
			log.Printf("Error executing task %d: %v\n", task.ID, err)
		} else {
			s.mu.Lock()
			s.completed[task.ID] = struct{}{}
			s.mu.Unlock()
		}
	}

}

func (s *Scheduler) Run() {
	for s.tasks.Len() > 0 {
		s.RunNext()
	}
}

func (s *Scheduler) AddDependency(t *Task, d *Task) {
	if t.ID == d.ID {
		log.Printf("Task cannot depend on itself. Task ID: %d", t.ID)
		return
	}
	t.dependencies = append(t.dependencies, d)
}
