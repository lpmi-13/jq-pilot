package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"jq-pilot/transforms"
	"jq-pilot/util"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-test/deep"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type JsonToJsonQuestion struct {
	Question map[string]interface{} `json:"question"`
	Answer   map[string]interface{} `json:"answer"`
}

type JsonToStringQuestion struct {
	Question map[string]interface{} `json:"question"`
	Answer   string                 `json:"answer"`
}

type JsonToIntQuestion struct {
	Question map[string]interface{} `json:"question"`
	Answer   int                    `json:"answer"`
}

// it's at this point that I feel like I want some generics, so I'll get to implementing
// that at some point (https://itnext.io/how-to-use-golang-generics-with-structs-8cabc9353d75)
type JsonToIntArrayQuestion struct {
	Question map[string][]util.FakePurchase `json:"question"`
	Answer   []int                          `json:"answer"`
}

type JsonToStringArrayQuestion struct {
	Question map[string][]util.FakePurchase `json:"question"`
	Answer   []string                       `json:"answer"`
}

const (
	jsonToJson        = "jsonToJson"
	jsonToString      = "jsonToString"
	jsonToInt         = "jsonToInt"
	jsonToStringArray = "jsonToStringArray"
	jsonToIntArray    = "jsonToIntArray"
)

var (
	peopleFunctionTypesToCall = []string{
		jsonToJson, jsonToString, jsonToInt,
	}
	purchaseFunctionTypesToCall = []string{
		jsonToStringArray, jsonToIntArray,
	}
	delay                         = 500
	totalJsonToJsonFunctions      = 3
	totalJsonToStringFunctions    = 1
	totalJsonToIntFunctions       = 1
	currentQuestionType           = util.SimplePeopleQuestions
	currentFunctionType           string
	personQuestionData            transforms.PureJson
	personAnswerDataJson          transforms.PureJson
	personAnswerDataString        string
	personAnswerDataInt           int
	purchaseQuestionData          transforms.PureJsonArray
	purchaseAnswerDataIntArray    []int
	purchaseAnswerDataStringArray []string
)

func generatePurchaseQuestionData() transforms.PureJsonArray {
	return transforms.PureJsonArray{"purchases": util.GeneratePurchaseList()}
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
		"age":            util.GeneratePossibleAges(),
		"id":             util.GeneratePurchaseID(),
		"name":           util.PickOneName(),
		"location":       util.PickOneLocation(),
		"favoriteColors": favoriteColorsInterface,
		"activities":     activitiesInterface,
	}
}

func main() {
	currentQuestionType = util.SimplePeopleQuestions
	currentFunctionType = jsonToJson
	personQuestionData = generatePersonQuestionData()
	personAnswerDataJson = transforms.DeleteOneKey(personQuestionData)

	// using standard library "flag" package
	flag.String("MODE", "dev", "whether we're running in dev or production mode")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	MODE := viper.GetString("MODE") // retrieve value from viper

	// don't bother overriding the mode when developing locally
	viper.SetDefault("PORT", "8000")

	router := gin.Default()

	// Serve static files to frontend if server is started in production environment
	if MODE == "prod" {
		gin.SetMode(gin.ReleaseMode)
		router.Use(static.Serve("/", static.LocalFile("./build", true)))
	}

	router.SetTrustedProxies(nil)

	// the websocket stuff
	router.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)

			return
		}
		defer ws.Close()
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)

				break
			}
			var response []byte
			if string(message) == "update" {
				var mixedResponse interface{}
				if currentFunctionType == jsonToJson {
					mixedResponse = JsonToJsonQuestion{
						Question: personQuestionData,
						Answer:   personAnswerDataJson,
					}
				} else if currentFunctionType == jsonToString {
					mixedResponse = JsonToStringQuestion{
						Question: personQuestionData,
						Answer:   personAnswerDataString,
					}
				} else if currentFunctionType == jsonToInt {
					mixedResponse = JsonToIntQuestion{
						Question: personQuestionData,
						Answer:   personAnswerDataInt,
					}
				} else if currentFunctionType == jsonToIntArray {
					mixedResponse = JsonToIntArrayQuestion{
						Question: purchaseQuestionData,
						Answer:   purchaseAnswerDataIntArray,
					}
				} else if currentFunctionType == jsonToStringArray {
					mixedResponse = JsonToStringArrayQuestion{
						Question: purchaseQuestionData,
						Answer:   purchaseAnswerDataStringArray,
					}
				}
				response, err = json.Marshal(mixedResponse)
				if err != nil {
					log.Println("could not marshall json")
				}
				err = ws.WriteMessage(mt, response)
				if err != nil {
					log.Println(err)

					break
				}
			}

			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	})

	router.GET("/question", getQuestion)
	router.POST("/answer", getAnswer)

	// for no matching routes
	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(http.StatusNotFound, gin.H{}) })

	port := viper.GetString("PORT")
	router.Run(":" + port)
}

