package imageutil

import (
	"os"
	"testing"
)

func TestExtractExif(t *testing.T) {
	f, err := os.Open("../../test/img-gps.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tags, err := ExtractExif(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(tags) == 0 {
		t.Fail()
	}

	for field, tag := range tags {
		t.Logf("%20s: %s\n", field, tag)
	}
}
