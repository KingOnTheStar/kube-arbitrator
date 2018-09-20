package scheduler_plugin

type SchedulerPlugin struct {
	HelloDLL          func(string)
	Init              func()
	GetResourceName   func() string
	OnAddNode         func(string, map[string]string)
	OnUpdateNode      func(string, map[string]string)
	OnDeleteNode      func(string)
	AssessTaskAndNode func(string, int) (int, map[string]string)
	OnAddTask         func(string, map[string]string)
	OnRemoveTask      func(string, map[string]string)
}
