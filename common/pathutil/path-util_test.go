package pathutil

import "testing"

func TestGetTotalFiles(t *testing.T) {
	var want int64 = 5760

	total, err := GetTotalFiles("/home/martins/Desenvolvimento/SPTC/files/benchmark/rotate")
	if err != nil {
		t.Fatal(err)
	}

	if total != want {
		t.Log(total)
		t.Errorf("%d, want %d", total, want)
	}
}
