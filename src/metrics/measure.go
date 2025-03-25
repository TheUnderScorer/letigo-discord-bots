package metrics

import "time"

type Measure struct {
	start *time.Time
	end   *time.Time
}

func NewMeasure() *Measure {
	return &Measure{}
}

func (m *Measure) Start() {
	now := time.Now()
	m.start = &now
}

func (m *Measure) End() {
	now := time.Now()
	m.end = &now
}

func (m *Measure) Duration() time.Duration {
	if m.end == nil || m.start == nil {
		return 0
	}

	return m.end.Sub(*m.start)
}
