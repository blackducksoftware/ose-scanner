package arbiter

import (
	"net/http"
	"testing"
)

func TestScanAbortLogic(t *testing.T) {
	statusCode, logMessage, _, _, err := scanAbortLogic("abc", "123", make(map[string]*controllerDaemon), make(map[string]*assignImage), 0)
	if statusCode != http.StatusNotFound {
		t.Error("wrong status code, expected StatusNotFound")
	}
	expectedMessage := "Unknown controller [abc] claimed abort for image: 123\n"
	if logMessage != expectedMessage {
		t.Error("wrong log message, got", logMessage, "instead of", expectedMessage)
	}
	if err == nil {
		t.Error("expected error, got nil")
	}
}
