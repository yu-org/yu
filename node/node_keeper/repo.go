package node_keeper

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

const CmdFileName = "cmd"

type Repo struct {
	Name     string `json:"name"`
	StartCmd string `json:"start_cmd"`
	Version  int    `json:"version"`
	Arch     string `json:"arch"`
}

func NewRepo(name string, files []string, dir string, version int, arch string) (*Repo, error) {
	repo := &Repo{
		Name:    name,
		Version: version,
		Arch:    arch,
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

func (r *Repo) encode() ([]byte, error) {
	return json.Marshal(r)
}

func decodeRepo(byt []byte) (r *Repo, err error) {
	err = json.Unmarshal(byt, r)
	return
}
