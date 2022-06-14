package main

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"time"
)

const (
	DOT = 1
)

type Buff struct {
	name     string
	duration time.Duration
	effect   int
	receiver *entity.Entity
	taker    *entity.Entity
}
