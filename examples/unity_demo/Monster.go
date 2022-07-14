package main

import (
	"fmt"
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
	defer func() { //defer就是把匿名函数压入到defer栈中，等到执行完毕后或者发生异常后调用匿名函数
		err := recover() //recover是内置函数，可以捕获到异常
		if err != nil {  //说明有错误
			gwlog.Errorf("cast skill error=", err)
			//当然这里可以把错误的详细位置发送给开发人员
			//send email to admin
		}
	}()

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

	defer func() { //defer就是把匿名函数压入到defer栈中，等到执行完毕后或者发生异常后调用匿名函数
		ticker.Stop()
		err := recover() //recover是内置函数，可以捕获到异常
		if err != nil {  //说明有错误
			gwlog.Errorf("line death penalty error=", err)
			//当然这里可以把错误的详细位置发送给开发人员
			//send email to admin
		}
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

func calcMatrix(vec entity.Vector3, yaw entity.Yaw, width float32, length float32) (pointList []entity.Vector3) {

	// 角动量
	yaw = yaw * math.Pi / 180
	width /= 2
	// 单位向量
	unitVec := entity.Vector3{X: entity.Coord(math.Cos(float64(yaw))), Z: entity.Coord(math.Sin(float64(yaw)))}
	// 顺时针90°
	unitVecP90 := entity.Vector3{X: unitVec.Z, Z: -unitVec.X}
	// 逆时针90°
	// UnitVecM90 := entity.Vector3{X: -unitVec.Z, Z: unitVec.X}
	l := entity.Vector3{X: unitVec.X * entity.Coord(length), Z: unitVec.Z * entity.Coord(length)}
	w := entity.Vector3{X: unitVecP90.X * entity.Coord(width), Z: unitVecP90.Z * entity.Coord(width)}

	pointList = append(pointList, vec.Add(l).Add(w))
	pointList = append(pointList, vec.Add(w))
	pointList = append(pointList, vec.Sub(w))
	pointList = append(pointList, vec.Add(l).Sub(w))

	fmt.Println(pointList)
	gwlog.Infof("calcMatrix, vec %s, yaw %f, width %f, length %f, pointList %s", vec, yaw, width, length, pointList)
	return pointList
}

func calcInMatrix(pointList []entity.Vector3, point entity.Vector3) bool {
	vec1 := pointList[1].Sub(pointList[0]) // AB
	vec2 := pointList[2].Sub(pointList[1]) // BC
	vec3 := pointList[3].Sub(pointList[2]) // CD
	vec4 := pointList[0].Sub(pointList[3]) // DA

	pointVec1 := point.Sub(pointList[0]) // OA
	pointVec2 := point.Sub(pointList[1]) // OB
	pointVec3 := point.Sub(pointList[2]) // OC
	pointVec4 := point.Sub(pointList[3]) // OD

	result1 := math.Trunc(float64(vec1.VectorProduct(pointVec1)*100)) / 100
	result2 := math.Trunc(float64(vec2.VectorProduct(pointVec2)*100)) / 100
	result3 := math.Trunc(float64(vec3.VectorProduct(pointVec3)*100)) / 100
	result4 := math.Trunc(float64(vec4.VectorProduct(pointVec4)*100)) / 100

	if (result1 >= 0 && result2 >= 0 && result3 >= 0 && result4 >= 0) ||
		(result1 <= 0 && result2 <= 0 && result3 <= 0 && result4 <= 0) {
		return true
	} else {
		return false
	}

}
