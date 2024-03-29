// Package roulette provides helpful functions to
// items picking by their probability
// https://en.wikipedia.org/wiki/Fitness_proportionate_selection
package roulette

import (
	"chattweiler/internal/repository/model"
	"math/rand"
	"time"
)

func Spin(phrases ...model.Phrase) *model.Phrase {
	if len(phrases) == 0 {
		return nil
	}

	totalWeight := 0
	for _, phrase := range phrases {
		totalWeight += phrase.Weight
	}

	if totalWeight == 0 {
		totalWeight++
	}

	rand.Seed(time.Now().UnixNano())
	bingo := rand.Intn(totalWeight)

	// from the last to the first || from the first element to the last
	direction := time.Now().Unix() % 2

	if direction == 1 {
		for _, phrase := range phrases {
			bingo -= phrase.Weight
			if bingo <= 0 {
				return &phrase
			}
		}
	} else {
		index := len(phrases) - 1
		for index >= 0 {
			phrase := phrases[index]
			bingo -= phrase.Weight
			if bingo <= 0 {
				return &phrase
			}
			index--
		}
	}

	return nil
}
