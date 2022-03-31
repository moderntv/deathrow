package deathrow

import (
	"container/heap"
	"math"
	"time"
)

type deathRow []*Item

func newDeathRow() *deathRow {
	return &deathRow{}
}

func (dr deathRow) Len() int { return len(dr) }

func (dr deathRow) Less(i, j int) bool {
	return dr[i].Deadline.Before(dr[j].Deadline)
}

func (dr deathRow) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
	dr[i].index = i
	dr[j].index = j
}

func (dr *deathRow) Push(x any) {
	n := len(*dr)
	item := x.(*Item)
	item.index = n

	*dr = append(*dr, item)
}

func (dr *deathRow) Pop() any {
	old := *dr
	n := len(old)

	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety

	*dr = old[0 : n-1]

	return item
}

func (dr *deathRow) Get(idx int) *Item {
	if idx < 0 || idx >= len(*dr) {
		return nil
	}

	return (*dr)[idx]
}

func (dr *deathRow) GetFirst() *Item {
	return dr.Get(0)
}

func (dr *deathRow) GetLast() *Item {
	return dr.Get(dr.Len() - 1)
}

func (dr *deathRow) prolong(item *Item, ttl time.Duration) {
	item.prolong(ttl)

	heap.Fix(dr, item.index)
}

func (dr *deathRow) drop(item *Item) {
	// set very low time to keep it at top position
	item.Deadline = time.Now().Add(-math.MaxInt)
	// remove will swap it to top and pop it - it should be poppable since its dead
	heap.Remove(dr, item.index)
}

func (dr *deathRow) canPop() bool {
	if dr.Len() <= 0 {
		return false
	}

	first := dr.GetFirst()

	if first == nil {
		return false
	}

	return first.IsDeadMan()
}
