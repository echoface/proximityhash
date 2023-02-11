package proximityhash

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func explodeGeoHashCodes(codes []string) (vs []string) {
	for _, code := range codes {
		vs = append(vs, GenNextPrecisionGeoHashCode(code)...)
	}
	return vs
}

func TestCreateGeohash(t *testing.T) {
	convey.Convey("test compression", t, func() {
		ids := CreateGeohash(116.334255, 40.027400, 100, 8)
		fmt.Println(ids)
	})
}

func TestCompressGeoHash(t *testing.T) {
	convey.Convey("test compression", t, func() {
		initCode := "wx4g0"
		codes := GenNextPrecisionGeoHashCode(initCode)
		codesNextLevel1 := explodeGeoHashCodes(codes)
		codesNextLevel2 := explodeGeoHashCodes(codesNextLevel1)

		codesNextLevel2 = append(codesNextLevel2, "wx4g1xwc")

		results := CompressGeoHash(codesNextLevel2, 5, 8)
		fmt.Println("result:", results)
		convey.So(len(results), convey.ShouldEqual, 2)
		convey.So(results, convey.ShouldContain, initCode)
		convey.So(results, convey.ShouldContain, "wx4g1xwc")
	})

	convey.Convey("test merge", t, func() {
		results := []string{"tdnu2"}
		combinations := []string{
			"tdnu20", "tdnu21", "tdnu22", "tdnu23", "tdnu24", "tdnu25", "tdnu26", "tdnu27",
			"tdnu28", "tdnu29", "tdnu2b", "tdnu2c", "tdnu2d", "tdnu2e", "tdnu2f", "tdnu2g",
			"tdnu2h", "tdnu2j", "tdnu2k", "tdnu2m", "tdnu2n", "tdnu2p", "tdnu2q", "tdnu2r",
			"tdnu2s", "tdnu2t", "tdnu2u", "tdnu2v", "tdnu2w", "tdnu2x", "tdnu2y", "tdnu2z",
		}
		res := CompressGeoHash(combinations, 3, 7)
		fmt.Println("result:", res)
		convey.So(res, convey.ShouldResemble, results)
	})

	convey.Convey("test cut off", t, func() {
		combinations := []string{
			"tdnu20", "tdnu21", "tdnu22", "tdnu23", "tdnu24", "tdnu25", "tdnu26", "tdnu27",
			"tdnu28", "tdnu29", "tdnu2b", "tdnu2c", "tdnu2d", "tdnu2e", "tdnu2f", "tdnu2g",
			"tdnu2h", "tdnu2j", "tdnu2k", "tdnu2m", "tdnu2n", "tdnu2p", "tdnu2q", "tdnu2r",
			"tdnu2s", "tdnu2t", "tdnu2u", "tdnu2v", "tdnu2w", "tdnu2x", "tdnu2y", "tdnu2z",
			"tdnu2z",   // duplicate
			"wx4g1xwc"} // should be cutoff

		res := CompressGeoHash(combinations, 3, 7)
		fmt.Println("result:", res)
		convey.So(len(res), convey.ShouldEqual, 2)
		convey.So(res, convey.ShouldContain, "tdnu2")
		convey.So(res, convey.ShouldContain, "wx4g1xw")
	})
}
