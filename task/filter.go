package task

type Filter map[string]bool

var globalResolverFilter = NewFilter()

func NewFilter() Filter {
	return make(Filter)
}

func (t Filter) Add(depend string) {
	t[depend] = true
}

func (t Filter) Contains(depend string) bool {
	return t[depend]
}
