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
	desc.DefineAttr("name", "AllClients")
	desc.DefineAttr("lv", "AllClients")
	desc.DefineAttr("hp", "AllClients")
	desc.DefineAttr("hpmax", "AllClients")
	desc.DefineAttr("action", "AllClients")
	desc.DefineAttr("radius", "AllClients")
}

func (monster *Monster) OnCreated() {
	monster.Attrs.SetDefaultInt("radius", 3)
	monster.setDefaultAttrs()
	gwlog.Infof("monster created", monster)
}

func (monster *Monster) OnEnterSpace() {
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

func (monster *Monster) getName() string {
	return monster.Attrs.GetStr("name")
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

	if !monster.BattleStarted && nearestPlayer != nil && nearestPlayer.TypeName == "Player" && nearestPlayer.DistanceTo(&monster.Entity) <= 8 {
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
	if !monster.BattleStarted {
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
			power:     0,
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
			power:        1,
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
		monster.Attrs.SetStr("action", "cast")
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
		if p.TypeName == "Player" {
			player := p.I.(*Player)
			players = append(players, player)
		}
	}

	switch skill.skillType {
	case AOE:
		for _, p := range players {
			p.TakeDamage(0)
			p.CallAllClients("DisplayAttacked", p.ID)
		}
	case MOON:

		for _, p := range players {
			gwlog.Debugf("skill pos: %s", skill.Position.String(), ", player pos: %s", p.Position.String())
			if p.Position.DistanceTo2D(skill.Position) < skill.radius {
				continue
			}
			p.TakeDamage(0)
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
			for _, p := range players {
				if p.TypeName != "Player" {
					continue
				}
				player := p.I.(*Player)
				if player.Position.DistanceTo2D(target.Position) > skill.radius {
					continue
				}
				player.TakeDamage(0)
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
					p.TakeDamage(10)
					p.CallAllClients("DisplayAttacked", p.ID)
				}
			}

			gwlog.Infof("DisPlayLineAttacked %s", player.ID)
		}

	case Apportion:
		for _, e := range skill.targets {
			var playerList []*Player

			if e.TypeName != "Player" {
				continue
			}
			playerList = append(playerList, e.I.(*Player))
			for otherE := range monster.Space.Entities {
				if otherE.TypeName != "Player" {
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
				skill.power /= len(playerList)
				for _, p := range playerList {
					p.TakeDamage(int64(skill.power))
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
			if e.TypeName != "Player" {
				continue
			}
			yaw := e.Position.Sub(monster.Position).DirToYaw()
			if math.Abs(float64(yaw-oldYaw)) <= 10 && e.Position.DistanceTo(monster.Position) <= oldDistance {
				skill.targets = []*entity.Entity{e}
				gwlog.Debugf("monster cast target %s", e.I.(*Player).Attrs.GetStr("name"))
			}
		}
		if skill.durationTime <= 0 && skill.targets != nil && len(skill.targets) > 0 {
			gwlog.Debugf("monster cast line finished", skill.targets[0])
			position := skill.targets[0].Position
			for e := range monster.Space.Entities {
				if e.TypeName != "Player" {
					continue
				}

				p := e.I.(*Player)
				if p.Position.DistanceTo(position) > 3 {
					continue
				}
				p.TakeDamage(0)
				p.CallAllClients("DisplayAttacked", p.ID)
			}
			return
		}
	}
}

//func (monster *Monster) lineDeathPenalty() {
//	if monster.IsDestroyed() {
//		return
//	}
//	monster.isCasting = true
//	monster.castingTarget = monster.attackingTarget
//	ticker := time.NewTicker(300 * time.Millisecond)
//	count := 0
//	defer func() {
//		monster.isCasting = false
//		ticker.Stop()
//		err := recover()
//		if err != nil {
//			gwlog.Errorf("line death penalty error=", err)
//		}
//	}()
//
//	for {
//		select {
//		case <-ticker.C:
//			if monster.IsDestroyed() {
//				return
//			}
//			gwlog.Debugf("monster cast line", monster.castingTarget)
//			if monster.castingTarget != nil {
//				monster.FaceTo(monster.castingTarget)
//				gwlog.Debugf("monster yaw", monster.GetYaw())
//			}
//			oldDistance := monster.DistanceTo(monster.castingTarget)
//			oldYaw := monster.GetYaw()
//
//			for e := range monster.InterestedIn {
//				if e.TypeName != "Player" {
//					continue
//				}
//				yaw := e.Position.Sub(monster.Position).DirToYaw()
//				if math.Abs(float64(yaw-oldYaw)) <= 10 && e.Position.DistanceTo(monster.Position) <= oldDistance {
//					monster.castingTarget = e
//					gwlog.Debugf("monster cast target %s", monster.castingTarget.Attrs.GetStr("name"))
//				}
//			}
//
//			count += 1
//			if count >= 10 {
//				if monster.castingTarget != nil {
//					gwlog.Debugf("monster cast line finished", monster.castingTarget)
//					player := monster.castingTarget.I.(*Player)
//					player.TakeDamage(0)
//					player.CallAllClients("DisplayAttacked", player.ID)
//
//					for e := range monster.Space.Entities {
//						if e.TypeName != "Player" {
//							continue
//						}
//						if e.ID == player.ID {
//							continue
//						}
//						p := e.I.(*Player)
//						if p.Position.DistanceTo(player.Position) > 3 {
//							continue
//						}
//						p.TakeDamage(0)
//						p.CallAllClients("DisplayAttacked", p.ID)
//					}
//				}
//				return
//			}
//		}
//	}
//
//}
