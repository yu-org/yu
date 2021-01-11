package smelter

import (
	"github.com/pkg/errors"
	. "yu/common"
	. "yu/context"
	. "yu/pod/interfaces"
)

type Smelter struct {
	pods map[string]Pod
}

func NewSmelter() *Smelter {
	return &Smelter{
		pods: make(map[string]Pod),
	}
}

func (s *Smelter) SetPods(pods ...Pod) {
	for _, pod := range pods {
		podName := pod.PodHeader().Name()
		s.pods[podName] = pod
	}
}

func (s *Smelter) Execute(c *Call) error {
	pod, ok := s.pods[c.PodName]
	if !ok {
		return errors.Errorf("POD (%s) not found", c.PodName)
	}
	ph := pod.PodHeader()
	fn := ph.GetExecFn(c.FuncName)
	if fn == nil {
		return errors.Errorf("Execution (%s) not found", c.FuncName)
	}
	ctx := NewContext()
	ctx.SetParams(c.Params)
	return fn(ctx)
}
