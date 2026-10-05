package runner

import "time"

type Usage struct {
	CPUSeconds float64
	PeakMemory int64
	AvgMemory  int64
}

// Sample starts measuring now. Call the returned function when the job ends.
func Sample(every time.Duration) func() Usage {
	cpuStart := CPUSeconds()
	done := make(chan struct{})
	finished := make(chan struct{})

	var peak, sum, n int64
	take := func() {
		m := MemoryCurrent()
		if m > peak {
			peak = m
		}
		sum += m
		n++
	}

	go func() {
		defer close(finished)
		t := time.NewTicker(every)
		defer t.Stop()
		take()
		for {
			select {
			case <-done:
				take()
				return
			case <-t.C:
				take()
			}
		}
	}()

	return func() Usage {
		close(done)
		<-finished
		u := Usage{CPUSeconds: CPUSeconds() - cpuStart, PeakMemory: peak}
		if n > 0 {
			u.AvgMemory = sum / n
		}
		return u
	}
}