package util

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
)

// for now, we'll just keep a list of the "types" of different activities that we're going
// to use for practice:
// - simple people-based objects for very basic filtering
// - simple arrays of purchases for filtering and basic aggregation and deduplications
// - simple arrays of lottery picks for finding uniq values and total number of values
// - more complicated stuff later...
const (
	SimpleLotteryQuestions  = "simpleLotteryQuestions"
	SimplePeopleQuestions   = "simplePeopleQuestions"
	SimplePurchaseQuestions = "simplePurchaseQuestions"
	SimpleGradesQuestions   = "simpleGradesQuestions"
)

// we currently just have two string arrays to hold what is essentially either a 2-dimensional
// array or a map of some sort. We have the top level question type and each question type
// has (at least currently) a set of function types. so something like:
// People Questions:
//   - filter one
//   - keep one
//   - etc.
//
// Purchase Questions:
//   - return all currency strings
//   - return all uniq currency strings
//   - etc
//
// ...however, at some point, it might be expedient to use the same function type (eg, return all
// unique int values from an array) for multiple different question types.
//
// So we'll keep it as two different string array for now and merge or refactor or whatever into
// something a bit cleaner eventually
var PossibleQuestionTypes = []string{
	SimpleGradesQuestions, SimpleLotteryQuestions, SimplePeopleQuestions, SimplePurchaseQuestions,
}

var possibleActivities = []string{
	"hiking", "golf", "fishing", "resting", "puzzles", "tennis", "fishing",
	"baking", "knitting", "reading", "hacking", "writing", "painting",
}

var PossibleAges = []float64{
	10, 33, 39, 13, 52, 19, 64, 24, 52, 44, 20, 84, 27, 63, 27, 62, 36,
}

var possibleColors = []string{
	"red", "orange", "yellow", "blue",
	"green", "indigo", "black", "white", "purple", "brown",
}

// we currently use faker for these, but we probably want to have times when it's a much
// smaller subset to set up scenarios like having duplicates that need to be de-duped
var possibleCurrencies = []string{
	"USD", "GBP", "CNY", "JPY", "EUR", "CAD", "AUS", "MNT", "THB", "IDR",
}

var PossibleIDs = []float64{
	23345, 383885, 2229494, 192929585, 22828385, 558585,
}

var possiblePurchaseItems = []string{
	"car", "truck", "van", "plane", "boat", "motorcycle", "bus",
}

var PossibleNames = []string{
	"Alice", "Bob", "Christine", "Dan",
	"Elsa", "Frank", "Greta", "Harry", "Ingrid",
	"Jack", "Kelly", "Liam", "Mary", "Nick", "Ollie",
	"Pat", "Quinn", "Ronnie", "Sophie", "Tyler", "Vivian",
	"William", "Yvonne",
}

var PossibleLocations = []string{
	"Chicago", "London", "Paris", "Shanghai", "Nairobi", "Amsterdam",
	"Venice", "Sao Paolo", "Santiago", "Los Angeles", "New Orleans",
	"Karachi", "Kigali", "Rabat", "Zagreb", "Tokyo",
}

var PossibleSubjects = []string{"math", "art", "history"}

// searches for a value in an array and returns true if found
func ContainsElement[T comparable](s []T, id T) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}

	return false
}

// there are a lot of places where we want "3 random student names" or "4 random items", and
// it would be nice to abstract that to a function...might be we only need to do this with strings
func GetNRandomValuesFromArray[T comparable](a []T, howMany int) []T {
	rand.Seed(time.Now().UnixNano())

	var arrayItems []T

	var randomIndex int

	for i := 0; i < howMany; i++ {
		for {
			randomIndex = rand.Intn(len(a))

			if !ContainsElement(arrayItems, a[randomIndex]) {
				break
			}
		}

		arrayItems = append(arrayItems, a[randomIndex])
	}

	return arrayItems
}

// we generate an array of floats because that's what we need to conform with expectations
// of the type in "wild" json
func MakeRangeFloats(min, max int) []float64 {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}

	b := make([]float64, max-min+1)

	for i := range a {
		b[i] = float64(a[i])
	}

	return b
}

// a generic function to take an array and return one of the items at random
func GeneratePossibleValue[T any](valuesArray []T) T {
	rand.Seed(time.Now().UnixNano())

	indexToChoose := rand.Intn(len(valuesArray))

	return valuesArray[indexToChoose]
}

type FakePurchase struct {
	PurchaseID       string `faker:"uuid_digit"`
	PurchaseCurrency string `faker:"currency"`
	PurchaseItem     string
	// literally just putting this int stuff in so we can have an exercise for an array of ints
	PurchaseCode  int     `faker:"oneof: 4, 9, 18, 55, 102, 188, 225, 801, 3997"`
	PurchasePrice float64 `faker:"oneof: 1.99, 3.50, 10.81, 12.18, 19.99,, 22.00, 25.50, 31.67, 38.50, 40.99, 45.00, 67.20, 69.99, 76.89, 85.15, 90.12"`
}

// previously using faker for this, but it sometimes output duplicate names
type FakeLotteryPick struct {
	Person  string `json:"person"`
	Numbers []int  `json:"numbers"`
	Winner  bool   `json:"winner"`
}

type Grades struct {
	Results map[string][]int `json:"results"`
}

type Student struct {
	Name   string `json:"name"`
	Grades Grades `json:"grades"`
}

type ComplexGradesObject struct {
	Students []Student `json:"students"`
}

