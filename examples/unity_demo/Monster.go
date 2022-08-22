package main

import (
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/action"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/eType"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/prop"
	"github.com/xiaonanln/goworld/examples/unity_demo/utils"
	"math"
	"time"
)

// Monster type
type Monster struct {
	entity.Entity   // Entity type should always inherit entity.Entity
	movingToTarget  *entity.Entity
	attackingTarget *entity.Entity
	// castingTarget   *entity.Entity
	lastTickTime time.Time

	attackCD       time.Duration
	lastAttackTime time.Time

	isCasting bool

	buffList []*Buff

	radius int64

	BattleStarted bool

	tick int

	castSKill *Skill
}

func (monster *Monster) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetUseAOI(true, 100)
	desc.DefineAttr(prop.NAME, "AllClients")
	desc.DefineAttr(prop.Level, "AllClients")
	desc.DefineAttr(prop.Hp, "AllClients")
	desc.DefineAttr(prop.HpMax, "AllClients")
	desc.DefineAttr(prop.Action, "AllClients")
	desc.DefineAttr(prop.Radius, "AllClients")
}

func (monster *Monster) OnCreated() {
	monster.Attrs.SetDefaultInt(prop.Radius, 3)
	monster.setDefaultAttrs()
	gwlog.Infof("monster created", monster)
}

func (monster *Monster) OnEnterSpace() {
	monster.AddTimer(time.Millisecond*100, "AI")
}

func (monster *Monster) setDefaultAttrs() {
	monster.Attrs.SetDefaultStr(prop.NAME, "default monster")
	monster.Attrs.SetDefaultInt(prop.Level, 1)
	monster.Attrs.SetDefaultInt(prop.Hp, 1000)
	monster.Attrs.SetDefaultInt(prop.HpMax, 1000)
	monster.Attrs.SetDefaultStr(prop.Action, action.Idle)
	monster.Attrs.SetDefaultInt(prop.Radius, 3)
	monster.attackCD = time.Second
	monster.lastAttackTime = time.Now()
}

