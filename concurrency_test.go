package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestConcurrentStateAccess exercises the shared game-state readers and the
// writer (generateNextQuestionAnswer) from many goroutines at once. Before the
// stateMu guard was added this reliably tripped the race detector (and could
// crash the process with "concurrent map read and map write"). Run with -race:
//
//	go test -race -run TestConcurrentStateAccess
func TestConcurrentStateAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Make sure there is some state to read before the goroutines start.
	generateNextQuestionAnswer()

	const workers = 16
	stop := make(chan struct{})
	var wg sync.WaitGroup

	worker := func(fn func()) {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				fn()
			}
		}
	}

	// Writers: regenerate the whole shared state.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(generateNextQuestionAnswer)
	}

	// Readers: GET /question.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/question", nil)
			getQuestion(c)
		})
	}

	// Readers: GET /prompt.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/prompt", nil)
			getPrompt(c)
		})
	}

	// Readers: the SSE payload builder + marshal (what the /sse stream does).
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(func() {
			resp, err := generateMixedResponse()
			if err != nil {
				return
			}
			_, _ = json.Marshal(resp)
		})
	}

	// Readers/writers: POST /answer with a (usually wrong) body. A correct answer
	// also triggers a regeneration from inside the handler.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/answer", strings.NewReader(`"probably-wrong"`))
			c.Request.Header.Set("Content-Type", "application/json")
			getAnswer(c)
		})
	}

	time.Sleep(300 * time.Millisecond)
	close(stop)
	wg.Wait()
}
