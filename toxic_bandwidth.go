package main

import "time"

// The BandwidthToxic passes data through at a limited rate
type BandwidthToxic struct {
	Enabled bool `json:"enabled"`
	// Rate in KB/s
	Rate int64 `json:"rate"`
}

func (t *BandwidthToxic) Name() string {
	return "bandwidth"
}

func (t *BandwidthToxic) IsEnabled() bool {
	return t.Enabled
}

func (t *BandwidthToxic) SetEnabled(enabled bool) {
	t.Enabled = enabled
}

func (t *BandwidthToxic) Pipe(stub *ToxicStub) {
	for {
		select {
		case <-stub.interrupt:
			return
		case p := <-stub.input:
			if p == nil {
				stub.Close()
				return
			}
			var sleep time.Duration
			if t.Rate <= 0 {
				sleep = 0
			} else {
				sleep = time.Duration(len(p.data)) * time.Millisecond / time.Duration(t.Rate)
			}
			// If the rate is low enough, split the packet up and send in 100 millisecond intervals
			for sleep > 100*time.Millisecond {
				select {
				case <-time.After(100 * time.Millisecond):
					stub.output <- &StreamChunk{p.data[:t.Rate*100], p.timestamp}
					p.data = p.data[t.Rate*100:]
					sleep -= 100 * time.Millisecond
				case <-stub.interrupt:
					stub.output <- p // Don't drop any data on the floor
					return
				}
			}
			select {
			case <-time.After(sleep):
				stub.output <- p
			case <-stub.interrupt:
				stub.output <- p // Don't drop any data on the floor
				return
			}
		}
	}
}
