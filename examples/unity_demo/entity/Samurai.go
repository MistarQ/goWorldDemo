package entity

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type Samurai struct {
	Job
}

func (s *Samurai) Attack(victim *entity.Entity) {
	gwlog.Infof("Samurai attack")
}

func (s *Samurai) Cast(victim *entity.Entity) {
	gwlog.Infof("Samurai cast")
}
