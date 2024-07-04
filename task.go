package schedule

import (
	"log"
)

type Task struct {
	ID           int
	Priority     int
	ExecuteFunc  func() error
	dependencies []*Task
	executed     bool
}

func (t *Task) Execute() error {
	// Dependencies might already have been executed so we don't need to execute them again
	if t.executed {
		return nil
	}

	for _, dep := range t.dependencies {
		log.Printf("Executing dependency %d for task %d", dep.ID, t.ID)
		if err := dep.Execute(); err != nil {
			return err
		}
	}
	err := t.ExecuteFunc()

	if err == nil {
		t.executed = true
	}
	return err
}
