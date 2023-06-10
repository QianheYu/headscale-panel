package util

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

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

// InsideControl returns a ProcessMode for inside control mode with the given appFile and configFile.
func InsideControl(appFile, configFile string) ProcessMode {
	return func(process *Process) {
		process.app = appFile
		process.p = newInsideProcess(appFile, configFile)
	}
}

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
	//log.Log.Info("refresh api key")
	fmt.Println("refresh api key")
	var err error
	if p.apikey != "" {
		//log.Log.Info("expire an old api key")
		fmt.Println("expire an old api key")
		cmd := exec.Command(p.app, "apikey", "expire", "-p", strings.Split(p.apikey, ".")[0], "-c", p.GetConfigPath())
		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("expire api key failed: %w", err)
			fmt.Printf("%s\n", err.Error())
		}
	}

	//log.Log.Info("create a new api key")
	fmt.Println("create a new api key")
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
	return p.err
}

// Stop the process
func (p *Process) Stop(ctx context.Context) error {
	p.err = p.p.Stop(ctx)
	return p.err
}

// operate is an interface for process operations.
type operate interface {
	IsRunning() bool
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetConfigPath() string
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
		fmt.Println("inside process started")
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
