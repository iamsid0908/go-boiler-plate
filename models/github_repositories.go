package models

import (
	"time"

	"github.com/lib/pq"
)

type GitHubRepository struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	InstallationID int64     `gorm:"column:installation_id;index;not null"`
	UserID         int64     `gorm:"column:user_id;index;not null"`
	GithubRepoID   int64     `gorm:"column:github_repo_id;uniqueIndex;not null"`
	Name           string    `gorm:"column:name;not null"`
	FullName       string    `gorm:"column:full_name;not null"`
	Private        bool      `gorm:"column:private;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

type GitHubRepositoryResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

type CommitActivity struct {
	CommitID         int64          `json:"commit_id"`
	CommitSha        string         `json:"commit_sha"`
	CommitMessage    string         `json:"commit_message"`
	GithubAuthorName string         `json:"github_author_name"`
	CommitDate       time.Time      `json:"commit_date"`
	FilesChanged     pq.StringArray `json:"files_changed"`
}

type CommitFileDetail struct {
	ID        int64  `json:"id"`
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
}

type CommitInfo struct {
	CommitSha   string `json:"commit_sha"`
	Author      string `json:"author"`
	Message     string `json:"message"`
	CommittedAt string `json:"committed_at"`
}

type CommitDetailsResponse struct {
	Commit CommitInfo         `json:"commit"`
	Files  []CommitFileDetail `json:"files"`
}
