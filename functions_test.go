package main

import (
	"testing"

	"jq-pilot/util"

	"github.com/go-test/deep"
)

var (
	stringArray = []string{"a", "b", "c", "d"}
	floatArray  = []float64{34, 22, 29, 20}
)

func TestContainsStringTrue(t *testing.T) {
	stringToFind := "b"

	actual := util.ContainsElement(stringArray, stringToFind)

	if expected := true; expected != actual {
		t.Errorf("Expected to find the string %s and did not", stringToFind)
	}
}

func TestContainsStringFalse(t *testing.T) {
	stringToFind := "f"

	actual := util.ContainsElement(stringArray, stringToFind)

	if expected := false; expected != actual {
		t.Errorf("Expected to not find the string %s and did", stringToFind)
	}
}

func TestContainsFloatTrue(t *testing.T) {
	floatToFind := float64(34)

	actual := util.ContainsElement(floatArray, floatToFind)

	if expected := true; expected != actual {
		t.Errorf("Expected to find the float %v and did not", floatToFind)
	}
}

func TestContainsFloatFalse(t *testing.T) {
	floatToFind := float64(88)

	actual := util.ContainsElement(floatArray, floatToFind)

	if expected := false; expected != actual {
		t.Errorf("Expected not to find the float %v and did", floatToFind)
	}
}

func TestUniqueRemovesDuplicates(t *testing.T) {
	testArray := []int{1, 2, 3, 4, 5, 5, 4, 3, 2, 1}
	expected := []int{1, 2, 3, 4, 5}
	actual := util.Unique(testArray)

	diff := deep.Equal(expected, actual)
	if diff != nil {
		t.Errorf("Expected %v to equal %v, but it didn't", expected, actual)
	}
}

func TestUniqueKeepsOnlyUniques(t *testing.T) {
	testArray := []int{1, 3, 9, 27}
	expected := []int{1, 3, 9, 27}
	actual := util.Unique(testArray)

	diff := deep.Equal(expected, actual)
	if diff != nil {
		t.Errorf("Expected %v to equal %v, but it didn't", expected, actual)
	}
}

func TestGetNRandomValuesFromArray(t *testing.T) {
	testArray := []string{"this", "that", "the other"}
	expected := 2
	randomValues := util.GetNRandomValuesFromArray(testArray, expected)

	if expected := 2; expected != len(randomValues) {
		t.Errorf("expected %v but got %v", expected, len(randomValues))
	}
}
