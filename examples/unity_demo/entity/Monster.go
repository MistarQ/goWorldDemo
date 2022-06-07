package entity

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/examples/unity_demo/utils"
	"time"
)

// Monster type
type Monster struct {
	entity.Entity   // Entity type should always inherit entity.Entity
	movingToTarget  *entity.Entity
	attackingTarget *entity.Entity
	lastTickTime    time.Time

	attackCD       time.Duration
	lastAttackTime time.Time

	CastCD       time.Duration
	CastRadius   entity.Coord
	lastCastTime time.Time

	skillChan chan *castSkill
}

type castSkill struct {
	name     string
	Position goworld.Vector3
}

func (monster *Monster) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetUseAOI(true, 100)
	desc.DefineAttr("name", "AllClients")
	desc.DefineAttr("lv", "AllClients")
	desc.DefineAttr("hp", "AllClients")
	desc.DefineAttr("hpmax", "AllClients")
	desc.DefineAttr("action", "AllClients")
	desc.DefineAttr("radius", "AllClients")
}

func (monster *Monster) OnCreated() {
	monster.Attrs.SetDefaultInt("radius", 3)
	monster.skillChan = make(chan *castSkill, 5)
	gwlog.Infof("monster created", monster)
}

func (monster *Monster) OnEnterSpace() {
	monster.setDefaultAttrs()
	monster.AddTimer(time.Millisecond*100, "AI")
	monster.lastTickTime = time.Now()
	monster.AddTimer(time.Millisecond*30, "Tick")
	go monster.skillCalc()
}

func (monster *Monster) setDefaultAttrs() {
	monster.Attrs.SetDefaultStr("name", "minion")
	monster.Attrs.SetDefaultInt("lv", 1)
	monster.Attrs.SetDefaultInt("hpmax", 100)
	monster.Attrs.SetDefaultInt("hp", 100)
	monster.Attrs.SetDefaultStr("action", "idle")

	monster.attackCD = time.Second
	monster.lastAttackTime = time.Now()

	monster.CastCD = 10 * time.Second
	monster.CastRadius = 3
	monster.lastAttackTime = time.Now()
}

func (monster *Monster) AI() {
	var nearestPlayer *entity.Entity
	for entity := range monster.InterestedIn {

		if entity.TypeName != "Player" {
			continue
		}

		if entity.GetInt("hp") <= 0 {
			// dead
			continue
		}

		if nearestPlayer == nil || nearestPlayer.DistanceTo(&monster.Entity) > entity.DistanceTo(&monster.Entity) {
			nearestPlayer = entity
		}
	}

	if nearestPlayer == nil {
		monster.Idling()
		return
	}

	if nearestPlayer.DistanceTo(&monster.Entity) > monster.GetAttackRange() {
		monster.MovingTo(nearestPlayer)
	} else {
		monster.Attacking(nearestPlayer)
	}
}

func (monster *Monster) Tick() {

	now := time.Now()

	if !now.Before(monster.lastCastTime.Add(monster.CastCD)) {
		monster.cast(monster.Position)
		monster.lastCastTime = now
		return
	}

	if monster.attackingTarget != nil && monster.IsInterestedIn(monster.attackingTarget) {
		if !now.Before(monster.lastAttackTime.Add(monster.attackCD)) {
			monster.FaceTo(monster.attackingTarget)
			monster.attack(monster.attackingTarget.I.(*Player))
			monster.lastAttackTime = now
		}

	}

	if monster.movingToTarget != nil && monster.IsInterestedIn(monster.movingToTarget) {
		mypos := monster.GetPosition()
		direction := monster.movingToTarget.GetPosition().Sub(mypos)
		direction.Y = 0

		t := direction.Normalized().Mul(monster.GetSpeed() * 30 / 1000.0)
		monster.SetPosition(mypos.Add(t))
		monster.FaceTo(monster.movingToTarget)
		return
	}

}

func (monster *Monster) GetSpeed() entity.Coord {
	return 2
}

func (monster *Monster) GetAttackRange() entity.Coord {
	return 3
}

func (monster *Monster) Idling() {
	if monster.movingToTarget == nil && monster.attackingTarget == nil {
		return
	}

	monster.movingToTarget = nil
	monster.attackingTarget = nil
	monster.Attrs.SetStr("action", "idle")
}

func (monster *Monster) MovingTo(player *entity.Entity) {
	if monster.movingToTarget == player {
		// moving target not changed
		return
	}

	monster.movingToTarget = player
	monster.attackingTarget = nil
	monster.Attrs.SetStr("action", "move")
}

func (monster *Monster) Attacking(player *entity.Entity) {
	if monster.attackingTarget == player {
		return
	}

	monster.movingToTarget = nil
	monster.attackingTarget = player
	monster.Attrs.SetStr("action", "move")
}

func (monster *Monster) cast(position goworld.Vector3) {
	monster.CallAllClients("DisplayCast", monster.ID)
	c := &castSkill{
		name:     "cast",
		Position: position}
	monster.skillChan <- c
}

func (monster *Monster) attack(player *Player) {
	monster.CallAllClients("DisplayAttack", player.ID)

	if player.GetInt("hp") <= 0 {
		return
	}

	player.TakeDamage(monster.GetDamage())
}

func (monster *Monster) GetDamage() int64 {
	return 0
}

func (monster *Monster) TakeDamage(damage int64) {
	hp := monster.GetInt("hp")
	hp = hp - damage
	if hp < 0 {
		hp = 0
	}

	monster.Attrs.SetInt("hp", hp)
	gwlog.Infof("%s TakeDamage %d => hp=%d", monster, damage, hp)
	if hp <= 0 {
		monster.Attrs.SetStr("action", "death")
		monster.Destroy()
	}
	monster.CallAllClients("DisplayAttacked", monster.ID)
}

func (monster *Monster) skillCalc() {
	defer func() {
		if err := recover(); err != nil {

			gwlog.Fatalf("skillCalc", err, utils.PrintStackTrace(err))
		}
	}()
	for {
		select {
		case x := <-monster.skillChan:
			if x.name == "cast" {
				time.Sleep(3 * time.Second)
				space := monster.Space
				players := space.Entities
				for p, _ := range players {
					if p.TypeName != "Player" {
						continue
					}
					player := p.I.(*Player)
					if player.Position.DistanceTo2D(x.Position) > monster.CastRadius {
						continue
					}
					player.TakeDamage(0)
					p.CallAllClients("DisplayAttacked", p.ID)
				}
			}
		}
	}
}
