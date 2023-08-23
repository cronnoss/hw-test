package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type pair struct {
	word string
	freq int
}

var result []string

func Top10(s string) []string {
	defer func() {
		result = nil
	}()

	// Split the input text into individual words. This assumes that words are separated by spaces.
	words := strings.Fields(s)

	// Iterate over the slice of words and use a map to store the frequency of each word
	freqMap := make(map[string]int)
	for _, word := range words {
		freqMap[word]++
	}

	// Transform the map into a slice of pairs that can be sorted.
	pairs := make([]pair, 0, len(freqMap))
	for word, freq := range freqMap {
		pairs = append(pairs, pair{word, freq})
	}

	// Sort the slice of pairs by frequency (in descending order),
	// then lexicographically (in ascending order) if frequencies match.
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].freq == pairs[j].freq {
			return pairs[i].word < pairs[j].word
		}

		return pairs[i].freq > pairs[j].freq
	})

	// Return the first 10 entries of the sorted slice of pairs.
	for _, pair := range pairs {
		result = append(result, pair.word)
	}

	if len(result) > 10 {
		result = result[:10]
	}

	return result
}
