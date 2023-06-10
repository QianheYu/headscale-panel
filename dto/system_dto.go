package dto

import "time"

type SystemInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	Branch    string `json:"branch"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
}

type SystemStatusDto struct {
	Headscale  Headscale `json:"headscale"`
	Disk       DiskSpace `json:"disk_space"`
	Memory     Memory    `json:"memory"`
	Swap       Swap      `json:"swap"`
	CPU        CPU       `json:"cpu"`
	Load       Load      `json:"load"`
	NetTraffic Net       `json:"net_traffic"`
	NetIO      NetIO     `json:"net_io"`
	Uptime     uint64    `json:"uptime"`
	T          time.Time `json:"t"`
}

type Headscale struct {
	Version     string `json:"version"`
	LastVersion string `json:"last_version"`
	Status      int    `json:"status"`
	Error       string `json:"error"`
}

type DiskSpace struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type Memory struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type Swap struct {
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	Total       uint64  `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type CPU struct {
	UsedPercent float64 `json:"used_percent"`
}

type Load struct {
	One     float64 `json:"one"`
	Five    float64 `json:"five"`
	Fifteen float64 `json:"fifteen"`
}

type Net struct {
	Recv uint64 `json:"received"`
	Sent uint64 `json:"sent"`
}

type NetIO struct {
	Up   uint64 `json:"up"`
	Down uint64 `json:"down"`
}
