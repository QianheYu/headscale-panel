package task

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"headscale-panel/log"
	"os/exec"
	"time"
)

// InsideControl returns a ProcessMode for inside control mode with the given appFile and configFile.
func InsideControl(appFile, configFile string) ProcessMode {
	return func(process *Process) {
		process.app = appFile
		process.p = newInsideProcess(appFile, configFile)
	}
}

// insideProcess represents an inside control process.
type insideProcess struct {
	app    string
	config string
	cmd    *exec.Cmd
	eg     *errgroup.Group
	ctx    context.Context
	cancel context.CancelFunc
}

// newInsideProcess creates a new inside control process with the given app and config.
func newInsideProcess(app, config string) operate {
	return &insideProcess{
		app:    app,
		config: config,
	}
}

// IsRunning checks if the inside process is running.
func (p *insideProcess) IsRunning() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	if p.cmd.ProcessState != nil {
		return false
	}

	return true
}

// Start starts the inside process.
func (p *insideProcess) Start(ctx context.Context) error {
	defer func() {
		log.Log.Info("inside process started")
	}()
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.eg, p.ctx = errgroup.WithContext(p.ctx)
	if p.IsRunning() {
		return errors.New("headscale has running")
	}
	p.cmd = exec.Command(p.app, "serve", "-c", p.config)
	p.eg.SetLimit(2)
	p.eg.Go(func() error {
		err := p.cmd.Run()
		return err
	})
	p.eg.Go(func() error {
		<-p.ctx.Done()
		if p.cmd == nil || p.cmd.Process == nil || !p.IsRunning() {
			return nil
		}
		return p.cmd.Process.Kill()
	})
	return nil
}

// Stop stops the inside process.
func (p *insideProcess) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if p.cancel == nil {
		return nil
	}
	p.cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stop headscale timeout: %w", ctx.Err())
		case <-p.ctx.Done():
			if err := p.ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("stop headscale failed: %w", err)
			}
			if p.eg == nil {
				return nil
			}
			if err := p.eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		}
	}
}

// GetConfigPath returns the configuration file path of the inside process.
func (p *insideProcess) GetConfigPath() string {
	return p.config
}
