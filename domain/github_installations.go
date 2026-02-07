package domain

import (
	"core/config"
	"core/models"
)

type GitHubInstallationsDomain interface {
	GetAllByWorkspaceId(workspaceID int64) ([]models.GitHubRepository, error)
}

type GitHubInstallationsDomainCtx struct{}

func (g *GitHubInstallationsDomainCtx) GetAllByWorkspaceId(workspaceID int64) ([]models.GitHubRepository, error) {
	db := config.DbManager()
	var repositories []models.GitHubRepository

	err := db.Table("git_hub_repository").
		Select("git_hub_repository.*").
		Joins("INNER JOIN github_installations ON git_hub_repository.installation_id = github_installations.installation_id").
		Where("github_installations.workspace_id = ?", workspaceID).
		Find(&repositories).Error

	if err != nil {
		return nil, err
	}

	return repositories, nil
}
