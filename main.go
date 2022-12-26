package main

import (
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// initialize this at the beginning
var (
	currentLevel  = 1
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
	router := gin.Default()

	// the websocket stuff
	router.GET("/", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)

			return
		}
		defer ws.Close()
		for {
			log.Println("in a loop!")
			mt, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)

				break
			}
			var response []byte
			if string(message) == "ping" {
				response = []byte("pong")
			}
			err = ws.WriteMessage(mt, response)
			if err != nil {
				log.Println(err)

				break
			}
		}
	})

	router.GET("/question", getQuestion)
	router.POST("/answer", getAnswer)
	router.Run("localhost:8000")
}

var upgrader = websocket.Upgrader{
	// this is just to set the upgrade to succeed
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getQuestion(c *gin.Context) {
	log.Println("current level here is: ", currentLevel)

	switch currentLevel {
	case 1:
		c.IndentedJSON(http.StatusOK, question1Data)
	case 2:
		c.IndentedJSON(http.StatusOK, question2Data)
	case 3:
		c.IndentedJSON(http.StatusOK, question3Data)
	default:
		c.IndentedJSON(http.StatusOK, "out of questions!")
	}
}

func getAnswer(c *gin.Context) {
	switch currentLevel {
	case 1:
		var actualAnswer Person

		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		if reflect.DeepEqual(actualAnswer, question1Answer) {
			log.Println("current level is: ", currentLevel)

			log.Println("you got the right answer!")

			currentLevel = currentLevel + 1
		} else {
			log.Println("nope, try again!")
		}
	case 2:
		var actualAnswer Colors
		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		if reflect.DeepEqual(actualAnswer, question2Answer) {
			log.Print("you got the right answer!")

			currentLevel = currentLevel + 1
		} else {
			log.Println("nope, try again")
		}
	case 3:
		var actualAnswer Name
		if err := c.BindJSON(&actualAnswer); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		if actualAnswer == question3Answer {
			log.Println("you got the right answer")
		} else {
			log.Println("nope, try again")
		}
	}
}
