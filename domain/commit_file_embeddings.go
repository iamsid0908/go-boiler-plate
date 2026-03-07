package domain

import (
	"core/config"
	"core/models"
	"fmt"
	"strconv"

	"github.com/pgvector/pgvector-go"
)

type CommitFileEmbeddingDomain interface {
	StoreEmbedding(embeddingRow models.CommitFileEmbedding) error
	GetRelatedCommitFiles(commitFileID string) ([]models.RelatedCommitFileResponse, error)
}

type CommitFileEmbeddingDomainCtx struct{}

func (c *CommitFileEmbeddingDomainCtx) StoreEmbedding(embeddingRow models.CommitFileEmbedding) error {
	db := config.DbManager()

	// Upsert: if a row with the same commit_file_id already exists, update the embedding
	// instead of creating a duplicate row.
	err := db.Exec(`
		INSERT INTO commit_file_embedding
			(embedding, commit_file_id, github_commit_id, github_repo_id,
			 installation_id, github_author_name, platform_user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (commit_file_id) DO UPDATE SET
			embedding          = EXCLUDED.embedding,
			github_commit_id   = EXCLUDED.github_commit_id,
			github_repo_id     = EXCLUDED.github_repo_id,
			installation_id    = EXCLUDED.installation_id,
			github_author_name = EXCLUDED.github_author_name,
			platform_user_id   = EXCLUDED.platform_user_id,
			created_at         = NOW()
	`,
		embeddingRow.Embedding,
		embeddingRow.CommitFileID,
		embeddingRow.GithubCommitID,
		embeddingRow.GithubRepoID,
		embeddingRow.InstallationID,
		embeddingRow.GithubAuthorName,
		embeddingRow.PlatformUserID,
	).Error

	if err != nil {
		return err
	}

	return nil
}

func (g *CommitFileEmbeddingDomainCtx) GetRelatedCommitFiles(commitFileID string) ([]models.RelatedCommitFileResponse, error) {

	db := config.DbManager()

	// Parse commitFileID to int64 to avoid Postgres type-mismatch on bigint column
	commitFileIDInt, err := strconv.ParseInt(commitFileID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid commit_file_id %q: %w", commitFileID, err)
	}

	// ── Debug: count total embeddings in the table ──
	var totalCount int64
	db.Table("commit_file_embedding").Count(&totalCount)
	fmt.Printf("[DEBUG-RELATED] Total rows in commit_file_embedding: %d\n", totalCount)

	// ── Debug: show min/max commit_file_id in the table ──
	var minMax struct {
		MinID int64
		MaxID int64
	}
	db.Raw("SELECT MIN(commit_file_id) AS min_id, MAX(commit_file_id) AS max_id FROM commit_file_embedding").Scan(&minMax)
	fmt.Printf("[DEBUG-RELATED] commit_file_id range in table: min=%d, max=%d (querying id=%d)\n", minMax.MinID, minMax.MaxID, commitFileIDInt)

	// Step 1: fetch the source embedding + context
	var source struct {
		Embedding      pgvector.Vector
		GithubRepoID   int64
		InstallationID int64
		CommitFileID   int64
	}

	tx := db.
		Table("commit_file_embedding").
		Select("embedding, github_repo_id, installation_id, commit_file_id").
		Where("commit_file_id = ?", commitFileIDInt).
		Scan(&source)

	err = tx.Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source embedding: %w", err)
	}

	fmt.Printf("[DEBUG-RELATED] Source lookup for commit_file_id=%d → rows_affected=%d, github_repo_id=%d, installation_id=%d\n",
		commitFileIDInt, tx.RowsAffected, source.GithubRepoID, source.InstallationID)

	if tx.RowsAffected == 0 {
		fmt.Printf("[DEBUG-RELATED] No embedding row found for commit_file_id=%d\n", commitFileIDInt)
		return []models.RelatedCommitFileResponse{}, nil
	}

	embSlice := source.Embedding.Slice()
	fmt.Printf("[DEBUG-RELATED] Embedding dimension=%d, first 3 values=%v\n", len(embSlice), embSlice[:min(3, len(embSlice))])

	if len(embSlice) == 0 {
		fmt.Printf("[DEBUG-RELATED] Embedding is empty for commit_file_id=%d\n", commitFileIDInt)
		return []models.RelatedCommitFileResponse{}, nil
	}

	// ── Debug: count how many other embeddings exist for same repo ──
	var sameRepoCount int64
	db.Table("commit_file_embedding").
		Where("github_repo_id = ? AND installation_id = ? AND commit_file_id != ?",
			source.GithubRepoID, source.InstallationID, commitFileIDInt).
		Count(&sameRepoCount)
	fmt.Printf("[DEBUG-RELATED] Other embeddings in same repo (excluding self): %d\n", sameRepoCount)

	if sameRepoCount == 0 {
		fmt.Println("[DEBUG-RELATED] No other embeddings exist for this repo — nothing to compare against")
		return []models.RelatedCommitFileResponse{}, nil
	}

	// ── Debug: check if JOINs would match ──
	var joinCount int64
	db.Raw(`
		SELECT COUNT(*)
		FROM commit_file_embedding cfe
		JOIN git_hub_commit_files f ON f.id = cfe.commit_file_id
		JOIN git_hub_commits c ON c.id = f.github_commit_id
		WHERE cfe.github_repo_id = ?
		  AND cfe.installation_id = ?
		  AND cfe.commit_file_id != ?
	`, source.GithubRepoID, source.InstallationID, commitFileIDInt).Scan(&joinCount)
	fmt.Printf("[DEBUG-RELATED] Rows surviving JOIN (excluding self): %d\n", joinCount)

	// Step 2: similarity search — all params are now proper int64, not string
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
		AND cfe.commit_file_id != ?
	ORDER BY cfe.embedding <=> ?
	LIMIT 5
`,
		source.Embedding,
		source.GithubRepoID,
		source.InstallationID,
		commitFileIDInt,
		source.Embedding,
	).Scan(&results).Error

	if err != nil {
		fmt.Printf("[DEBUG-RELATED] Similarity query ERROR: %v\n", err)
		return nil, fmt.Errorf("failed to fetch related commit files: %w", err)
	}

	fmt.Printf("[DEBUG-RELATED] Similarity search returned %d results\n", len(results))
	for i, r := range results {
		fmt.Printf("[DEBUG-RELATED]   result[%d]: file_id=%d, filename=%s, similarity=%.4f\n", i, r.CommitFileID, r.Filename, r.Similarity)
	}

	return results, nil
}
