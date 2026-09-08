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
	"github.com/gin-contrib/gzip"
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

// stateMu guards all of the shared game state declared above: the
// currentQuestionType / currentFunctionType selectors, the prompt, and every
// question/answer global. This stays a single, world-readable shared game, so
// the state remains global on purpose; the mutex only makes the concurrent
// access safe. gin serves each request in its own goroutine, so without this
// two concurrent POSTs (or an SSE read racing a regeneration) would be a data
// race on these maps/slices/strings and could crash the process.
//
// Writers take the write lock (generateNextQuestionAnswer). Readers take the
// read lock, either directly or via snapshot below.
var stateMu sync.RWMutex

// snapshot returns a copy of a shared-state value read under the read lock, so a
// handler never touches a global concurrently with a writer. Maps and slices are
// copied by header, which is safe because published state is never mutated in
// place: generateNextQuestionAnswer only ever reassigns these globals to freshly
// built values, so a header captured here always points at an immutable snapshot.
func snapshot[T any](p *T) T {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return *p
}

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

    // Add gzip as fallback, excluding SSE endpoint
    router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/sse"})))

    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8000"}
    config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
    config.ExposeHeaders = []string{"Content-Length"}
    config.AllowCredentials = true
    router.Use(cors.New(config))

	if MODE == "prod" {
		router.Use(static.Serve("/", static.LocalFile("./build", true)))
	} else {
		// Add skip endpoint only in development mode
                router.POST("/skip", handleSkip)
	}

	router.GET("/sse", handleSSE)

	router.GET("/question", getQuestion)
	router.POST("/answer", getAnswer)
	router.GET("/prompt", getPrompt)

	router.SetTrustedProxies(nil)
	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(http.StatusNotFound, gin.H{}) })

	port := viper.GetString("PORT")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
		// ReadHeaderTimeout guards against Slowloris-style slow-header attacks.
		ReadHeaderTimeout: 10 * time.Second,
		// IdleTimeout reaps idle keep-alive connections.
		IdleTimeout: 120 * time.Second,
		// ReadTimeout/WriteTimeout are intentionally left unset: /sse streams
		// long-lived responses and a write deadline would sever those connections.
	}

	log.Fatal(srv.ListenAndServe())
}

var (
    clientsMutex sync.RWMutex
    clients      = make(map[chan struct{}]bool)
)

func registerClient(ch chan struct{}) {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()
    clients[ch] = true
}

func unregisterClient(ch chan struct{}) {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()
    delete(clients, ch)
    close(ch)
}

// SSE connection limiting: /sse holds long-lived streaming connections, and on
// Cloud Run each open connection is billed as an in-flight request for its whole
// duration. Capping simultaneous connections stops a flood of open connections
// from pinning the instance (driving up cost) or exhausting request concurrency.
// Keep this below containerConcurrency (80) so other routes keep headroom.
const maxConcurrentSSEConnections = 50

var (
    sseConnMutex sync.Mutex
    sseConnCount int
)

// acquireSSESlot reserves a connection slot, returning false if we're at the cap.
func acquireSSESlot() bool {
    sseConnMutex.Lock()
    defer sseConnMutex.Unlock()
    if sseConnCount >= maxConcurrentSSEConnections {
        return false
    }
    sseConnCount++
    return true
}

func releaseSSESlot() {
    sseConnMutex.Lock()
    defer sseConnMutex.Unlock()
    sseConnCount--
}

func handleSkip(c *gin.Context) {
    generateNextQuestionAnswer()
    c.Status(http.StatusOK)
}

func notifyStateChange() {
    clientsMutex.RLock()
    defer clientsMutex.RUnlock()

    for ch := range clients {
        select {
        case ch <- struct{}{}:
        default:
            // Skip if channel is blocked
        }
    }
}

