package main

import (
	"encoding/json"
	"flag"
	"io"
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
	Prompt   string                 `json:"prompt"`
}

type JsonToStringQuestion struct {
	Question map[string]interface{} `json:"question"`
	Answer   string                 `json:"answer"`
	Prompt   string                 `json:"prompt"`
}

type JsonToIntQuestion struct {
	Question map[string]interface{} `json:"question"`
	Answer   int                    `json:"answer"`
	Prompt   string                 `json:"prompt"`
}

// it's at this point that I feel like I want some generics, so I'll get to implementing
// that at some point (https://itnext.io/how-to-use-golang-generics-with-structs-8cabc9353d75)
type JsonToIntArrayQuestion struct {
	Question map[string][]util.FakePurchase `json:"question"`
	Answer   []int                          `json:"answer"`
	Prompt   string                         `json:"prompt"`
}

type JsonToStringArrayQuestion struct {
	Question map[string][]util.FakePurchase `json:"question"`
	Answer   []string                       `json:"answer"`
	Prompt   string                         `json:"prompt"`
}

type JsonToJsonArrayQuestion struct {
	Question map[string][]util.FakePurchase `json:"question"`
	Answer   []util.FakePurchase            `json:"answer"`
	Prompt   string                         `json:"prompt"`
}

type JsonToIntArrayLotteryQuestion struct {
	Question map[string][]util.FakeLotteryPick `json:"question"`
	Answer   []int                             `json:"answer"`
	Prompt   string                            `json:"prompt"`
}

type JsonToIntLotteryQuestion struct {
	Question map[string][]util.FakeLotteryPick `json:"question"`
	Answer   int                               `json:"answer"`
	Prompt   string                            `json:"prompt"`
}

