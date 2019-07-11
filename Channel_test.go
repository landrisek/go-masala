package masala

import (
	"testing"
)

/** go test -v masala */
func TestLastEventID(test *testing.T) {
	test.Run("LastEventID", func(test *testing.T) {
		channel := NewChannel("test")
		eventID := channel.LastEventID()
		if len(eventID) > 0 {
			test.Fatal(eventID + " is not empty")
		}
	})
}