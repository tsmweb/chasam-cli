package hashutil

import (
	"os"
	"testing"
)

func TestHashSHA1(t *testing.T) {
	f, err := os.Open("../../test/img-large.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	t.Log(f.Name())

	h, err := HashSHA1(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(h)
}

func TestHashED2K(t *testing.T) {
	f, err := os.Open("../../test/img-large.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	h, err := HashED2K(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(h)
}
