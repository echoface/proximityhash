package proximityhash

import (
	"fmt"
	"math"

	"github.com/mmcloughlin/geohash"
)

func inCircleCheck(latitude, longitude, centreLat, centreLon, radius float64) bool {
	xDiff := longitude - centreLon
	yDiff := latitude - centreLat

	return math.Pow(xDiff, 2)+math.Pow(yDiff, 2) <= math.Pow(radius, 2)
}

func getCentroID(latitude, longitude, height, width float64) (float64, float64) {
	return longitude + (width / 2), latitude + (height / 2)
}

func convertToLatLon(y float64, x float64, latitude float64, longitude float64) (float64, float64) {
	rEarth := 6371000.0

	latDiff := (y / rEarth) * (180 / math.Pi)
	lonDiff := (x / rEarth) * (180 / math.Pi) / math.Cos(latitude*math.Pi/180)

	finalLat := latitude + latDiff
	finalLon := longitude + lonDiff

	if finalLon > 180 {
		finalLon = finalLon - 360
	}
	if finalLat < -180 {
		finalLon = 360 + finalLon
	}

	return finalLat, finalLon
}

var (
	gridWidth  = [12]float64{5009400.0, 1252300.0, 156500.0, 39100.0, 4900.0, 1200.0, 152.9, 38.2, 4.8, 1.2, 0.149, 0.0370}
	gridHeight = [12]float64{4992600.0, 624100.0, 156000.0, 19500.0, 4900.0, 609.4, 152.4, 19.0, 4.8, 0.595, 0.149, 0.0199}
)

// CreateGeohash Get the list of geohashed id(uint64) that approximate a circle
func CreateGeohash(latitude, longitude, radius float64, precision uint) []string {
	if precision > 12 || precision == 0 {
		panic(fmt.Errorf("invalid precision:%d", precision))
	}

	geohashes := make([]string, 0, 128)

	height := (gridHeight[precision-1]) / 2
	width := (gridWidth[precision-1]) / 2

	latMoves := int(math.Ceil(radius / height))
	lonMoves := int(math.Ceil(radius / width))

	for i := 0; i < latMoves; i++ {

		tempLat := height * float64(i)
		for j := 0; j < lonMoves; j++ {

			tempLon := width * float64(j)
			if inCircleCheck(tempLat, tempLon, 0, 0, radius) {
				var lat, lon float64
				xCen, yCen := getCentroID(tempLat, tempLon, height, width)

				lat, lon = convertToLatLon(yCen, xCen, latitude, longitude)
				geohashes = append(geohashes, geohash.EncodeWithPrecision(lat, lon, precision))

				lat, lon = convertToLatLon(-yCen, xCen, latitude, longitude)

				geohashes = append(geohashes, geohash.EncodeWithPrecision(lat, lon, precision))

				lat, lon = convertToLatLon(yCen, -xCen, latitude, longitude)
				geohashes = append(geohashes, geohash.EncodeWithPrecision(lat, lon, precision))

				lat, lon = convertToLatLon(-yCen, -xCen, latitude, longitude)
				geohashes = append(geohashes, geohash.EncodeWithPrecision(lat, lon, precision))
			}
		}
	}
	return geohashes
}

/*
# Combination generator for a given geohash at the next level
def get_combinations(string):
base32 = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "b", "c", "d", "e", "f", "g", "h", "j", "k", "m",
"n", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"]
return [string + "{0}".format(i) for i in base32]
*/
var base32 = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"b", "c", "d", "e", "f", "g", "h", "j", "k", "m", "n",
	"p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
}

func GenNextPrecisionGeoHashCode(hashCode string) (results []string) {
	for _, char := range base32 {
		results = append(results, hashCode+char)
	}
	return
}

// CompressGeoHash merge geohash code as far as possible,
// minPrecision: minimum length of geohash code after merge
// cutoffPrecision: cutoff precision geohash codes that can't be merged
func CompressGeoHash(codes []string, minPrecision, cutoffPrecision int) []string {
	PanicIf(minPrecision < 1 || minPrecision > 12, "need 0 < minPrecision <= 12")
	PanicIf(cutoffPrecision < 1 || cutoffPrecision > 12, "need 0 < cutoffPrecision <= 12")

	if len(codes) == 0 {
		return codes
	}

	codeLevelMax := 0
	codesMap := map[string]struct{}{} // map<code, finalResult>
	for _, code := range codes {
		codesMap[code] = struct{}{}
		codeLevelMax = MaxInt(len(code), codeLevelMax)
	}

	if codeLevelMax <= minPrecision {
		return codes
	}
	cutoffPrecision = MinInt(codeLevelMax, cutoffPrecision)

	resultMap := map[string]struct{}{}

	// try compress from maxLevel to maxLevel-1 level util minPrecision reached
	for targetPrecision := codeLevelMax - 1; targetPrecision >= minPrecision && len(codesMap) >= 32; targetPrecision-- {

		candidatePrecision := targetPrecision + 1
		targetPrecisionData := map[string][]string{}

		for code := range codesMap {
			if codeLen := len(code); codeLen == candidatePrecision {
				shorterCode := code[:targetPrecision]
				targetPrecisionData[shorterCode] = append(targetPrecisionData[shorterCode], code)
			} else if codeLen > candidatePrecision {
				delete(codesMap, code)
				if codeLen > cutoffPrecision {
					code = code[:cutoffPrecision]
				}
				resultMap[code] = struct{}{}
			}
		}

		// if shortCode has a 32 subset geohash code,
		// it's meaning can merge into a bigger one
		for targetPrecisionCode, subSetCodes := range targetPrecisionData {
			canMerge := len(subSetCodes) >= 32
			if canMerge {
				codesMap[targetPrecisionCode] = struct{}{}
			}

			for _, code := range subSetCodes {
				delete(codesMap, code)

				if !canMerge {
					if len(code) > cutoffPrecision {
						code = code[:cutoffPrecision]
					}
					resultMap[code] = struct{}{}
				}
			}
		}
	}
	for code := range codesMap {
		if len(code) > cutoffPrecision {
			code = code[:cutoffPrecision]
		}
		resultMap[code] = struct{}{}
	}
	results := make([]string, 0, len(codes)/2)
	for code := range resultMap {
		results = append(results, code)
	}
	return results
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func PanicIf(ok bool, ss string, v ...interface{}) {
	if !ok {
		return
	}
	panic(fmt.Errorf(ss, v...))
}