func handleSSE(c *gin.Context) {
    // Reject the connection if we're already at the concurrent-connection cap.
    // Each open SSE connection pins billable instance time on Cloud Run, so this
    // bounds both cost and request-concurrency exhaustion.
    if !acquireSSESlot() {
        c.AbortWithStatus(http.StatusTooManyRequests)
        return
    }
    defer releaseSSESlot()

    // Set required SSE headers.
    // CORS is deliberately NOT set here: the global cors middleware already emits a
    // valid, per-origin Access-Control-Allow-Origin (with credentials) for the
    // allowlisted origins. Emitting "Access-Control-Allow-Origin: *" together with
    // credentials here produced an invalid combination that browsers reject.
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no") // Disable buffering for nginx

    // Create a context that's cancelled when the client disconnects
    ctx, cancel := context.WithCancel(c.Request.Context())
    defer cancel()

    // Create a channel for this client
    clientChan := make(chan string, 1)
    defer close(clientChan)

    // Create a ticker for keepalive messages so we can run on cloudrun with
	// CPU allocated only during request processing
    keepAliveTicker := time.NewTicker(30 * time.Second)
    defer keepAliveTicker.Stop()

    // Send initial state immediately
    go func() {
        mixedResponse, err := generateMixedResponse()
        if err != nil {
            log.Printf("Error generating initial mixed response: %v", err)
            return
        }

        response, err := json.Marshal(mixedResponse)
        if err != nil {
            log.Printf("Error marshalling initial JSON: %v", err)
            return
        }

        select {
        case clientChan <- string(response):
        case <-ctx.Done():
            return
        }
    }()

    // Create a channel for state updates
    stateChan := make(chan struct{}, 1)
    registerClient(stateChan)
    defer unregisterClient(stateChan)

    // Start streaming
    c.Stream(func(w io.Writer) bool {
        select {
        case <-ctx.Done():
            return false
        case <-keepAliveTicker.C:
            // Send keepalive comment to prevent connection timeout
            c.SSEvent("keepalive", time.Now().Unix())
            return true
        case <-stateChan:
            // State has changed, generate and send new response
            mixedResponse, err := generateMixedResponse()
            if err != nil {
                log.Printf("Error generating mixed response: %v", err)
                return true // Continue streaming despite error
            }

            response, err := json.Marshal(mixedResponse)
            if err != nil {
                log.Printf("Error marshalling JSON: %v", err)
                return true // Continue streaming despite error
            }

            c.SSEvent("message", string(response))
            return true
        case msg := <-clientChan:
            c.SSEvent("message", msg)
            return true
        }
    })
}

func generateMixedResponse() (interface{}, error) {
	// Held for the whole assembly so the selectors and the data they point at are
	// read as one consistent snapshot. The returned value only holds header copies
	// of the (immutable) published state, so the caller can marshal it after the
	// lock is released.
	stateMu.RLock()
	defer stateMu.RUnlock()

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
	// Read the selector and its matching data as one consistent snapshot, then
	// release the lock before serializing/writing to the (possibly slow) client.
	stateMu.RLock()
	var data any
	switch currentQuestionType {
	case util.SimplePeopleQuestions:
		data = personQuestionData
	case util.SimplePurchaseQuestions:
		data = purchaseQuestionData
	case util.SimpleLotteryQuestions:
		data = lotteryQuestionData
	case util.SimpleGradesQuestions:
		data = gradesQuestionData
	case util.SimpleTagsQuestionsArrayToDict:
		data = tagsQuestionDataArray
	case util.SimpleTagsQuestionsDictToArray:
		data = tagsQuestionDataDict
	}
	stateMu.RUnlock()

	if data != nil {
		c.IndentedJSON(http.StatusOK, data)
	}
}

