package schedule

import (
	"container/heap"
	"log"
	"sync"
)

type Scheduler struct {
	tasks  PriorityQueue
	nextId int
	mu     sync.Mutex
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		tasks:  make(PriorityQueue, 0),
		nextId: 0,
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

func (s *Scheduler) Run() {

	for s.tasks.Len() > 0 {
		task := heap.Pop(&s.tasks).(*Task)
		err := task.Execute()

		if err != nil {
			log.Printf("Error executing task %d: %v\n", task.ID, err)
		}
	}
}

func (s *Scheduler) AddDependency(t *Task, d *Task) {
	if t.ID == d.ID {
		log.Printf("Task cannot depend on itself. Task ID: %d", t.ID)
		return
	}
	t.dependencies = append(t.dependencies, d)
}
