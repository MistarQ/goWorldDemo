package main

import (
	"fmt"
	"github.com/xiaonanln/goworld/engine/entity"
	"testing"
)

func TestName(t *testing.T) {
	pointList := calcMatrix(entity.Vector3{X: 0.5, Z: 0}, 90, 1, 1)

	// fmt.Printf("%t ", calcInMatrix(pointList, entity.Vector3{X: 0.6, Z:0}))
	for x := -0.1; x <= 1.1; x += 0.1 {
		for y := 1.1; y >= -0.1; y -= 0.1 {
			fmt.Printf("%t ", calcInMatrix(pointList, entity.Vector3{X: entity.Coord(x), Z: entity.Coord(y)}))

		}
		fmt.Println("")
	}
}
