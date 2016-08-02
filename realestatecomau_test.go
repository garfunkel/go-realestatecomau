package realestatecomau

import (
	"log"
	"testing"
)

func TestGetInfo(t *testing.T) {
	info, err := GetInfo("North Sydney, NSW 2060")

	if err != nil {
		log.Fatal(err)
	}

	err = info.GetInspections()

	if err != nil {
		log.Fatal(err)
	}

	err = info.GetImages()

	if err != nil {
		log.Fatal(err)
	}
}