func generateNextQuestionAnswer() {
	stateMu.Lock()
	defer stateMu.Unlock()

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

// maxAnswerBodyBytes bounds the size of a POST /answer body so an oversized
// payload can't exhaust the instance's memory (only 128Mi).
const maxAnswerBodyBytes = 1 << 20 // 1 MiB

func processAnswer[T any](context *gin.Context, expectedAnswer T) {
	context.Request.Body = http.MaxBytesReader(context.Writer, context.Request.Body, maxAnswerBodyBytes)

	var actualAnswer T
	if err := context.BindJSON(&actualAnswer); err != nil {
		log.Println(err)
		context.AbortWithStatus(http.StatusBadRequest)
		return
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
	c.IndentedJSON(http.StatusOK, snapshot(&prompt))
}

func getAnswer(c *gin.Context) {
	// Read the two selectors as one consistent snapshot. Each expected answer is
	// then snapshotted individually (via snapshot below) so we never hold the read
	// lock while processAnswer reads the request body or calls
	// generateNextQuestionAnswer (which needs the write lock).
	stateMu.RLock()
	questionType := currentQuestionType
	functionType := currentFunctionType
	stateMu.RUnlock()

	if questionType == util.SimplePurchaseQuestions {
		// we keep the current function type in state so we know how to compare the answer
		if functionType == jsonToStringArray {
			processAnswer[[]string](c, snapshot(&purchaseAnswerDataStringArray))
		} else if functionType == jsonToIntArray {
			processAnswer[[]int](c, snapshot(&purchaseAnswerDataIntArray))
		} else if functionType == jsonToJson {
			processAnswer[[]util.FakePurchase](c, snapshot(&purchaseAnswerDataJsonArray))
		} else if functionType == jsonToElaboratedJson {
			processAnswer[[]util.FakePurchaseVerified](c, snapshot(&purchaseAnswerDataJsonVerified))
		} else if functionType == jsonToLowercaseJson {
			processAnswer[[]map[string]any](c, snapshot(&purchaseAnswerDataJsonLowercase))
		} else if functionType == jsonToGroupedJson {
			processAnswer[util.FakePurchaseGrouped](c, snapshot(&purchaseAnswerDataGrouped))
		}
	} else if questionType == util.SimplePeopleQuestions {
		if functionType == jsonToString {
			processAnswer[string](c, snapshot(&personAnswerDataString))
		} else if functionType == jsonToInt {
			processAnswer[int](c, snapshot(&personAnswerDataInt))
		} else if functionType == jsonToJson {
			processAnswer[transforms.PureJson](c, snapshot(&personAnswerDataJson))
		}
	} else if questionType == util.SimpleLotteryQuestions {
		if functionType == jsonToJson {
			processAnswer[util.FakeLotteryPick](c, snapshot(&lotteryAnswerDataJson))
		} else if functionType == jsonToSmallerJson {
			processAnswer[[]util.FakeLotteryPick](c, snapshot(&lotteryAnswerDataSmallerJson))
		} else if functionType == jsonToDict {
			processAnswer[map[string]int](c, snapshot(&lotteryAnswerFreqDist))
		} else if functionType == jsonToIntArray {
			processAnswer[[]int](c, snapshot(&lotteryAnswerDataIntArray))
		} else if functionType == jsonToInt {
			processAnswer[int](c, snapshot(&lotteryAnswerDataInt))
		}
	} else if questionType == util.SimpleGradesQuestions {
		if functionType == jsonToInt {
			processAnswer[int](c, snapshot(&gradesAnswerDataInt))
		} else if functionType == jsonToJson {
			processAnswer[[]util.SimplerStudent](c, snapshot(&gradesAnswerDataJson))
		} else if functionType == jsonToRidicJson {
			processAnswer[util.Student](c, snapshot(&gradesAnswerDataRidicJson))
		}
	} else if questionType == util.SimpleTagsQuestionsArrayToDict {
		if functionType == jsonToDict {
			processAnswer[map[string]string](c, snapshot(&tagsAnswerDictData))
		} else {
			log.Fatal("didn't match a function type")
		}
	} else if questionType == util.SimpleTagsQuestionsDictToArray {
		if functionType == jsonToArray {
			processAnswer[[]util.Tag](c, snapshot(&tagsAnswerArrayData))
		} else {
			log.Fatal("didn't match a function type here either")
		}
	} else {
		log.Println("No current question type...sad day")
	}
}
