package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
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

// these generics are hashtag delicious, and I'm here for it
type JsonQuestion[T any, V any] struct {
	Question T      `json:"question"`
	Answer   V      `json:"answer"`
	Prompt   string `json:"prompt"`
}

const (
	jsonToJson = "jsonToJson"
	// this just commemorates when I solved a particularly tricky jq slicing problem
	jsonToRidicJson   = "jsonToRidicJson"
	jsonToString      = "jsonToString"
	jsonToInt         = "jsonToInt"
	jsonToStringArray = "jsonToStringArray"
	jsonToIntArray    = "jsonToIntArray"
	jsonToDict        = "jsonToDict"
)

var (
	peopleFunctionTypesToCall = []string{
		jsonToJson, jsonToString, jsonToInt,
	}
	purchaseFunctionTypesToCall = []string{
		jsonToStringArray, jsonToIntArray, jsonToJson,
	}
	lotteryFunctionTypesToCall = []string{
		jsonToJson, jsonToIntArray, jsonToInt,
	}
	gradesFunctionTypesToCall = []string{
		jsonToInt, jsonToJson, jsonToRidicJson,
	}
	delay                         = 500
	totalJsonToJsonFunctions      = 3
	currentQuestionType           = util.SimplePeopleQuestions
	currentFunctionType           string
	personQuestionData            transforms.PureJson
	personAnswerDataJson          transforms.PureJson
	personAnswerDataString        string
	personAnswerDataInt           int
	purchaseQuestionData          transforms.PureJsonArrayPurchases
	purchaseAnswerDataIntArray    []int
	purchaseAnswerDataStringArray []string
	purchaseAnswerDataJsonArray   []util.FakePurchase
	lotteryQuestionData           transforms.PureJsonArrayLottery
	lotteryAnswerDataIntArray     []int
	lotteryAnswerDataInt          int
	lotteryAnswerDataJson         util.FakeLotteryPick
	lotteryAnswerFreqDist         map[string]int
	gradesQuestionData            util.ComplexGradesObject
	gradesAnswerDataInt           int
	gradesAnswerDataJson          []util.SimplerStudent
	gradesAnswerDataRidicJson     util.Student
	prompt                        = "please do stuff!"
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
	// this should be dynamic to set the first exercise instead of the same one every time
	currentQuestionType = util.SimplePeopleQuestions
	currentFunctionType = jsonToJson
	personQuestionData = generatePersonQuestionData()
	personAnswerDataJson, prompt = transforms.DeleteOneKey(personQuestionData)

	// using standard library "flag" package
	flag.String("MODE", "dev", "whether we're running in dev or production mode")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	MODE := viper.GetString("MODE") // retrieve value from viper

	// don't bother overriding the mode when developing locally
	viper.SetDefault("PORT", "8000")

	// this is two different checks for MODE == "prod", because we need to set the mode
	// before we intialize the router, but we can't call router.Use until after we initialize the router
	if MODE == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Serve static files to frontend if server is started in production environment
	if MODE == "prod" {
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
				if currentQuestionType == util.SimplePeopleQuestions {
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
						log.Println("couldn't match function type for people question")
					}
				} else if currentQuestionType == util.SimplePurchaseQuestions {
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
					}
				} else if currentQuestionType == util.SimpleLotteryQuestions {
					if currentFunctionType == jsonToJson {
						mixedResponse = JsonQuestion[map[string][]util.FakeLotteryPick, util.FakeLotteryPick]{
							Question: lotteryQuestionData,
							Answer:   lotteryAnswerDataJson,
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
					}
				} else if currentQuestionType == util.SimpleGradesQuestions {
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
						log.Fatal("no obvious function type")
					}
				} else {
					log.Fatal("couldn't get question type")
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

				time.Sleep(time.Duration(delay) * time.Millisecond)
			} else {
				log.Fatal("no dice")
			}
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

// this is crap and should be more general
func getQuestion(c *gin.Context) {
	if currentQuestionType == util.SimplePeopleQuestions {
		c.IndentedJSON(http.StatusOK, personQuestionData)
	} else if currentQuestionType == util.SimplePurchaseQuestions {
		c.IndentedJSON(http.StatusOK, purchaseQuestionData)
	} else if currentQuestionType == util.SimpleLotteryQuestions {
		c.IndentedJSON(http.StatusOK, lotteryQuestionData)
	} else if currentQuestionType == util.SimpleGradesQuestions {
		c.IndentedJSON(http.StatusOK, gradesQuestionData)
	}
}

func generateNextQuestionAnswer() {
	// we need to know what type of question we want so that we can use that to determine the subset
	// of function types to use to create the activity
	// currentQuestionType = util.GeneratePossibleValue(util.PossibleQuestionTypes)
	currentQuestionType = util.SimpleGradesQuestions

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
	}
}

func processAnswer[T any](context *gin.Context, expectedAnswer T) {
	var actualAnswer T
	if err := context.BindJSON(&actualAnswer); err != nil {
		context.AbortWithStatus(http.StatusBadRequest)
	}

	diff := deep.Equal(actualAnswer, expectedAnswer)

	if diff == nil {
		log.Println("correct")
		generateNextQuestionAnswer()
	} else {
		log.Println("wrong answer, please try again")
	}
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
		}
	} else {
		log.Println("No current function type...sad day")
	}
}
