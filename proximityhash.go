package geoholder

import (
	"fmt"
	"math"

	"github.com/mmcloughlin/geohash"
)

func InCircleCheck(latitude, longitude, centreLat, centreLon, radius float64) bool {
	xDiff := longitude - centreLon
	yDiff := latitude - centreLat

	if math.Pow(xDiff, 2)+math.Pow(yDiff, 2) <= math.Pow(radius, 2) {
		return true
	}

	return false
}

func getCentroid(latitude, longitude, height, width float64) (float64, float64) {
	yCen := latitude + (height / 2)
	xCen := longitude + (width / 2)

	return xCen, yCen
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
	gridWidth = [12]float64{5009400.0, 1252300.0, 156500.0, 39100.0, 4900.0, 1200.0, 152.9, 38.2, 4.8, 1.2, 0.149, 0.0370}
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
			if InCircleCheck(tempLat, tempLon, 0, 0, radius) {
				var lat, lon float64
				xCen, yCen := getCentroid(tempLat, tempLon, height, width)

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

// CompressGeoHash 要求输入的level都是一样的？
func CompressGeoHash(codes []string, min int) []string {
	if len(codes) == 0 || min <= 0 {
		return codes
	}

	codesMap := map[string]struct{}{}
	codeLevelMax := 0
	for _, code := range codes {
		codesMap[code] = struct{}{}
		codeLevelMax = MaxInt(len(code), codeLevelMax)
	}

	if codeLevelMax <= min {
		return codes
	}

	results := make([]string, 0, len(codes)/2)
	// 逐一尝试从maxLevel 向 maxLevel - 1
	// i: 合并后的目标长度
	// 将 codeLevelMax长度的code合并成长度为i的方块
	type prefixData struct {
		prefixCnt       int
		samePrefixCodes []string
	}

	for i := codeLevelMax - 1; i >= min && len(codesMap) >= 32; i-- {

		mergeLevel := i + 1
		cntMap := map[string]*prefixData{}

		for code := range codesMap {
			if len(code) == mergeLevel {
				shorterCode := code[:i]
				data, ok := cntMap[shorterCode]
				if !ok {
					data = &prefixData{}
					cntMap[shorterCode] = data
				}
				data.prefixCnt++
				data.samePrefixCodes = append(data.samePrefixCodes, code)

			} else if len(code) > mergeLevel { // 复用逻辑,大于目标合并长度的直接添加到结果集中，并从map中删除
				delete(codesMap, code)
				results = append(results, code)
			}
		}
		// 如果shorterCode的数量有32个说明可以合并成更大的方块
		// 不足32个则说明这些方块无法做任何合并,直接加入结果集
		for shortCode, data := range cntMap {
			var merged bool
			for _, code := range data.samePrefixCodes {
				delete(codesMap, code)

				if data.prefixCnt >= 32 && !merged { // 合并方块
					merged = true
					codesMap[shortCode] = struct{}{}
				} else if data.prefixCnt < 32 { // 不能合并成大方块，则直接加入结果集
					results = append(results, code)
				} else {
					// >= 32 && merged, do nothing
				}
			}
		}
	}
	for code := range codesMap {
		results = append(results, code)
	}
	return results
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
