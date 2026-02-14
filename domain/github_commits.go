package domain

import (
	"core/config"
	"core/models"
)

type GitHubCommitsDomain interface {
	StoreCommit(params models.GitHubCommits) (int64, error)
	GetCommitDetails(repoID int64, commitSHA string) (models.CommitDetailsResponse, error)
	GetRepoCommitsByRepoId(param models.GetRepoCommitsReqs) (models.GetRepoCommitsPaginatedResponse, error)
}

type GitHubCommitsDomainCtx struct{}

func (g *GitHubCommitsDomainCtx) StoreCommit(params models.GitHubCommits) (int64, error) {
	db := config.DbManager()

	err := db.Create(&params).Error

	if err != nil {
		return 0, err
	}

	return params.ID, nil

}

func (g *GitHubCommitsDomainCtx) GetCommitDetails(repoID int64, commitSHA string) (models.CommitDetailsResponse, error) {
	db := config.DbManager()

	var commit models.CommitInfo
	err := db.Table("git_hub_commits").
		Select("commit_sha, github_author_name as author, commit_message as message, committed_at").
		Where("github_repository_id = ? AND commit_sha = ?", repoID, commitSHA).
		First(&commit).Error

	if err != nil {
		return models.CommitDetailsResponse{}, err
	}

	var files []models.CommitFileDetail
	err = db.Table("git_hub_commit_files cf").
		Select("cf.id, cf.filename, cf.status, cf.additions, cf.deletions, cf.patch").
		Joins("JOIN git_hub_commits c ON cf.github_commit_id = c.id").
		Where("c.github_repository_id = ? AND c.commit_sha = ?", repoID, commitSHA).
		Find(&files).Error

	if err != nil {
		return models.CommitDetailsResponse{}, err
	}

	response := models.CommitDetailsResponse{
		Commit: commit,
		Files:  files,
	}

	return response, nil
}

func (g *GitHubCommitsDomainCtx) GetRepoCommitsByRepoId(param models.GetRepoCommitsReqs) (models.GetRepoCommitsPaginatedResponse, error) {
	db := config.DbManager()

	var commits []models.GitHubCommitResponse
	var totalCount int64

	// Get total count for pagination
	err := db.Table("git_hub_commits").
		Where("github_repository_id = ?", param.RepoID).
		Count(&totalCount).Error

	if err != nil {
		return models.GetRepoCommitsPaginatedResponse{}, err
	}

	// Calculate offset for pagination
	offset := (param.Page - 1) * param.Limit

	// Fetch paginated commits
	err = db.Table("git_hub_commits").
		Select("id, commit_sha as sha, commit_message as message, github_author_name, author_email, committed_at").
		Where("github_repository_id = ?", param.RepoID).
		Order("committed_at DESC").
		Limit(param.Limit).
		Offset(offset).
		Find(&commits).Error

	if err != nil {
		return models.GetRepoCommitsPaginatedResponse{}, err
	}

	// Calculate total pages
	totalPages := int64(totalCount) / int64(param.Limit)
	if totalCount%int64(param.Limit) != 0 {
		totalPages++
	}

	response := models.GetRepoCommitsPaginatedResponse{
		Commits: commits,
		Meta: models.MetaPagination{
			PageNumber:   int64(param.Page),
			PageSize:     int64(param.Limit),
			TotalPages:   totalPages,
			TotalRecords: totalCount,
		},
	}

	return response, nil
}
