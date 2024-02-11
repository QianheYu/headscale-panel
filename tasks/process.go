package task

import (
	"context"
	"fmt"
	"headscale-panel/log"
	"os/exec"
	"runtime"
	"strings"
)

// operate is an interface for process operations.
type operate interface {
	IsRunning() bool
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetConfigPath() string
}

// ProcessOption is a function type for processing options.
type ProcessOption func(process *Process)

// ProcessMode is a function type for defining the process mode.
type ProcessMode func(process *Process)

// Process represents a process with configuration and application details.
type Process struct {
	p       operate
	version string
	apikey  string
	app     string
	err     error
}

// stop the process used by runtime.SetFinalizer to GC
func stop(process *Process) {
	process.Stop(context.Background())
}

// NewProcess creates a new process with a specified mode and a list of ProcessOption for expanding the process.
func NewProcess(mode ProcessMode, options ...ProcessOption) *Process {
	p := &Process{version: "Unknown"}
	runtime.SetFinalizer(p, stop)
	mode(p)
	for _, opt := range options {
		opt(p)
	}
	return p
}

// GetVersion get the version of the app(headscale)
func (p *Process) GetVersion() string {
	return p.version
}

// GetApplication get the application file path
func (p *Process) GetApplication() string {
	return p.app
}

// GetConfigPath returns the configuration file path of the process.
func (p *Process) GetConfigPath() string {
	return p.p.GetConfigPath()
}

// GetErr returns the error of the process.
func (p *Process) GetErr() error {
	return p.err
}

// IsRunning check if the process is running
func (p *Process) IsRunning() bool {
	return p.p.IsRunning()
	//if p.cmd == nil || p.cmd.Process == nil {
	//	return false
	//} else if p.cmd.ProcessState == nil {
	//	return true
	//} else {
	//	return false
	//}
}

// refreshVersion refresh the application version
func (p *Process) refreshVersion() {
	cmd := exec.Command(p.app, "version")
	data, err := cmd.Output()
	if err != nil {
		p.version = "Unknown"
	} else {
		p.version = strings.TrimSpace(string(data))
	}
}

// RefreshApiKey refreshes the API key of the process.
func (p *Process) RefreshApiKey() (string, error) {
	log.Log.Info("refresh api key")
	var err error
	if p.apikey != "" {
		log.Log.Info("expire an old api key")
		cmd := exec.Command(p.app, "apikey", "expire", "-p", strings.Split(p.apikey, ".")[0], "-c", p.GetConfigPath())
		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("expire api key failed: %w", err)
			log.Log.Error(err)
		}
	}

	log.Log.Info("create a new api key")
	cmd := exec.Command(p.app, "apikey", "create", "-c", p.GetConfigPath())
	data, err := cmd.Output()
	if err != nil {
		p.apikey = ""
		return p.apikey, fmt.Errorf("create api key failed: command: %s, error: %w", cmd.String(), err)
	}
	lines := strings.Split(string(data), "\n")
	p.apikey = strings.TrimSpace(lines[len(lines)-2])
	return p.apikey, err
}

// GetApiKey returns the API key of the process.
func (p *Process) GetApiKey() string {
	return p.apikey
}

// Start the process
func (p *Process) Start() error {
	p.refreshVersion() // get headscale version
	p.err = p.p.Start(context.Background())

	log.Log.Debugf("start headscale error: %v", p.err)
	return p.err
}

// Stop the process
func (p *Process) Stop(ctx context.Context) error {
	p.err = p.p.Stop(ctx)
	log.Log.Debugf("stop headscale error: %v", p.err)
	return p.err
}
