package sysinfo

import (
	"fmt"
	"runtime"
)

// WithCgroup2GroupPath is a no-op on android, where container cgroups are
// provided by the runtime rather than the host kernel.
func WithCgroup2GroupPath(string) Opt {
	return func(*SysInfo) {}
}

// New reports the container capabilities provided by the runtime on android
// instead of probing the host's cgroup hierarchy, which an app cannot read.
func New(options ...Opt) *SysInfo {
	ncpu := max(runtime.NumCPU(), 1)
	cpuSets := make(map[int]struct{}, ncpu)
	for i := range ncpu {
		cpuSets[i] = struct{}{}
	}
	cpus := "0"
	if ncpu > 1 {
		cpus = fmt.Sprintf("0-%d", ncpu-1)
	}

	info := &SysInfo{
		cgroupMemInfo: cgroupMemInfo{
			MemoryLimit:       true,
			MemoryReservation: true,
		},
		cgroupCPUInfo: cgroupCPUInfo{
			CPUShares: true,
			CPUCfs:    true,
		},
		cgroupCpusetInfo: cgroupCpusetInfo{
			Cpuset:  true,
			Cpus:    cpus,
			CPUSets: cpuSets,
			Mems:    "0",
			MemSets: map[int]struct{}{0: {}},
		},
		cgroupPids: cgroupPids{
			PidsLimit: true,
		},
		CgroupNamespaces: true,
	}
	applyNetworkingInfo(info)
	for _, o := range options {
		o(info)
	}
	return info
}
