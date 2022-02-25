package roulette

import (
	"chattweiler/pkg/repository/model"
	"testing"
)

func TestSpinWithEmptyPhrases(t *testing.T) {
	phrases := []model.Phrase{}
	bingo := Spin(phrases...)

	if bingo != nil {
		t.Errorf("Spin result with empty input is unexpected")
	}
}

func TestSpinWithOnePhrase(t *testing.T) {
	phrase := model.Phrase{}
	phrase.Text = "first phrase"
	phrase.Weight = 100

	bingo := Spin(phrase)

	if *bingo != phrase {
		t.Errorf("Spin result with a single phrase not the same")
	}
}
