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
	people := []Person{
		{ID: "1", Name: "Alice", FavoriteColors: []string{"green", "yellow"}},
		{ID: "2", Name: "Bob", FavoriteColors: []string{"green", "purple", "red"}},
		{ID: "3", Name: "Sue", FavoriteColors: []string{"red", "blue"}},
	}
	c.IndentedJSON(http.StatusOK, people)
}

func getAnswer(c *gin.Context) {
	var actualAnswer Person

	expectedAnswer := Person{ID: "1", Name: "Alice", FavoriteColors: []string{"green", "yellow"}}

	if err := c.BindJSON(&actualAnswer); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if reflect.DeepEqual(actualAnswer, expectedAnswer) {
		log.Println("you got the right answer!")
	} else {
		log.Println("nope, try again!")
	}
}
