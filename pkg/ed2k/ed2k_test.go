package ed2k

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestGetED2KHashFromFile(t *testing.T) {
	f, err := os.Open("file")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	eh := New()
	_, err = rd.WriteTo(eh)
	if err != nil {
		t.Fatal(err)
	}

	rh := fmt.Sprintf("%x", eh.Sum(nil))
	t.Log(rh)
	if rh != "hash here" {
		t.Fail()
	}
}
