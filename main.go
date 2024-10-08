package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"jq-pilot/transforms"
	"jq-pilot/util"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-test/deep"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// these generics are hashtag delicious, and I'm here for it
type JsonQuestion[T any, V any] struct {
	Question T      `json:"question"`
	Answer   V      `json:"answer"`
	Prompt   string `json:"prompt"`
}

const (
	jsonToJson           = "jsonToJson"
	jsonToElaboratedJson = "jsonToElaboratedJson"
	jsonToLowercaseJson  = "jsonToLowercaseJson"
	jsonToGroupedJson    = "jsonToGroupedJson"
	// this just commemorates when I solved a particularly tricky jq slicing problem
	jsonToRidicJson   = "jsonToRidicJson"
	jsonToString      = "jsonToString"
	jsonToInt         = "jsonToInt"
	jsonToStringArray = "jsonToStringArray"
	jsonToIntArray    = "jsonToIntArray"
	jsonToSmallerJson = "jsonToSmallerJson"
	jsonToDict        = "jsonToDict"
	jsonToArray       = "jsonToArray"
)

var (
	peopleFunctionTypesToCall = []string{
		jsonToJson, jsonToString, jsonToInt,
	}
	purchaseFunctionTypesToCall = []string{
		jsonToStringArray, jsonToIntArray, jsonToJson, jsonToElaboratedJson, jsonToLowercaseJson, jsonToGroupedJson,
	}
	lotteryFunctionTypesToCall = []string{
		jsonToJson, jsonToIntArray, jsonToInt, jsonToSmallerJson, jsonToDict,
	}
	gradesFunctionTypesToCall = []string{
		jsonToInt, jsonToJson, jsonToRidicJson,
	}
	// there are two of these function types because we needed to break up the
	// tags questions into two different starting data types
	tagsArrayFunctionTypesToCall = []string{
		jsonToDict,
	}
	tagsDictFunctionTypesToCall = []string{
		jsonToArray,
	}
	delay                           = 500
	totalJsonToJsonFunctions        = 3
	currentQuestionType             = util.SimplePeopleQuestions
	currentFunctionType             string
	personQuestionData              transforms.PureJson
	personAnswerDataJson            transforms.PureJson
	personAnswerDataString          string
	personAnswerDataInt             int
	purchaseQuestionData            transforms.PureJsonArrayPurchases
	purchaseAnswerDataIntArray      []int
	purchaseAnswerDataStringArray   []string
	purchaseAnswerDataJsonArray     []util.FakePurchase
	purchaseAnswerDataJsonVerified  []util.FakePurchaseVerified
	purchaseAnswerDataJsonLowercase []map[string]any
	purchaseAnswerDataGrouped       util.FakePurchaseGrouped
	lotteryQuestionData             transforms.PureJsonArrayLottery
	lotteryAnswerDataIntArray       []int
	lotteryAnswerDataInt            int
	lotteryAnswerDataJson           util.FakeLotteryPick
	lotteryAnswerDataSmallerJson    []util.FakeLotteryPick
	lotteryAnswerFreqDist           map[string]int
	gradesQuestionData              util.ComplexGradesObject
	gradesAnswerDataInt             int
	gradesAnswerDataJson            []util.SimplerStudent
	gradesAnswerDataRidicJson       util.Student
	tagsQuestionDataArray           []util.Tag
	tagsQuestionDataDict            map[string]string
	tagsAnswerDictData              map[string]string
	tagsAnswerArrayData             []util.Tag
	prompt                          string
)

func generateLotteryPickQuestionData() transforms.PureJsonArrayLottery {
	return transforms.PureJsonArrayLottery{"lotteryPicks": util.GenerateLotteryPicks()}
}

func generatePurchaseQuestionData() transforms.PureJsonArrayPurchases {
	return transforms.PureJsonArrayPurchases{"purchases": util.GeneratePurchaseList()}
}

// this has a lot of misdirection and complexity, so it should be simplified
func generateGradesQuestionData() util.ComplexGradesObject {
	return util.GenerateComplexGradesObject()
}

func generateTagsQuestionDataArray() []util.Tag {
	rand.Seed(time.Now().UnixNano())

	// we always want at least 2 tags, and no more than 6
	numberOfTags := rand.Intn(4) + 2

	// big fan of this pre-allocation instead of append
	tagsArray := make([]util.Tag, numberOfTags)

	for i := 0; i < numberOfTags; i++ {
		tagsArray[i].Label = util.GetSingleRandomValueFromArray[string](util.PossibleLabels)
		tagsArray[i].Value = util.GetSingleRandomValueFromArray[string](util.PossibleValues)
	}

	return tagsArray
}

