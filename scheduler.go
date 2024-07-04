package goschedule

import (
	"container/heap"
	"log/slog"
	"sync"
)

type SchedulerOption func(*Scheduler)

func WithLogger(logger *slog.Logger) SchedulerOption {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

type Scheduler struct {
	tasks     PriorityQueue
	nextId    int
	completed map[int]struct{}
	mu        sync.Mutex
	logger    *slog.Logger
}

func NewScheduler(options ...SchedulerOption) *Scheduler {
	s := &Scheduler{
		tasks:     make(PriorityQueue, 0),
		nextId:    0,
		completed: make(map[int]struct{}),
		logger:    slog.Default(),
	}
	heap.Init(&s.tasks)

	for _, option := range options {
		option(s)
	}

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

		s.logger.Info("Executing task", "ID", task.ID, "priority", task.Priority)
		if len(task.dependencies) != 0 {
			s.logger.Info("Task has dependencies. Executing dependencies...")
		}

		for _, dep := range task.dependencies {
			// TODO if any dependencies fail, we should bail on the dependent task
			s.logger.Info("Executing dependency", "ID", dep.ID)
			if err := dep.Execute(); err != nil {
				s.logger.Error("Error executing dependency", "ID", dep.ID, "Error", err)
				s.logger.Error("Task will not be executed due to failed dependencies", "ID", task.ID)
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
			s.logger.Error("Error executing task", "ID", task.ID, "Error", err)
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

func (s *Scheduler) AddDependency(task *Task, dependency *Task) {
	if task.ID == dependency.ID {
		s.logger.Error("Task cannot depend on itself", "ID", task.ID)
		return
	}
	task.dependencies = append(task.dependencies, dependency)
}
