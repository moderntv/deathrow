package deathrow

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

const DEFAULT_POPPER_RESOLUTION = 100 * time.Millisecond //nolint: revive

// Prison takes care of its prisoners (items) and their executions (timeouts).
// Prison is the main structure in this package. It contains a priority queue based on the
// deadlines of its items as well as backreferences to the items
// in the queue, which makes accessing the dead items as well as specific items easy and efficient.
type Prison[K comparable] struct {
	mu    sync.Mutex
	dr    *deathRow[K]
	items map[K]Item[K]
}

// NewPrison creates new Prison without any items.
func NewPrison[K comparable]() *Prison[K] {
	return &Prison[K]{
		dr:    newDeathRow[K](),
		items: map[K]Item[K]{},
	}
}

// Push adds new item to the Prison. If the item already exists, its TTL is prolonged by `ttl`.
func (p *Prison[K]) Push(itemID K, ttl time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// extend only for existing items
	item, ok := p.items[itemID]
	if ok {
		p.dr.prolong(item, ttl)
		return
	}

	item = NewItem(itemID, ttl)
	heap.Push(p.dr, item)
	p.items[itemID] = item
}

// Pop pops all expired items in Prison.
// If there are no such items, it returns empty slice.
func (p *Prison[K]) Pop() (items []Item[K]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	items = []Item[K]{}
	for p.dr.canPop() {
		itemI := heap.Pop(p.dr)
		if itemI == nil {
			continue
		}

		item := itemI.(Item[K])
		items = append(items, item)
	}

	// delete from prison
	for _, item := range items {
		delete(p.items, item.ID())
	}

	return
}

// Drop removes an item from the Prison. It doesn't have to be expired.
func (p *Prison[K]) Drop(itemID K) {
	p.mu.Lock()
	defer p.mu.Unlock()

	item, ok := p.items[itemID]
	if !ok {
		return
	}

	p.dr.drop(item)
	delete(p.items, itemID)
}

// Popper is the same as PopperWithResolution but with the default resolution.
func (p *Prison[K]) Popper(ctx context.Context) <-chan Item[K] {
	return p.PopperWithResolution(ctx, DEFAULT_POPPER_RESOLUTION)
}

// PopperWithResolution returns a new channel
// into which newly expired popped items are periodically (every `resolution`) pushed.
// This loop ends when `ctx` is cancelled.
func (p *Prison[K]) PopperWithResolution(ctx context.Context, resolution time.Duration) <-chan Item[K] {
	ch := make(chan Item[K])

	go func(res time.Duration) {
		t := time.NewTicker(res)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case <-t.C:
				poppedItems := p.Pop()
				for _, item := range poppedItems {
					ch <- item
				}
			}
		}
	}(resolution)

	return ch
}
