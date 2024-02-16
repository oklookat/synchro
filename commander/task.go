package commander

import (
	"context"
	"time"

	"github.com/oklookat/synchro/shared"
)

type OnTasker interface {
	Callback()
}

var (
	_taskInProgress  = false
	_cancelTask      context.CancelFunc
	_onTask          OnTasker
	_onTaskDone      OnTasker
	_onTaskCancelled OnTasker
)

func SetOnTask(cb OnTasker) {
	_onTask = cb
}

func SetOnTaskDone(cb OnTasker) {
	_onTaskDone = cb
}

func SetOnTaskCancelled(cb OnTasker) {
	_onTaskCancelled = cb
}

func CancelTask() {
	_log.Debug("CancelTask()")
	if _cancelTask != nil {
		_cancelTask()
	}
	_cancelTask = nil
}

func execTask(deadlineSeconds int, cb func(context.Context) error) error {
	if _taskInProgress {
		return ErrAnotherTaskInProgess
	}
	if _onTask != nil {
		_onTask.Callback()
	}

	_taskInProgress = true
	defer func() {
		_taskInProgress = false
		if _onTaskDone != nil {
			_onTaskDone.Callback()
		}
	}()

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if deadlineSeconds > 0 {
		deadline := time.Now().Add(time.Duration(deadlineSeconds) * time.Second)
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	_cancelTask = cancel

	err := cb(ctx)
	if err != nil &&
		shared.IsContextError(err) &&
		_onTaskCancelled != nil {
		_onTaskCancelled.Callback()
	}
	return err
}
