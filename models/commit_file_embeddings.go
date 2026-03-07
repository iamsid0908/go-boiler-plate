package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type CommitFileEmbedding struct {
	ID               int64           `json:"id" gorm:"primaryKey"`
	Embedding        pgvector.Vector `gorm:"type:vector(1536);not null"`
	CommitFileID     int64           `json:"commit_file_id" gorm:"uniqueIndex;not null"`
	GithubCommitID   int64           `json:"github_commit_id" gorm:"not null"`
	GithubRepoID     int64           `json:"github_repo_id" gorm:"index;not null"`
	InstallationID   int64           `json:"installation_id" gorm:"index;not null"`
	GithubAuthorName *string         `json:"github_author_name,omitempty"`
	PlatformUserID   *int64          `json:"platform_user_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

type EmbedCommitFile2 struct {
	CommitFileId   int64  `json:"commit_file_id"`
	GitHubRepoId   int64  `json:"github_repo_id"`
	InstallationID int64  `json:"installation_id"`
	GithubCommitID int64  `json:"github_commit_id"`
	FullName       string `json:"full_name"`
	Filename       string `json:"filename"`
	Author         string `json:"author"`
	Message        string `json:"message"`
	Patch          string `json:"patch"`
}
