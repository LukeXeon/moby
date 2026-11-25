package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// getSystemCPUUsage returns the CPU time of the online CPUs since boot and
// their number. Apps cannot read /proc/stat on android. Its total counts
// every tick of every online CPU, busy or idle, so the boot time clock times
// the number of online CPUs advances the same way, and the CPU percentage
// only uses the difference between two samples.
func getSystemCPUUsage() (cpuUsage uint64, cpuNum uint32, _ error) {
	online, err := os.ReadFile("/sys/devices/system/cpu/online")
	if err != nil {
		return 0, 0, err
	}
	for _, cpus := range strings.Split(strings.TrimSpace(string(online)), ",") {
		first, last, isRange := strings.Cut(cpus, "-")
		if !isRange {
			last = first
		}
		low, err := strconv.ParseUint(first, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("parsing online CPUs %q: %w", online, err)
		}
		high, err := strconv.ParseUint(last, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("parsing online CPUs %q: %w", online, err)
		}
		cpuNum += uint32(high - low + 1)
	}
	var now unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &now); err != nil {
		return 0, 0, err
	}
	return uint64(now.Nano()) * uint64(cpuNum), cpuNum, nil
}
