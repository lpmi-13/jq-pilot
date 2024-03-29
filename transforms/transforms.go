package transforms

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"jq-pilot/util"
)

type PureJson map[string]interface{}

type PureJsonArrayPurchases map[string][]util.FakePurchase

type PureJsonArrayLottery map[string][]util.FakeLotteryPick

// these are the Simple Person Question functions
// (we'll split these different question types out into separate files eventually, but keeping
// it all in one place for now)
func DeleteOneKey(jsonInput PureJson) (PureJson, string) {
	// we want to delete one of the keys (just assume we always have more than one key...
	// maybe we can look at making this more robust for all cases later)
	rand.Seed(time.Now().UnixNano())

	keyToDelete := rand.Intn(len(jsonInput))

	var keyNameToDelete string

	copiedJson := PureJson{}

	count := 0
	for key := range jsonInput {
		if count == keyToDelete {
			keyNameToDelete = key
		}
		count++
	}

	for k := range jsonInput {
		copiedJson[k] = jsonInput[k]
	}

	delete(copiedJson, keyNameToDelete)

	return copiedJson, fmt.Sprintf("delete: %s", keyNameToDelete)
}

func DeleteRandomKeys(jsonInput PureJson) (PureJson, string) {
	rand.Seed(time.Now().UnixNano())

	// for now, let's just assume there's always more than 1 key
	minimumToDelete := 1
	maximumToDelete := len(jsonInput) - 1

	howManyKeysToDelete := rand.Intn(maximumToDelete - minimumToDelete)

	var keyNamesToDelete []string

	copiedJson := make(map[string]interface{})

	count := 0
	for key := range jsonInput {
		// this could be smarter about picking random keys to delete, but this is fast
		// to get working for now
		if count <= howManyKeysToDelete {
			keyNamesToDelete = append(keyNamesToDelete, key)
		}
		count++
	}

	for k := range jsonInput {
		copiedJson[k] = jsonInput[k]
	}

	for _, keyName := range keyNamesToDelete {
		delete(copiedJson, keyName)
	}

	return copiedJson, fmt.Sprintf("delete: %s", strings.Join(keyNamesToDelete, "/"))
}

// this is all incredibly repetitive, so refactoring to be a generic function
// will be a good learning experience later :tada:
func GetOneKeyStringValue(jsonInput PureJson) (string, string) {
	rand.Seed(time.Now().UnixNano())

	keysWithStringValues := []string{"location", "name"}

	keyIndexToPick := rand.Intn(len(keysWithStringValues))
	keyToPick := keysWithStringValues[keyIndexToPick]

	var valueToReturn string

	for key, value := range jsonInput {
		if key == keyToPick {
			str, ok := value.(string)
			if !ok {
				log.Printf("Error, not a string: %#v\n", value)
			}

			valueToReturn = str
		}
	}

	return valueToReturn, fmt.Sprintf("get the raw value of: %s", keyToPick)
}

func GetOneKeyIntValue(jsonInput PureJson) (int, string) {
	rand.Seed(time.Now().UnixNano())

	keysWithIntValues := []string{"id", "age"}

	keyIndexToPick := rand.Intn(len(keysWithIntValues))
	keyToPick := keysWithIntValues[keyIndexToPick]

	var valueToReturn int

	for key, value := range jsonInput {
		if key == keyToPick {
			// this is a little hacky, since all numbers that come
			// into golang from json are floats by default
			floatValue, ok := value.(float64)
			if !ok {
				log.Printf("Error, not a number: %#v", value)
			}

			valueToReturn = int(floatValue)
		}
	}

	return valueToReturn, fmt.Sprintf("get the value of: %v", keyToPick)
}

func KeepOneKey(jsonInput PureJson) (PureJson, string) {
	rand.Seed(time.Now().UnixNano())

	// for now, let's just assume there's always more than 1 key
	minimumToKeep := 1
	maximumToKeep := len(jsonInput) - 1

	whichKeyToKeep := rand.Intn(maximumToKeep - minimumToKeep)

	var keyValue string

	copiedJson := make(map[string]interface{})

	count := 0
	for key := range jsonInput {
		// this could be smarter about picking random keys to delete, but this is fast
		// to get working for now
		if count == whichKeyToKeep {
			copiedJson[key] = jsonInput[key]
			// this is also crap, but we can clean it up later
			keyValue = key
		}
		count++
	}

	return copiedJson, fmt.Sprintf("filter to keep: %s", keyValue)
}

