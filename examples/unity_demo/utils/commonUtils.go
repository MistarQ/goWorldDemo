package utils

import (
	"bytes"
	"fmt"
	"github.com/xiaonanln/goworld"
	"math"
	"runtime"
)

func CalcDistance(entity1 goworld.Vector3, entity2 goworld.Vector3) float64 {
	x := math.Pow(float64(entity1.X-entity2.X), 2)
	z := math.Pow(float64(entity1.Z-entity2.Z), 2)

	return math.Sqrt(x + z)
}

func PrintStackTrace(err interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", err)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}

func Normalize(x float32, z float32) (normalizedX float32, normalizedZ float32) {
	if x == 0 {
		if z > 0 {
			return 0, 1
		} else {
			return 0, -1
		}

	} else if z == 0 {
		if x > 0 {
			return 1, 0
		} else {
			return -1, 0
		}
	} else {
		idx := float32(math.Sqrt(float64(x*x + z*z)))
		return x / idx, z / idx
	}

}
