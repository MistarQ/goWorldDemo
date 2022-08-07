package utils

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"math"
)

func CalcMatrix(vec entity.Vector3, yaw entity.Yaw, width float32, length float32) (pointList []entity.Vector3) {

	// 角动量
	yaw = yaw * math.Pi / 180
	width /= 2
	// 单位向量
	unitVec := entity.Vector3{X: entity.Coord(math.Sin(float64(yaw))), Z: entity.Coord(math.Cos(float64(yaw)))}
	// 顺时针90°
	unitVecP90 := entity.Vector3{X: unitVec.Z, Z: -unitVec.X}
	// 逆时针90°
	// UnitVecM90 := entity.Vector3{X: -unitVec.Z, Z: unitVec.X}
	l := entity.Vector3{X: unitVec.X * entity.Coord(length), Z: unitVec.Z * entity.Coord(length)}
	w := entity.Vector3{X: unitVecP90.X * entity.Coord(width), Z: unitVecP90.Z * entity.Coord(width)}

	pointList = append(pointList, vec.Add(l).Add(w))
	pointList = append(pointList, vec.Add(w))
	pointList = append(pointList, vec.Sub(w))
	pointList = append(pointList, vec.Add(l).Sub(w))

	gwlog.Infof("calcMatrix, vec %s, yaw %f, width %f, length %f, pointList %s", vec, yaw, width, length, pointList)
	return pointList
}

func CalcInMatrix(pointList []entity.Vector3, point entity.Vector3) bool {
	vec1 := pointList[1].Sub(pointList[0]) // AB
	vec2 := pointList[2].Sub(pointList[1]) // BC
	vec3 := pointList[3].Sub(pointList[2]) // CD
	vec4 := pointList[0].Sub(pointList[3]) // DA

	pointVec1 := point.Sub(pointList[0]) // OA
	pointVec2 := point.Sub(pointList[1]) // OB
	pointVec3 := point.Sub(pointList[2]) // OC
	pointVec4 := point.Sub(pointList[3]) // OD

	result1 := math.Trunc(float64(vec1.VectorProduct(pointVec1)*100)) / 100
	result2 := math.Trunc(float64(vec2.VectorProduct(pointVec2)*100)) / 100
	result3 := math.Trunc(float64(vec3.VectorProduct(pointVec3)*100)) / 100
	result4 := math.Trunc(float64(vec4.VectorProduct(pointVec4)*100)) / 100

	if (result1 >= 0 && result2 >= 0 && result3 >= 0 && result4 >= 0) ||
		(result1 <= 0 && result2 <= 0 && result3 <= 0 && result4 <= 0) {
		return true
	} else {
		return false
	}
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
