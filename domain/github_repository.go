package domain

import (
	"core/config"
	"core/models"
)

type GitHubRepositoryDomain interface {
	StoreRepository(params models.GitHubRepository) (int64, error)
	GetAllByWorkspaceId(workspaceID int64) ([]models.GitHubRepository, error)
	GetRepositoryActivity(repoID, days int64) ([]models.CommitActivity, error)
}

type GitHubRepositoryDomainCtx struct{}

func (g *GitHubRepositoryDomainCtx) StoreRepository(params models.GitHubRepository) (int64, error) {
	db := config.DbManager()

	err := db.Create(&params).Error

	if err != nil {
		return 0, err
	}

	return params.ID, nil

}

func (g *GitHubRepositoryDomainCtx) GetAllByWorkspaceId(workspaceID int64) ([]models.GitHubRepository, error) {
	db := config.DbManager()
	var repositories []models.GitHubRepository

	err := db.Where("installation_id = ?", workspaceID).Find(&repositories).Error
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

func (g *GitHubRepositoryDomainCtx) GetRepositoryActivity(
	repoID, days int64,
) ([]models.CommitActivity, error) {

	db := config.DbManager()

	var activities []models.CommitActivity

	err := db.
		Table("git_hub_commits c").
		Select(`
		c.id AS commit_id,
		c.commit_sha,
		c.commit_message,
		c.github_author_name,
		DATE(c.committed_at) AS commit_date,
		COALESCE(
			ARRAY_AGG(f.filename) FILTER (WHERE f.filename IS NOT NULL),
			'{}'
		) AS files_changed
	`).
		Joins("LEFT JOIN git_hub_commit_files f ON f.github_commit_id = c.id").
		Where(
			"c.github_repository_id = ? AND c.committed_at >= NOW() - make_interval(days => ?)",
			repoID,
			days,
		).
		Group("c.id, commit_date").
		Order("commit_date DESC, c.committed_at DESC").
		Scan(&activities).Error

	if err != nil {
		return nil, err
	}

	return activities, nil
}
