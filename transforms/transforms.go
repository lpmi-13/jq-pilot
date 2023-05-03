package transforms

import (
	"log"
	"math/rand"
	"sort"
	"time"

	"jq-pilot/util"
)

type PureJson map[string]interface{}

type PureJsonArrayPurchases map[string][]util.FakePurchase

type PureJsonArrayLottery map[string][]util.FakeLotteryPick

// these are the Simple Person Question functions
// (we'll split these different question types out into separate files eventually, but keeping
// it all in one place for now)
func DeleteOneKey(jsonInput PureJson) PureJson {
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

	return copiedJson
}

func DeleteRandomKeys(jsonInput PureJson) PureJson {
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

	return copiedJson
}

// this is all incredibly repetitive, so refactoring to be a generic function
// will be a good learning experience later :tada:
func GetOneKeyStringValue(jsonInput PureJson) string {
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

	return valueToReturn
}

func GetOneKeyIntValue(jsonInput PureJson) int {
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

	return valueToReturn
}

func KeepOneKey(jsonInput PureJson) PureJson {
	rand.Seed(time.Now().UnixNano())

	// for now, let's just assume there's always more than 1 key
	minimumToKeep := 1
	maximumToKeep := len(jsonInput) - 1

	whichKeyToKeep := rand.Intn(maximumToKeep - minimumToKeep)

	copiedJson := make(map[string]interface{})

	count := 0
	for key := range jsonInput {
		// this could be smarter about picking random keys to delete, but this is fast
		// to get working for now
		if count == whichKeyToKeep {
			copiedJson[key] = jsonInput[key]
		}
		count++
	}

	return copiedJson
}

// these are the Simple Purchase Question functions
func GetAllArrayStringValues(jsonInput PureJsonArrayPurchases) []string {
	rand.Seed(time.Now().UnixNano())

	// just grab the currency for now...this is nasty

	// keysToGrabArrayValues := []string{"purchaseID", "purchaseCurrency"}

	// totalKeys := len(keysToGrabArrayValues)
	// keyToGrab := rand.Intn(totalKeys)
	// keyToMatch := keysToGrabArrayValues[keyToGrab]
	valuesArray := []string{}

	nestedPurchases := jsonInput["purchases"]

	for i := range nestedPurchases {
		valuesArray = append(valuesArray, nestedPurchases[i].PurchaseCurrency)
	}

	log.Println(valuesArray)

	return valuesArray
}

// this also feels highly duplicated from the above, and should be generalized
func GetAllArrayIntValues(jsonInput PureJsonArrayPurchases) []int {
	rand.Seed(time.Now().UnixNano())

	// keysToGrabArrayValues := []string{"purchaseCode"}

	// totalKeys := len(keysToGrabArrayValues)
	// keyToGrab := rand.Intn(totalKeys)
	// keyToMatch := keysToGrabArrayValues[keyToGrab]
	valuesArray := []int{}

	// I feel bad about this, and you should too
	nestedPurchases := jsonInput["purchases"]
	for i := range nestedPurchases {
		valuesArray = append(valuesArray, nestedPurchases[i].PurchaseCode)
	}

	log.Println(valuesArray)

	return valuesArray
}

// here are some fuctions to transform the lottery picks stuff
func GetAllUniqueArrayIntValues(jsonInput PureJsonArrayLottery) []int {
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
	// to figure out how to do a unique without a sort, which is much more cumbersom
	sort.Ints(uniqueValues)
	log.Println("unique values:", uniqueValues)

	return uniqueValues
}

// this could definitely be more generic, since there are a lot of potential
// applications for "find the total number of things/keys/etc"
func GetNumberOfPicks(jsonInput PureJsonArrayLottery) int {
	nestedLotteryPicks := jsonInput["lotteryPicks"]

	// this is SUPES basic, like "just find the number of lottery picks"
	return len(nestedLotteryPicks)
}
