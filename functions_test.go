package main

import (
	"testing"

	"jq-pilot/util"
)

var (
	stringArray = []string{"a", "b", "c", "d"}
	floatArray  = []float64{34, 22, 29, 20}
)

func TestContainsStringTrue(t *testing.T) {
	stringToFind := "b"

	actual := util.ContainsString(stringArray, stringToFind)

	if expected := true; expected != actual {
		t.Error("Expected to find the string and did not")
	}
}

func TestContainsStringFalse(t *testing.T) {
	stringToFind := "f"

	actual := util.ContainsString(stringArray, stringToFind)

	if expected := false; expected != actual {
		t.Error("Expected to not find the string and did")
	}
}

func TestContainsFloatTrue(t *testing.T) {
	floatToFind := float64(34)

	actual := util.ContainsFloat(floatArray, floatToFind)

	if expected := true; expected != actual {
		t.Error("Expected to find the float and did not")
	}
}

func TestContainsFloatFalse(t *testing.T) {
	floatToFind := float64(88)

	actual := util.ContainsFloat(floatArray, floatToFind)

	if expected := false; expected != actual {
		t.Error("Expected not to find the float and did")
	}
}
