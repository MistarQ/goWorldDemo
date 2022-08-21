package main

import (
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"strconv"
	"time"
)

const (
	_SPACE_DESTROY_CHECK_INTERVAL = time.Minute * 5
)

// MySpace is the custom space type
type MySpace struct {
	goworld.Space // Space type should always inherit from entity.Space

	destroyCheckTimer entity.EntityTimerID
}

// OnSpaceCreated is called when the space is created
func (space *MySpace) OnSpaceCreated() {
	// notify the SpaceService that it's ok
	space.EnableAOI(100)

	goworld.CallServiceShardKey("SpaceService", strconv.Itoa(space.Kind), "NotifySpaceLoaded", space.Kind, space.ID)
	space.AddTimer(time.Second*5, "DumpEntityStatus")
	// space.AddTimer(time.Second*5, "SummonMonsters")
	space.AddTimer(time.Second, "EdgeDetection")
	space.AddTimer(time.Second*5, "ResetScene")
	space.SummonMonsters()
	// space.AddTimer(time.Second*5, "SummonMonsters")

}

func (space *MySpace) DumpEntityStatus() {
	space.ForEachEntity(func(e *entity.Entity) {
		gwlog.Debugf(">>> %s @ position %s, neighbors=%d", e, e.GetPosition(), len(e.InterestedIn))
	})
}

func (space *MySpace) SummonMonsters() {
	monster1 := space.CreateEntity("Monster", entity.Vector3{5, 0, 5})
	monster1.Attrs.SetStr("name", "Ser Grinnaux")
	monster1.SetYaw(180)
	monster2 := space.CreateEntity("Monster", entity.Vector3{-5, 0, 5})
	monster2.SetYaw(180)
	monster2.Attrs.SetStr("name", "Ser Adelphel")
	gwlog.Infof("SummerMonsters", monster1.Position, monster2.Position, monster1.I.(*Monster).BattleStarted, monster2.I.(*Monster).BattleStarted)
}

func (space *MySpace) EdgeDetection() {
	for e := range space.Entities {
		if e.TypeName == "Player" {

			if (e.Position.X > 20) ||
				e.Position.X < -20 ||
				e.Position.Z > 20 ||
				e.Position.Z < -20 {

				e.I.(*Player).TakeDamage(99999)
			}
		}
	}
}

func (space *MySpace) ResetScene() {
	for e := range space.Entities {
		if e.TypeName == "Player" {
			if e.I.(*Player).Attrs.GetBool("alive") {
				return
			}
		}
	}

	for e := range space.Entities {
		if e.TypeName == "Player" {
			p := e.I.(*Player)
			p.ResetAttr()
			p.CallAllClients("ResetCoord", p.Position.X, p.Position.Y, p.Position.Z)
		}
	}
	for e := range space.Entities {
		if e.TypeName == "Monster" {
			e.I.(*Monster).Destroy()
		}
	}
	space.SummonMonsters()
}

// OnEntityEnterSpace is called when entity enters space
func (space *MySpace) OnEntityEnterSpace(entity *entity.Entity) {
	if entity.TypeName == "Player" {
		space.onPlayerEnterSpace(entity)
	}
}

func (space *MySpace) onPlayerEnterSpace(entity *entity.Entity) {
	gwlog.Debugf("Player %s enter space %s, total avatar count %d", entity, space, space.CountEntities("Player"))
	gwlog.Debugf("player position", entity.Position)
	space.clearDestroyCheckTimer()
}

// OnEntityLeaveSpace is called when entity leaves space
func (space *MySpace) OnEntityLeaveSpace(entity *entity.Entity) {
	if entity.TypeName == "Player" {
		space.onPlayerLeaveSpace(entity)
	}
}

func (space *MySpace) onPlayerLeaveSpace(entity *entity.Entity) {
	gwlog.Infof("Player %s leave space %s, left avatar count %d", entity, space, space.CountEntities("Player"))
	if space.CountEntities("Player") == 0 {
		// no avatar left, start destroying space
		space.setDestroyCheckTimer()
	}
}

func (space *MySpace) setDestroyCheckTimer() {
	if space.destroyCheckTimer != 0 {
		return
	}

	space.destroyCheckTimer = space.AddTimer(_SPACE_DESTROY_CHECK_INTERVAL, "CheckForDestroy")
}

// CheckForDestroy checks if the space should be destroyed
func (space *MySpace) CheckForDestroy() {
	avatarCount := space.CountEntities("Player")
	if avatarCount != 0 {
		gwlog.Panicf("Player count should be 0, but is %d", avatarCount)
	}

	goworld.CallServiceShardKey("SpaceService", strconv.Itoa(space.Kind), "RequestDestroy", space.Kind, space.ID)
}

func (space *MySpace) clearDestroyCheckTimer() {
	if space.destroyCheckTimer == 0 {
		return
	}

	space.CancelTimer(space.destroyCheckTimer)
	space.destroyCheckTimer = 0
}

// ConfirmRequestDestroy is called by SpaceService to confirm that the space
func (space *MySpace) ConfirmRequestDestroy(ok bool) {
	if ok {
		if space.CountEntities("Player") != 0 {
			gwlog.Panicf("%s ConfirmRequestDestroy: avatar count is %d", space, space.CountEntities("Player"))
		}
		space.Destroy()
	}
}

// OnGameReady is called when the game server is ready
func (space *MySpace) OnGameReady() {
	timer.AddCallback(time.Millisecond*1000, checkServerStarted)
}

func checkServerStarted() {
	ok := isAllServicesReady()
	gwlog.Infof("checkServerStarted: %v", ok)
	if ok {
		onAllServicesReady()
	} else {
		timer.AddCallback(time.Millisecond*1000, checkServerStarted)
	}
}

func isAllServicesReady() bool {
	for _, serviceName := range _SERVICE_NAMES {
		if !goworld.CheckServiceEntitiesReady(serviceName) {
			gwlog.Infof("%s entities are not ready ...", serviceName)
			return false
		}
	}
	return true
}

func onAllServicesReady() {
	gwlog.Infof("All services are ready!")
}
