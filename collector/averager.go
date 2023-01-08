package collector

type Averager struct {
	Total float64
	Count int
}

func (a *Averager) Add(value float64) {
	a.Total += value
	a.Count++
}

func (a *Averager) Average() float64 {
	value := a.Total / float64(a.Count)
	a.Total = 0
	a.Count = 0
	return value
}
