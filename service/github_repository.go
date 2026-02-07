package service

import (
	"core/domain"
	"core/models"
)

type GitHubRepositoryService struct {
	GitHubRepositoryDomain domain.GitHubRepositoryDomain
	GitHubCommitsDomain    domain.GitHubCommitsDomain
}

func (g *GitHubRepositoryService) GetRepositoryActivity(repoID, days int64) ([]models.CommitActivity, error) {
	data, err := g.GitHubRepositoryDomain.GetRepositoryActivity(repoID, days)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *GitHubRepositoryService) GetCommitDetails(repoID int64, commitSHA string) (models.CommitDetailsResponse, error) {
	data, err := g.GitHubCommitsDomain.GetCommitDetails(repoID, commitSHA)
	if err != nil {
		return models.CommitDetailsResponse{}, err
	}
	return data, nil
}
