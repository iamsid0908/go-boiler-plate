package domain

import (
	"core/config"
	"core/models"
)

type GitHubCommitFilesDomain interface {
	StoreCommitFile(params models.GitHubCommitFiles) (int64, error)
	GetGitHubCommitFilesByID(commitFileID int64) (models.GitHubCommitFiles, error)
	GetCommitFilesDetailsByCommitId(param models.GitHubCommitFiles) (models.GitHubCommitFiles, error)
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

func (g *GitHubCommitFilesDomainCtx) GetGitHubCommitFilesByID(commitFileID int64) (models.GitHubCommitFiles, error) {
	db := config.DbManager()

	var commitFile models.GitHubCommitFiles

	err := db.Where("id = ?", commitFileID).First(&commitFile).Error

	if err != nil {
		return models.GitHubCommitFiles{}, err
	}

	return commitFile, nil

}

func (g *GitHubCommitFilesDomainCtx) GetCommitFilesDetailsByCommitId(param models.GitHubCommitFiles) (models.GitHubCommitFiles, error) {
	db := config.DbManager()

	var commitFileDetails models.GitHubCommitFiles

	err := db.Where("github_commit_id = ?", param.GithubCommitID).Find(&commitFileDetails).Error

	if err != nil {
		return models.GitHubCommitFiles{}, err
	}

	return commitFileDetails, nil
}
