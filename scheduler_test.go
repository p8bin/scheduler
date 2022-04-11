package scheduler

import (
	"context"
	"fmt"
	"testing"

	"github.com/nibi8/scheduler/models"
	
	"github.com/nibi8/dlocker"
	"github.com/nibi8/dlocker/storageproviders/testsp"
)

// todo: add tests

func TestNewScheduler(t *testing.T) {
	sp := testsp.NewStorageProvider()
	locker := dlocker.NewLocker(sp)
	sc := NewScheduler(locker)

	job, err := models.NewJobEx(
		"unique-job-name",
		30,
		10,
		func(ctx context.Context, job models.Job) error {
			fmt.Println("start job action")
			fmt.Println("end before ctx.Done()")
			return nil
		}, 0, 0, nil,
	)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = sc.RunJob(ctx, job)
	if err != nil {
		t.Error(err)
	}
}
