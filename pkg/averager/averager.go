package averager

type Number interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 |
		~float32 | ~float64
}

type Averager[T Number] struct {
	total T
	count int
}

func (a *Averager[T]) Add(value T) {
	a.total += value
	a.count++
}

func (a *Averager[T]) Count() int {
	return a.count
}

func (a *Averager[T]) Average() T {
	value := float64(a.total) / float64(a.count)
	a.total = 0
	a.count = 0
	return T(value)
}
