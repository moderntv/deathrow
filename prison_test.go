package deathrow

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewPrison(t *testing.T) {
	tests := []struct {
		ID   string
		want *Prison
	}{
		{
			ID: "basic",
			want: &Prison{
				mu:    sync.Mutex{},
				dr:    newDeathRow(),
				items: map[string]*Item{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			if got := NewPrison(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPrison() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_Prison_Push(t *testing.T) {
	type push struct {
		itemID string
		ttl    time.Duration
	}
	tests := []struct {
		ID     string
		pushes []push
	}{
		{
			ID: "single",
			pushes: []push{
				{itemID: "item1", ttl: 0},
			},
		},
		{
			ID: "various",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
				{itemID: "item3", ttl: 1 * time.Second},
				{itemID: "item4", ttl: 1 * time.Second},
			},
		},
		{
			ID: "single-multiple",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item1", ttl: 1 * time.Second},
			},
		},
		{
			ID: "mix",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
				{itemID: "item3", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item1", ttl: 1 * time.Second},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			t.Parallel()
			p := NewPrison()

			uniqueitems := map[string]bool{}
			for _, push := range tt.pushes {
				p.Push(push.itemID, push.ttl)
				uniqueitems[push.itemID] = true
			}

			ugl := len(uniqueitems)
			pl := len(p.items)
			hl := p.dr.Len()
			if ugl != pl || ugl != hl || pl != hl {
				t.Errorf("expected `%d` items in heap, got `%d` in heap wrapper and `%d` in heap", ugl, pl, hl)
			}
		})
	}
}

func Test_Prison_Pop(t *testing.T) {
	type push struct {
		itemID string
		ttl    time.Duration

		// NOT a real push - will wait some time during the pushing phase
		justWaitDuration time.Duration
	}
	tests := []struct {
		ID           string
		pushes       []push
		waitDuration time.Duration
		wantItemIDs  []string
	}{
		{
			ID:           "empty",
			pushes:       []push{},
			waitDuration: 0,
			wantItemIDs:  []string{},
		},
		{
			ID:           "single-expired",
			pushes:       []push{{itemID: "item1", ttl: 1 * time.Second}},
			waitDuration: 2 * time.Second,
			wantItemIDs:  []string{"item1"},
		},
		{
			ID:           "single-nonexpired",
			pushes:       []push{{itemID: "item1", ttl: 5 * time.Second}},
			waitDuration: 2 * time.Second,
			wantItemIDs:  []string{},
		},
		{
			ID: "nonexpired-expired",
			pushes: []push{
				{itemID: "item1", ttl: 5 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
			},
			waitDuration: 2 * time.Second,
			wantItemIDs:  []string{"item2"},
		},
		{
			ID: "expired-nonexpired",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 5 * time.Second},
			},
			waitDuration: 2 * time.Second,
			wantItemIDs:  []string{"item1"},
		},
		{
			ID: "keep-alive-single",
			pushes: []push{
				{itemID: "item1", ttl: 3 * time.Second},
				{justWaitDuration: 2 * time.Second},
				{itemID: "item1", ttl: 3 * time.Second},
			},
			waitDuration: 2 * time.Second,
			wantItemIDs:  []string{},
		},
		{
			// push both at the same time, wait until almost expired, prolong one
			ID: "keep-alive-mix",
			pushes: []push{
				{itemID: "item1", ttl: 3 * time.Second},
				{itemID: "item2", ttl: 3 * time.Second},
				{justWaitDuration: 2 * time.Second},
				// at this point both have ttl 1s
				{itemID: "item1", ttl: 3 * time.Second},
				// g1 has ttl 3s
				// g2 has ttl 1s
			},
			waitDuration: 1 * time.Second,
			wantItemIDs:  []string{"item2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			t.Parallel()
			p := NewPrison()

			for _, push := range tt.pushes {
				if push.justWaitDuration > 0 {
					time.Sleep(push.justWaitDuration)
					continue
				}

				p.Push(push.itemID, push.ttl)
			}

			time.Sleep(tt.waitDuration)

			gotItems := p.Pop()
			gotItemIDs := []string{}
			for _, item := range gotItems {
				t.Logf("is item %s dead? %v", item.ID, item.IsDeadMan())
				gotItemIDs = append(gotItemIDs, item.ID)
			}

			if !reflect.DeepEqual(gotItemIDs, tt.wantItemIDs) {
				t.Errorf("Prison.Pop() = %+v, want %+v", gotItemIDs, tt.wantItemIDs)
			}
		})
	}
}

func Test_Prison_Drop(t *testing.T) {
	type push struct {
		itemID string
		ttl    time.Duration
	}
	type drop struct {
		itemID string
		ttl    time.Duration
	}
	tests := []struct {
		ID     string
		pushes []push
		drops  []drop
	}{
		{
			ID:     "nonexisting",
			pushes: []push{},
			drops:  []drop{{itemID: "item1"}},
		},
		{
			ID:     "single",
			pushes: []push{{itemID: "item1", ttl: 1 * time.Second}},
			drops:  []drop{{itemID: "item1"}},
		},
		{
			ID: "many-drop-first",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
				{itemID: "item3", ttl: 1 * time.Second},
			},
			drops: []drop{
				{itemID: "item1"},
			},
		},
		{
			ID: "many-drop-last",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
				{itemID: "item3", ttl: 1 * time.Second},
			},
			drops: []drop{
				{itemID: "item3"},
			},
		},
		{
			ID: "many-drop-middle",
			pushes: []push{
				{itemID: "item1", ttl: 1 * time.Second},
				{itemID: "item2", ttl: 1 * time.Second},
				{itemID: "item3", ttl: 1 * time.Second},
			},
			drops: []drop{
				{itemID: "item2"},
			},
		},
		{
			ID: "many-drop-middle2",
			pushes: []push{
				{itemID: "group1", ttl: 1 * time.Second},
				{itemID: "group2", ttl: 1 * time.Second},
				{itemID: "group3", ttl: 1 * time.Second},
				{itemID: "group4", ttl: 1 * time.Second},
				{itemID: "group5", ttl: 1 * time.Second},
			},
			drops: []drop{
				{itemID: "group2"},
			},
		},
		{
			ID: "many-drop-expired",
			pushes: []push{
				{itemID: "group1", ttl: 1 * time.Second},
				{itemID: "group2", ttl: -1 * time.Second},
				{itemID: "group3", ttl: 1 * time.Second},
			},
			drops: []drop{
				{itemID: "group2"},
			},
		},
		{
			ID: "many-drop-while-expired-stays",
			pushes: []push{
				{itemID: "group1", ttl: 1 * time.Second},
				{itemID: "group2", ttl: -1 * time.Second},
				{itemID: "group3", ttl: 1 * time.Second},
			},
			drops: []drop{
				{itemID: "group3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			t.Parallel()
			p := NewPrison()

			uniqueGroups := map[string]bool{}
			for _, push := range tt.pushes {
				p.Push(push.itemID, push.ttl)
				uniqueGroups[push.itemID] = true
			}

			for _, drop := range tt.drops {
				p.Drop(drop.itemID)
				delete(uniqueGroups, drop.itemID)
			}

			ugl := len(uniqueGroups)
			gehl := len(p.items)
			hl := p.dr.Len()
			if ugl != gehl || ugl != hl || gehl != hl {
				t.Errorf("expected `%d` groups in heap, got `%d` in heap wrapper and `%d` in heap", ugl, gehl, hl)
			}

			for i, item := range *p.dr {
				if i != item.index {
					t.Errorf("group at %d has index %d", i, item.index)
				}
			}
		})
	}
}

func TestPrisonComplex(t *testing.T) {
	p := NewPrison()

	// push a lot at once
	batchN := 10
	for i := 0; i < batchN; i++ {
		p.Push(fmt.Sprintf("item%d", i), 1*time.Second)
	}

	dropped := 3
	p.Drop("item0")
	p.Drop(fmt.Sprintf("item%d", int(batchN/2)))
	p.Drop(fmt.Sprintf("item%d", int(batchN/2)+1))

	p.mu.Lock()
	for i, item := range *p.dr {
		if i != item.index {
			t.Errorf("group at %d has index %d", i, item.index)
		}
	}
	p.mu.Unlock()

	// wait for the first batch to expire
	time.Sleep(2 * time.Second)

	popped := p.Pop()
	pl := len(popped)
	if pl != batchN-dropped {
		t.Errorf("did not pop all items: got %d, want %d", pl, batchN-dropped)
	}
}

func TestPrisonPopper(t *testing.T) {
	p := NewPrison()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batchN := 10
	for i := 0; i < batchN; i++ {
		p.Push(fmt.Sprintf("item%d", i), time.Duration(i/2)*time.Second)
	}

	now := time.Now()
	c := p.Popper(ctx)
	for i := 0; i < batchN; i++ {
		item := <-c
		t.Logf("popped item `%s` after %+v", item.ID, time.Since(now))
	}
}
