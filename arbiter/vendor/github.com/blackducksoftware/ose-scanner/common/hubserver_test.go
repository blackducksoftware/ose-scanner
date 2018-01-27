package common

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Although this test is slow, it should run okay in a parallel suite.
func TestClientDoesntHang(t *testing.T) {
	log.SetOutput(os.Stdout)

	// Add some cushion to the hangtime.
	totalAllowedTime := time.Second * time.Duration(hangtimeBeforeTimingOutOnTheHub.Seconds()+1)

	// so that we know to end the test.
	testResult := make(chan bool, 1)

	// so we know when to kill the server.
	requestCompleted := make(chan string, 1)

	t.Logf("hangtime = %v, allowed = %v", hangtimeBeforeTimingOutOnTheHub, totalAllowedTime)

	// Should complete in < 'hangtime' seconds
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// simulate a hub response that never, ever, ever returns.
		for {
			select {
			// shutdown hook for the unit test.
			case <-testResult:
				log.Println("Test result obtained. Exiting inifinte request time simulator.")
				return
			// simulate 'the hub' being down.
			default:
				log.Println("Test still running.", time.Now())
				time.Sleep(1 * time.Second)
			}
		}
	}))
	defer func() {
		svr.Close()
	}()

	webGet := func() {
		t.Logf("Creating server and running Get operation.  This should hang.")
		_, err := NewHubServer(nil).client.Get(svr.URL)
		if err != nil {
			// finished the Get operation, returning.
			requestCompleted <- fmt.Sprintf("yay, i completed with an error %v", err)
			return
		} else {
			// This should never happen unless this test is borked.
			t.Fatal("Web server should always return an error for this test!!!")
			t.Fail()
		}
	}

	// Start the get.  Should return long before one minute.
	go webGet()
	for {
		select {
		case <-requestCompleted:
			testResult <- true
			return
		case <-time.After(totalAllowedTime):
			testResult <- false
			t.Fail()
			return
		}
	}

}
