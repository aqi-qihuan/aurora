package service

import (
	"context"
	"fmt"

	"github.com/aurora-go/aurora/internal/dto"
	"github.com/aurora-go/aurora/internal/errors"
	"github.com/aurora-go/aurora/internal/model"
	"github.com/aurora-go/aurora/internal/vo"
	"gorm.io/gorm"
)

// MenuService 菜单业务逻辑 (对标 Java MenuServiceImpl)
// 菜单用于: 1) 后台导航菜单 2) 动态路由(前端根据菜单生成路由)
type MenuService struct {
	db *gorm.DB
}

func NewMenuService(db *gorm.DB) *MenuService {
	return &MenuService{db: db}
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(ctx context.Context, vo vo.MenuVO) (*model.Menu, error) {
	menu := model.Menu{
		Name:       vo.Name,
		Path:       vo.Path,
		Component:   vo.Component,
		Icon:        vo.Icon,
		Sort:        vo.Sort,
		Type:        vo.Type,
		Hidden:      0,
	}

	if vo.ParentID > 0 {
		menu.ParentID = &vo.ParentID
	}
	if vo.Permission != "" {
		menu.Permission = vo.Permission
	}
	if vo.Hidden != nil {
		menu.Hidden = *vo.Hidden
	}
	if vo.OrderNum != nil {
		menu.OrderNum = vo.OrderNum
	}

	if err := s.db.WithContext(ctx).Create(&menu).Error; err != nil {
		return nil, fmt.Errorf("创建菜单失败: %w", err)
	}
	return &menu, nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(ctx context.Context, id uint, vo vo.MenuVO) error {
	var menu model.Menu
	if err := s.db.WithContext(ctx).First(&menu, id).Error; err != nil {
		return errors.ErrMenuNotFound
	}

	updates := map[string]interface{}{
		"name":        vo.Name,
		"path":        vo.Path,
		"component":    vo.Component,
		"icon":         vo.Icon,
		"sort":         vo.Sort,
		"type":         vo.Type,
		"permission":   vo.Permission,
	}
	if vo.ParentID > 0 {
		updates["parent_id"] = vo.ParentID
	} else if menu.ParentID != nil && *menu.ParentID > 0 {
		updates["parent_id"] = nil // 取消父级
	}
	if vo.Hidden != nil {
		updates["hidden"] = *vo.Hidden
	}
	if vo.OrderNum != nil {
		updates["order_num"] = *vo.OrderNum
	}

	return s.db.WithContext(ctx).Model(&menu).Updates(updates).Error
}

// DeleteMenu 删除菜单 (级联删除子菜单 + 清理角色关联)
func (s *MenuService) DeleteMenu(ctx context.Context, id uint) error {
	var menu model.Menu

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&menu, id).Error; err != nil {
			return errors.ErrMenuNotFound
		}

		// 级联删除子菜单
		tx.Delete(&model.Menu{}, "parent_id = ?", id)

		// 清理角色-菜单关联
		tx.Exec("DELETE FROM t_role_menu WHERE menu_id IN (SELECT id FROM t_menu WHERE id = ? OR parent_id = ?)", id, id)

		if err := tx.Delete(&menu).Error; err != nil {
			return fmt.Errorf("删除菜单失败: %w", err)
		}
		return nil
	})
}

// GetMenuTree 获取完整菜单树 (后台管理用)
func (s *MenuService) GetMenuTree(ctx context.Context) ([]dto.MenuTreeDTO, error) {
	var menus []model.Menu

	err := s.db.WithContext(ctx).
		Order("sort ASC, order_num ASC").
		Find(&menus).Error

	if err != nil {
		return nil, fmt.Errorf("查询菜单失败: %w", err)
	}

	return s.buildMenuTree(menus), nil
}

// GetUserMenus 获取用户的菜单树 (动态路由用)
func (s *MenuService) GetUserMenus(ctx context.Context, userID uint) ([]dto.MenuTreeDTO, error) {
	var menus []model.Menu

	err := s.db.WithContext(ctx).
		Distinct().
		Joins("JOIN t_role_menu ON t_role_menu.menu_id = t_menu.id").
		Joins("JOIN t_user_role ON t_user_role.role_id = t_role_menu.role_id").
		Where("t_user_role.user_id = ? AND t_menu.hidden = 0", userID).
		Order("sort ASC, order_num ASC").
		Find(&menus).Error

	if err != nil {
		return nil, fmt.Errorf("查询用户菜单失败: %w", err)
	}

	return s.buildMenuTree(menus), nil
}

// ListAllMenus 后台管理获取所有菜单(扁平列表)
func (s *MenuService) ListAllMenus(ctx context.Context) ([]dto.MenuDTO, error) {
	var menus []model.Menu

	err := s.db.WithContext(ctx).
		Order("sort ASC").
		Find(&menus).Error

	if err != nil {
		return nil, fmt.Errorf("查询菜单列表失败: %w", err)
	}

	list := make([]dto.MenuDTO, len(menus))
	for i, m := range menus {
		list[i] = dto.MenuDTO{
			ID:         m.ID,
			Name:       m.Name,
			Path:       m.Path,
			Component:   m.Component,
			Icon:        m.Icon,
			Sort:        m.Sort,
			Type:        m.Type,
			Permission:  m.Permission,
			Hidden:      m.Hidden,
			OrderNum:     m.OrderNum,
			CreateTime:   m.CreateTime,
		}
		if m.ParentID != nil {
			parentID := *m.ParentID
			list[i].ParentID = &parentID
		}
	}
	return list, nil
}

// ===== 内部方法 =====

func (s *MenuService) buildMenuTree(menus []model.Menu) []dto.MenuTreeDTO {
	menuMap := make(map[uint]*dto.MenuTreeDTO)
	var roots []dto.MenuTreeDTO

	for _, m := range menus {
		dto := dto.MenuTreeDTO{
			ID:         m.ID,
			Name:       m.Name,
			Path:       m.Path,
			Component:   m.Component,
			Icon:        m.Icon,
			Type:        m.Type,
			Permission:  m.Permission,
			Hidden:      m.Hidden,
			Sort:        m.Sort,
			OrderNum:     m.OrderNum,
			Children:    []dto.MenuTreeDTO{},
		}
		if m.ParentID != nil {
			parentID := *m.ParentID
			dto.ParentID = &parentID
		}
		menuMap[m.ID] = &dto
	}

	for _, m := range menus {
		node := menuMap[m.ID]
		if m.ParentID == nil || *m.ParentID == 0 || menuMap[*m.ParentID] == nil {
			roots = append(roots, *node)
		} else {
			parent := menuMap[*m.ParentID]
			parent.Children = append(parent.Children, *node)
		}
	}

	return roots
}
