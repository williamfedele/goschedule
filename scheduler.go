package goschedule

import (
	"container/heap"
	"fmt"
	"log/slog"
	"sync"
)

type SchedulerOption func(*Scheduler)

func WithLogger(logger *slog.Logger) SchedulerOption {
	return func(s *Scheduler) {
		if logger != nil {
			s.logger = logger
		}
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
		status:      NotStarted,
	}
}

func (s *Scheduler) RunNext() {
	if s.tasks.Len() > 0 {
		task := heap.Pop(&s.tasks).(*Task)

		s.logger.Info("Next task", "ID", task.ID, "priority", task.Priority)
		if task.status == Completed {
			s.logger.Info("Task already executed", "ID", task.ID)
			return
		}
		if len(task.dependencies) != 0 {
			s.logger.Info("Task has dependencies. Executing dependencies...")
		}

		for _, dep := range task.dependencies {
			if dep.status == Failed {
				s.logger.Error("Dependency is in failed state. Task will not be executed", "ID", task.ID)
				return
			}
			if dep.status == Completed {
				s.logger.Warn("Dependency already executed", "ID", dep.ID)
				continue
			}

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

		s.logger.Info("Executing task", "ID", task.ID, "priority", task.Priority)
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

func (s *Scheduler) AddDependency(task *Task, dependency *Task) error {
	if task.ID == dependency.ID {
		s.logger.Error("Task cannot depend on itself", "ID", task.ID)
		return fmt.Errorf("Task cannot depend on itself")
	}
	task.dependencies = append(task.dependencies, dependency)
	return nil
}
