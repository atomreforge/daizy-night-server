package dbware

import (
	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
)

type RepoMarkdown struct {
	pDB abstract.InterfaceProviderDB
}

func NewRepoMarkdown(pDB abstract.InterfaceProviderDB) *RepoMarkdown {
	return &RepoMarkdown{pDB: pDB}
}

//func (r *RepoMarkdown) CreateItem(md *model.Markdown) error {}
