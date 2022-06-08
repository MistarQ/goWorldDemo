package entity

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type Mage struct {
	Job
}

func (m *Mage) Attack(victim *entity.Entity) {
	gwlog.Infof("mage attack")
}

func (m *Mage) Cast(victim *entity.Entity) {
	gwlog.Infof("check job", m.Job)
	if victim.TypeName != "Monster" {
		return
	}
	monster := victim.I.(*Monster)
	gwlog.Infof("mage cast")
	monster.TakeDamage(m.Job.Atk)
}

func (m *Mage) ChangeATK(atk int64) {
	m.Atk = 50
}
