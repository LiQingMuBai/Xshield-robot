package tools

import (
	"testing"
	logger "ushield_bot/internal/logger"
)

func TestCompareNumberStrings(t *testing.T) {

	a := "4000"
	b := "4000.001"
	value, err := CompareNumberStrings(a, b)
	if err != nil {

		t.Error(err.Error())
	}

	logger.Println(value)
}

func TestSubtractStringNumbers(t *testing.T) {

	a := "47.99999999999999"
	b := "5.2"
	n := 1
	value, err := SubtractStringNumbers(a, b, float64(n))
	if err != nil {

		t.Error(err.Error())
	}

	logger.Println(value)
}

func TestMultiplyStringNumbers(t *testing.T) {

	energy_cost_2x, err1 := StringMultiply("1", 2)

	if err1 != nil {
		t.Error(err1.Error())
	}
	logger.Println(energy_cost_2x)
	energy_cost_10x, err2 := StringMultiply("1", 10)

	if err2 != nil {
		t.Error(err2.Error())
	}
	logger.Println(energy_cost_10x)

}
