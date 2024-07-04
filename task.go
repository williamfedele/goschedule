package goschedule

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

	err := t.ExecuteFunc()

	if err == nil {
		t.executed = true
	}
	return err
}
