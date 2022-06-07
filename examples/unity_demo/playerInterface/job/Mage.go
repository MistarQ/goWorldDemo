package job

import "github.com/xiaonanln/goworld/engine/gwlog"

type Mage struct {
}

func (m *Mage) Attack() {
	gwlog.Infof("Samurai attack")
}

func (m *Mage) Cast() {
	gwlog.Infof("Samurai cast")
}
