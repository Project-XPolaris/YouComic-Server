package services

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/sirupsen/logrus"
)

type RemoveEmptyTagTask struct {
	*task.BaseTask
	stopFlag   bool
	TaskOutput *RemoveEmptyTagTaskOutput
}

func (t *RemoveEmptyTagTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

type RemoveEmptyTagTaskOutput struct {
	CurrentTag *model.Tag
	Total      int
	Current    int
}

func (t *RemoveEmptyTagTask) Stop() error {
	t.stopFlag = true
	return nil
}

func (t *RemoveEmptyTagTask) Start() error {
	go func() {
		var tags []model.Tag
		database.Instance.Find(&tags)
		t.TaskOutput.Total = len(tags)
		for _, tag := range tags {
			t.TaskOutput.Current += 1
			t.TaskOutput.CurrentTag = &tag
			ass := database.Instance.Model(&tag).Association("Books")
			if ass.Count() == 0 {
				err := database.Instance.Unscoped().Delete(&tag).Error
				if err != nil {
					logrus.Error(err)
				}
			}
		}
		t.Status = StatusComplete
	}()
	return nil
}
func (p *TaskPool) NewRemoveEmptyTagTask() (*RemoveEmptyTagTask, error) {
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*RemoveEmptyTagTask); ok {
			if task.Status == StatusRunning {
				return true
			}
		}
		return false
	})
	if exist != nil {
		return exist.(*RemoveEmptyTagTask), nil
	}
	info := task.NewBaseTask("RemoveEmptyTag", "0", StatusRunning)
	task := &RemoveEmptyTagTask{
		BaseTask:   info,
		TaskOutput: &RemoveEmptyTagTaskOutput{},
	}
	task.Status = StatusRunning
	module.Task.Pool.AddTask(task)
	return task, nil
}
