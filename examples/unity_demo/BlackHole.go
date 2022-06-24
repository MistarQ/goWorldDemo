package main

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"time"
)

type BlackHole struct {
	entity.Entity
}

func (blackHole *BlackHole) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetUseAOI(true, 100)
	desc.DefineAttr("name", "AllClients")

}

func (blackHole *BlackHole) OnCreated() {
	blackHole.Attrs.SetDefaultStr("name", "blackHole")
	gwlog.Infof("blackHole created", blackHole)
}

func (blackHole *BlackHole) OnEnterSpace() {
	time.Sleep(30 * time.Second)
	blackHole.Destroy()
}

func (blackHole *BlackHole) Cast() {

	for e := range blackHole.InterestedIn {
		if e.TypeName != "Player" {
			continue
		}
		player := e.I.(*Player)
		player.TakeDamage(99999)
		player.CallAllClients("DisPlayAttacked", player.ID)
	}
	gwlog.Infof("blackHole cast")
}
