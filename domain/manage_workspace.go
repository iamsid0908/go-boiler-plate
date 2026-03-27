package domain

import (
	"core/config"
	"core/models"
)

type ManageWorkspaceDomain interface {
	Create(params models.ManageWorkspace) (models.ManageWorkspace, error)
	GetByWorkspaceIdAndUserId(params models.ManageWorkspace) (models.ManageWorkspace, error)
	GetByWorkspaceIdAndUser(params models.ManageWorkspace) (models.ManageWorkspace, error)
	CountOfUsersInWorkspace(params models.ManageWorkspace) (int64, error)
	UpdateIsAccepted(params models.ManageWorkspace) error
	GetMembersByWorkspaceId(params models.Workspace) ([]models.WorkspaceMembersResp, error)
	GetAllWorkspaceByUserId(userId int64) []models.GetAllWorkspaceByUserIdResp
}

type ManageWorkspaceDomainCtx struct {
}

func (c *ManageWorkspaceDomainCtx) Create(params models.ManageWorkspace) (models.ManageWorkspace, error) {
	db := config.DbManager()
	err := db.Create(&params).Error
	if err != nil {
		return models.ManageWorkspace{}, err
	}
	return params, nil
}

func (c *ManageWorkspaceDomainCtx) GetByWorkspaceIdAndUserId(params models.ManageWorkspace) (models.ManageWorkspace, error) {
	db := config.DbManager()
	err := db.Where("workspace_id = ? AND joined_user_id = ?", params.WorkspaceID, params.JoinedUserID).First(&params)
	if err != nil {
		return models.ManageWorkspace{}, err.Error
	}

	return params, nil
}

func (c *ManageWorkspaceDomainCtx) GetByWorkspaceIdAndUser(params models.ManageWorkspace) (models.ManageWorkspace, error) {
	db := config.DbManager()

	var result models.ManageWorkspace
	err := db.Where("workspace_id = ? AND joined_user_id = ?", params.WorkspaceID, params.JoinedUserID).First(&result).Error
	if err != nil {
		return models.ManageWorkspace{}, err
	}

	return result, nil
}

func (c *ManageWorkspaceDomainCtx) CountOfUsersInWorkspace(params models.ManageWorkspace) (int64, error) {
	var count int64
	db := config.DbManager()
	err := db.Model(&models.ManageWorkspace{}).Where("workspace_id = ? AND is_accepted = ?", params.WorkspaceID, true).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *ManageWorkspaceDomainCtx) UpdateIsAccepted(params models.ManageWorkspace) error {
	db := config.DbManager()
	if params.IsAccepted {
		err := db.Model(&models.ManageWorkspace{}).Where("workspace_id = ? AND joined_user_id = ?", params.WorkspaceID, params.JoinedUserID).Update("is_accepted", params.IsAccepted).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ManageWorkspaceDomainCtx) GetMembersByWorkspaceId(params models.Workspace) ([]models.WorkspaceMembersResp, error) {
	db := config.DbManager()
	var members []models.WorkspaceMembersResp

	err := db.Table("manage_workspace").
		Select("users.id as user_id, users.name as name, users.email as email, manage_workspace.role as role").
		Joins("JOIN users ON manage_workspace.joined_user_id = users.id").
		Where("manage_workspace.workspace_id = ? AND manage_workspace.is_accepted = ?", params.ID, true).
		Scan(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

func (c *ManageWorkspaceDomainCtx) GetAllWorkspaceByUserId(userId int64) []models.GetAllWorkspaceByUserIdResp {
	db := config.DbManager()

	var userWorkspaces []models.ManageWorkspace
	err := db.Table("manage_workspace").
		Where("joined_user_id = ? AND is_accepted = ?", userId, true).
		Find(&userWorkspaces).Error
	if err != nil {
		return nil
	}

	if len(userWorkspaces) == 0 {
		return []models.GetAllWorkspaceByUserIdResp{}
	}

	workspaceIDs := make([]int64, 0, len(userWorkspaces))
	rolesByWorkspaceID := make(map[int64]string, len(userWorkspaces))
	for _, userWorkspace := range userWorkspaces {
		workspaceIDs = append(workspaceIDs, userWorkspace.WorkspaceID)
		rolesByWorkspaceID[userWorkspace.WorkspaceID] = userWorkspace.Role
	}

	var workspaceRows []models.Workspace
	err = db.Table("workspace").Where("id IN ?", workspaceIDs).Find(&workspaceRows).Error
	if err != nil {
		return nil
	}

	workspaceByID := make(map[int64]models.Workspace, len(workspaceRows))
	for _, workspace := range workspaceRows {
		workspaceByID[workspace.ID] = workspace
	}

	workspaces := make([]models.GetAllWorkspaceByUserIdResp, 0, len(userWorkspaces))
	for _, userWorkspace := range userWorkspaces {
		workspace, exists := workspaceByID[userWorkspace.WorkspaceID]
		if !exists {
			continue
		}

		workspaces = append(workspaces, models.GetAllWorkspaceByUserIdResp{
			ID:        workspace.ID,
			Name:      workspace.Name,
			OwnerID:   workspace.OwnerID,
			Type:      workspace.Type,
			Role:      rolesByWorkspaceID[userWorkspace.WorkspaceID],
			CreatedAt: workspace.CreatedAt,
			UpdatedAt: workspace.UpdatedAt,
		})
	}

	return workspaces
}
