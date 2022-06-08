package main

import "github.com/xiaonanln/goworld/engine/gwlog"

type Job struct {
	Atk         int64
	AttackRange int64
}

func (j *Job) Attack() {
	gwlog.Infof("Attack", j.Atk)
}
