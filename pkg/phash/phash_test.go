package phash

import (
	"github.com/tsmweb/chasam/common/mediautil"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDiff(t *testing.T) {
	root := "../../../files/lenna"
	img, err := loadImage(filepath.Join(root, "lenna.jpg"))
	if err != nil {
		t.Fatal(err)
	}

	dHash, err := DifferenceHash(img)
	if err != nil {
		t.Fatal(err)
	}

	dHashV, err := DifferenceHashVertical(img)
	if err != nil {
		t.Fatal(err)
	}

	dHashA := []uint64{dHash, dHashV}

	files, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		img, err = loadImage(filepath.Join(root, f.Name()))
		if err != nil {
			t.Log(err.Error())
			continue
		}

		// D-HASH
		dh, err := DifferenceHash(img)
		if err != nil {
			t.Log(err.Error())
			continue
		}

		dis, err := Distance(dHash, dh)
		if err != nil {
			t.Log(err.Error())
			continue
		}
		t.Logf("DHash: %s - %d\n", f.Name(), dis)

		// D-HASH-VERTICAL
		dhv, err := DifferenceHashVertical(img)
		if err != nil {
			t.Log(err.Error())
			continue
		}

		dis, err = Distance(dHashV, dhv)
		if err != nil {
			t.Log(err.Error())
			continue
		}
		t.Logf("DHashV: %s - %d\n", f.Name(), dis)

		// D-HASH-ALL
		dha := []uint64{dh, dhv}

		dis, err = ExtDistance(dHashA, dha)
		if err != nil {
			t.Log(err.Error())
			continue
		}
		t.Logf("DHashA: %s - %d\n", f.Name(), dis)
	}
}

func TestDHash(t *testing.T) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := DifferenceHash(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Meu D-HASH: %s\n", FormatToHex(hash1))

	hash2, err := DifferenceHash_(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Lib D-HASH: %s\n", FormatToHex(hash2))
}

func TestDHashVertical(t *testing.T) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := DifferenceHashVertical(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Meu DifferenceHashVertical: %s\n", FormatToHex(hash1))
}

func BenchmarkDHash(b *testing.B) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = DifferenceHash(img)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDHashVertical(b *testing.B) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = DifferenceHashVertical(img)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDifferenceHash(b *testing.B) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = DifferenceHash_(img)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func loadImage(uri string) (image.Image, error) {
	f, err := os.Open(uri)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	contentType, err := mediautil.GetContentType(f)
	if err != nil {
		return nil, err
	}

	img, err := mediautil.Decode(f, contentType)
	if err != nil {
		return nil, err
	}
	return img, nil
}
