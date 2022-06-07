package job

import "github.com/xiaonanln/goworld/engine/gwlog"

type Samurai struct {
}

func (s *Samurai) Attack() {
	gwlog.Infof("Samurai attack")
}

func (s *Samurai) Cast() {
	gwlog.Infof("Samurai cast")
}