// the three structs below are crap, but will work for now
type SimplerGrades struct {
	Results map[string]int `json:"results"`
}

type SimplerStudent struct {
	Name   string        `json:"name"`
	Grades SimplerGrades `json:"grades"`
}

type SimplerGradesObject struct {
	Students []SimplerStudent `json:"students"`
}

func generateGradeResults() []int {
	var scores []int

	scoreUpperRange := 100
	scoreLowerRange := 50

	for i := 0; i < 3; i++ {
		rand.Seed(time.Now().UnixNano())
		score := rand.Intn(scoreUpperRange-scoreLowerRange) + scoreLowerRange

		scores = append(scores, score)
	}

	return scores
}

func GenerateComplexGradesObject() ComplexGradesObject {
	rand.Seed(time.Now().UnixNano())

	// just hardcode to always have 3
	var numberOfStudents int = 3

	studentNames := GetNRandomValuesFromArray(PossibleNames, numberOfStudents)

	var studentGradesArray []Student

	for i := 0; i < len(studentNames); i++ {
		studentGradesArray = append(studentGradesArray, Student{
			Name: studentNames[i], Grades: Grades{
				Results: map[string][]int{
					"art": generateGradeResults(), "math": generateGradeResults(), "history": generateGradeResults(),
				},
			},
		})
	}

	return ComplexGradesObject{
		Students: studentGradesArray,
	}
}

func GenerateLotteryPicks() []FakeLotteryPick {
	rand.Seed(time.Now().UnixNano())

	var fakePicksArray []FakeLotteryPick

	var peopleArray []string

	maxNumberOfPicks := 6

	// so we always have at least 2 picks
	randomAmountOfPicks := rand.Intn(maxNumberOfPicks) + 2

	for i := 0; i < randomAmountOfPicks; i++ {
		var numbersArray []int

		var lotteryPick FakeLotteryPick

		var randomPeopleIndex int

		for {
			randomPeopleIndex = rand.Intn(len(PossibleNames))

			if !ContainsElement(peopleArray, PossibleNames[randomPeopleIndex]) {
				break
			}
		}

		peopleArray = append(peopleArray, PossibleNames[randomPeopleIndex])
		lotteryPick.Person = PossibleNames[randomPeopleIndex]

		numberOfPicks := 5
		maxLotteryNumber := 10

		for i := 0; i < numberOfPicks; i++ {
			numbersArray = append(numbersArray, rand.Intn(maxLotteryNumber))
		}

		lotteryPick.Numbers = numbersArray
		lotteryPick.Winner = false
		fakePicksArray = append(fakePicksArray, lotteryPick)
	}

	return fakePicksArray
}

func GeneratePurchaseList() []FakePurchase {
	rand.Seed(time.Now().UnixNano())

	var purchasesArray []FakePurchase

	// hacking this so we always have at least 2 purchases...for now
	numberOfPurchases := rand.Intn(4) + 2
	for i := 0; i < numberOfPurchases; i++ {
		var purchaseItem FakePurchase

		err := faker.FakeData(&purchaseItem)
		if err != nil {
			log.Fatal(err)
		}

		purchasesArray = append(purchasesArray, purchaseItem)
	}

	// we might end up substituting something like a bounded string array of values and
	// then let faker take care of the randomization
	for i := range purchasesArray {
		purchasesArray[i].PurchaseItem = GeneratePossibleValue(possiblePurchaseItems)
		// we truncate this because it messes with the UI to have a code block that wide
		purchasesArray[i].PurchaseID = purchasesArray[i].PurchaseID[:12]
	}

	return purchasesArray
}

func PickActivities() map[string]string {
	rand.Seed(time.Now().UnixNano())

	activitiesArray := GetNRandomValuesFromArray(possibleActivities, 7)

	activitiesBase := make(map[string]string)

	daysArray := []string{
		"monday", "tuesday", "wednesday", "thursday", "friday",
		"saturday", "sunday",
	}

	// we could do this in the step above, since it's the same number of
	// iterations, but it's easier to read this way
	for i := 0; i < 7; i++ {
		activitiesBase[daysArray[i]] = activitiesArray[i]
	}

	return activitiesBase
}

func PickFavoriteColors() []string {
	rand.Seed(time.Now().UnixNano())

	// we always want at least one color
	howManyColorsToPick := rand.Intn(4) + 1

	colorsArray := GetNRandomValuesFromArray(possibleColors, howManyColorsToPick)

	return colorsArray
}

// this also feels like it should be somewhere in the standard lib, but I will shamelessly
// copy it from the internet until it is!
func Unique[T comparable](inputSlice []T) []T {
	keys := make(map[T]bool)
	list := []T{}

	for _, entry := range inputSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true

			list = append(list, entry)
		}
	}

	return list
}

func GenerateRandomBoolean() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(2) == 1
}

func GetAverageOfInts(a []int) int {
	var total int = 0

	for _, value := range a {
		total += value
	}

	return total / len(a)
}

// apparently we can't use greater than with generics...oh well
func GetHighestIntValue(a []int) int {
	max := 0

	for _, value := range a {
		if value > max {
			max = value
		}
	}

	return max
}

// shameless steal from the internet
func GetRandomKeyFromMap(m map[string][]int) string {
	rand.Seed(time.Now().UnixNano())

	k := rand.Intn(len(m))
	i := 0

	for key := range m {
		if i == k {
			return key
		}
		i++
	}

	panic("unreachable")
}
