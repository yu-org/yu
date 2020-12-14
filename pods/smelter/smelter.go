package smelter

import (
	. "yu/pods/interfaces"
)

type Smelter struct {
	pods []Pod
}

func New() *Smelter {
	return &Smelter{
		pods: make([]Pod, 0),
	}
}