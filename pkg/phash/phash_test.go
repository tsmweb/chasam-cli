package phash

import (
	"github.com/tsmweb/chasam/common/mediautil"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestAverageHash(t *testing.T) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := AverageHash(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("A-HASH: %s\n", FormatToHex(hash1))
}

func TestGenerateDiff(t *testing.T) {
	root := "../../../files/ambiente"
	img, err := loadImage(filepath.Join(root, "img-a.jpeg"))
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

func TestDifferenceHash(t *testing.T) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := DifferenceHash(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("D-HASH: %s\n", FormatToHex(hash1))
}

func TestDifferenceHashVertical(t *testing.T) {
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

func TestPerceptionHash(t *testing.T) {
	img, err := loadImage("../../../files/test/lenna.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := PerceptionHash_(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("P-HASH: %s\n", FormatToHex(hash1))

	img2, err := loadImage("../../../files/test/lenna-blur.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := PerceptionHash_(img2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("P-HASH2: %s\n", FormatToHex(hash2))

	dh, err := Distance(hash1, hash2)
	if err != nil {
		t.Error(err)
	}
	t.Logf("HAMMING: %d\n", dh)
}

func TestWaveletHash(t *testing.T) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash, err := WaveletHash(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("WaveletHash: %v\n", hash)
}

func BenchmarkAverageHash(b *testing.B) {
	img, err := loadImage("../../../files/test/Alyson.jpg")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = AverageHash(img)
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
		_, err = DifferenceHash(img)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDifferenceHashVertical(b *testing.B) {
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

func BenchmarkPerceptualHash(b *testing.B) {
	img, err := loadImage("../../../files/test/lenna.jpg")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = PerceptionHash(img)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPerceptualHashLib(b *testing.B) {
	img, err := loadImage("../../../files/test/lenna.jpg")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = PerceptionHash_(img)
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
