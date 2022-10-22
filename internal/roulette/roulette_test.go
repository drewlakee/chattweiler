package roulette

import (
	"chattweiler/internal/repository/model"
	"testing"
)

func TestSpinWithEmptyPhrases(t *testing.T) {
	var phrases []model.Phrase
	bingo := Spin(phrases...)

	if bingo != nil {
		t.Errorf("Spin result with empty input is unexpected")
	}
}

func TestSpinWithOnePhrase(t *testing.T) {
	phrase := model.PhraseCsv{}
	phrase.Text = "first phrase"
	phrase.Weight = 100

	bingo := Spin(phrase)

	if bingo != phrase {
		t.Errorf("Spin result with a single phrase not the same")
	}
}
