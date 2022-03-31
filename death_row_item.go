package deathrow

import "time"

// item contains the information about the expiration
type item struct {
	ID       string
	Deadline time.Time
	index    int // index in heap

	ttl time.Duration
}

func newItem(id string, ttl time.Duration) *item {
	return &item{
		ID:       id,
		Deadline: time.Now().Add(ttl),
		index:    0,

		ttl: ttl,
	}
}

// IsDeadMan decides whether this item is after its deadline
func (ag *item) IsDeadMan() bool {
	return time.Now().After(ag.Deadline)
}

func (ag *item) prolongDefault() {
	ag.prolong(ag.ttl)
}

func (ag *item) prolong(ttl time.Duration) {
	ag.Deadline = time.Now().Add(ttl)
}