// these are the Simple Purchase Question functions
func GetAllArrayStringValues(jsonInput PureJsonArrayPurchases) ([]string, string) {
	rand.Seed(time.Now().UnixNano())

	nestedPurchases := jsonInput["purchases"]

	// just grab the currency for now...this is nasty
	valuesArray := make([]string, len(nestedPurchases))

	for i := range nestedPurchases {
		valuesArray[i] = nestedPurchases[i].PurchaseCurrency
	}

	return valuesArray, "get all the purchase currencies"
}

// this also feels highly duplicated from the above, and should be generalized
func GetAllArrayIntValues(jsonInput PureJsonArrayPurchases) ([]int, string) {
	rand.Seed(time.Now().UnixNano())

	// I feel bad about this, and you should too
	nestedPurchases := jsonInput["purchases"]
	valuesArray := make([]int, len(nestedPurchases))

	for i := range nestedPurchases {
		valuesArray[i] = nestedPurchases[i].PurchaseCode
	}

	return valuesArray, "get all the purchase codes"
}

func AddVerifiedToEachPurchase(jsonInput PureJsonArrayPurchases) ([]util.FakePurchaseVerified, string) {
	justPurchases := jsonInput["purchases"]
	newPurchases := make([]util.FakePurchaseVerified, len(justPurchases))

	for i := range justPurchases {
		// can't range over a struct...oh well
		newPurchases[i].PurchaseCode = justPurchases[i].PurchaseCode
		newPurchases[i].PurchaseCurrency = justPurchases[i].PurchaseCurrency
		newPurchases[i].PurchaseID = justPurchases[i].PurchaseID
		newPurchases[i].PurchaseItem = justPurchases[i].PurchaseItem
		newPurchases[i].PurchasePrice = justPurchases[i].PurchasePrice
		newPurchases[i].Verified = true
	}

	return newPurchases, "make all the purchases verified"
}

func MakeAllFieldsLowercase(jsonInput PureJsonArrayPurchases) ([]map[string]any, string) {
	justPurchases := jsonInput["purchases"]
	newPurchases := make([]map[string]any, len(justPurchases))

	for i := range justPurchases {
		// there's definitely a function that would do this automaticaly, but this was quick and easy
		intermediateMap := make(map[string]any)
		// we need to cast this to float64 since golang expects all json to be floats
		intermediateMap["purchasecode"] = float64(justPurchases[i].PurchaseCode)
		intermediateMap["purchasecurrency"] = justPurchases[i].PurchaseCurrency
		intermediateMap["purchaseid"] = justPurchases[i].PurchaseID
		intermediateMap["purchaseitem"] = justPurchases[i].PurchaseItem
		intermediateMap["purchaseprice"] = justPurchases[i].PurchasePrice

		newPurchases[i] = intermediateMap
	}

	return newPurchases, "make all the keys lowercase"
}

// this is largely going to be a duplication of the function below it, but we'll get it working for now...
func GetGroupByPurchasePrice(jsonInput PureJsonArrayPurchases) (util.FakePurchaseGrouped, string) {
	var minPurchasePrice float64 = 1000

	var maxPurchasePrice float64 = 0

	justPurchases := jsonInput["purchases"]

	for i := range justPurchases {
		if justPurchases[i].PurchasePrice > maxPurchasePrice {
			maxPurchasePrice = justPurchases[i].PurchasePrice
		}

		if justPurchases[i].PurchasePrice < minPurchasePrice {
			minPurchasePrice = justPurchases[i].PurchasePrice
		}
	}

	differenceBetweenMinMax := maxPurchasePrice - minPurchasePrice

	middlePrice := maxPurchasePrice - (differenceBetweenMinMax / 2)

	copiedArrayHigher := []util.FakePurchase{}
	copiedArrayLower := []util.FakePurchase{}

	for i := range justPurchases {
		if jsonInput["purchases"][i].PurchasePrice > middlePrice {
			copiedArrayHigher = append(copiedArrayHigher, jsonInput["purchases"][i])
		} else {
			copiedArrayLower = append(copiedArrayLower, jsonInput["purchases"][i])
		}
	}

	var groupedArray util.FakePurchaseGrouped
	groupedArray.HigherPrice = copiedArrayHigher
	groupedArray.LowerPrice = copiedArrayLower

	return groupedArray, fmt.Sprintf("group the purchases into higher and lower than %.2f", middlePrice)
}

