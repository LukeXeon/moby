package sysinfo

import (
	"fmt"
	"strconv"
	"strings"
)

// defaultMaxCPUs is the normal maximum number of CPUs on Linux.
const defaultMaxCPUs = 8192

func isCpusetListAvailable(requested string, available map[int]struct{}) (bool, error) {
	// Start with the normal maximum number of CPUs on Linux, but accept
	// more if we actually have more CPUs available.
	//
	// This limit was added in f8e876d7616469d07b8b049ecb48967eeb8fa7a5
	// to address CVE-2018-20699:
	//
	// Using a value such as `--cpuset-mems=1-9223372036854775807` would cause
	// dockerd to run out of memory allocating a map of the values in the
	// validation code. Set limits to the normal limit of the number of CPUs.
	//
	// More details in https://github.com/docker-archive/engine/pull/70#issuecomment-458458288
	maxCPUs := defaultMaxCPUs
	for m := range available {
		if m > maxCPUs {
			maxCPUs = m
		}
	}
	parsedRequested, err := parseUintList(requested, maxCPUs)
	if err != nil {
		return false, err
	}
	for k := range parsedRequested {
		if _, ok := available[k]; !ok {
			return false, nil
		}
	}
	return true, nil
}

// parseUintList parses and validates the specified string as the value
// found in some cgroup file (e.g. `cpuset.cpus`, `cpuset.mems`), which could be
// one of the formats below. Note that duplicates are actually allowed in the
// input string. It returns a `map[int]bool` with available elements from `val`
// set to `true`. Values larger than `maximum` cause an error if max is non-zero,
// in order to stop the map becoming excessively large.
// Supported formats:
//
//	7
//	1-6
//	0,3-4,7,8-10
//	0-0,0,1-7
//	03,1-3      <- this is gonna get parsed as [1,2,3]
//	3,2,1
//	0-2,3,1
func parseUintList(val string, maximum int) (map[int]struct{}, error) {
	if val == "" {
		return map[int]struct{}{}, nil
	}

	availableInts := make(map[int]struct{})
	errInvalidFormat := fmt.Errorf("invalid format: %s", val)

	for r := range strings.SplitSeq(val, ",") {
		if !strings.Contains(r, "-") {
			v, err := strconv.Atoi(r)
			if err != nil {
				return nil, errInvalidFormat
			}
			if maximum != 0 && v > maximum {
				return nil, fmt.Errorf("value of out range, maximum is %d", maximum)
			}
			availableInts[v] = struct{}{}
		} else {
			minS, maxS, _ := strings.Cut(r, "-")
			minAvailable, err := strconv.Atoi(minS)
			if err != nil {
				return nil, errInvalidFormat
			}
			maxAvailable, err := strconv.Atoi(maxS)
			if err != nil {
				return nil, errInvalidFormat
			}
			if maxAvailable < minAvailable {
				return nil, errInvalidFormat
			}
			if maximum != 0 && maxAvailable > maximum {
				return nil, fmt.Errorf("value of out range, maximum is %d", maximum)
			}
			for i := minAvailable; i <= maxAvailable; i++ {
				availableInts[i] = struct{}{}
			}
		}
	}
	return availableInts, nil
}
