package services

import (
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/rs/xid"
	"sync"
	"time"
)

var DefaultTaskPool = TaskPool{
	Tasks: []Task{},
}

const (
	TaskStatusInit    = "Init"
	StatusRunning     = "Running"
	StatusComplete    = "Complete"
	StatusStop        = "Stop"
	StatusError       = "Error"
	ScanStatusAnalyze = "Analyze"
	ScanStatusAdd     = "Add"
)

type TaskPool struct {
	Tasks []Task
	sync.Mutex
}

func (p *TaskPool) AddTask(task Task) {
	p.Lock()
	defer p.Unlock()
	p.Tasks = append(p.Tasks, task)
}
func (p *TaskPool) StopTask(id string) error {
	task := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		return i.(Task).GetBaseInfo().ID == id
	}).(Task)
	return task.Stop()
}

type Task interface {
	Stop() error
	Start() error
	GetBaseInfo() *BaseTask
}
type BaseTask struct {
	ID      string
	Status  string
	Created time.Time
}

func (t *BaseTask) GetBaseInfo() *BaseTask {
	return t
}
func NewBaseTask() BaseTask {
	return BaseTask{
		ID:      xid.New().String(),
		Status:  TaskStatusInit,
		Created: time.Now(),
	}
}
