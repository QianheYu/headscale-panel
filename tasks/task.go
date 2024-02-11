package task

import (
	"context"
	"github.com/robfig/cron/v3"
	"headscale-panel/config"
	"headscale-panel/dto"
	"headscale-panel/log"
	"time"
)

type Task interface {
	Restart(force bool)
	Start() error
	Stop(ctx context.Context)
}

var (
	lastStatus        *dto.SystemStatusDto
	lastGetStatusTime time.Time
)

func GetStatus() *dto.SystemStatusDto {
	lastGetStatusTime = time.Now()
	return lastStatus
}

type task struct {
	cron *cron.Cron
}

// InitTasks initializes the tasks
// and it will start some corn, like check work status and version
func InitTasks() (Task, error) {
	lastGetStatusTime = time.Now()
	t := &task{
		cron: cron.New(cron.WithSeconds()),
	}
	t.cron.Start()
	// get status
	if _, err := t.cron.AddFunc("@every 2s", func() {
		now := time.Now()
		if now.Sub(lastGetStatusTime) > time.Minute*3 {
			return
		}
		lastStatus = refreshHostStatus(lastStatus)
	}); err != nil {
		return nil, err
	}

	// Check headscale run status and version on standalone deployments only
	if config.GetMode() < config.MULTI {
		// check the headscale weather it is running
		if _, err := t.cron.AddFunc("@every 30s", h.checkProcess); err != nil {
			return nil, err
		}
		// check a new version
		if _, err := t.cron.AddFunc("@daily", h.checkNewVersion); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// Restart the task
// force param is used to force to restart the task
func (t *task) Restart(force bool) {
	if p == nil || !p.IsRunning() {
		return
	}

	if !force && p.IsRunning() {
		return
	}
	t.Stop(context.Background())
	t.Start()
}

// Start the tasks after you set them.
func (t *task) Start() error {
	// Need to start headscale for standalone deployments only
	if config.GetMode() >= config.MULTI {
		return nil
	}

	if err := h.Start(); err != nil {
		return err
	}
	h.checkNewVersion()
	h.checkProcess()
	return nil
}

// Stop the tasks with context
func (t *task) Stop(ctx context.Context) {
	t.cron.Stop()
	// Need to stop headsclae for standalone deployments only
	if config.GetMode() >= config.MULTI {
		return
	}
	if err := h.Stop(ctx); err != nil {
		log.Log.Error(err)
	}
}
