package domain

import (
	"core/config"
	"core/models"
)

type GitHubCommitFilesDomain interface {
	StoreCommitFile(params models.GitHubCommitFiles) (int64, error)
}

type GitHubCommitFilesDomainCtx struct{}

func (g *GitHubCommitFilesDomainCtx) StoreCommitFile(params models.GitHubCommitFiles) (int64, error) {
	db := config.DbManager()

	err := db.Create(&params).Error

	if err != nil {
		return 0, err
	}

	return params.ID, nil

}
