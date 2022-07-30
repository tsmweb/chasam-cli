package hash

import (
	"github.com/tsmweb/chasam/common/mediautil"
	"image"
	"os"
	"testing"
)

func TestSha1Hash(t *testing.T) {
	f, err := os.Open("../../test/img.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	t.Log(f.Name())

	h, err := Sha1Hash(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(h)
}

func TestEd2kHash(t *testing.T) {
	f, err := os.Open("../../test/img.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	h, err := Ed2kHash(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(h)
}

func TestAverageHash(t *testing.T) {
	img, err := loadImage("../../test/img.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := AverageHash(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("A-HASH: %s\n", FormatToHex(hash1))
}

func TestDifferenceHash(t *testing.T) {
	img, err := loadImage("../../test/img.jpg")
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
	img, err := loadImage("../../test/img.jpg")
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
	img, err := loadImage("../../test/img.jpg")
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := PerceptionHash(img)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("P-HASH: %s\n", FormatToHex(hash1))
}

func TestWaveletHash(t *testing.T) {
	img, err := loadImage("../../test/img.jpg")
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
	img, err := loadImage("../../test/img.jpg")
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
	img, err := loadImage("../../test/img.jpg")
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
	img, err := loadImage("../../test/img.jpg")
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
	img, err := loadImage("../../test/img.jpg")
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
