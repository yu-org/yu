package node_keeper

import "os/exec"

type Repo struct {
	StartCmd       string
	BinaryFilename string
	Version        int
}

func (r *Repo) Start() error {
	cmd := exec.Command(r.StartCmd, r.BinaryFilename)
	return cmd.Start()
}