var upgrader = websocket.Upgrader{
	// this is just to set the upgrade to succeed
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getQuestion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, personQuestionData)
}

func generateNextQuestionAnswer() {
	// we need to know what type of question we want so that we can use that to determine the subset
	// of function types to use to create the activity
	currentQuestionType = util.GenerateQuestionType()

	log.Println("current question type is:", currentQuestionType)

	// there's probably a better way to structure this hierarchy, but we'll just go with
	// something dumb and verbose for now
	if currentQuestionType == util.SimplePeopleQuestions {
		totalFunctionTypes := len(peopleFunctionTypesToCall)
		functionArrayIndex := rand.Intn(totalFunctionTypes)
		currentFunctionType = peopleFunctionTypesToCall[functionArrayIndex]
	} else if currentQuestionType == util.SimplePurchaseQuestions {
		totalFunctionTypes := len(purchaseFunctionTypesToCall)
		functionArrayIndex := rand.Intn(totalFunctionTypes)
		currentFunctionType = purchaseFunctionTypesToCall[functionArrayIndex]
	} else {
		log.Fatal("this blew up because we couldn't determine the currentFunctionType")
	}

	log.Println("and the current function type is:", currentFunctionType)

	// this should really by something like nextFunctionType, but we can refactor later
	switch currentFunctionType {
	// this is where the Simple Purchase Question exercises are generated
	case jsonToIntArray:
		var jsonToIntArrayFunction func(transforms.PureJsonArray) []int

		functionToCall := 0

		switch functionToCall {
		case 0:
			jsonToIntArrayFunction = transforms.GetAllArrayIntValues
		default:
			log.Fatal("blow it all to hell!")
		}

		purchaseQuestionData = generatePurchaseQuestionData()
		purchaseAnswerDataIntArray = jsonToIntArrayFunction(purchaseQuestionData)
	case jsonToStringArray:
		var jsonToStringArrayFunction func(transforms.PureJsonArray) []string

		functionToCall := 0

		switch functionToCall {
		case 0:
			jsonToStringArrayFunction = transforms.GetAllArrayStringValues
		default:
			log.Fatal("whoa...this massively blew up")
		}

		purchaseQuestionData = generatePurchaseQuestionData()
		purchaseAnswerDataStringArray = jsonToStringArrayFunction(purchaseQuestionData)

		// this is where the Simple Person Question exercises are generated
	case jsonToInt:
		var jsonToIntFunction func(transforms.PureJson) int

		// only one of these at the moment, same as below
		functionToCall := 0

		switch functionToCall {
		case 0:
			jsonToIntFunction = transforms.GetOneKeyIntValue
		default:
			log.Fatal("this deffo blew up")
		}

		personQuestionData = generatePersonQuestionData()
		personAnswerDataInt = jsonToIntFunction(personQuestionData)
	case jsonToString:
		var jsonToStringFunction func(transforms.PureJson) string

		// we should have more of these, but for now, we just hardcode to 0
		// functionCall := rand.Intn(totalJsonToStringFunctions)
		functionToCall := 0

		switch functionToCall {
		case 0:
			jsonToStringFunction = transforms.GetOneKeyStringValue
		default:
			log.Fatal("this blew up!")
		}

		personQuestionData = generatePersonQuestionData()
		personAnswerDataString = jsonToStringFunction(personQuestionData)

	case jsonToJson:
		var jsonToJsonFunction func(transforms.PureJson) transforms.PureJson

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

		personQuestionData = generatePersonQuestionData()
		personAnswerDataJson = jsonToJsonFunction(personQuestionData)
	default:
		log.Fatal("blew the F up!")
	}
}

func getAnswer(c *gin.Context) {
	// we keep the current function type in state so we know how to compare the answer
	if currentFunctionType == jsonToStringArray {
		// not implemented
	} else if currentFunctionType == jsonToIntArray {
		// not implemented
	} else if currentFunctionType == jsonToInt {
		response, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
		}

		result, err := strconv.Atoi(string(response))
		if err != nil {
			log.Fatal(err)
		}

		if result == personAnswerDataInt {
			log.Println("you got that int, bruh!")
			generateNextQuestionAnswer()
		} else {
			log.Println("wrong answer, please try again")
		}
	} else if currentFunctionType == jsonToString {
		response, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
		}

		if string(response) == personAnswerDataString {
			log.Println("you got it!")
			generateNextQuestionAnswer()
		} else {
			log.Println("wrong answer, please try again")
		}
	} else if currentFunctionType == jsonToJson {
		var actualAnswer transforms.PureJson

		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		diff := deep.Equal(actualAnswer, personAnswerDataJson)
		if diff == nil {
			log.Println("noice, bruh!")
			generateNextQuestionAnswer()
		} else {
			log.Println("wrong answer, please try again")
		}
	} else {
		log.Println("No current function type...sad day")
	}
}
