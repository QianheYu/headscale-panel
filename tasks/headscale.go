package task

import (
	"context"
	"errors"
	"fmt"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/util"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const baseURL = "https://api.github.com/repos/juanfont/headscale/releases"

var p *Process
var latestVersion = Release{}
var checkTimes int

var h HeadscaleService

type HeadscaleService struct{}

// Start the headscale process
func (h *HeadscaleService) Start() error {
	// check only on first run
	if p == nil {
		log.Log.Debug("create process")

		// check the inside or outside to init process
		if config.Conf.Headscale.Controller.Inside {
			log.Log.Debug("start inside control process")
			p = NewProcess(InsideControl(config.Conf.Headscale.App, config.Conf.Headscale.Config))
		} else {
			log.Log.Debug("start outside control process")
			p = NewProcess(
				OutsideControl(
					config.Conf.Headscale.App,
					config.Conf.Headscale.Config,
					config.Conf.Headscale.Controller.Command.Start,
					config.Conf.Headscale.Controller.Command.Stop,
				),
			)
		}

		// if application not exists then download and create it else return the error
		if info, err := os.Stat(config.Conf.Headscale.App); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			log.Log.Info("Download and install the headscale")
			if err = h.Update(); err != nil {
				return err
			}
		} else {
			if info.IsDir() {
				if err = os.Remove(config.Conf.Headscale.App); err != nil {
					return err
				}
				log.Log.Info("Download and install the headscale")
				if err := h.Update(); err != nil {
					return err
				}
			}
		}
	}

	// check the headscale config file. if not exists, then download and create it
	if info, err := os.Stat(config.Conf.Headscale.Config); err != nil || info.IsDir() {
		if errors.Is(err, os.ErrNotExist) {
			if err = h.downloadConfig(); err != nil {
				return err
			}
			return nil
		}
		if info.IsDir() {
			return errors.New("headscale config is dir")
		}
		return err
	}

	// start headscale
	log.Log.Info("start headscale")
	if err := p.Start(); err != nil {
		log.Log.Error("headscale run error:", err)
		return err
	}
	return nil
}

// Stop the headscale process
func (h *HeadscaleService) Stop(ctx context.Context) error {
	log.Log.Info("stop headscale")
	if p == nil || !p.IsRunning() {
		return nil
	}
	if err := p.Stop(ctx); err != nil {
		log.Log.Error("headscale stop error:", err)
		return err
	}
	return nil
}

// GetErr get the error
func (h *HeadscaleService) GetErr() error {
	return p.GetErr()
}

// IsRunning check if the headscale process is running
func (h *HeadscaleService) IsRunning() bool {
	if p != nil && p.IsRunning() {
		return true
	}
	return false
}

// RefreshApiKey refresh and get the api key
func (h *HeadscaleService) RefreshApiKey() (string, error) {
	return p.RefreshApiKey()
}

// GetApiKey get the api key
func (h *HeadscaleService) GetApiKey() string {
	return p.GetApiKey()
}

// GetVersion get the headscale version
func (h *HeadscaleService) GetVersion() string {
	if p == nil {
		return "Unknown"
	}
	return p.GetVersion()
}

type Release struct {
	ID         uint     `json:"id"`
	TagName    string   `json:"tag_name"`
	Assets     []Assets `json:"assets"`
	newVersion bool
}

type Assets struct {
	DownloadUrl string `json:"browser_download_url"`
}

// checkProcess check the headscale process status
func (h *HeadscaleService) checkProcess() {
	if p == nil {
		return
	}
	ok := p.IsRunning()
	if ok {
		checkTimes = 0
		return
	}

	checkTimes++
	if checkTimes > 2 {
		checkTimes = 0
		log.Log.Warn("headscale process is not running")
		// Restart the process
		if err := p.Stop(context.Background()); err != nil {
			log.Log.Errorf("deamon stop headscale process error: %v", err)
			return
		}
		log.Log.Info("restart headscale")
		if err := p.Start(); err != nil {
			log.Log.Errorf("deamon start headscale process error: %v", err)
		}
	}
}

// Gets the headscale versions from GitHub.
func (h *HeadscaleService) GetVersions() (releases []*Release, err error) {
	err = util.RequestJson2Struct(baseURL, &releases)
	return
}

// GetLatestVersion will get the latest version information of headscale from GitHub
func (h *HeadscaleService) getLatestVersion() (*Release, error) {
	releases, err := h.GetVersions()
	if err != nil {
		return nil, err
	}
	return releases[0], nil
}

// checkNewVersion check the new version
func (h *HeadscaleService) checkNewVersion() {
	release, err := h.getLatestVersion()
	if err != nil {
		log.Log.Errorf("get latest version error: %v", err)
	}
	if release != nil && release.ID > latestVersion.ID {
		latestVersion = *release
		latestVersion.newVersion = true
	}
}

// GetLatestVersion get the new version of headscale it after CheckNewVersion
func (h *HeadscaleService) GetLatestVersion() string {
	return latestVersion.TagName
}

// HaveNewVersion check if the new version is available
func (h *HeadscaleService) HaveNewVersion() bool {
	return latestVersion.newVersion
}

// Installs a specific version of headscale.
func (h *HeadscaleService) Install(id uint) error {
	release := &Release{}
	uri, err := url.JoinPath(baseURL, fmt.Sprintf("%d", id))
	if err != nil {
		return err
	}
	if err = util.RequestJson2Struct(uri, release); err != nil {
		return err
	}

	// Choose the right version according to the system architecture
	osArch := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	for _, version := range release.Assets {
		if strings.Contains(version.DownloadUrl, osArch) {
			uri = version.DownloadUrl
			break
		}
	}

	file, err := util.Download(uri, "/tmp/headscale/app", 0755)
	if err != nil {
		return fmt.Errorf("download error: %v", err)
	}

	err = util.Update(file, p.GetApplication())
	if err != nil {
		return fmt.Errorf("update error: %v", err)
	}
	return nil
}

// Update will download the latest version of headscale from GitHub
// And move it to /tmp/headscale, then update the app file in workdir
func (h *HeadscaleService) Update() error {
	release, err := h.getLatestVersion()
	if err != nil {
		return err
	}

	if err = h.Install(release.ID); err != nil {
		return err
	}
	latestVersion = *release
	return nil
}

// Rollback will delete the old version of headscale
func (h *HeadscaleService) Rollback() error {
	oldVersion := fmt.Sprintf("%s.back", p.GetApplication())
	if _, err := os.Stat(oldVersion); err != nil {
		return errors.New("not found old version on server")
	}

	if err := os.Remove(p.GetApplication()); err != nil {
		return err
	}

	if err := os.Rename(oldVersion, p.GetApplication()); err != nil {
		return err
	}
	return nil
}

// Downloads the headscale config file from GitHub.
func (h *HeadscaleService) downloadConfig() error {
	file, err := util.Download("https://raw.githubusercontent.com/juanfont/headscale/master/config-example.yaml", "/tmp/headscale/config.yaml", 0644)
	if err != nil {
		return fmt.Errorf("download config error: %v", err)
	}

	err = util.Update(file, p.GetConfigPath())
	if err != nil {
		return fmt.Errorf("update config error: %v", err)
	}
	return nil
}