type JsonToJsonGradesQuestion struct {
	Question util.ComplexGradesObject `json:"question"`
	Answer   util.Student             `json:"answer"`
	Prompt   string                   `json:"prompt"`
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
		jsonToStringArray, jsonToIntArray, jsonToJson,
	}
	lotteryFunctionTypesToCall = []string{
		jsonToIntArray, jsonToInt,
	}
	gradesFunctionTypesToCall = []string{
		jsonToJson,
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
	purchaseQuestionData          transforms.PureJsonArrayPurchases
	purchaseAnswerDataIntArray    []int
	purchaseAnswerDataStringArray []string
	purchaseAnswerDataJsonArray   []util.FakePurchase
	lotteryQuestionData           transforms.PureJsonArrayLottery
	lotteryAnswerDataIntArray     []int
	lotteryAnswerDataInt          int
	gradesQuestionData            util.ComplexGradesObject
	gradesAnswerData              util.Student
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
						mixedResponse = JsonToJsonQuestion{
							Question: personQuestionData,
							Answer:   personAnswerDataJson,
							Prompt:   prompt,
						}
					} else if currentFunctionType == jsonToString {
						mixedResponse = JsonToStringQuestion{
							Question: personQuestionData,
							Answer:   personAnswerDataString,
							Prompt:   prompt,
						}
					} else if currentFunctionType == jsonToInt {
						mixedResponse = JsonToIntQuestion{
							Question: personQuestionData,
							Answer:   personAnswerDataInt,
							Prompt:   prompt,
						}
					} else {
						log.Println("couldn't match function type for people question")
					}
				} else if currentQuestionType == util.SimplePurchaseQuestions {
					if currentFunctionType == jsonToIntArray {
						mixedResponse = JsonToIntArrayQuestion{
							Question: purchaseQuestionData,
							Answer:   purchaseAnswerDataIntArray,
							Prompt:   prompt,
						}
					} else if currentFunctionType == jsonToStringArray {
						mixedResponse = JsonToStringArrayQuestion{
							Question: purchaseQuestionData,
							Answer:   purchaseAnswerDataStringArray,
							Prompt:   prompt,
						}
					} else if currentFunctionType == jsonToJson {
						mixedResponse = JsonToJsonArrayQuestion{
							Question: purchaseQuestionData,
							Answer:   purchaseAnswerDataJsonArray,
							Prompt:   prompt,
						}
					}
				} else if currentQuestionType == util.SimpleLotteryQuestions {
					if currentFunctionType == jsonToIntArray {
						mixedResponse = JsonToIntArrayLotteryQuestion{
							Question: lotteryQuestionData,
							Answer:   lotteryAnswerDataIntArray,
							Prompt:   prompt,
						}
					} else if currentFunctionType == jsonToInt {
						mixedResponse = JsonToIntLotteryQuestion{
							Question: lotteryQuestionData,
							Answer:   lotteryAnswerDataInt,
							Prompt:   prompt,
						}
					}
				} else if currentQuestionType == util.SimpleGradesQuestions {
					if currentFunctionType == jsonToJson {
						mixedResponse = JsonToJsonGradesQuestion{
							Question: gradesQuestionData,
							Answer:   gradesAnswerData,
							Prompt:   prompt,
						}
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
	} else {
		log.Fatal("this blew up because we couldn't determine the currentFunctionType")
	}

	switch currentQuestionType {
	case util.SimpleLotteryQuestions:
		switch currentFunctionType {
		case jsonToIntArray:
			var jsonToIntArrayFunction func(transforms.PureJsonArrayLottery) ([]int, string)

			// only one so far
			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToIntArrayFunction = transforms.GetAllUniqueArrayIntValues
			}

			lotteryQuestionData = generateLotteryPickQuestionData()
			lotteryAnswerDataIntArray, prompt = jsonToIntArrayFunction(lotteryQuestionData)
		case jsonToInt:

			var jsonToIntFunction func(transforms.PureJsonArrayLottery) (int, string)

			// only one so far here too
			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToIntFunction = transforms.GetNumberOfPicks
			}

			lotteryQuestionData = generateLotteryPickQuestionData()
			lotteryAnswerDataInt, prompt = jsonToIntFunction(lotteryQuestionData)
		}

	case util.SimplePeopleQuestions:
		// this should really by something like nextFunctionType, but we can refactor later
		switch currentFunctionType {
		// this is where the Simple Person Question exercises are generated
		case jsonToInt:
			var jsonToIntFunction func(transforms.PureJson) (int, string)

			// only one of these at the moment, same as below
			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToIntFunction = transforms.GetOneKeyIntValue
			default:
				log.Fatal("this deffo blew up")
			}

			personQuestionData = generatePersonQuestionData()
			personAnswerDataInt, prompt = jsonToIntFunction(personQuestionData)
		case jsonToString:
			var jsonToStringFunction func(transforms.PureJson) (string, string)

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
			personAnswerDataString, prompt = jsonToStringFunction(personQuestionData)

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

			personQuestionData = generatePersonQuestionData()
			personAnswerDataJson, prompt = jsonToJsonFunction(personQuestionData)
		default:
			log.Fatal("blew the F up!")
		}
	case util.SimplePurchaseQuestions:
		switch currentFunctionType {
		// this is where the Simple Purchase Question exercises are generated
		case jsonToIntArray:
			var jsonToIntArrayFunction func(transforms.PureJsonArrayPurchases) ([]int, string)

			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToIntArrayFunction = transforms.GetAllArrayIntValues
			default:
				log.Fatal("blow it all to hell!")
			}

			purchaseQuestionData = generatePurchaseQuestionData()
			purchaseAnswerDataIntArray, prompt = jsonToIntArrayFunction(purchaseQuestionData)
		case jsonToStringArray:
			var jsonToStringArrayFunction func(transforms.PureJsonArrayPurchases) ([]string, string)

			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToStringArrayFunction = transforms.GetAllArrayStringValues
			default:
				log.Fatal("whoa...this massively blew up")
			}

			purchaseQuestionData = generatePurchaseQuestionData()
			purchaseAnswerDataStringArray, prompt = jsonToStringArrayFunction(purchaseQuestionData)
		case jsonToJson:
			var jsonToJsonArrayFunction func(transforms.PureJsonArrayPurchases) ([]util.FakePurchase, string)

			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToJsonArrayFunction = transforms.GetFilteredByPurchasePrice
			default:
				log.Fatal("couldn't figure out which function to use")
			}

			purchaseQuestionData = generatePurchaseQuestionData()
			purchaseAnswerDataJsonArray, prompt = jsonToJsonArrayFunction(purchaseQuestionData)
		}
	case util.SimpleGradesQuestions:
		switch currentFunctionType {
		case jsonToJson:
			var jsonToJsonFunction func(util.ComplexGradesObject) (util.Student, string)

			// this is getting a bit ridic, since I initially assumed we might want to have different types of transforms
			// possibly returning the same type of data structure, but what it's turning into is just one type of
			// data structure per function type, so I'll probably get rid of this hard coding later
			functionToCall := 0

			switch functionToCall {
			case 0:
				jsonToJsonFunction = transforms.GetHighestResultInOneSubject
			default:
				log.Fatal("something bad happened")
			}

			gradesQuestionData = generateGradesQuestionData()
			gradesAnswerData, prompt = jsonToJsonFunction(gradesQuestionData)
		default:
			log.Println("fell into the default question type...for...reasons...")
		}
	}
}

