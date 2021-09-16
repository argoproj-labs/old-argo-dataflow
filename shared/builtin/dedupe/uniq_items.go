package dedupe

import (
	"container/heap"
	"time"
)

type uniqItems struct {
	ids map[string]*item
	items
}

func (is *uniqItems) update(id string) bool {
	i, ok := is.ids[id]
	if ok {
		i.lastObserved = time.Now()
		heap.Fix(&is.items, i.index)
	} else {
		i = &item{id: id, lastObserved: time.Now()}
		heap.Push(&is.items, i)
		is.ids[id] = i
	}
	return ok
}

func (is *uniqItems) shrink() {
	i := heap.Pop(&is.items).(*item)
	delete(is.ids, i.id)
}

func (is *uniqItems) size() int {
	return len(is.ids)
}
