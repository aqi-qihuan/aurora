package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// RoleService 角色权限业务逻辑 (对标 Java RoleServiceImpl)
type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(ctx context.Context, vo vo.RoleVO) (*model.Role, error) {
	role := model.Role{
		RoleName:    vo.RoleName,
		RoleLabel:   vo.RoleLabel,
		Description:  vo.Description,
		Sort:        vo.Sort,
		IsDisable:   0, // 默认启用
	}

	if err := s.db.WithContext(ctx).Create(&role).Error; err != nil {
		if errors.IsStd(err, gorm.ErrDuplicatedKey) {
			return nil, errors.ErrRoleNameExists
		}
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	// 分配菜单权限
	if len(vo.MenuIDs) > 0 {
		var menus []model.Menu
		s.db.WithContext(ctx).Find(&menus, vo.MenuIDs)
		s.db.WithContext(ctx).Model(&role).Association("Menus").Replace(menus)
	}

	return &role, nil
}

// UpdateRole 更新角色 (含权限变更)
func (s *RoleService) UpdateRole(ctx context.Context, id uint, vo vo.RoleVO) error {
	var role model.Role
	if err := s.db.WithContext(ctx).First(&role, id).Error; err != nil {
		return errors.ErrRoleNotFound
	}

	updates := map[string]interface{}{
		"role_name":    vo.RoleName,
		"role_label":   vo.RoleLabel,
		"description":   vo.Description,
		"sort":         vo.Sort,
	}
	if err := s.db.WithContext(ctx).Model(&role).Updates(updates).Error; err != nil {
		if errors.IsStd(err, gorm.ErrDuplicatedKey) {
			return errors.ErrRoleNameExists
		}
		return fmt.Errorf("更新角色失败: %w", err)
	}

	// 更新菜单权限
	if vo.MenuIDs != nil {
		if len(vo.MenuIDs) > 0 {
			var menus []model.Menu
			s.db.Find(&menus, vo.MenuIDs)
			s.db.Model(&role).Association("Menus").Replace(menus)
		} else {
			s.db.Model(&role).Association("Menus").Clear()
		}
	}

	slog.Info("角色更新完成", "role_id", id, "name", vo.RoleName)
	return nil
}

// DeleteRole 删除角色 (检查是否有关联用户)
func (s *RoleService) DeleteRole(ctx context.Context, id uint) error {
	var role model.Role

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&role, id).Error; err != nil {
			return errors.ErrRoleNotFound
		}

		// 检查是否是默认角色(不可删除)
		if role.IsDefault == 1 {
			return errors.ErrCannotDeleteDefaultRole
		}

		// 检查是否有关联用户
		var userCount int64
		tx.Table("t_user_role").Where("role_id = ?", id).Count(&userCount)
		if userCount > 0 {
			return errors.ErrRoleHasUsers
		}

		// 清理关联
		tx.Exec("DELETE FROM t_role_menu WHERE role_id = ?", id)

		if err := tx.Delete(&role).Error; err != nil {
			return fmt.Errorf("删除角色失败: %w", err)
		}
		return nil
	})
}

// ListRoles 获取所有角色列表
func (s *RoleService) ListRoles(ctx context.Context) ([]dto.RoleDTO, error) {
	var roles []model.Role

	err := s.db.WithContext(ctx).
		Preload("Menus").
		Order("sort ASC").
		Find(&roles).Error

	if err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}

	list := make([]dto.RoleDTO, len(roles))
	for i, r := range roles {
		menuIDs := make([]uint, len(r.Menus))
		for j, m := range r.Menus {
			menuIDs[j] = m.ID
		}
		list[i] = dto.RoleDTO{
			ID:          r.ID,
			RoleName:    r.RoleName,
			RoleLabel:   r.RoleLabel,
			Description:  r.Description,
			IsDisable:   r.IsDisable,
			IsDefault:   r.IsDefault,
			MenuIDs:     menuIDs,
			CreateTime:   r.CreateTime,
		}
	}
	return list, nil
}

// GetRoleByID 根据ID获取角色详情(含菜单树)
func (s *RoleService) GetRoleByID(ctx context.Context, id uint) (*dto.RoleDetailDTO, error) {
	var role model.Role

	err := s.db.WithContext(ctx).
		Preload("Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}).
		First(&role, id).Error

	if err != nil {
		if errors.IsStd(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("查询角色详情失败: %w", err)
	}

	dto := &dto.RoleDetailDTO{
		ID:          role.ID,
		RoleName:    role.RoleName,
		RoleLabel:   role.RoleLabel,
		Description:  role.Description,
		IsDisable:   role.IsDisable,
		IsDefault:   role.IsDefault,
	}

	menuIDs := make([]uint, len(role.Menus))
	for i, m := range role.Menus {
		menuIDs[i] = m.ID
	}
	dto.MenuIDs = menuIDs
	dto.Menus = role.Menus // 直接返回Menu结构用于前端渲染

	return dto, nil
}

// AssignRoleToUser 为用户分配角色
func (s *RoleService) AssignRoleToUser(ctx context.Context, userID uint, roleIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先清除原有角色
		tx.Exec("DELETE FROM t_user_role WHERE user_id = ?", userID)

		// 分配新角色
		for _, roleID := range roleIDs {
			err := tx.Exec("INSERT INTO t_user_role(user_id, role_id) VALUES (?, ?)", userID, roleID).Error
			if err != nil {
				return fmt.Errorf("分配角色[%d]失败: %w", roleID, err)
			}
		}

		slog.Info("分配角色完成", "user_id", userID, "roles", roleIDs)
		return nil
	})
}

// SetRoleStatus 设置角色禁用/启用
func (s *RoleService) SetRoleStatus(ctx context.Context, id uint, isDisable int8) error {
	result := s.db.WithContext(ctx).
		Model(&model.Role{}).
		Where("id = ?", id).
		Update("is_disable", isDisable)

	if result.Error != nil {
		return fmt.Errorf("设置角色状态失败: %w", result.Error)
	}
	return nil
}
