package goschedule

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
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

func createErrorTask(id int, priority int) *Task {
	return &Task{
		ID:       id,
		Priority: priority,
		ExecuteFunc: func() error {
			return fmt.Errorf("Error message!")
		},
	}
}

func TestCreateTask(t *testing.T) {
	s := NewScheduler()

	task1 := s.CreateTask(3, func() error {
		return nil
	})

	task2 := s.CreateTask(5, func() error {
		return nil
	})

	assert.Equal(t, 1, task1.ID)
	assert.Equal(t, 3, task1.Priority)

	assert.Equal(t, 2, task2.ID)
	assert.Equal(t, 5, task2.Priority)

}

func TestScheduler(t *testing.T) {
	s := NewScheduler()

	tasks := generateTasks(10)
	for _, task := range tasks {
		s.AddTask(task)
	}

	s.Run()

	for _, task := range tasks {
		assert.Equal(t, Completed, task.status, fmt.Sprintf("Task %d was not marked as completed", task.ID))
		_, ok := s.completed[task.ID]
		assert.True(t, ok, fmt.Sprintf("Task %d was not marked as completed", task.ID))
	}
}

func TestSchedulerWithDependencies(t *testing.T) {
	s := NewScheduler()

	// task1 (dep) -> task3 -> task2
	task1 := createTestTask(1, 2)
	task2 := createTestTask(2, 3)
	task3 := createTestTask(3, 5)
	s.AddDependency(task3, task1)

	s.AddTask(task1)
	s.AddTask(task2)
	s.AddTask(task3)

	s.RunNext()

	assert.Equal(t, Completed, task1.status, "Task 1 was marked as started")
	_, ok := s.completed[task1.ID]
	assert.True(t, ok, "Task 1 was not marked as completed")

	assert.Equal(t, Completed, task3.status, "Task 3 was marked as started")
	_, ok = s.completed[task3.ID]
	assert.True(t, ok, "Task 3 was not marked as completed")

	s.RunNext()

	assert.Equal(t, Completed, task2.status, "Task 2 was marked as started")
	_, ok = s.completed[task2.ID]
	assert.True(t, ok, "Task 2 was not marked as completed")
}

func TestSchedulerWithError(t *testing.T) {
	s := NewScheduler()

	// executed should be false, ID should not be in completed map
	task1 := createErrorTask(1, 2)

	s.AddTask(task1)

	s.RunNext()
	assert.Equal(t, Failed, task1.status, "Task 1 was not marked as failed")
	_, ok := s.completed[task1.ID]
	assert.False(t, ok, "Task 1 was marked as completed")
}

func TestSchedulerWithErrorDependency(t *testing.T) {
	s := NewScheduler()

	// task1 fails, task2 should not be executed
	task1 := createErrorTask(1, 2)
	task2 := createTestTask(2, 3)
	s.AddDependency(task2, task1)

	s.AddTask(task1)
	s.AddTask(task2)

	s.Run()

	assert.Equal(t, Failed, task1.status, "Task 1 was not marked as failed")
	_, ok := s.completed[task1.ID]
	assert.False(t, ok, "Task 1 was marked as completed")

	assert.Equal(t, NotStarted, task2.status, "Task 2 was marked as started")
	_, ok = s.completed[task2.ID]
	assert.False(t, ok, "Task 2 was marked as completed")
}

func TestTaskSelfDependent(t *testing.T) {
	s := NewScheduler()

	task1 := createTestTask(1, 2)
	s.AddDependency(task1, task1)

	s.AddTask(task1)

	s.Run()

	assert.Equal(t, len(task1.dependencies), 0, "Task 1 has dependencies")
	assert.Equal(t, Completed, task1.status, "Task 1 was not marked as completed")
	_, ok := s.completed[task1.ID]
	assert.True(t, ok, "Task 1 was not marked as completed")
}
