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

// searches for an ID in an array and returns true if found
func ContainsElement[T comparable](s []T, id T) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}

	return false
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

type FakeLotteryPick struct {
	Person string `faker:"first_name"`
	// we'll fill this in with pure golang
	Numbers []int
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

func generateGradeResults() []int {
	var scores []int

	scoreUpperRange := 100
	scoreLowerRange := 50

	for i := 0; i < 3; i++ {
		rand.Seed(time.Now().UnixNano())
		score := rand.Intn(scoreUpperRange) + scoreLowerRange

		scores = append(scores, score)
	}

	return scores
}

func GenerateComplexGradesObject() ComplexGradesObject {
	// this deeply nested map stuff is wild, but fun
	return ComplexGradesObject{
		Students: []Student{
			{Name: "Joe", Grades: Grades{Results: map[string][]int{
				"art": generateGradeResults(), "math": generateGradeResults(), "history": generateGradeResults(),
			}}},
			{Name: "Susan", Grades: Grades{Results: map[string][]int{
				"art": generateGradeResults(), "math": generateGradeResults(), "history": generateGradeResults(),
			}}},
			{Name: "Cameron", Grades: Grades{Results: map[string][]int{
				"art": generateGradeResults(), "math": generateGradeResults(), "history": generateGradeResults(),
			}}},
		},
	}
}

func GenerateLotteryPicks() []FakeLotteryPick {
	rand.Seed(time.Now().UnixNano())

	var fakePicksArray []FakeLotteryPick

	maxNumberOfPicks := 6

	// so we always have at least 2 picks
	randomAmountOfPicks := rand.Intn(maxNumberOfPicks) + 2

	for i := 0; i < randomAmountOfPicks; i++ {
		var numbersArray []int

		var lotteryPick FakeLotteryPick

		numberOfPicks := 5
		maxLotteryNumber := 10

		err := faker.FakeData(&lotteryPick)
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < numberOfPicks; i++ {
			numbersArray = append(numbersArray, rand.Intn(maxLotteryNumber))
		}

		lotteryPick.Numbers = numbersArray
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

	var activitiesArray []string

	totalActivities := len(possibleActivities)

	// hardcoded because that's how many days in the week
	for i := 0; i < 7; i++ {
		var randomActivitiesIndex int

		for {
			randomActivitiesIndex = rand.Intn(totalActivities)

			if !ContainsElement(activitiesArray, possibleActivities[randomActivitiesIndex]) {
				break
			}
		}

		activitiesArray = append(activitiesArray, possibleActivities[randomActivitiesIndex])
	}

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
	totalColors := 10

	rand.Seed(time.Now().UnixNano())

	// we always want at least one color
	howManyColorsToPick := rand.Intn(4) + 1

	var colorsArray []string

	// for however many colors get chosen to be picked, loop so we can confirm that
	// we don't add the same color twice
	// ...on the other hand, sometimes, we DO want duplicate values for things, so
	// maybe we'll end up with a more general function that takes values and either
	// returns one with only unique values (like here), or one that's totally random
	// (possibly powered by a bounded set using faker or something)
	for i := 0; i < howManyColorsToPick; i++ {
		var randomColorIndex int

		for {
			randomColorIndex = rand.Intn(totalColors)
			if !ContainsElement(colorsArray, possibleColors[randomColorIndex]) {
				break
			}
		}

		colorsArray = append(colorsArray, possibleColors[randomColorIndex])
	}

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
