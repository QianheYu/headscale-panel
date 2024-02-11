package task

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"headscale-panel/config"
	"headscale-panel/dto"
	"headscale-panel/log"
	"time"
)

// Constants for system status
const (
	Connected = iota - 2
	Running
	Stop
	Error
	Disconnected
)

// Refreshes the system status and returns a SystemStatusDto
func refreshHostStatus(lastStatus *dto.SystemStatusDto) *dto.SystemStatusDto {
	now := time.Now()
	status := &dto.SystemStatusDto{
		T: time.Now(),
	}

	// Get CPU usage percentage
	percents, err := cpu.Percent(0, false)
	if err != nil {
		log.Log.Warn("get cpu percent failed:", err)
	} else {
		status.CPU.UsedPercent = percents[0]
	}

	// Get system uptime
	upTime, err := host.Uptime()
	if err != nil {
		log.Log.Warn("get uptime failed:", err)
	} else {
		status.Uptime = upTime
	}

	// Get virtual memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Log.Warn("get virtual memory failed:", err)
	} else {
		status.Memory.Used = memInfo.Used
		status.Memory.Total = memInfo.Total
		status.Memory.Free = memInfo.Free
		status.Memory.UsedPercent = memInfo.UsedPercent
	}

	// Get swap memory usage
	swapInfo, err := mem.SwapMemory()
	if err != nil {
		log.Log.Warn("get swap memory failed:", err)
	} else {
		status.Swap.Used = swapInfo.Used
		status.Swap.Free = swapInfo.Free
		status.Swap.Total = swapInfo.Total
		status.Swap.UsedPercent = swapInfo.UsedPercent
	}

	// Get disk usage
	distInfo, err := disk.Usage("/")
	if err != nil {
		log.Log.Warn("get dist usage failed:", err)
	} else {
		status.Disk.Used = distInfo.Used
		status.Disk.Total = distInfo.Total
		status.Disk.Free = distInfo.Free
		status.Disk.UsedPercent = distInfo.UsedPercent
	}

	// Get load average
	avgState, err := load.Avg()
	if err != nil {
		log.Log.Warn("get load avg failed:", err)
	} else {
		status.Load.One = avgState.Load1
		status.Load.Five = avgState.Load5
		status.Load.Fifteen = avgState.Load15
	}

	// Get network I/O counters
	ioStats, err := net.IOCounters(false)
	if err != nil {
		log.Log.Warn("get io counters failed:", err)
	} else if len(ioStats) > 0 {
		ioStat := ioStats[0]
		status.NetTraffic.Sent = ioStat.BytesSent
		status.NetTraffic.Recv = ioStat.BytesRecv

		// Calculate network I/O rate
		if lastStatus != nil {
			duration := now.Sub(lastStatus.T)
			seconds := float64(duration) / float64(time.Second)
			up := uint64(float64(status.NetTraffic.Sent-lastStatus.NetTraffic.Sent) / seconds)
			down := uint64(float64(status.NetTraffic.Recv-lastStatus.NetTraffic.Recv) / seconds)
			status.NetIO.Up = up
			status.NetIO.Down = down
		}
	} else {
		log.Log.Warn("can not find io counters")
	}

	// Get Headscale status
	if config.GetMode() < config.MULTI {
		// The case of stand-alone deployments
		if h.IsRunning() {
			status.Headscale.Status = Running
			status.Headscale.Error = ""
		} else {
			err := h.GetErr()
			log.Log.Errorf("status error: %v", err)
			if err != nil {
				status.Headscale.Status = Error
				status.Headscale.Error = err.Error()
			} else {
				status.Headscale.Status = Stop
			}
		}
		status.Headscale.Version = h.GetVersion()
		status.Headscale.LastVersion = h.GetLatestVersion()
	} else {
		// Separate deployments
		if HeadscaleControl == nil || HeadscaleControl.conn == nil || HeadscaleControl.status < 0 {
			status.Headscale.Status = Disconnected
		} else {
			status.Headscale.Status = Connected
		}
		status.Headscale.Error = ""
	}
	return status
}
