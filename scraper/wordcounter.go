package scraper

type WordCounter struct {
	words map[string]int
}

func (w *WordCounter) Add(word string) {
	if w.words == nil {
		w.words = make(map[string]int)
	}
	count := w.words[word]
	w.words[word] = count + 1
}

func (w *WordCounter) GetMostUsed() string {
	var result string
	var max int

	for word, count := range w.words {
		if count > max {
			result = word
			max = count
		}
	}

	w.words = make(map[string]int)
	return result
}
