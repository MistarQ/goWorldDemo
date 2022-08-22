package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/entity"
	"time"
)

const (
	AOE              = 1 // 全屏aoe
	IRON             = 2 // 钢铁 需远离
	MOON             = 3 // 月环 需靠近
	DeathPenaltyAOE  = 4 // aoe死刑
	LineBlackHole    = 5 // 直线aoe 附带黑洞
	Apportion        = 6 // 分摊
	LineDeathPenalty = 7 // 接线死刑
)

type Skill struct {
	name         string
	power        int64
	caster       *entity.Entity
	Position     goworld.Vector3
	skillType    int32
	castTime     time.Duration
	delayTime    time.Duration
	startTIme    time.Time
	durationTime time.Duration
	targets      []*entity.Entity
	radius       entity.Coord
}

const (
	CAST     = 1 // 施法
	ULTIMATE = 2 // 大招
)

func CalcDmgFactor(skillType int) int64 {
	switch skillType {
	case CAST:
		return 1
	case ULTIMATE:
		return 5
	}
	return 0
}
