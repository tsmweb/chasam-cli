package imageutil

import (
	"github.com/rwcarlsen/goexif/exif"
	"os"
)

var exifTags = []exif.FieldName{
	exif.Make,
	exif.Model,
	exif.Software,
	exif.GPSAltitudeRef,
	exif.GPSAltitude,
	exif.GPSLatitudeRef,
	exif.GPSLatitude,
	exif.GPSLongitudeRef,
	exif.GPSLongitude,
	exif.GPSDateStamp,
	exif.DateTime,
	exif.DateTimeDigitized,
	exif.DateTimeOriginal,
}

func ExtractExif(file *os.File) (map[string]string, error) {
	x, err := exif.Decode(file)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]string)

	for _, field := range exifTags {
		tag, _ := x.Get(field)
		tagMap[string(field)] = tag.String()
	}

	return tagMap, nil
}
