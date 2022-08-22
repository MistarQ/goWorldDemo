package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/action"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/eType"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/prop"
	"math/rand"
	"strconv"
	"time"
)

// Player 对象代表一名玩家
type Player struct {
	entity.Entity
	*rand.Rand
}

func (player *Player) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetPersistent(true).SetUseAOI(true, 100)
	desc.DefineAttr(prop.NAME, "AllClients", "Persistent")
	desc.DefineAttr(prop.Level, "AllClients", "Persistent")
	desc.DefineAttr(prop.Hp, "AllClients")
	desc.DefineAttr(prop.HpMax, "AllClients")
	desc.DefineAttr(prop.Action, "AllClients")
	desc.DefineAttr(prop.SpaceKind, "Persistent")
	desc.DefineAttr(prop.AttackRange, "AllClients")
	desc.DefineAttr(prop.Atk, "AllClients")
	desc.DefineAttr(prop.Crit, "AllClients")
	desc.DefineAttr(prop.CritIndex, "AllClients")
	desc.DefineAttr(prop.Alive, "AllClients")
	desc.DefineAttr(prop.UltimatePoint, "client")
}

// OnCreated 在Player对象创建后被调用
func (player *Player) OnCreated() {
	player.Entity.OnCreated()
	// 应该从account service 获取
	player.setDefaultAttrs()
	player.Rand = rand.New(rand.NewSource(time.Now().Unix()))
}

// setDefaultAttrs 设置玩家的一些默认属性
func (player *Player) setDefaultAttrs() {
	// 应该从account service 获取
	player.Attrs.SetDefaultInt(prop.SpaceKind, 1)
	player.Attrs.SetDefaultInt(prop.Level, 1)
	player.Attrs.SetDefaultInt(prop.Hp, 100)
	player.Attrs.SetDefaultInt(prop.HpMax, 100)
	player.Attrs.SetDefaultStr(prop.Action, action.Idle)
	player.Attrs.SetDefaultInt(prop.AttackRange, 5)
	player.Attrs.SetDefaultInt(prop.Atk, 30)
	player.Attrs.SetDefaultInt(prop.Crit, 10)
	player.Attrs.SetDefaultInt(prop.CritIndex, 2)
	player.Attrs.SetBool(prop.Alive, true)
	player.Attrs.SetInt(prop.UltimatePoint, 0)
	player.SetClientSyncing(true)
}

func (player *Player) ResetAttr() {
	player.Attrs.SetInt(prop.SpaceKind, 1)
	player.Attrs.SetInt(prop.Level, 1)
	player.Attrs.SetInt(prop.Hp, 100)
	player.Attrs.SetInt(prop.HpMax, 100)
	player.Attrs.SetStr(prop.Action, action.Idle)
	player.Attrs.SetInt(prop.AttackRange, 5)
	player.Attrs.SetInt(prop.Atk, 30)
	player.Attrs.SetInt(prop.Crit, 10)
	player.Attrs.SetInt(prop.CritIndex, 2)
	player.Attrs.SetBool(prop.Alive, true)
	player.Attrs.SetInt(prop.UltimatePoint, 0)
	player.Position.X = 0
	player.Position.Y = 0
	player.Position.Z = -10
	player.SetClientSyncing(true)
}

// GetSpaceID 获得玩家的场景ID并发给调用者
func (player *Player) GetSpaceID(callerID common.EntityID) {
	player.Call(callerID, "OnGetPlayerSpaceID", player.ID, player.Space.ID)
}

func (player *Player) enterSpace(spaceKind int) {
	if player.Space.Kind == spaceKind {
		return
	}
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s enter space from %d => %d", player, player.Space.Kind, spaceKind)
	}
	goworld.CallServiceShardKey("SpaceService", strconv.Itoa(spaceKind), "EnterSpace", player.ID, spaceKind)
}

// OnClientConnected is called when client is connected
func (player *Player) OnClientConnected() {
	gwlog.Infof("%s client connected", player)
	player.enterSpace(int(player.GetInt(prop.SpaceKind)))
}

// OnClientDisconnected is called when client is lost
func (player *Player) OnClientDisconnected() {
	gwlog.Infof("%s client disconnected", player)
	player.Destroy()
}

// EnterSpace_Client is enter space RPC for client
func (player *Player) EnterSpace_Client(kind int) {
	player.enterSpace(kind)
}

// DoEnterSpace is called by SpaceService to notify avatar entering specified space
func (player *Player) DoEnterSpace(kind int, spaceID common.EntityID) {
	// let the avatar enter space with spaceID
	gwlog.Infof("do enter space", spaceID)
	player.EnterSpace(spaceID, entity.Vector3{Z: -10})
}

// OnEnterSpace is called when avatar enters a space
func (player *Player) OnEnterSpace() {
	gwlog.Infof("%s ENTER SPACE %s", player, player.Space)
	player.SetClientSyncing(true)
}

func (player *Player) Cast_Client(victimID common.EntityID) {
	victim := player.Space.GetEntity(victimID)
	if victim == nil {
		gwlog.Warnf("Cast %s, but monster not found", victimID)
		return
	}
	if !eType.IsPlayer(victim.TypeName) {
		return
	}
	if victim.Attrs.GetInt(prop.Hp) <= 0 {
		return
	}

	monster := victim.I.(*Monster)
	dmg, isCrit := player.CalcDmg(monster, CAST)
	monster.TakeDamage(dmg, isCrit)
}

func (player *Player) SetAction_Client(action string) {
	if player.GetInt(prop.Hp) <= 0 { // dead already
		return
	}

	player.Attrs.SetStr(prop.Action, action)
}

func (player *Player) Ultimate_Client(victimID common.EntityID) {
	victim := player.Space.GetEntity(victimID)
	if victim == nil {
		gwlog.Warnf("Ultimate %s, but monster not found", victimID)
		return
	}
	if !eType.IsPlayer(victim.TypeName) {
		return
	}
	if victim.Attrs.GetInt(prop.Hp) <= 0 {
		return
	}
	monster := victim.I.(*Monster)
	dmg, isCrit := player.CalcDmg(monster, ULTIMATE)
	monster.TakeDamage(dmg, isCrit)
}

func (player *Player) GetUltimate_Client(playerID common.EntityID) {
	ultimatePoint := player.Attrs.GetInt(prop.UltimatePoint)
	if ultimatePoint > 2 {
		return
	}
	player.Attrs.SetInt(prop.UltimatePoint, ultimatePoint+1)
}

func (player *Player) CalcDmg(monster *Monster, skillType int) (dmg int64, isCrit bool) {
	r := player.Intn(100) + 1
	dmg = player.Attrs.GetInt(prop.Atk)

	dmg *= CalcDmgFactor(skillType)

	if int64(r) > player.Attrs.GetInt(prop.Crit) {
		dmg *= player.Attrs.GetInt(prop.CritIndex)
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

	hp := player.GetInt(prop.Hp)

	if hp <= 0 {
		return
	}

	hp = hp - damage
	if hp < 0 {
		hp = 0
	}
	// gwlog.Infof("player take damage", player.Attrs.String())
	player.Attrs.SetInt(prop.Hp, hp)
	gwlog.Infof("player take damage", player.Attrs.String())
	if hp <= 0 {
		// now player dead ...
		player.Attrs.SetStr(prop.Action, action.Death)
		player.Attrs.SetBool(prop.Alive, false)
		player.SetClientSyncing(false)
	}

}
