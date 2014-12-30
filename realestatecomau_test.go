package realestatecomau

import (
	"log"
	"testing"
)

func TestGetInfo(t *testing.T) {
	_, err := GetInfo("59/47 Hampstead Road Homebush West NSW 2140")

	if err != nil {
		log.Fatal(err)
	}
}
