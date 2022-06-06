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
