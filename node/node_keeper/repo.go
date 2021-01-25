package node_keeper

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

const CmdFileName = "cmd"

type Repo struct {
	Name     string
	StartCmd string
	Version  int
}

func NewRepo(name string, files []string, dir string, version int) (*Repo, error) {
	repo := &Repo{
		Name:    name,
		Version: version,
	}
	for _, file := range files {
		if file == filepath.Join(dir, CmdFileName) {
			byt, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, err
			}
			repo.StartCmd = string(byt)
		}
	}

	return repo, nil
}

func (r *Repo) Start() error {
	cmd := exec.Command(r.StartCmd)
	return cmd.Start()
}
