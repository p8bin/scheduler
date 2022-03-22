package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nibi8/scheduler/models"
)

// Scheme
// Job "super-job" with 5 min duration:
//  Instace 1: [set lock success] [running within the duration of the lock (5 minutes)] [try set lock] ...
//		Must complete execution before the lock expires.
//  Instace 2:       [set lock fail] [sleep during lock (5 min)                       ] [try set lock] ...
// Execution of other jobs is not affected.

type SchedulerImp struct {
	storage *Storage
}

func NewScheduler(
	sp StorageProvider,
) Scheduler {
	svc := SchedulerImp{
		storage: NewStorage(sp),
	}
	return &svc
}

func (s *SchedulerImp) RunJob(
	ctx context.Context,
	job models.Job,
) {
	// todo: ? catch panic
	go func() {
		for {
			err := s.storage.SetLock(ctx, job)
			if err != nil {
				if errors.Is(err, models.ErrNoLuck) {
					// no luck
					// continue
				} else {
					job.ErrHandler(ctx, job, err)
					time.Sleep(time.Duration(job.ErrTimeoutSec) * time.Second)
				}
			} else {
				_, err := s.runJobAction(ctx, job)
				if err != nil {
					job.ErrHandler(ctx, job, err)
					time.Sleep(time.Duration(job.ErrTimeoutSec) * time.Second)
				}
			}
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(job.PeekTimeoutSec) * time.Second)
			}
		}
	}()
}

func (s *SchedulerImp) runJobAction(
	ctx context.Context,
	job models.Job,
) (cancel context.CancelFunc, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recover panic %v", r)
		}
	}()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(job.ExecutionDurationSec)*time.Second)

	err = job.Action(ctx, job)
	if err != nil {
		return cancel, err
	}

	return cancel, nil
}
