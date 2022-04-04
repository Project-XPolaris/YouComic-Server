package services

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/sirupsen/logrus"
)

type RemoveEmptyTagTask struct {
	BaseTask
	CurrentTag *model.Tag
	Total      int
	Current    int
	stopFlag   bool
}

func (t *RemoveEmptyTagTask) Stop() error {
	t.stopFlag = true
	return nil
}

func (t *RemoveEmptyTagTask) Start() error {
	go func() {
		var tags []model.Tag
		database.Instance.Find(&tags)
		t.Total = len(tags)
		for _, tag := range tags {
			t.Current += 1
			t.CurrentTag = &tag
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
	task := &RemoveEmptyTagTask{
		BaseTask: NewBaseTask(),
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}
