package models

import "time"

// BackfillCommitFileRow is returned by GetUnembeddedCommitFiles — a join
// of commit_files + commits + repos that have no embedding yet.
type BackfillCommitFileRow struct {
	CommitFileID   int64  `gorm:"column:commit_file_id"`
	GithubCommitID int64  `gorm:"column:github_commit_id"`
	GithubRepoID   int64  `gorm:"column:github_repo_id"`
	InstallationID int64  `gorm:"column:installation_id"`
	FullName       string `gorm:"column:full_name"`
	Filename       string `gorm:"column:filename"`
	Patch          string `gorm:"column:patch"`
	Author         string `gorm:"column:author"`
	Message        string `gorm:"column:message"`
}

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
