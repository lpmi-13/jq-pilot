package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"reflect"
	"sort"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Person struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	FavoriteColors []string `json:"favoriteColors"`
}

type Colors struct {
	FavoriteColors []string `json:"favoriteColors"`
}

type Name struct {
	Name string `json:"name"`
}

type Question1 struct {
	Question []Person `json:"question"`
	Answer   Person   `json:"answer"`
}

type Question2 struct {
	Question []Person `json:"question"`
	Answer   Colors   `json:"answer"`
}

type Question3 struct {
	Question []Person `json:"question"`
	Answer   Name     `json:"answer"`
}

// initialize this at the beginning
var (
	currentLevel  = 1
	level1        = 1
	level2        = 2
	level3        = 3
	delay         = 500
	question1Data = []Person{
		{ID: "1", Name: "Alice", FavoriteColors: []string{"green", "yellow"}},
		{ID: "2", Name: "Bob", FavoriteColors: []string{"green", "purple", "red"}},
		{ID: "3", Name: "Sue", FavoriteColors: []string{"red", "blue"}},
	}
	question1Answer = Person{ID: "1", Name: "Alice", FavoriteColors: []string{"green", "yellow"}}
	question2Data   = []Person{
		{ID: "1", Name: "Rachel", FavoriteColors: []string{"blue"}},
		{ID: "2", Name: "Thomas", FavoriteColors: []string{"red", "orange"}},
		{ID: "3", Name: "Sarah", FavoriteColors: []string{"blue", "green"}},
		{ID: "4", Name: "Max", FavoriteColors: []string{"yellow", "black"}},
		{ID: "5", Name: "Rudi", FavoriteColors: []string{"red", "blue", "white"}},
	}
	question2Answer = Colors{
		FavoriteColors: []string{"blue", "red", "orange", "green", "yellow", "black", "white"},
	}
	question3Data = []Person{
		{ID: "1", Name: "Joe", FavoriteColors: []string{"blue", "black"}},
		{ID: "2", Name: "Alex", FavoriteColors: []string{"pink", "purple"}},
		{ID: "3", Name: "Jessie", FavoriteColors: []string{"red"}},
		{ID: "4", Name: "Phil", FavoriteColors: []string{"orange"}},
	}
	question3Answer = Name{Name: "Jessie"}
)

func main() {
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
			if string(message) == "ping" {
				response = []byte("pong")
			}
			if string(message) == "update" {
				switch currentLevel {
				case level1:
					mixedResponse := Question1{
						Question: question1Data,
						Answer:   question1Answer,
					}
					response, err = json.Marshal(mixedResponse)
					if err != nil {
						log.Println("could not marshall json for level:", currentLevel)
					}
				case level2:
					mixedResponse := Question2{
						Question: question2Data,
						Answer:   question2Answer,
					}
					response, err = json.Marshal(mixedResponse)
					if err != nil {
						log.Println("could not marshall json for level:", currentLevel)
					}
				case level3:
					mixedResponse := Question3{
						Question: question3Data,
						Answer:   question3Answer,
					}
					response, err = json.Marshal(mixedResponse)
					if err != nil {
						log.Println("could not marshall json for level:", currentLevel)
					}
				default:
					response = []byte("no level set")
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
	switch currentLevel {
	case level1:
		c.IndentedJSON(http.StatusOK, question1Data)
	case level2:
		c.IndentedJSON(http.StatusOK, question2Data)
	case level3:
		c.IndentedJSON(http.StatusOK, question3Data)
	default:
		c.IndentedJSON(http.StatusOK, "out of questions!")
	}
}

func getAnswer(c *gin.Context) {
	switch currentLevel {
	case level1:
		var actualAnswer Person

		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		if reflect.DeepEqual(actualAnswer, question1Answer) {
			log.Println("you got the right answer!")

			currentLevel = currentLevel + 1
		} else {
			if viper.GetString("MODE") == "dev" {
				log.Println("got:", actualAnswer)
				log.Println("wanted:", question1Answer)
			} else {
				log.Println("wrong answer, please try again")
			}
		}
	case level2:
		var actualAnswer Colors
		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		sort.Strings(actualAnswer.FavoriteColors)
		sort.Strings(question2Answer.FavoriteColors)

		if reflect.DeepEqual(actualAnswer, question2Answer) {
			log.Print("you got the right answer!")

			currentLevel = currentLevel + 1
		} else {
			if viper.GetString("MODE") == "dev" {
				log.Println("got:", actualAnswer)
				log.Println("wanted:", question2Answer)
			} else {
				log.Println("wrong answer, please try again")
			}
		}
	case level3:
		var actualAnswer Name
		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		if actualAnswer == question3Answer {
			log.Println("you got the right answer")
		} else {
			if viper.GetString("MODE") == "dev" {
				log.Println("got:", actualAnswer)
				log.Println("wanted:", question3Answer)
			} else {
				log.Println("wrong answer, please try again")
			}
		}
	}
}
