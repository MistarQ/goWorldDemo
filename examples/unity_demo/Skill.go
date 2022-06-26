package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/entity"
	"time"
)

const (
	AOE             = 1
	IRON            = 2
	MOON            = 3
	DeathPenaltyAOE = 4
	LineBlackHole   = 5
	Apportion       = 6
)

type Skill struct {
	name         string
	power        int
	Position     goworld.Vector3
	skillType    int32
	castTime     time.Duration
	delayTime    time.Duration
	startTIme    time.Time
	durationTime time.Duration
	targets      []*entity.Entity
	radius       entity.Coord
}
