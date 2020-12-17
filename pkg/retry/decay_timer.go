package retry

import (
	"math/rand"
	"time"
)

type DecayTimer struct {
	C        chan time.Time
	backoff  float32
	duration time.Duration
	done     chan struct{}
}

func NewDecayTimer(d time.Duration, b float32) *DecayTimer {
	t := &DecayTimer{
		C:        make(chan time.Time, 1),
		backoff:  b,
		duration: d,
		done:     make(chan struct{}, 1),
	}
	go t.start()
	return t
}

func (t *DecayTimer) start() {
	d := t.duration
	for {
		select {
		case <-t.done:
			return
		default:
			if len(t.C) == 0 {
				t.C <- time.Now()
			}
			time.Sleep(d)
		}
		d = time.Duration(float32(d)*t.backoff + float32(d)*rand.Float32()/10.)
	}
}

func (t *DecayTimer) Stop() {
	t.done <- struct{}{}
}
