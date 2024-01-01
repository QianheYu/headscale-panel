package task

import (
	"context"
	"os/exec"
	"strings"
)

// OutsideControl returns a ProcessMode for outside control mode with the given app, config, start, and stop commands.
func OutsideControl(app, config, start, stop string) ProcessMode {
	return func(process *Process) {
		process.app = app
		strs := strings.Split(process.app, "/")
		if strs != nil {
			app = strs[len(strs)-1]
		}
		process.p = newOutsideProcess(app, config, start, stop)
	}
}

// outsideProcess represents an outside control process.
type outsideProcess struct {
	app    string
	config string
	start  string
	stop   string
	status string
	cmd    *exec.Cmd
}

// newOutsideProcess creates a new outside control process with the given app, config, start, and stop commands.
func newOutsideProcess(app, config, start, stop string) operate {
	return &outsideProcess{
		app:    app,
		config: config,
		start:  start,
		stop:   stop,
	}
}

// IsRunning checks if the outside process is running.
func (o *outsideProcess) IsRunning() bool {
	o.cmd = exec.Command("pgrep", o.app)
	output, err := o.cmd.Output()
	if err != nil {
		return false
	}
	if output != nil && len(output) > 0 {
		return true
	}
	return false
}

// Start starts the outside process.
func (o *outsideProcess) Start(ctx context.Context) error {
	o.cmd = exec.Command(strings.Split(o.start, " ")[0], strings.Split(o.start, " ")[1:]...)
	return o.cmd.Run()
}

// Stop stops the outside process.
func (o *outsideProcess) Stop(ctx context.Context) error {
	o.cmd = exec.Command(strings.Split(o.stop, " ")[0], strings.Split(o.stop, " ")[1:]...)
	return o.cmd.Run()
}

// GetConfigPath returns the configuration file path of the outside process.
func (o *outsideProcess) GetConfigPath() string {
	return o.config
}