// we need two of these tags question generators because they don't both start from the
// same structure
func generateTagsQuestionDataDict() map[string]string {
	rand.Seed(time.Now().UnixNano())

	// we always want at least 2 tags, and no more than 6
	numberOfTags := rand.Intn(4) + 2

	tagsDict := make(map[string]string)

	for i := 0; i < numberOfTags; i++ {
		tagsDict[util.GetSingleRandomValueFromArray[string](util.PossibleLabels)] = util.GetSingleRandomValueFromArray[string](util.PossibleValues)
	}

	return tagsDict
}

func generatePersonQuestionData() transforms.PureJson {
	realActivities := util.PickActivities()
	realFavoriteColors := util.PickFavoriteColors()

	// these hacks are beyond delicious...they're needed for the comparison later with
	// the data structure returned from the user, which has a bunch of []interface{}
	favoriteColorsInterface := make([]interface{}, len(realFavoriteColors))

	activitiesInterface := make(map[string]interface{}, len(realActivities))

	for k := range realActivities {
		activitiesInterface[k] = realActivities[k]
	}

	for i := range realFavoriteColors {
		favoriteColorsInterface[i] = realFavoriteColors[i]
	}

	// the coercion here (and above) into floats is super annoying, but all integers are
	// floats by default when go sees json, so here we are!
	return transforms.PureJson{
		"age":            util.GeneratePossibleValue(util.PossibleAges),
		"id":             util.GeneratePossibleValue(util.PossibleIDs),
		"name":           util.GeneratePossibleValue(util.PossibleNames),
		"location":       util.GeneratePossibleValue(util.PossibleLocations),
		"favoriteColors": favoriteColorsInterface,
		"activities":     activitiesInterface,
	}
}