func (monster *Monster) AI() {

	var nearestPlayer *entity.Entity
	for entity := range monster.InterestedIn {

		if !eType.IsPlayer(entity.TypeName) {
			continue
		}

		if entity.GetInt(prop.Hp) <= 0 {
			// dead
			continue
		}

		if nearestPlayer == nil || nearestPlayer.DistanceTo(&monster.Entity) > entity.DistanceTo(&monster.Entity) {
			nearestPlayer = entity
		}
	}

	if !monster.BattleStarted && nearestPlayer != nil && eType.IsPlayer(nearestPlayer.TypeName) && nearestPlayer.DistanceTo(&monster.Entity) <= 8 {
		gwlog.Infof("start battle ", monster.Position, nearestPlayer.Position)
		monster.startBattle()
	}

	if !monster.BattleStarted {
		return
	}

	if nearestPlayer == nil {
		monster.Idling()
		return
	}

	// 施法时无动作
	if monster.isCasting {
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
	monster.Attrs.SetStr(prop.Action, action.Idle)
}

func (monster *Monster) MovingTo(player *entity.Entity) {
	if monster.movingToTarget == player {
		// moving targets not changed
		return
	}

	monster.movingToTarget = player
	monster.attackingTarget = nil
	monster.Attrs.SetStr(prop.Action, action.Move)
}

func (monster *Monster) Attacking(player *entity.Entity) {
	if monster.attackingTarget == player {
		return
	}

	monster.movingToTarget = nil
	monster.attackingTarget = player
	monster.Attrs.SetStr(prop.Action, action.Move)
}

func (monster *Monster) attack(player *Player) {
	monster.CallAllClients("DisplayAttack", player.ID)

	if player.GetInt(prop.Hp) <= 0 {
		return
	}

	player.TakeDamage(monster.GetDamage())
}

func (monster *Monster) GetDamage() int64 {
	return 1
}

func (monster *Monster) TakeDamage(damage int64, isCrit bool) {
	if !monster.BattleStarted {
		monster.startBattle()
	}

	hp := monster.GetInt(prop.Hp)
	hp = hp - damage
	if hp < 0 {
		hp = 0
	}

	monster.Attrs.SetInt(prop.Hp, hp)
	gwlog.Infof("%s TakeDamage %d => Hp=%d", monster, damage, hp)
	if hp <= 0 {
		monster.Attrs.SetStr(prop.Action, action.Death)
		monster.Destroy()
	}
	monster.CallAllClients("DisplayAttacked", monster.ID, isCrit)
}

func (monster *Monster) startBattle() {

	monster.BattleStarted = true
	monster.lastTickTime = time.Now()
	monster.AddTimer(time.Millisecond*30, "Tick")
	// 计算技能
	monster.AddTimer(time.Second, "SkillTimeline")
}

func (monster *Monster) SkillTimeline() {
	// 理论上是从配置文件种读取时间轴配置
	monster.tick += 1

	gwlog.Infof("SkillTimeline, tick is", monster.tick, monster.Position)

	switch monster.tick {
	case 20:
		s0 := &Skill{
			name:         "Hollest of Holy",
			Position:     monster.Position,
			skillType:    AOE,
			castTime:     3 * time.Second,
			delayTime:    0,
			power:        10,
			startTIme:    time.Now(),
			durationTime: 0}
		monster.castSKill = s0
	case 30:
		s1 := &Skill{
			name:         "Empty dimension",
			Position:     monster.Position,
			skillType:    MOON,
			castTime:     3 * time.Second,
			delayTime:    0,
			power:        30,
			startTIme:    time.Now(),
			durationTime: 0,
			radius:       3}
		monster.castSKill = s1
	case 40:
		s2 := &Skill{
			name:      "Heaven Blaze",
			skillType: Apportion,
			castTime:  3 * time.Second,
			startTIme: time.Now(),
			radius:    3,
			power:     20,
			targets:   []*entity.Entity{monster.attackingTarget}}
		monster.castSKill = s2
	case 5:
		s3 := &Skill{
			name:         "接线",
			skillType:    LineDeathPenalty,
			Position:     monster.Position,
			caster:       &monster.Entity,
			startTIme:    time.Now(),
			durationTime: 3 * time.Second,
			power:        15,
			targets:      []*entity.Entity{monster.attackingTarget},
		}
		monster.castSKill = s3
	case 50:
		s4 := &Skill{
			name:         "Hyper Dimensional Slash",
			Position:     monster.Position,
			skillType:    LineBlackHole,
			castTime:     3 * time.Second,
			delayTime:    0,
			startTIme:    time.Now(),
			power:        25,
			durationTime: 0,
			targets:      []*entity.Entity{monster.attackingTarget},
		}
		monster.castSKill = s4
	}

	if monster.castSKill != nil {
		gwlog.Infof("castSkill is", monster.castSKill)
		if monster.isCasting {
			monster.castSKill.castTime -= time.Second
		}
		if !monster.isCasting {
			monster.CallAllClients("DisplayCastBar", float32(monster.castSKill.castTime.Seconds()), monster.castSKill.skillType, monster.castSKill.name, monster.ID)
		}
		monster.isCasting = true
		monster.Attrs.SetStr(prop.Action, action.Cast)
		if monster.castSKill.castTime <= 0 {
			monster.castSkill(monster.castSKill)
			if monster.castSKill.durationTime > 0 {
				monster.castSKill.durationTime -= time.Second
				gwlog.Infof("duration time", monster.castSKill.durationTime)
			} else {
				monster.isCasting = false
				monster.castSKill = nil
			}
		}
	}

}

func (monster *Monster) castSkill(skill *Skill) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.Errorf("cast skill error=", err)
		}
	}()

	if monster.IsDestroyed() {
		return
	}
	space := monster.Space

	var players []*Player
	for p, _ := range space.Entities {
		if eType.IsPlayer(p.TypeName) {
			player := p.I.(*Player)
			players = append(players, player)
		}
	}

	switch skill.skillType {
	case AOE:
		for _, p := range players {
			p.TakeDamage(skill.power)
			p.CallAllClients("DisplayAttacked", p.ID)
		}
	case MOON:

		for _, p := range players {
			gwlog.Debugf("skill pos: %s", skill.Position.String(), ", player pos: %s", p.Position.String())
			if p.Position.DistanceTo2D(skill.Position) < skill.radius {
				continue
			}
			p.TakeDamage(skill.power)
			p.CallAllClients("DisplayAttacked", p.ID)
		}
	case DeathPenaltyAOE:
		if skill.targets == nil {
			return
		}

		for _, e := range skill.targets {
			target := e.I.(*Player)
			target.TakeDamage(skill.power)
			target.CallAllClients("DisplayAttacked", target.ID)
			for _, p := range players {
				if eType.IsPlayer(p.TypeName) {
					continue
				}
				player := p.I.(*Player)
				if player.Position.DistanceTo2D(target.Position) > skill.radius {
					continue
				}
				player.TakeDamage(skill.power)
				player.CallAllClients("DisplayAttacked", player.ID)
			}
		}
	case LineBlackHole:
		for _, e := range skill.targets {
			gwlog.Debugf("line black hole", e.Attrs)
			monster.CallAllClients("DisPlayLine", monster.Position.X, monster.Position.Z, e.Position.X, e.Position.Z)

			X := float32(e.Position.X - monster.Position.X)
			Z := float32(e.Position.Z - monster.Position.Z)

			vecX, vecZ := utils.Normalize(X, Z)
			for X < 100 && X > -100 && Z < 100 && Z > -100 {
				X += vecX
				Z += vecZ
			}
			gwlog.Infof("black hole pos, %f, %f", X, Z)
			// monster.Space.CreateEntity("BlackHole", entity.Vector3{X: entity.Coord(X), Z: entity.Coord(Z)})
			player := e.I.(*Player)
			gwlog.Infof("monster yaw %f", monster.GetYaw())
			gwlog.Infof("dis yaw %f ", (player.Position.Sub(monster.Position)).DirToYaw())
			// 计算矩形
			pointList := utils.CalcMatrix(monster.Position, (player.Position.Sub(monster.Position)).DirToYaw(), 2, 10)

			monster.CallAllClients("DisplayMatrix", pointList[0].X, pointList[0].Y, pointList[0].Z,
				pointList[1].X, pointList[1].Y, pointList[1].Z,
				pointList[2].X, pointList[2].Y, pointList[2].Z,
				pointList[3].X, pointList[3].Y, pointList[3].Z)
			gwlog.Infof("pointList", pointList)

			for _, p := range players {
				gwlog.Infof("position", p.Position)
				if utils.CalcInMatrix(pointList, p.Position) {
					p.TakeDamage(skill.power)
					p.CallAllClients("DisplayAttacked", p.ID)
				}
			}

			gwlog.Infof("DisPlayLineAttacked %s", player.ID)
		}

	case Apportion:
		for _, e := range skill.targets {
			var playerList []*Player

			if !eType.IsPlayer(e.TypeName) {
				continue
			}
			playerList = append(playerList, e.I.(*Player))
			for otherE := range monster.Space.Entities {
				if !eType.IsPlayer(otherE.TypeName) {
					continue
				}
				if otherE.ID == e.ID {
					continue
				}
				if otherE.Position.DistanceTo2D(e.Position) > skill.radius {
					continue
				}
				otherP := e.I.(*Player)
				playerList = append(playerList, otherP)
			}
			if playerList != nil {
				skill.power /= int64(len(playerList))
				for _, p := range playerList {
					p.TakeDamage(skill.power)
					p.CallAllClients("DisplayAttacked", p.ID)
				}
			}
		}
	case LineDeathPenalty:
		if monster.IsDestroyed() {
			return
		}
		if skill.targets == nil || len(skill.targets) <= 0 {
			gwlog.Errorf("line death penalty failed, no targets", skill)
			return
		}
		e := skill.targets[0]
		monster.FaceTo(e)
		oldDistance := monster.DistanceTo(e)
		oldYaw := monster.GetYaw()

		for e := range monster.InterestedIn {
			if !eType.IsPlayer(e.TypeName) {
				continue
			}
			yaw := e.Position.Sub(monster.Position).DirToYaw()
			if math.Abs(float64(yaw-oldYaw)) <= 10 && e.Position.DistanceTo(monster.Position) <= oldDistance {
				skill.targets = []*entity.Entity{e}
				gwlog.Debugf("monster cast target %s", e.I.(*Player).Attrs.GetStr(prop.NAME))
			}
		}
		if skill.durationTime <= 0 && skill.targets != nil && len(skill.targets) > 0 {
			gwlog.Debugf("monster cast line finished", skill.targets[0])
			position := skill.targets[0].Position
			for e := range monster.Space.Entities {
				if !eType.IsPlayer(e.TypeName) {
					continue
				}

				p := e.I.(*Player)
				if p.Position.DistanceTo(position) > 3 {
					continue
				}
				p.TakeDamage(skill.power)
				p.CallAllClients("DisplayAttacked", p.ID)
			}
			return
		}
	}
}