func getAnswer(c *gin.Context) {
	if currentQuestionType == util.SimplePurchaseQuestions {
		// we keep the current function type in state so we know how to compare the answer
		if currentFunctionType == jsonToStringArray {
			var actualAnswer []string

			if err := c.BindJSON(&actualAnswer); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
			}

			diff := deep.Equal(actualAnswer, purchaseAnswerDataStringArray)

			if diff == nil {
				log.Println("ye olde string slice is all good!")
				generateNextQuestionAnswer()
			} else {
				log.Println("wrong answer, please try again")
			}
		} else if currentFunctionType == jsonToIntArray {
			var actualAnswer []int

			if err := c.BindJSON(&actualAnswer); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
			}

			diff := deep.Equal(actualAnswer, purchaseAnswerDataIntArray)
			if diff == nil {
				log.Println("get that int slice!")
				generateNextQuestionAnswer()
			} else {
				log.Println("wrong answer, please try again")
			}
		} else if currentFunctionType == jsonToJson {
			var actualAnswer []util.FakePurchase

			if err := c.BindJSON(&actualAnswer); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
			}

			diff := deep.Equal(actualAnswer, purchaseAnswerDataJsonArray)
			if diff == nil {
				log.Println("nice work, all filtered!")
				generateNextQuestionAnswer()
			} else {
				log.Println("not quite...try again")
			}
		}
	} else if currentQuestionType == util.SimplePeopleQuestions {
		if currentFunctionType == jsonToString {
			response, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Fatal(err)
			}

			if string(response) == personAnswerDataString {
				log.Println("you got it!")
				generateNextQuestionAnswer()
			} else {
				log.Println("wrong answer, please try again")
			}
		} else if currentFunctionType == jsonToInt {
			response, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Fatal(err)
			}

			result, err := strconv.Atoi(string(response))
			if err != nil {
				log.Fatal(err)
			}

			if result == personAnswerDataInt {
				log.Println("get that int for the people!")
				generateNextQuestionAnswer()
			} else {
				log.Println("wrong answer, please try again")
			}
		} else if currentFunctionType == jsonToJson {
			var actualAnswer transforms.PureJson

			log.Println("got to here")
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
		}
	} else if currentQuestionType == util.SimpleLotteryQuestions {
		// these are the exact same implementation as above, so BE BETTER!
		if currentFunctionType == jsonToIntArray {
			var actualAnswer []int

			if err := c.BindJSON(&actualAnswer); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
			}

			diff := deep.Equal(actualAnswer, lotteryAnswerDataIntArray)
			if diff == nil {
				log.Println("get that int slice of lottery picks!")
				generateNextQuestionAnswer()
			} else {
				log.Println("wrong answer, please try again")
			}
		} else if currentFunctionType == jsonToInt {
			response, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Fatal(err)
			}

			result, err := strconv.Atoi(string(response))
			if err != nil {
				log.Fatal(err)
			}

			if result == lotteryAnswerDataInt {
				log.Println("get that int for total lottery picks!")
				generateNextQuestionAnswer()
			} else {
				log.Println("wrong answer, please try again")
			}
		}
	} else if currentQuestionType == util.SimpleGradesQuestions {
		if currentFunctionType == jsonToJson {
			var actualAnswer util.Student

			if err := c.BindJSON(&actualAnswer); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
			}

			log.Println("the submitted answer: ", actualAnswer)
			log.Println("the answer we want: ", gradesAnswerData)

			diff := deep.Equal(actualAnswer, gradesAnswerData)
			if diff == nil {
				log.Println("you found the student!")
				generateNextQuestionAnswer()
			} else {
				log.Println("didn't find the student")
			}
		}
	} else {
		log.Println("No current function type...sad day")
	}
}
