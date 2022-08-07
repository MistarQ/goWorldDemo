package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"math/rand"
	"strconv"
	"time"
)

// Player 对象代表一名玩家
type Player struct {
	entity.Entity
	*rand.Rand
}

func (a *Player) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetPersistent(true).SetUseAOI(true, 100)
	desc.DefineAttr("name", "AllClients", "Persistent")
	desc.DefineAttr("lv", "AllClients", "Persistent")
	desc.DefineAttr("hp", "AllClients")
	desc.DefineAttr("hpmax", "AllClients")
	desc.DefineAttr("action", "AllClients")
	desc.DefineAttr("spaceKind", "Persistent")
	desc.DefineAttr("attackRange", "AllClients")
	desc.DefineAttr("atk", "AllClients")
	desc.DefineAttr("crit", "AllClients")
	desc.DefineAttr("critIndex", "AllClients")
	desc.DefineAttr("alive", "AllClients")
}

// OnCreated 在Player对象创建后被调用
func (a *Player) OnCreated() {
	a.Entity.OnCreated()
	// 应该从account service 获取
	a.setDefaultAttrs()
	a.Rand = rand.New(rand.NewSource(time.Now().Unix()))
}

// setDefaultAttrs 设置玩家的一些默认属性
func (a *Player) setDefaultAttrs() {
	// 应该从account service 获取
	a.Attrs.SetDefaultInt("spaceKind", 1)
	a.Attrs.SetDefaultInt("lv", 1)
	a.Attrs.SetDefaultInt("hp", 100)
	a.Attrs.SetDefaultInt("hpmax", 100)
	a.Attrs.SetDefaultStr("action", "idle")
	a.Attrs.SetDefaultInt("attackRange", 5)
	a.Attrs.SetDefaultInt("atk", 30)
	a.Attrs.SetDefaultInt("crit", 10)
	a.Attrs.SetDefaultInt("critIndex", 2)
	a.Attrs.SetBool("alive", true)
	a.SetClientSyncing(true)
}

func (a *Player) ResetAttr() {
	a.Attrs.SetInt("spaceKind", 1)
	a.Attrs.SetInt("lv", 1)
	a.Attrs.SetInt("hp", 100)
	a.Attrs.SetInt("hpmax", 100)
	a.Attrs.SetStr("action", "idle")
	a.Attrs.SetInt("attackRange", 5)
	a.Attrs.SetInt("atk", 30)
	a.Attrs.SetInt("crit", 10)
	a.Attrs.SetInt("critIndex", 2)
	a.Attrs.SetBool("alive", true)
	a.Position.X = 0
	a.Position.Y = 0
	a.Position.Z = -10
	a.SetClientSyncing(true)
}

// GetSpaceID 获得玩家的场景ID并发给调用者
func (a *Player) GetSpaceID(callerID common.EntityID) {
	a.Call(callerID, "OnGetPlayerSpaceID", a.ID, a.Space.ID)
}

func (p *Player) enterSpace(spaceKind int) {
	if p.Space.Kind == spaceKind {
		return
	}
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s enter space from %d => %d", p, p.Space.Kind, spaceKind)
	}
	goworld.CallServiceShardKey("SpaceService", strconv.Itoa(spaceKind), "EnterSpace", p.ID, spaceKind)
}

// OnClientConnected is called when client is connected
func (a *Player) OnClientConnected() {
	gwlog.Infof("%s client connected", a)
	a.enterSpace(int(a.GetInt("spaceKind")))
}

// OnClientDisconnected is called when client is lost
func (a *Player) OnClientDisconnected() {
	gwlog.Infof("%s client disconnected", a)
	a.Destroy()
}

// EnterSpace_Client is enter space RPC for client
func (a *Player) EnterSpace_Client(kind int) {
	a.enterSpace(kind)
}

// DoEnterSpace is called by SpaceService to notify avatar entering specified space
func (a *Player) DoEnterSpace(kind int, spaceID common.EntityID) {
	// let the avatar enter space with spaceID
	gwlog.Infof("do enter space", spaceID)
	a.EnterSpace(spaceID, entity.Vector3{Z: -10})
}

// OnEnterSpace is called when avatar enters a space
func (a *Player) OnEnterSpace() {
	gwlog.Infof("%s ENTER SPACE %s", a, a.Space)
	a.SetClientSyncing(true)
}

func (a *Player) SetAction_Client(action string) {
	if a.GetInt("hp") <= 0 { // dead already
		return
	}

	a.Attrs.SetStr("action", action)
}

func (a *Player) Cast_Client(victimID common.EntityID) {
	// a.CallAllClients("Cast")
	victim := a.Space.GetEntity(victimID)
	if victim == nil {
		gwlog.Warnf("Cast %s, but monster not found", victimID)
		return
	}
	if victim.Attrs.GetInt("hp") <= 0 {
		return
	}

	monster := victim.I.(*Monster)
	dmg, isCrit := a.CalcDmg(monster)
	monster.TakeDamage(dmg, isCrit)
}

func (a *Player) CalcDmg(monster *Monster) (dmg int64, isCrit bool) {
	r := a.Intn(100) + 1
	dmg = a.Attrs.GetInt("atk")

	if int64(r) > a.Attrs.GetInt("crit") {
		dmg *= 2
		isCrit = true
	}
	return dmg, isCrit
}

func (player *Player) TakeDamage(damage int64) {

	defer func() {
		err := recover()
		if err != nil {
			gwlog.Errorf("take damage error=", err)
		}
	}()

	hp := player.GetInt("hp")

	if hp <= 0 {
		return
	}

	hp = hp - damage
	if hp < 0 {
		hp = 0
	}
	// gwlog.Infof("player take damage", player.Attrs.String())
	player.Attrs.SetInt("hp", hp)
	gwlog.Infof("player take damage", player.Attrs.String())
	if hp <= 0 {
		// now player dead ...
		player.Attrs.SetStr("action", "death")
		player.Attrs.SetBool("alive", false)
		player.SetClientSyncing(false)
	}

}

//func (a *Player) ShootMiss_Client() {
//	a.Attrs.SetStr("action", "attack")
//	a.CallAllClients("Shoot")
//}

//func (a *Player) ShootHit_Client(victimID common.EntityID) {
//	a.CallAllClients("Shoot")
//	victim := a.Space.GetEntity(victimID)
//	if victim == nil {
//		gwlog.Warnf("Shoot %s, but monster not found", victimID)
//		return
//	}
//
//	if victim.Attrs.GetInt("hp") <= 0 {
//		return
//	}
//
//	monster := victim.I.(*Monster)
//	monster.TakeDamage(50)
//}
