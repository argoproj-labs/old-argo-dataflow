package dedupe

type items []*item

func (is items) Len() int { return len(is) }

func (is items) Less(i, j int) bool {
	return is[i].lastObserved.After(is[j].lastObserved)
}

func (is items) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
	is[i].index = i
	is[j].index = j
}

func (is *items) Push(x interface{}) {
	n := len(*is)
	item := x.(*item)
	item.index = n
	*is = append(*is, item)
}

func (is *items) Pop() interface{} {
	old := *is
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*is = old[0 : n-1]
	return item
}
