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
	"os"
	"reflect"
	"sync"
	"time"

	"jq-pilot/transforms"
	"jq-pilot/util"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// jsonQuestion is the payload shape returned to the client for every exercise.
// The concrete Go types of the question and answer vary per exercise, but they
// are only ever marshaled to JSON, so plain any fields serialize identically to
// the old per-type generic struct while letting a single struct serve them all.
type jsonQuestion struct {
	Question any    `json:"question"`
	Answer   any    `json:"answer"`
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
	totalJsonToJsonFunctions = 3
	currentQuestionType      = util.SimplePeopleQuestions
	currentFunctionType      string

	// currentQuestionData / currentAnswerData hold the single live exercise: the
	// data shown to the player and its expected answer, each boxed in an any.
	// generateNextQuestionAnswer builds both from the same freshly generated data,
	// so they always agree on the (question type, function type) in effect. The
	// response path marshals them as-is; getAnswer type-asserts the answer back to
	// its concrete type to compare against the player's submission.
	currentQuestionData any
	currentAnswerData   any
	prompt              string
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

func main() {
	generateNextQuestionAnswer()

	modeFlag := flag.String("MODE", "dev", "whether we're running in dev or production mode")
	flag.Parse()

	MODE := *modeFlag

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

	// PORT is injected by the platform (e.g. Cloud Run sets it to the configured
	// container port); fall back to 8000 for local runs and the Docker image.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

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
	// Read the live question, answer and prompt as one consistent snapshot. The
	// returned struct only holds header copies of the (immutable) published state,
	// so the caller can marshal it after the lock is released.
	stateMu.RLock()
	defer stateMu.RUnlock()

	if currentAnswerData == nil {
		return nil, errors.New("no question has been generated yet")
	}

	return jsonQuestion{
		Question: currentQuestionData,
		Answer:   currentAnswerData,
		Prompt:   prompt,
	}, nil
}

func getQuestion(c *gin.Context) {
	// Snapshot the live question data, then release the lock before serializing and
	// writing to the (possibly slow) client.
	data := snapshot(&currentQuestionData)
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

	// pick the subset of function types available for the chosen question type
	switch currentQuestionType {
	case util.SimplePeopleQuestions:
		currentFunctionType = util.GeneratePossibleValue(peopleFunctionTypesToCall)
	case util.SimplePurchaseQuestions:
		currentFunctionType = util.GeneratePossibleValue(purchaseFunctionTypesToCall)
	case util.SimpleLotteryQuestions:
		currentFunctionType = util.GeneratePossibleValue(lotteryFunctionTypesToCall)
	case util.SimpleGradesQuestions:
		currentFunctionType = util.GeneratePossibleValue(gradesFunctionTypesToCall)
	case util.SimpleTagsQuestionsArrayToDict:
		currentFunctionType = util.GeneratePossibleValue(tagsArrayFunctionTypesToCall)
	case util.SimpleTagsQuestionsDictToArray:
		currentFunctionType = util.GeneratePossibleValue(tagsDictFunctionTypesToCall)
	default:
		log.Fatal("this blew up because we couldn't determine the currentFunctionType")
	}

	log.Println("the function type is:", currentFunctionType)

	// Each case generates fresh question data (kept in a local so the transforms can
	// take its concrete type) and publishes it, along with the expected answer, into
	// the currentQuestionData / currentAnswerData boxes read by the response and
	// answer-checking paths.
	switch currentQuestionType {
	case util.SimpleLotteryQuestions:
		questionData := generateLotteryPickQuestionData()
		currentQuestionData = questionData

		switch currentFunctionType {
		case jsonToJson:
			currentAnswerData, prompt = transforms.PickAWinner(questionData)
		case jsonToSmallerJson:
			currentAnswerData, prompt = transforms.GetNRangePicks(questionData)
		case jsonToDict:
			currentAnswerData, prompt = transforms.GetLotteryPickFrequencyDistribution(questionData)
		case jsonToIntArray:
			currentAnswerData, prompt = transforms.GetAllUniqueArrayIntValues(questionData)
		case jsonToInt:
			currentAnswerData, prompt = transforms.GetNumberOfPicks(questionData)
		}

	case util.SimplePeopleQuestions:
		questionData := generatePersonQuestionData()
		currentQuestionData = questionData
		// this should really by something like nextFunctionType, but we can refactor later
		switch currentFunctionType {
		// this is where the Simple Person Question exercises are generated
		case jsonToInt:
			currentAnswerData, prompt = transforms.GetOneKeyIntValue(questionData)
		case jsonToString:
			currentAnswerData, prompt = transforms.GetOneKeyStringValue(questionData)
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

			currentAnswerData, prompt = jsonToJsonFunction(questionData)
		default:
			log.Fatal("blew the F up!")
		}
	case util.SimplePurchaseQuestions:
		questionData := generatePurchaseQuestionData()
		currentQuestionData = questionData

		switch currentFunctionType {
		// this is where the Simple Purchase Question exercises are generated
		case jsonToIntArray:
			currentAnswerData, prompt = transforms.GetAllArrayIntValues(questionData)
		case jsonToStringArray:
			currentAnswerData, prompt = transforms.GetAllArrayStringValues(questionData)
		case jsonToJson:
			currentAnswerData, prompt = transforms.GetFilteredByPurchasePrice(questionData)
		case jsonToElaboratedJson:
			currentAnswerData, prompt = transforms.AddVerifiedToEachPurchase(questionData)
		case jsonToLowercaseJson:
			currentAnswerData, prompt = transforms.MakeAllFieldsLowercase(questionData)
		case jsonToGroupedJson:
			currentAnswerData, prompt = transforms.GetGroupByPurchasePrice(questionData)
		}
	case util.SimpleGradesQuestions:
		questionData := generateGradesQuestionData()
		currentQuestionData = questionData

		switch currentFunctionType {
		case jsonToJson:
			currentAnswerData, prompt = transforms.GetHighestScoreForEachSubject(questionData)
		case jsonToRidicJson:
			currentAnswerData, prompt = transforms.GetHighestResultInOneSubject(questionData)
		case jsonToInt:
			currentAnswerData, prompt = transforms.GetHighestScoreForPersonInSubject(questionData)
		default:
			log.Println("fell into the default question type...for...reasons...")
		}
	case util.SimpleTagsQuestionsArrayToDict:
		switch currentFunctionType {
		case jsonToDict:
			questionData := generateTagsQuestionDataArray()
			currentQuestionData = questionData
			currentAnswerData, prompt = transforms.GetDictFromArray(questionData)
		}
	case util.SimpleTagsQuestionsDictToArray:
		switch currentFunctionType {
		case jsonToArray:
			questionData := generateTagsQuestionDataDict()
			currentQuestionData = questionData
			currentAnswerData, prompt = transforms.GetArrayFromDict(questionData)
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

	if reflect.DeepEqual(actualAnswer, expectedAnswer) {
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
	// Snapshot the selectors and the expected answer together under one read lock so
	// they describe the same exercise, then release it before processAnswer reads the
	// request body or calls generateNextQuestionAnswer (which needs the write lock).
	// Because the triple is consistent, the concrete type asserted below always
	// matches what generateNextQuestionAnswer stored for this exercise.
	stateMu.RLock()
	questionType := currentQuestionType
	functionType := currentFunctionType
	answer := currentAnswerData
	stateMu.RUnlock()

	switch questionType {
	case util.SimplePurchaseQuestions:
		// we keep the current function type in state so we know how to compare the answer
		switch functionType {
		case jsonToStringArray:
			processAnswer(c, answer.([]string))
		case jsonToIntArray:
			processAnswer(c, answer.([]int))
		case jsonToJson:
			processAnswer(c, answer.([]util.FakePurchase))
		case jsonToElaboratedJson:
			processAnswer(c, answer.([]util.FakePurchaseVerified))
		case jsonToLowercaseJson:
			processAnswer(c, answer.([]map[string]any))
		case jsonToGroupedJson:
			processAnswer(c, answer.(util.FakePurchaseGrouped))
		}
	case util.SimplePeopleQuestions:
		switch functionType {
		case jsonToString:
			processAnswer(c, answer.(string))
		case jsonToInt:
			processAnswer(c, answer.(int))
		case jsonToJson:
			processAnswer(c, answer.(transforms.PureJson))
		}
	case util.SimpleLotteryQuestions:
		switch functionType {
		case jsonToJson:
			processAnswer(c, answer.(util.FakeLotteryPick))
		case jsonToSmallerJson:
			processAnswer(c, answer.([]util.FakeLotteryPick))
		case jsonToDict:
			processAnswer(c, answer.(map[string]int))
		case jsonToIntArray:
			processAnswer(c, answer.([]int))
		case jsonToInt:
			processAnswer(c, answer.(int))
		}
	case util.SimpleGradesQuestions:
		switch functionType {
		case jsonToInt:
			processAnswer(c, answer.(int))
		case jsonToJson:
			processAnswer(c, answer.([]util.SimplerStudent))
		case jsonToRidicJson:
			processAnswer(c, answer.(util.Student))
		}
	case util.SimpleTagsQuestionsArrayToDict:
		if functionType == jsonToDict {
			processAnswer(c, answer.(map[string]string))
		} else {
			log.Fatal("didn't match a function type")
		}
	case util.SimpleTagsQuestionsDictToArray:
		if functionType == jsonToArray {
			processAnswer(c, answer.([]util.Tag))
		} else {
			log.Fatal("didn't match a function type here either")
		}
	default:
		log.Println("No current question type...sad day")
	}
}
