package runner

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const cgroupDir = "/sys/fs/cgroup"

// CPUSeconds returns total CPU time used by this container so far.
func CPUSeconds() float64 {
	f, err := os.Open(cgroupDir + "/cpu.stat")
	if err != nil {
		return 0
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), " ")
		if ok && k == "usage_usec" {
			us, _ := strconv.ParseFloat(v, 64)
			return us / 1e6
		}
	}
	return 0
}

// PeakMemory returns the highest memory use in bytes (kernel 5.19 or newer).
func PeakMemory() int64 {
	b, err := os.ReadFile(cgroupDir + "/memory.peak")
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	return n
}