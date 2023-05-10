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

// these tests should use formatting to print out the expected and received valz

func TestContainsStringTrue(t *testing.T) {
	stringToFind := "b"

	actual := util.ContainsElement(stringArray, stringToFind)

	if expected := true; expected != actual {
		t.Error("Expected to find the string and did not")
	}
}

func TestContainsStringFalse(t *testing.T) {
	stringToFind := "f"

	actual := util.ContainsElement(stringArray, stringToFind)

	if expected := false; expected != actual {
		t.Error("Expected to not find the string and did")
	}
}

func TestContainsFloatTrue(t *testing.T) {
	floatToFind := float64(34)

	actual := util.ContainsElement(floatArray, floatToFind)

	if expected := true; expected != actual {
		t.Error("Expected to find the float and did not")
	}
}

func TestContainsFloatFalse(t *testing.T) {
	floatToFind := float64(88)

	actual := util.ContainsElement(floatArray, floatToFind)

	if expected := false; expected != actual {
		t.Error("Expected not to find the float and did")
	}
}

func TestUniqueRemovesDuplicates(t *testing.T) {
	testArray := []int{1, 2, 3, 4, 5, 5, 4, 3, 2, 1}
	expected := []int{1, 2, 3, 4, 5}
	actual := util.Unique(testArray)

	diff := deep.Equal(expected, actual)
	if diff != nil {
		t.Error("Expected to filter unique values, but didn't")
	}
}

func TestUniqueKeepsOnlyUniques(t *testing.T) {
	testArray := []int{1, 3, 9, 27}
	expected := []int{1, 3, 9, 27}
	actual := util.Unique(testArray)

	diff := deep.Equal(expected, actual)
	if diff != nil {
		t.Error("didn't keep all the unique values")
	}
}
