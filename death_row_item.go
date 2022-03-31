package deathrow

import "time"

type Item interface {
	ID() string
	Deadline() time.Time
	ShouldExecute() bool
	Prolong(ttl time.Duration)

	Index() int
	SetIndex(int)
}

// Item contains the information about the expiration
type item struct {
	id       string
	deadline time.Time
	index    int // index in heap

	ttl time.Duration
}

func NewItem(id string, ttl time.Duration) Item {
	return &item{
		id:       id,
		deadline: time.Now().Add(ttl),
		index:    0,

		ttl: ttl,
	}
}

func (i *item) ID() string          { return i.id }
func (i *item) Deadline() time.Time { return i.deadline }
func (i *item) Index() int          { return i.index }
func (i *item) SetIndex(idx int)    { i.index = idx }

// ShouldExecute decides whether this item is after its deadline
func (i *item) ShouldExecute() bool {
	return time.Now().After(i.deadline)
}

func (i *item) Prolong(ttl time.Duration) {
	i.deadline = time.Now().Add(ttl)
}
