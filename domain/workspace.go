package domain

import (
	"core/config"
	"core/models"
)

type WorkspaceDomain interface {
	Create(params models.Workspace) (models.Workspace, error)
	GetById(params models.Workspace) (models.Workspace, error)
	GetAllWorkspaceByUserId(userId int64) []models.GetAllWorkspaceByUserIdResp
	GetWorkspaceAndUserById(params models.Workspace) (models.GetWorkspaceDetailsResp, error)
}

type WorkspaceDomainCtx struct {
}

func (c *WorkspaceDomainCtx) Create(params models.Workspace) (models.Workspace, error) {
	db := config.DbManager()
	err := db.Create(&params).Error
	if err != nil {
		return models.Workspace{}, err
	}
	return params, nil
}
func (c *WorkspaceDomainCtx) GetById(params models.Workspace) (models.Workspace, error) {
	db := config.DbManager()
	var workspace models.Workspace
	result := db.Where("id = ?", params.ID).First(&workspace)
	if result.Error != nil {
		return models.Workspace{}, result.Error
	}

	return workspace, nil
}

func (c *WorkspaceDomainCtx) GetAllWorkspaceByUserId(userId int64) []models.GetAllWorkspaceByUserIdResp {
	db := config.DbManager()
	var workspaces []models.GetAllWorkspaceByUserIdResp
	err := db.Table("workspace").
		Select("workspace.id, workspace.name, workspace.owner_id, workspace.type, manage_workspace.role, workspace.created_at, workspace.updated_at").
		Joins("JOIN manage_workspace ON manage_workspace.workspace_id = workspace.id").
		Where("manage_workspace.joined_user_id = ? AND manage_workspace.is_accepted = ?", userId, true).
		Scan(&workspaces).Error
	if err != nil {
		return nil
	}
	return workspaces
}

func (c *WorkspaceDomainCtx) GetWorkspaceAndUserById(params models.Workspace) (models.GetWorkspaceDetailsResp, error) {
	db := config.DbManager()
	var result models.GetWorkspaceDetailsResp

	err := db.Table("workspace").
		Select("workspace.id as workspace_id, workspace.name as workspace, workspace.owner_id as user_id, users.email as owner_email, users.name as owner_name").
		Joins("JOIN users ON workspace.owner_id = users.id").
		Where("workspace.id = ?", params.ID).
		Scan(&result).Error

	if err != nil {
		return models.GetWorkspaceDetailsResp{}, err
	}

	return result, nil
}
