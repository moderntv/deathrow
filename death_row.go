package deathrow

import (
	"container/heap"
	"math"
	"time"
)

type deathRow[K comparable] []Item[K]

func newDeathRow[K comparable]() *deathRow[K] {
	return &deathRow[K]{}
}

func (dr deathRow[K]) Len() int { return len(dr) }

func (dr deathRow[K]) Less(i, j int) bool {
	return dr[i].Deadline().Before(dr[j].Deadline())
}

func (dr deathRow[K]) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
	dr[i].SetIndex(i)
	dr[j].SetIndex(j)
}

func (dr *deathRow[K]) Push(x any) {
	n := len(*dr)
	item := x.(Item[K])
	item.SetIndex(n)

	*dr = append(*dr, item)
}

func (dr *deathRow[K]) Pop() any {
	old := *dr
	n := len(old)

	item := old[n-1]
	old[n-1] = nil    // avoid memory leak
	item.SetIndex(-1) // for safety

	*dr = old[0 : n-1]

	return item
}

func (dr *deathRow[K]) Get(idx int) Item[K] {
	if idx < 0 || idx >= len(*dr) {
		return nil
	}

	return (*dr)[idx]
}

func (dr *deathRow[K]) GetFirst() Item[K] {
	return dr.Get(0)
}

func (dr *deathRow[K]) GetLast() Item[K] {
	return dr.Get(dr.Len() - 1)
}

func (dr *deathRow[K]) prolong(item Item[K], ttl time.Duration) {
	item.Prolong(ttl)

	heap.Fix(dr, item.Index())
}

func (dr *deathRow[K]) drop(item Item[K]) {
	// set very low time to keep it at top position
	item.Prolong(-math.MaxInt)
	// remove will swap it to top and pop it - it should be poppable since its dead
	heap.Remove(dr, item.Index())
}

func (dr *deathRow[K]) canPop() bool {
	if dr.Len() <= 0 {
		return false
	}

	first := dr.GetFirst()

	if first == nil {
		return false
	}

	return first.ShouldExecute()
}
