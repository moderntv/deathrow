package deathrow

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

const DEFAULT_POPPER_RESOLUTION = 100 * time.Millisecond

type Prison[K comparable] struct {
	mu    sync.Mutex
	dr    *deathRow[K]
	items map[K]Item[K]
}

func NewPrison[K comparable]() *Prison[K] {
	return &Prison[K]{
		dr:    newDeathRow[K](),
		items: map[K]Item[K]{},
	}
}

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

func (p *Prison[K]) Popper(ctx context.Context) <-chan Item[K] {
	return p.PopperWithResolution(ctx, DEFAULT_POPPER_RESOLUTION)
}

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

				// check when's the next item and sleep for such duration
				// this is to not waste CPU cycles on swapping non-poppable items
				gotLock := p.mu.TryLock()
				if !gotLock {
					// couldn't lock -> ignore checking
					continue
				}

				first := p.dr.GetFirst()
				if first == nil {
					p.mu.Unlock()
					continue
				}
				nextT := time.Until(first.Deadline())
				if nextT > res {
					time.Sleep(nextT)
				}

				p.mu.Unlock()
			}
		}
	}(resolution)

	return ch
}
