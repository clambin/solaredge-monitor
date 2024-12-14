package scraper

type weatherStates []string

func (w weatherStates) mostFrequent() string {
	if len(w) == 0 {
		return "" // Return an empty string if the slice is empty
	}

	// Count occurrences of each string
	frequency := make(map[string]int)
	for _, str := range w {
		frequency[str]++
	}

	// Find the string with the highest frequency
	var mostFrequent string
	var maxCount int
	for str, count := range frequency {
		if count > maxCount {
			mostFrequent = str
			maxCount = count
		}
	}
	return mostFrequent
}
