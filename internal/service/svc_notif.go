package service

import "github.com/atomreforge/daizy-night-server/internal/dbware"

type ServiceNotif struct {
	RepoMarkdown *dbware.RepoMarkdown
}

func NewServiceNotif(repoMarkdown *dbware.RepoMarkdown) *ServiceNotif {
	return &ServiceNotif{
		RepoMarkdown: repoMarkdown,
	}
}
