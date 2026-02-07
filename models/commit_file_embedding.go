package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type CommitFileEmbedding struct {
	ID               int64           `json:"id" db:"id"`
	Embedding        pgvector.Vector `json:"embedding" db:"embedding"`
	CommitFileID     int64           `json:"commit_file_id" db:"commit_file_id"`
	GithubCommitID   int64           `json:"github_commit_id" db:"github_commit_id"`
	GithubRepoID     int64           `json:"github_repo_id" db:"github_repo_id"`
	InstallationID   int64           `json:"installation_id" db:"installation_id"`
	GithubAuthorName *string         `json:"github_author_name,omitempty" db:"github_author_name"`
	PlatformUserID   *int64          `json:"platform_user_id,omitempty" db:"platform_user_id"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}
