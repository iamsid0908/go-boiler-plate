package domain

import (
	"core/config"
	"core/models"
)

type CommitFileEmbeddingDomain interface {
	StoreEmbedding(embeddingRow models.CommitFileEmbedding) error
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
