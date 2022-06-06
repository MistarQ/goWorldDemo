package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// SpaceService is the service entity for space management
type SpaceService struct {
	entity.Entity
	spaceID common.EntityID
}

func (s *SpaceService) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

// OnCreated is called when entity is created
func (s *SpaceService) OnCreated() {
	gwlog.Infof("Registering SpaceService ...")
	s.spaceID = goworld.CreateSpaceAnywhere(1)
}

// 获取场景ID
func (s *SpaceService) GetSpaceID(callerID common.EntityID) {
	s.Call(callerID, "OnGetSpaceID", s.spaceID)
}