func generatePeopleQuestion() (interface{}, error) {
	var mixedResponse interface{}
	if currentFunctionType == jsonToJson {
		mixedResponse = JsonQuestion[map[string]interface{}, map[string]interface{}]{
			Question: personQuestionData,
			Answer:   personAnswerDataJson,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToString {
		mixedResponse = JsonQuestion[map[string]interface{}, string]{
			Question: personQuestionData,
			Answer:   personAnswerDataString,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToInt {
		mixedResponse = JsonQuestion[map[string]interface{}, int]{
			Question: personQuestionData,
			Answer:   personAnswerDataInt,
			Prompt:   prompt,
		}
	} else {
		return nil, errors.New("couldn't match function type for people question")
	}

	return mixedResponse, nil
}

func generatePurchaseQuestion() (interface{}, error) {
	var mixedResponse interface{}
	if currentFunctionType == jsonToIntArray {
		mixedResponse = JsonQuestion[map[string][]util.FakePurchase, []int]{
			Question: purchaseQuestionData,
			Answer:   purchaseAnswerDataIntArray,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToStringArray {
		mixedResponse = JsonQuestion[map[string][]util.FakePurchase, []string]{
			Question: purchaseQuestionData,
			Answer:   purchaseAnswerDataStringArray,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToJson {
		mixedResponse = JsonQuestion[map[string][]util.FakePurchase, []util.FakePurchase]{
			Question: purchaseQuestionData,
			Answer:   purchaseAnswerDataJsonArray,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToElaboratedJson {
		mixedResponse = JsonQuestion[map[string][]util.FakePurchase, []util.FakePurchaseVerified]{
			Question: purchaseQuestionData,
			Answer:   purchaseAnswerDataJsonVerified,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToLowercaseJson {
		mixedResponse = JsonQuestion[map[string][]util.FakePurchase, []map[string]any]{
			Question: purchaseQuestionData,
			Answer:   purchaseAnswerDataJsonLowercase,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToGroupedJson {
		mixedResponse = JsonQuestion[map[string][]util.FakePurchase, util.FakePurchaseGrouped]{
			Question: purchaseQuestionData,
			Answer:   purchaseAnswerDataGrouped,
			Prompt:   prompt,
		}
	} else {
		return nil, errors.New("couldn't match function type for purchase question")
	}

	return mixedResponse, nil
}

func generateTagsQuestionArray() (interface{}, error) {
	var mixedResponse interface{}
	if currentFunctionType == jsonToDict {
		mixedResponse = JsonQuestion[[]util.Tag, map[string]string]{
			Question: tagsQuestionDataArray,
			Answer:   tagsAnswerDictData,
			Prompt:   prompt,
		}
	} else {
		log.Fatal("didn't match the function type")
	}

	return mixedResponse, nil
}

func generateTagsQuestionDict() (interface{}, error) {
	var mixedResponse interface{}
	if currentFunctionType == jsonToArray {
		mixedResponse = JsonQuestion[map[string]string, []util.Tag]{
			Question: tagsQuestionDataDict,
			Answer:   tagsAnswerArrayData,
			Prompt:   prompt,
		}
	} else {
		log.Fatal("function type match go boom")
	}

	return mixedResponse, nil
}

func generateGradesQuestion() (interface{}, error) {
	var mixedResponse interface{}
	if currentFunctionType == jsonToInt {
		mixedResponse = JsonQuestion[util.ComplexGradesObject, int]{
			Question: gradesQuestionData,
			Answer:   gradesAnswerDataInt,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToRidicJson {
		mixedResponse = JsonQuestion[util.ComplexGradesObject, util.Student]{
			Question: gradesQuestionData,
			Answer:   gradesAnswerDataRidicJson,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToJson {
		mixedResponse = JsonQuestion[util.ComplexGradesObject, []util.SimplerStudent]{
			Question: gradesQuestionData,
			Answer:   gradesAnswerDataJson,
			Prompt:   prompt,
		}
	} else {
		return nil, errors.New("couldn't match function type for grades question")
	}

	return mixedResponse, nil
}

func generateLotteryQuestion() (interface{}, error) {
	var mixedResponse interface{}
	if currentFunctionType == jsonToJson {
		mixedResponse = JsonQuestion[map[string][]util.FakeLotteryPick, util.FakeLotteryPick]{
			Question: lotteryQuestionData,
			Answer:   lotteryAnswerDataJson,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToSmallerJson {
		mixedResponse = JsonQuestion[map[string][]util.FakeLotteryPick, []util.FakeLotteryPick]{
			Question: lotteryQuestionData,
			Answer:   lotteryAnswerDataSmallerJson,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToDict {
		mixedResponse = JsonQuestion[map[string][]util.FakeLotteryPick, map[string]int]{
			Question: lotteryQuestionData,
			Answer:   lotteryAnswerFreqDist,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToIntArray {
		mixedResponse = JsonQuestion[map[string][]util.FakeLotteryPick, []int]{
			Question: lotteryQuestionData,
			Answer:   lotteryAnswerDataIntArray,
			Prompt:   prompt,
		}
	} else if currentFunctionType == jsonToInt {
		mixedResponse = JsonQuestion[map[string][]util.FakeLotteryPick, int]{
			Question: lotteryQuestionData,
			Answer:   lotteryAnswerDataInt,
			Prompt:   prompt,
		}
	} else {
		return nil, errors.New("couldn't match function type for lottery question")
	}

	return mixedResponse, nil
}

func main() {
	generateNextQuestionAnswer()

	flag.String("MODE", "dev", "whether we're running in dev or production mode")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	MODE := viper.GetString("MODE")

	viper.SetDefault("PORT", "8000")

	if MODE == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8000"}
    config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
    config.ExposeHeaders = []string{"Content-Length"}
    config.AllowCredentials = true
    router.Use(cors.New(config))

	if MODE == "prod" {
		router.Use(static.Serve("/", static.LocalFile("./build", true)))
	}

    router.OPTIONS("/sse", func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type")
        c.Header("Access-Control-Max-Age", "86400")
        c.Status(http.StatusOK)
    })

	router.GET("/sse", handleSSE)

	router.GET("/question", getQuestion)
	router.POST("/answer", getAnswer)
	router.GET("/prompt", getPrompt)

	router.SetTrustedProxies(nil)
	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(http.StatusNotFound, gin.H{}) })

	port := viper.GetString("PORT")
	router.Run(":" + port)
}

var (
    clientsMutex sync.Mutex
    clients      []chan struct{}
)

func registerClient(ch chan struct{}) {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()
    clients = append(clients, ch)
}

func unregisterClient(ch chan struct{}) {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()
    for i, c := range clients {
        if c == ch {
            clients = append(clients[:i], clients[i+1:]...)
            break
        }
    }
}

func notifyStateChange() {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()
    for _, ch := range clients {
        select {
        case ch <- struct{}{}:
        default:
            // If the channel is full, we skip this client
        }
    }
}

func handleSSE(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Headers", "Content-Type")
    c.Header("Access-Control-Allow-Credentials", "true")

    ctx, cancel := context.WithCancel(c.Request.Context())
    defer cancel()

    clientChan := make(chan string)
    defer close(clientChan)

    // Create a channel to receive state updates
    stateChan := make(chan struct{}, 1)
    defer close(stateChan)

    // Register this client's channel to receive state updates
    registerClient(stateChan)
    defer unregisterClient(stateChan)

    // Send initial state
    go func() {
        mixedResponse, err := generateMixedResponse()
        if err != nil {
            log.Println("Error generating initial mixed response:", err)
            return
        }

        response, err := json.Marshal(mixedResponse)
        if err != nil {
            log.Println("Error marshalling initial JSON:", err)
            return
        }

        select {
        case clientChan <- string(response):
        case <-ctx.Done():
            return
        }
    }()

    go func() {
        defer cancel()
        for {
            select {
            case <-ctx.Done():
                return
            case <-stateChan:
                // State has changed, generate and send new response
                mixedResponse, err := generateMixedResponse()
                if err != nil {
                    log.Println("Error generating mixed response:", err)
                    continue
                }

                response, err := json.Marshal(mixedResponse)
                if err != nil {
                    log.Println("Error marshalling JSON:", err)
                    continue
                }

                select {
                case clientChan <- string(response):
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    c.Stream(func(w io.Writer) bool {
        select {
        case msg := <-clientChan:
            c.SSEvent("message", msg)
            return true
        case <-ctx.Done():
            return false
        }
    })
}

func generateMixedResponse() (interface{}, error) {
	var mixedResponse interface{}
	var err error

	switch currentQuestionType {
	case util.SimplePeopleQuestions:
		mixedResponse, err = generatePeopleQuestion()
	case util.SimplePurchaseQuestions:
		mixedResponse, err = generatePurchaseQuestion()
	case util.SimpleLotteryQuestions:
		mixedResponse, err = generateLotteryQuestion()
	case util.SimpleGradesQuestions:
		mixedResponse, err = generateGradesQuestion()
	case util.SimpleTagsQuestionsArrayToDict:
		mixedResponse, err = generateTagsQuestionArray()
	case util.SimpleTagsQuestionsDictToArray:
		mixedResponse, err = generateTagsQuestionDict()
	default:
		return nil, errors.New("couldn't get question type")
	}

	if err != nil {
		return nil, err
	}

	return mixedResponse, nil
}

// this is crap and should be more general
func getQuestion(c *gin.Context) {
	switch currentQuestionType {
	case util.SimplePeopleQuestions:
		c.IndentedJSON(http.StatusOK, personQuestionData)
	case util.SimplePurchaseQuestions:
		c.IndentedJSON(http.StatusOK, purchaseQuestionData)
	case util.SimpleLotteryQuestions:
		c.IndentedJSON(http.StatusOK, lotteryQuestionData)
	case util.SimpleGradesQuestions:
		c.IndentedJSON(http.StatusOK, gradesQuestionData)
	case util.SimpleTagsQuestionsArrayToDict:
		c.IndentedJSON(http.StatusOK, tagsQuestionDataArray)
	case util.SimpleTagsQuestionsDictToArray:
		c.IndentedJSON(http.StatusOK, tagsQuestionDataDict)
	}
}

func generateNextQuestionAnswer() {
	// we need to know what type of question we want so that we can use that to determine the subset
	// of function types to use to create the activity
	currentQuestionType = util.GeneratePossibleValue(util.PossibleQuestionTypes)

	// there's probably a better way to structure this hierarchy, but we'll just go with
	// something dumb and verbose for now
	if currentQuestionType == util.SimplePeopleQuestions {
		currentFunctionType = util.GeneratePossibleValue(peopleFunctionTypesToCall)
	} else if currentQuestionType == util.SimplePurchaseQuestions {
		currentFunctionType = util.GeneratePossibleValue(purchaseFunctionTypesToCall)
	} else if currentQuestionType == util.SimpleLotteryQuestions {
		currentFunctionType = util.GeneratePossibleValue(lotteryFunctionTypesToCall)
	} else if currentQuestionType == util.SimpleGradesQuestions {
		currentFunctionType = util.GeneratePossibleValue(gradesFunctionTypesToCall)
	} else if currentQuestionType == util.SimpleTagsQuestionsArrayToDict {
		currentFunctionType = util.GeneratePossibleValue(tagsArrayFunctionTypesToCall)
	} else if currentQuestionType == util.SimpleTagsQuestionsDictToArray {
		currentFunctionType = util.GeneratePossibleValue(tagsDictFunctionTypesToCall)
	} else {
		log.Fatal("this blew up because we couldn't determine the currentFunctionType")
	}

	log.Println("the function type is:", currentFunctionType)

	switch currentQuestionType {
	case util.SimpleLotteryQuestions:
		lotteryQuestionData = generateLotteryPickQuestionData()

		switch currentFunctionType {
		case jsonToJson:
			lotteryAnswerDataJson, prompt = transforms.PickAWinner(lotteryQuestionData)
		case jsonToSmallerJson:
			lotteryAnswerDataSmallerJson, prompt = transforms.GetNRangePicks(lotteryQuestionData)
		case jsonToDict:
			lotteryAnswerFreqDist, prompt = transforms.GetLotteryPickFrequencyDistribution(lotteryQuestionData)
		case jsonToIntArray:
			lotteryAnswerDataIntArray, prompt = transforms.GetAllUniqueArrayIntValues(lotteryQuestionData)
		case jsonToInt:
			lotteryAnswerDataInt, prompt = transforms.GetNumberOfPicks(lotteryQuestionData)
		}

	case util.SimplePeopleQuestions:
		personQuestionData = generatePersonQuestionData()
		// this should really by something like nextFunctionType, but we can refactor later
		switch currentFunctionType {
		// this is where the Simple Person Question exercises are generated
		case jsonToInt:
			personAnswerDataInt, prompt = transforms.GetOneKeyIntValue(personQuestionData)
		case jsonToString:
			personAnswerDataString, prompt = transforms.GetOneKeyStringValue(personQuestionData)
		case jsonToJson:
			var jsonToJsonFunction func(transforms.PureJson) (transforms.PureJson, string)

			functionToCall := rand.Intn(totalJsonToJsonFunctions)

			switch functionToCall {
			case 0:
				jsonToJsonFunction = transforms.DeleteRandomKeys
			case 1:
				jsonToJsonFunction = transforms.DeleteOneKey
			case 2:
				jsonToJsonFunction = transforms.KeepOneKey
			default:
				log.Fatal("this blew up because we couldn't match the functionToCall here")
			}

			personAnswerDataJson, prompt = jsonToJsonFunction(personQuestionData)
		default:
			log.Fatal("blew the F up!")
		}
	case util.SimplePurchaseQuestions:
		purchaseQuestionData = generatePurchaseQuestionData()

		switch currentFunctionType {
		// this is where the Simple Purchase Question exercises are generated
		case jsonToIntArray:
			purchaseAnswerDataIntArray, prompt = transforms.GetAllArrayIntValues(purchaseQuestionData)
		case jsonToStringArray:
			purchaseAnswerDataStringArray, prompt = transforms.GetAllArrayStringValues(purchaseQuestionData)
		case jsonToJson:
			purchaseAnswerDataJsonArray, prompt = transforms.GetFilteredByPurchasePrice(purchaseQuestionData)
		case jsonToElaboratedJson:
			purchaseAnswerDataJsonVerified, prompt = transforms.AddVerifiedToEachPurchase(purchaseQuestionData)
		case jsonToLowercaseJson:
			purchaseAnswerDataJsonLowercase, prompt = transforms.MakeAllFieldsLowercase(purchaseQuestionData)
		case jsonToGroupedJson:
			purchaseAnswerDataGrouped, prompt = transforms.GetGroupByPurchasePrice(purchaseQuestionData)
		}
	case util.SimpleGradesQuestions:
		gradesQuestionData = generateGradesQuestionData()

		switch currentFunctionType {
		case jsonToJson:
			gradesAnswerDataJson, prompt = transforms.GetHighestScoreForEachSubject(gradesQuestionData)
		case jsonToRidicJson:
			gradesAnswerDataRidicJson, prompt = transforms.GetHighestResultInOneSubject(gradesQuestionData)
		case jsonToInt:
			gradesAnswerDataInt, prompt = transforms.GetHighestScoreForPersonInSubject(gradesQuestionData)
		default:
			log.Println("fell into the default question type...for...reasons...")
		}
	case util.SimpleTagsQuestionsArrayToDict:

		switch currentFunctionType {
		case jsonToDict:
			tagsQuestionDataArray = generateTagsQuestionDataArray()
			tagsAnswerDictData, prompt = transforms.GetDictFromArray(tagsQuestionDataArray)
		}
	case util.SimpleTagsQuestionsDictToArray:
		switch currentFunctionType {
		case jsonToArray:
			tagsQuestionDataDict = generateTagsQuestionDataDict()
			tagsAnswerArrayData, prompt = transforms.GetArrayFromDict(tagsQuestionDataDict)
		}
	}
	notifyStateChange()
}

func processAnswer[T any](context *gin.Context, expectedAnswer T) {
	var actualAnswer T
	if err := context.BindJSON(&actualAnswer); err != nil {
		log.Println(err)
		context.AbortWithStatus(http.StatusBadRequest)
	}

	diff := deep.Equal(actualAnswer, expectedAnswer)

	if diff == nil {
		log.Println("correct")
		generateNextQuestionAnswer()
	} else {
		log.Println("wrong answer, please try again")
		context.JSON(http.StatusBadRequest, gin.H{"message": "wrong answer, please try again"})
	}
}

func getPrompt(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, prompt)
}

func getAnswer(c *gin.Context) {
	if currentQuestionType == util.SimplePurchaseQuestions {
		// we keep the current function type in state so we know how to compare the answer
		if currentFunctionType == jsonToStringArray {
			processAnswer[[]string](c, purchaseAnswerDataStringArray)
		} else if currentFunctionType == jsonToIntArray {
			processAnswer[[]int](c, purchaseAnswerDataIntArray)
		} else if currentFunctionType == jsonToJson {
			processAnswer[[]util.FakePurchase](c, purchaseAnswerDataJsonArray)
		} else if currentFunctionType == jsonToElaboratedJson {
			processAnswer[[]util.FakePurchaseVerified](c, purchaseAnswerDataJsonVerified)
		} else if currentFunctionType == jsonToLowercaseJson {
			processAnswer[[]map[string]any](c, purchaseAnswerDataJsonLowercase)
		} else if currentFunctionType == jsonToGroupedJson {
			processAnswer[util.FakePurchaseGrouped](c, purchaseAnswerDataGrouped)
		}
	} else if currentQuestionType == util.SimplePeopleQuestions {
		if currentFunctionType == jsonToString {
			processAnswer[string](c, personAnswerDataString)
		} else if currentFunctionType == jsonToInt {
			processAnswer[int](c, personAnswerDataInt)
		} else if currentFunctionType == jsonToJson {
			processAnswer[transforms.PureJson](c, personAnswerDataJson)
		}
	} else if currentQuestionType == util.SimpleLotteryQuestions {
		if currentFunctionType == jsonToJson {
			processAnswer[util.FakeLotteryPick](c, lotteryAnswerDataJson)
		} else if currentFunctionType == jsonToSmallerJson {
			processAnswer[[]util.FakeLotteryPick](c, lotteryAnswerDataSmallerJson)
		} else if currentFunctionType == jsonToDict {
			processAnswer[map[string]int](c, lotteryAnswerFreqDist)
		} else if currentFunctionType == jsonToIntArray {
			processAnswer[[]int](c, lotteryAnswerDataIntArray)
		} else if currentFunctionType == jsonToInt {
			processAnswer[int](c, lotteryAnswerDataInt)
		}
	} else if currentQuestionType == util.SimpleGradesQuestions {
		if currentFunctionType == jsonToInt {
			processAnswer[int](c, gradesAnswerDataInt)
		} else if currentFunctionType == jsonToJson {
			processAnswer[[]util.SimplerStudent](c, gradesAnswerDataJson)
		} else if currentFunctionType == jsonToRidicJson {
			processAnswer[util.Student](c, gradesAnswerDataRidicJson)
		}
	} else if currentQuestionType == util.SimpleTagsQuestionsArrayToDict {
		if currentFunctionType == jsonToDict {
			processAnswer[map[string]string](c, tagsAnswerDictData)
		} else {
			log.Fatal("didn't match a function type")
		}
	} else if currentQuestionType == util.SimpleTagsQuestionsDictToArray {
		if currentFunctionType == jsonToArray {
			processAnswer[[]util.Tag](c, tagsAnswerArrayData)
		} else {
			log.Fatal("didn't match a function type here either")
		}
	} else {
		log.Println("No current question type...sad day")
	}
}
