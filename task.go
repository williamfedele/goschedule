package goschedule

type Task struct {
	ID           int
	Priority     int
	ExecuteFunc  func() error
	dependencies []*Task
	status       Status
}

type Status int

const (
	NotStarted Status = iota
	Running
	Completed
	Failed
)

func (t *Task) Execute() error {
	// Dependencies might already have been executed so we don't need to execute them again
	if t.status == Completed {
		return nil
	}

	t.status = Running
	err := t.ExecuteFunc()

	if err != nil {
		t.status = Failed
	} else {
		t.status = Completed
	}
	return err
}
