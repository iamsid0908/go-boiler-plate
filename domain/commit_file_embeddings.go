package domain

import (
	"core/config"
	"core/models"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

type CommitFileEmbeddingDomain interface {
	StoreEmbedding(embeddingRow models.CommitFileEmbedding) error
	GetRelatedCommitFiles(commitFileID string) ([]models.RelatedCommitFileResponse, error)
}

type CommitFileEmbeddingDomainCtx struct{}

func (c *CommitFileEmbeddingDomainCtx) StoreEmbedding(embeddingRow models.CommitFileEmbedding) error {
	db := config.DbManager()

	err := db.Create(&embeddingRow).Error

	if err != nil {
		return err
	}

	return nil

}

func (g *CommitFileEmbeddingDomainCtx) GetRelatedCommitFiles(commitFileID string) ([]models.RelatedCommitFileResponse, error) {

	db := config.DbManager()

	// Step 1: fetch the source embedding + context
	var source struct {
		Embedding      pgvector.Vector
		GithubRepoID   int64
		InstallationID int64
	}

	tx := db.
		Table("commit_file_embedding").
		Select("embedding, github_repo_id, installation_id").
		Where("commit_file_id = ?", commitFileID).
		Scan(&source)

	err := tx.Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source embedding: %w", err)
	}

	// No embedding row (or empty embedding) means no related files can be computed.
	// Return empty result so caller can continue with only the main commit file.
	if tx.RowsAffected == 0 || len(source.Embedding.Slice()) == 0 {
		return []models.RelatedCommitFileResponse{}, nil
	}

	// Step 2: similarity search
	var results []models.RelatedCommitFileResponse

	err = db.Raw(`
	SELECT
		cfe.commit_file_id,
		c.commit_sha,
		f.filename,
		c.github_author_name AS author,
		c.committed_at,
		1 - (cfe.embedding <=> ?) AS similarity
	FROM commit_file_embedding cfe
	JOIN git_hub_commit_files f
		ON f.id = cfe.commit_file_id
	JOIN git_hub_commits c
		ON c.id = f.github_commit_id
	WHERE
		cfe.github_repo_id = ?
		AND cfe.installation_id = ?
		AND cfe.commit_file_id <> ?
	ORDER BY cfe.embedding <=> ?
	LIMIT 5
`,
		source.Embedding,
		source.GithubRepoID,
		source.InstallationID,
		commitFileID,
		source.Embedding,
	).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch related commit files: %w", err)
	}

	return results, nil
}
