package main

import (
	"flag"
	"fmt"
	"github.com/echoface/proximityhash"
	"strings"
)

var flagLat float64
var flagLon float64
var flagRadius float64
var flagPrecision int = 7
var flagGeoraptor bool = false
var flagGeoraptorMin int = 3
var flagGeoraptorMax int = 7

func init() {
	flag.Float64Var(&flagLat, "lat", 361, "--lat=xxx.xx")
	flag.Float64Var(&flagLon, "lon", 361, "--lon=xxx.xx")
	flag.Float64Var(&flagRadius, "radius", 1000, "--radius=x(m)")
	flag.IntVar(&flagPrecision, "chars", 7, "--chars=7")
	flag.BoolVar(&flagGeoraptor, "georaptor", false, "--georaptor=xxx.xx")
	flag.IntVar(&flagGeoraptorMin, "min", 3, "--min=v  need: 1 < v < 12")
	flag.IntVar(&flagGeoraptorMax, "max", 7, "--max=v  need: min < v < 12")
}

func validLongitude(v float64) bool {
	return v >= -180 && v <= 180
}
func validLatitude(v float64) bool {
	return v >= -90 && v <= 90
}

func main() {
	flag.Parse()

	fmt.Println(flagLat, flagLon, flagRadius, flagPrecision, flagGeoraptor, flagGeoraptorMin, flagGeoraptorMax)
	proximityhash.PanicIf(!validLongitude(flagLon), "lon need in [-180, 180]")
	proximityhash.PanicIf(!validLongitude(flagLat), "lat need in [-90, 90]")
	codes := proximityhash.CreateGeohash(flagLat, flagLon, flagRadius, uint(flagPrecision))
	if !flagGeoraptor {
		fmt.Println(strings.Join(codes, ","))
		fmt.Println("total geohash code count:", len(codes))
		return
	}
	before := len(codes)

	codes = proximityhash.CompressGeoHash(codes, flagGeoraptorMin, flagGeoraptorMax)
	fmt.Println(strings.Join(codes, ","))
	fmt.Println("total geohash code count:", len(codes))
	fmt.Println("compression reduce codes:", before-len(codes))
}
