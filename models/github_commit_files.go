package models

import "time"

type GitHubCommitFiles struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	GithubCommitID int64     `gorm:"column:github_commit_id;index;not null"`
	Filename       string    `gorm:"column:filename;not null"`
	Status         string    `gorm:"column:status;not null"` // added, modified, removed
	Additions      int       `gorm:"column:additions;not null"`
	Deletions      int       `gorm:"column:deletions;not null"`
	Patch          string    `gorm:"column:patch;type:text"`
	GithubRepoID   int64     `gorm:"column:github_repo_id;index;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}
