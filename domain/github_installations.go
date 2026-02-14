package domain

import (
	"core/config"
	"core/models"
)

type GitHubInstallationsDomain interface {
	GetAllByWorkspaceId(workspaceID int64) ([]models.GitHubRepository, error)
	GetOrgDetailsByWorkspaceId(workspaceID int64) (models.OrgDetailsResponse, error)
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

func (g *GitHubInstallationsDomainCtx) GetOrgDetailsByWorkspaceId(workspaceID int64) (models.OrgDetailsResponse, error) {
	db := config.DbManager()
	var installation models.GitHubInstallation

	// Get GitHub installation details by workspace_id
	err := db.Table("github_installations").
		Where("workspace_id = ?", workspaceID).
		First(&installation).Error

	if err != nil {
		return models.OrgDetailsResponse{}, err
	}

	// Get all repositories associated with this installation
	var repositories []models.GitHubRepository
	err = db.Table("git_hub_repository").
		Where("installation_id = ?", installation.InstallationID).
		Find(&repositories).Error

	if err != nil {
		return models.OrgDetailsResponse{}, err
	}

	// Map repositories to response format
	var repoResponses []models.GitHubRepositoryResponse
	for _, repo := range repositories {
		repoResponses = append(repoResponses, models.GitHubRepositoryResponse{
			ID:       repo.ID,
			Name:     repo.Name,
			FullName: repo.FullName,
			Private:  repo.Private,
		})
	}

	// Build response
	orgDetails := models.OrgDetailsResponse{
		ID:             installation.ID,
		InstallationID: installation.InstallationID,
		AccountLogin:   installation.AccountLogin,
		AccountType:    installation.AccountType,
		WorkspaceID:    installation.WorkspaceID,
		Repositories:   repoResponses,
	}

	return orgDetails, nil
}
