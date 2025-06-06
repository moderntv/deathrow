package deathrow

import "time"

type Item[K comparable] interface {
	ID() K
	Deadline() time.Time
	ShouldExecute() bool
	Prolong(ttl time.Duration)

	Index() int
	SetIndex(index int)
}

// Item contains the information about the expiration.
type item[K comparable] struct {
	id       K
	deadline time.Time
	index    int // index in heap

	ttl time.Duration
}

func NewItem[K comparable](id K, ttl time.Duration) Item[K] {
	return &item[K]{
		id:       id,
		deadline: time.Now().Add(ttl),
		index:    0,

		ttl: ttl,
	}
}

func (i *item[K]) ID() K               { return i.id }
func (i *item[K]) Deadline() time.Time { return i.deadline }
func (i *item[K]) Index() int          { return i.index }
func (i *item[K]) SetIndex(idx int)    { i.index = idx }

// ShouldExecute decides whether this item is after its deadline.
func (i *item[K]) ShouldExecute() bool {
	return time.Now().After(i.deadline)
}

func (i *item[K]) Prolong(ttl time.Duration) {
	i.deadline = time.Now().Add(ttl)
}
