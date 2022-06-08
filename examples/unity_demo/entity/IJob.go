package entity

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type IJob interface {
	Attack(victim *entity.Entity)

	Cast(victim *entity.Entity)

	ChangeATK(atk int64)
}

type Job struct {
	AttackRange float64

	Atk int64
}

func (j *Job) Attack(victim *entity.Entity) {
	gwlog.Infof("job attack")
}

func (j *Job) Cast(victim *entity.Entity) {
	gwlog.Infof(" job cast")

}
func (j *Job) ChangeATK(atk int64) {
	gwlog.Infof(" job change atk")

}
