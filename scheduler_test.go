package schedule

import (
	"fmt"
	"math/rand"
	"testing"
)

func createTestTask(id int, priority int) *Task {
	return &Task{
		ID:       id,
		Priority: priority,
		ExecuteFunc: func() error {
			fmt.Printf("Task %d executed\n", id)
			return nil
		},
	}
}

func generateTasks(count int) []*Task {
	tasks := make([]*Task, count)
	for i := 0; i < count; i++ {
		priority := rand.Intn(10)
		tasks[i] = createTestTask(i+1, priority)
	}
	return tasks
}

func createDependentTask(id int, priority int, dependencies []*Task) *Task {
	return &Task{
		ID:           id,
		Priority:     priority,
		dependencies: dependencies,
		ExecuteFunc: func() error {
			fmt.Printf("Dependent task %d executed\n", id)
			return nil
		},
	}
}

func createErrorTask(id int, priority int) *Task {
	return &Task{
		ID:       id,
		Priority: priority,
		ExecuteFunc: func() error {
			return fmt.Errorf("Error executing task %d", id)
		},
	}
}

func TestSchedulerWithDependencies(t *testing.T) {
	s := NewScheduler()

	tasks := generateTasks(10)
	for _, t := range tasks {
		s.AddTask(t)
	}

	depTask := createDependentTask(11, 5, []*Task{tasks[0], tasks[1]})
	s.AddTask(depTask)

	s.Run()

	// TODO modify the scheduler to track completed execution order for testing
}