func GetFilteredByPurchasePrice(jsonInput PureJsonArrayPurchases) ([]util.FakePurchase, string) {
	// hacky boolean to decide whether it's finding purchases above or below a certain price
	filterForHigher := util.GenerateRandomBoolean()

	// hacky constant for now
	var minPurchasePrice float64 = 1000

	var maxPurchasePrice float64 = 0

	for i := range jsonInput["purchases"] {
		if jsonInput["purchases"][i].PurchasePrice > maxPurchasePrice {
			maxPurchasePrice = jsonInput["purchases"][i].PurchasePrice
		}

		if jsonInput["purchases"][i].PurchasePrice < minPurchasePrice {
			minPurchasePrice = jsonInput["purchases"][i].PurchasePrice
		}
	}

	differenceBetweenMinMax := maxPurchasePrice - minPurchasePrice

	middlePrice := maxPurchasePrice - (differenceBetweenMinMax / 2)

	copiedArray := []util.FakePurchase{}

	// now that we have a price in the middle of the highest and lowest prices, just
	// return either the purchases with a higher price if "filterForHigher" is true,
	// or ones with a lower price if it's false
	for i := range jsonInput["purchases"] {
		if filterForHigher {
			if jsonInput["purchases"][i].PurchasePrice > middlePrice {
				copiedArray = append(copiedArray, jsonInput["purchases"][i])
			}
		} else {
			if jsonInput["purchases"][i].PurchasePrice < middlePrice {
				copiedArray = append(copiedArray, jsonInput["purchases"][i])
			}
		}
	}

	var promptString string

	if filterForHigher {
		promptString = fmt.Sprintf("find all purchases with a price above: %.2f", middlePrice)
	} else {
		promptString = fmt.Sprintf("find all purchases with a price below: %.2f", middlePrice)
	}

	return copiedArray, promptString
}

// here are some fuctions to transform the lottery picks stuff
func GetAllUniqueArrayIntValues(jsonInput PureJsonArrayLottery) ([]int, string) {
	nestedLotteryPicks := jsonInput["lotteryPicks"]
	totalValuesArray := []int{}

	// get each number from each of the lotter picks object in the array
	for i := range nestedLotteryPicks {
		for j := range nestedLotteryPicks[i].Numbers {
			totalValuesArray = append(totalValuesArray, nestedLotteryPicks[i].Numbers[j])
		}
	}

	uniqueValues := util.Unique(totalValuesArray)
	// we do this to make the jq operations easier...since the `unique` native function in
	// jq does a sort automatically, it's easier to just let users use it without needing
	// to figure out how to do a unique without a sort, which is much more cumbersome
	sort.Ints(uniqueValues)

	return uniqueValues, "get all unique lottery pick numbers"
}

// this could definitely be more generic, since there are a lot of potential
// applications for "find the total number of things/keys/etc"
func GetNumberOfPicks(jsonInput PureJsonArrayLottery) (int, string) {
	nestedLotteryPicks := jsonInput["lotteryPicks"]

	// this is SUPES basic, like "just find the number of lottery picks"
	return len(nestedLotteryPicks), "find the number of lottery picks"
}

// this is doing double duty for both a starting range and ending range rather than two
// separate functions
func GetNRangePicks(jsonInput PureJsonArrayLottery) ([]util.FakeLotteryPick, string) {
	rand.Seed(time.Now().UnixNano())
	// just hardcode this, since we want at least 2 and not more than 4
	randomNumberOfPicks := rand.Intn(3) + 1
	updatedArray := make([]util.FakeLotteryPick, len(jsonInput["lotteryPicks"]))

	rangeFromBeginning := util.GenerateRandomBoolean()

	for i := range jsonInput["lotteryPicks"] {
		var numberOfPicks []int
		if rangeFromBeginning {
			numberOfPicks = jsonInput["lotteryPicks"][i].Numbers[:randomNumberOfPicks]
		} else {
			numberOfPicks = jsonInput["lotteryPicks"][i].Numbers[randomNumberOfPicks:]
		}

		updatedArray[i] = jsonInput["lotteryPicks"][i]
		updatedArray[i].Numbers = numberOfPicks
	}

	var modifier string

	var picksPrompt int

	// my kingdom for a ternary!
	if rangeFromBeginning {
		modifier = "first"
		picksPrompt = randomNumberOfPicks
	} else {
		modifier = "last"
		picksPrompt = randomNumberOfPicks + 1
	}

	return updatedArray, fmt.Sprintf("find the %s %d picks for each person", modifier, picksPrompt)
}

func PickAWinner(jsonInput PureJsonArrayLottery) (util.FakeLotteryPick, string) {
	// this is super basic just to practice updating fields in objects
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(jsonInput["lotteryPicks"]))

	var winner util.FakeLotteryPick

	for i := range jsonInput["lotteryPicks"] {
		if i == randomIndex {
			winner = jsonInput["lotteryPicks"][i]
		}
	}

	winner.Winner = true

	return winner, fmt.Sprintf("make %s the winner", winner.Person)
}

