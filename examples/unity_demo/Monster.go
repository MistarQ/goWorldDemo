package main

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/examples/unity_demo/utils"
	"math"
	"time"
)

// Monster type
type Monster struct {
	entity.Entity   // Entity type should always inherit entity.Entity
	movingToTarget  *entity.Entity
	attackingTarget *entity.Entity
	castingTarget   *entity.Entity
	lastTickTime    time.Time

	attackCD       time.Duration
	lastAttackTime time.Time

	isCasting bool

	buffList []*Buff

	battleStarted bool

	radius int64
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
	gwlog.Infof("monster created", monster)
}

func (monster *Monster) OnEnterSpace() {
	monster.setDefaultAttrs()
	monster.AddTimer(time.Millisecond*100, "AI")
}

func (monster *Monster) setDefaultAttrs() {
	monster.Attrs.SetDefaultStr("name", "minion")
	monster.Attrs.SetDefaultInt("lv", 1)
	monster.Attrs.SetDefaultInt("hpmax", 100)
	monster.Attrs.SetDefaultInt("hp", 100)
	monster.Attrs.SetDefaultStr("action", "idle")
	monster.Attrs.SetDefaultInt("radius", 3)
	monster.attackCD = time.Second
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

	if !monster.battleStarted && nearestPlayer != nil && nearestPlayer.DistanceTo(&monster.Entity) <= 8 {
		monster.startBattle()
	}

	if !monster.battleStarted {
		return
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

	if !monster.battleStarted {
		return
	}

	now := time.Now()

	// 施法时无动作
	if monster.isCasting {
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
		// moving targets not changed
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

func (monster *Monster) TakeDamage(damage int64, isCrit bool) {
	if !monster.battleStarted {
		monster.startBattle()
	}

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
	monster.CallAllClients("DisplayAttacked", monster.ID, isCrit)
}

func (monster *Monster) startBattle() {
	monster.battleStarted = true
	monster.lastTickTime = time.Now()
	monster.AddTimer(time.Millisecond*30, "Tick")
	// 计算技能
	go monster.skillTimeline()
}

func (monster *Monster) skillTimeline() {
	if monster.IsDestroyed() {
		return
	}
	// 理论上是从配置文件种读取时间轴配置
	time.Sleep(10 * time.Second)

	s0 := &Skill{
		name:         "Hollest of Holy",
		Position:     monster.Position,
		skillType:    AOE,
		castTime:     3 * time.Second,
		delayTime:    0,
		startTIme:    time.Now(),
		durationTime: 0}
	go monster.calcSkill(s0)
	time.Sleep(10 * time.Second)

	s1 := &Skill{
		name:         "Empty dimension",
		Position:     monster.Position,
		skillType:    MOON,
		castTime:     3 * time.Second,
		delayTime:    0,
		startTIme:    time.Now(),
		durationTime: 0,
		radius:       3}
	go monster.calcSkill(s1)
	time.Sleep(10 * time.Second)

	s2 := &Skill{
		name:         "HEAVEN BLAZE",
		Position:     monster.Position,
		skillType:    DeathPenaltyAOE,
		castTime:     3 * time.Second,
		delayTime:    0,
		startTIme:    time.Now(),
		durationTime: 0,
		radius:       3,
		targets:      []*entity.Entity{monster.attackingTarget}}
	go monster.calcSkill(s2)
	time.Sleep(10 * time.Second)

	go monster.lineDeathPenalty()
	time.Sleep(10 * time.Second)

	s3 := &Skill{
		name:         "Hyper Dimensional Slash",
		Position:     monster.Position,
		skillType:    LineBlackHole,
		castTime:     3 * time.Second,
		delayTime:    0,
		startTIme:    time.Now(),
		durationTime: 0,
		targets:      []*entity.Entity{monster.attackingTarget},
	}
	go monster.castSkill(s3)
}

func (monster *Monster) calcSkill(skill *Skill) {
	if monster.IsDestroyed() {
		return
	}
	if skill.castTime > 0 {
		monster.CallAllClients("DisplayCastBar", float32(skill.castTime.Seconds()), skill.skillType, skill.name, monster.ID)
		monster.isCasting = true
		monster.Attrs.SetStr("action", "cast")
		time.Sleep(skill.castTime)
		monster.isCasting = false
		monster.Attrs.SetStr("action", "idle")
	}
	if skill.delayTime > 0 {
		time.Sleep(skill.delayTime)
		monster.castSkill(skill)
	} else {
		monster.castSkill(skill)
	}
	// 持续性技能
	if skill.durationTime > 0 {
		monster.durationSkill(skill)
	}
}

func (monster *Monster) durationSkill(skill *Skill) {
	for skill.durationTime > 0 {
		skill.durationTime -= 1
		monster.castSkill(skill)
	}
}

func (monster *Monster) castSkill(skill *Skill) {
	if monster.IsDestroyed() {
		return
	}
	space := monster.Space
	players := space.Entities
	switch skill.skillType {
	case AOE:
		for p, _ := range players {
			if p.TypeName != "Player" {
				continue
			}
			player := p.I.(*Player)
			player.TakeDamage(0)
			p.CallAllClients("DisplayAttacked", p.ID)
		}
	case IRON:
		for p, _ := range players {
			if p.TypeName != "Player" {
				continue
			}
			player := p.I.(*Player)
			if player.Position.DistanceTo2D(skill.Position) > skill.radius {
				continue
			}
			player.TakeDamage(0)
			p.CallAllClients("DisplayAttacked", p.ID)
		}
	case DeathPenaltyAOE:
		if skill.targets == nil {
			return
		}

		for _, e := range skill.targets {
			target := e.I.(*Player)
			target.TakeDamage(0)
			target.CallAllClients("DisplayAttacked", target.ID)
			for p, _ := range players {
				if p.TypeName != "Player" {
					continue
				}
				player := p.I.(*Player)
				if player.Position.DistanceTo2D(target.Position) > skill.radius {
					continue
				}
				player.TakeDamage(0)
				p.CallAllClients("DisplayAttacked", p.ID)
			}
		}
	case LineBlackHole:
		for _, e := range skill.targets {
			monster.CallAllClients("DisPlayLine", monster.Position.X, monster.Position.Z, e.Position.X, e.Position.Z)

			X := float32(e.Position.X - monster.Position.X)
			Z := float32(e.Position.Z - monster.Position.Z)

			vecX, vecZ := utils.Normalize(X, Z)
			for X < 100 && X > -100 && Z < 100 && Z > -100 {
				X += vecX
				Z += vecZ
			}
			monster.Space.CreateEntity("BlackHole", entity.Vector3{X: entity.Coord(X), Z: entity.Coord(Z)})
			target := e.I.(*Player)
			target.TakeDamage(0)
			target.CallAllClients("DisPlayAttacked", target.ID)
		}
	}
}

func (monster *Monster) lineDeathPenalty() {
	if monster.IsDestroyed() {
		return
	}
	monster.castingTarget = monster.attackingTarget
	ticker := time.NewTicker(300 * time.Millisecond)
	count := 0
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if monster.IsDestroyed() {
				return
			}
			monster.isCasting = true
			if monster.castingTarget != nil {
				monster.FaceTo(monster.castingTarget)
			}
			oldDistance := monster.DistanceTo(monster.castingTarget)
			oldYaw := monster.GetYaw()

			for e := range monster.InterestedIn {
				if e.TypeName != "Player" {
					continue
				}
				yaw := e.Position.Sub(monster.Position).DirToYaw()
				if math.Abs(float64(yaw-oldYaw)) <= 10 && e.Position.DistanceTo(monster.Position) < oldDistance {
					monster.castingTarget = e
				}
			}

			count += 1
			if count >= 10 {
				return
			}
		}
	}

}
