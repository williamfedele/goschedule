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

func (s *Scheduler) executeTask(t *Task) error {
	s.logger.Info("Executing task", "ID", t.ID, "priority", t.Priority)

	if err := t.Execute(); err != nil {
		s.logger.Error("Task failed", "ID", t.ID, "Error", err)
		return err
	}

	s.logger.Info("Task completed", "ID", t.ID)
	s.mu.Lock()
	s.completed[t.ID] = struct{}{}
	s.mu.Unlock()
	return nil
}

func (s *Scheduler) RunNext() {
	if s.tasks.Len() > 0 {
		task := heap.Pop(&s.tasks).(*Task)

		s.logger.Info("Next scheduled task", "ID", task.ID, "priority", task.Priority)
		if task.status == Failed {
			s.logger.Error("Scheduled task has failed. Will not be executed", "ID", task.ID)
			return
		}
		if task.status == Completed {
			s.logger.Warn("Scheduled task has already been executed", "ID", task.ID)
			return
		}

		for _, dep := range task.dependencies {
			if dep.status == Failed {
				s.logger.Error("Dependency is in failed state. Scheduled task will not be executed", "ID", task.ID)
				return
			}
			if dep.status == Completed {
				s.logger.Warn("Dependency already executed", "ID", dep.ID)
				continue
			}

			if err := s.executeTask(dep); err != nil {
				s.logger.Error("Scheduled task will not be executed due to failed dependencies", "ID", task.ID)
				return
			}
		}
		s.executeTask(task)
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