func GetLotteryPickFrequencyDistribution(jsonInput PureJsonArrayLottery) (map[string]int, string) {
	var totalNumbers []int

	for i := range jsonInput["lotteryPicks"] {
		numbers := jsonInput["lotteryPicks"][i].Numbers
		for numberIndex := range numbers {
			totalNumbers = append(totalNumbers, numbers[numberIndex])
		}
	}

	numbersDict := make(map[string]int)

	for _, number := range totalNumbers {
		numbersDict[fmt.Sprint(number)] = numbersDict[fmt.Sprint(number)] + 1
	}

	return numbersDict, "find out how often each number was chosen"
}

// these are functions for working with the grades data

// it's very possible that we'll want this to be a general function to get both the highest and lowest scorers
// but we can just get the highest result version working for now

// I continue to bow in respect to Senseis Thomas Franklin and Rudi Tooty Fresh and Fruity for their jq skillz
func GetHighestResultInOneSubject(jsonInput util.ComplexGradesObject) (util.Student, string) {
	rand.Seed(time.Now().UnixNano())

	subjectIndex := rand.Intn(len(util.PossibleSubjects))
	selectedSubject := util.PossibleSubjects[subjectIndex]

	// we'll hold the calculations of which student had the highest score for the particular subject
	subjectResultsArray := make(map[string]int)

	for i := range jsonInput.Students {
		studentName := jsonInput.Students[i].Name
		subjectResult := util.GetAverageOfInts(jsonInput.Students[i].Grades.Results[selectedSubject])
		subjectResultsArray[studentName] = subjectResult
	}

	maxScore := 0

	var winnerName string

	for name, score := range subjectResultsArray {
		if score > maxScore {
			winnerName = name
			maxScore = score
		}
	}

	var selectedStudent util.Student

	for i := range jsonInput.Students {
		if jsonInput.Students[i].Name == winnerName {
			selectedStudent = jsonInput.Students[i]
		}
	}

	return selectedStudent, fmt.Sprintf("find the best at: %s", selectedSubject)
}

func GetHighestScoreForPersonInSubject(jsonInput util.ComplexGradesObject) (int, string) {
	rand.Seed(time.Now().UnixNano())

	// first we pick a random student
	randomStudentIndex := rand.Intn(len(jsonInput.Students))

	selectedStudentName := jsonInput.Students[randomStudentIndex].Name

	randomSubject := util.GetRandomKeyFromMap(jsonInput.Students[0].Grades.Results)

	var grades []int

	for i := range jsonInput.Students {
		if jsonInput.Students[i].Name == selectedStudentName {
			grades = jsonInput.Students[i].Grades.Results[randomSubject]
		}
	}

	// a bit unnecessary to create some state here, but go linter says it's more readable
	highestGrade := util.GetHighestIntValue(grades)

	return highestGrade, fmt.Sprintf("get the highest %s grade for %s", randomSubject, selectedStudentName)
}

func GetHighestScoreForEachSubject(jsonInput util.ComplexGradesObject) ([]util.SimplerStudent, string) {
	studentArray := make([]util.SimplerStudent, len(jsonInput.Students))

	// would be nicer to be agnostic about key names here, but we can clean that up later
	for i := range studentArray {
		studentArray[i].Name = jsonInput.Students[i].Name
		intermediateGrades := jsonInput.Students[i].Grades.Results
		artResult := util.GetHighestIntValue(intermediateGrades["art"])
		mathResults := util.GetHighestIntValue(intermediateGrades["math"])
		historyResults := util.GetHighestIntValue(intermediateGrades["history"])
		studentArray[i].Grades = util.SimplerGrades{
			Results: map[string]int{"art": artResult, "history": historyResults, "math": mathResults},
		}
	}

	return studentArray, "get the top scores for each student in each subject"
}

// this is the stuff for the tags questions

func GetDictFromArray(a []util.Tag) (map[string]string, string) {
	newMap := make(map[string]string)
	for i := range a {
		newMap[a[i].Label] = a[i].Value
	}

	return newMap, "convert the array into a dict"
}

// I can't imagine why anyone would want to do this, but for completeness sake
func GetArrayFromDict(a map[string]string) ([]util.Tag, string) {
	newArray := make([]util.Tag, len(a))

	count := 0

	for k, v := range a {
		newArray[count] = util.Tag{
			Label: k,
			Value: v,
		}
		count++
	}

	// we sort this array, since that makes it deterministic and an easier target
	// of the exercise (ie, we just need a sort_by() at the end of the jq expression)
	sort.SliceStable(newArray, func(i, j int) bool {
		return newArray[i].Label < newArray[j].Label
	})

	return newArray, "convert the dict into an array"
}
