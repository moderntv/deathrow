package deathrow

import "time"

// Item contains information about aggregation group expiration
type Item struct {
	ID       string
	Deadline time.Time
	index    int // index in heap

	ttl time.Duration
}

func NewItem(id string, ttl time.Duration) *Item {
	return &Item{
		ID:       id,
		Deadline: time.Now().Add(ttl),
		index:    0,

		ttl: ttl,
	}
}

// IsDeadMan decides whether this item is after its deadline
func (ag *Item) IsDeadMan() bool {
	return time.Now().After(ag.Deadline)
}

func (ag *Item) prolongDefault() {
	ag.prolong(ag.ttl)
}

func (ag *Item) prolong(ttl time.Duration) {
	ag.Deadline = time.Now().Add(ttl)
}
