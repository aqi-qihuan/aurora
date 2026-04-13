package service

import (
	"context"
	"fmt"
	"strings"

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
		Name:      vo.Name,
		Path:      vo.Path,
		Component: vo.Component,
		Icon:      vo.Icon,
		OrderNum:  vo.OrderNum,
		IsHidden:  0,
	}

	if vo.ParentID > 0 {
		menu.ParentID = &vo.ParentID
	}
	if vo.IsHidden != nil {
		menu.IsHidden = *vo.IsHidden
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
		"name":      vo.Name,
		"path":      vo.Path,
		"component": vo.Component,
		"icon":      vo.Icon,
		"order_num": vo.OrderNum,
	}
	if vo.ParentID > 0 {
		updates["parent_id"] = vo.ParentID
	} else if menu.ParentID != nil && *menu.ParentID > 0 {
		updates["parent_id"] = nil // 取消父级
	}
	if vo.IsHidden != nil {
		updates["is_hidden"] = *vo.IsHidden
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
		Where("is_hidden = 0"). // 过滤隐藏的菜单
		Order("order_num ASC").
		Find(&menus).Error

	if err != nil {
		return nil, fmt.Errorf("查询菜单失败: %w", err)
	}

	return s.convertUserMenuList(menus), nil // 使用Java的转换逻辑
}

// GetUserMenus 获取用户的菜单树 (动态路由用，对标Java convertUserMenuList)
func (s *MenuService) GetUserMenus(ctx context.Context, userID uint) ([]dto.MenuTreeDTO, error) {
	var menus []model.Menu

	err := s.db.WithContext(ctx).
		Distinct().
		Joins("JOIN t_role_menu ON t_role_menu.menu_id = t_menu.id").
		Joins("JOIN t_user_role ON t_user_role.role_id = t_role_menu.role_id").
		Where("t_user_role.user_id = ? AND t_menu.is_hidden = 0", userID).
		Order("order_num ASC").
		Find(&menus).Error

	if err != nil {
		return nil, fmt.Errorf("查询用户菜单失败: %w", err)
	}

	// 使用 Java 的 convertUserMenuList 逻辑
	return s.convertUserMenuList(menus), nil
}

// ListMenus 后台管理获取菜单列表（树形结构，对标Java listMenus，不过滤isHidden）
func (s *MenuService) ListMenus(ctx context.Context) ([]dto.MenuDTO, error) {
	var menus []model.Menu

	err := s.db.WithContext(ctx).
		Order("order_num ASC").
		Find(&menus).Error

	if err != nil {
		return nil, fmt.Errorf("查询菜单列表失败: %w", err)
	}

	// 构建目录（一级菜单）和子菜单映射，对标Java listCatalogs + getMenuMap
	catalogs := make([]model.Menu, 0)
	childrenMap := make(map[uint][]model.Menu)

	for _, m := range menus {
		if m.ParentID == nil || *m.ParentID == 0 {
			catalogs = append(catalogs, m)
		} else {
			childrenMap[*m.ParentID] = append(childrenMap[*m.ParentID], m)
		}
	}

	// 转换为DTO并组装树
	var menuDTOs []dto.MenuDTO
	for _, catalog := range catalogs {
		catalogDTO := dto.MenuDTO{
			ID:         catalog.ID,
			Name:       catalog.Name,
			Path:       catalog.Path,
			Component:  catalog.Component,
			Icon:       catalog.Icon,
			OrderNum:   catalog.OrderNum,
			IsHidden:   catalog.IsHidden,
			CreateTime: catalog.CreateTime,
			UpdateTime: catalog.UpdateTime,
			Children:   []dto.MenuDTO{},
		}

		if children, ok := childrenMap[catalog.ID]; ok {
			for _, child := range children {
				childDTO := dto.MenuDTO{
					ID:         child.ID,
					Name:       child.Name,
					Path:       child.Path,
					Component:  child.Component,
					Icon:       child.Icon,
					OrderNum:   child.OrderNum,
					IsHidden:   child.IsHidden,
					CreateTime: child.CreateTime,
					UpdateTime: child.UpdateTime,
				}
				if child.ParentID != nil {
					parentID := *child.ParentID
					childDTO.ParentID = &parentID
				}
				catalogDTO.Children = append(catalogDTO.Children, childDTO)
			}
			delete(childrenMap, catalog.ID)
		}
		menuDTOs = append(menuDTOs, catalogDTO)
	}

	// 处理孤儿节点（父菜单已删除的情况）
	for _, children := range childrenMap {
		for _, child := range children {
			childDTO := dto.MenuDTO{
				ID:         child.ID,
				Name:       child.Name,
				Path:       child.Path,
				Component:  child.Component,
				Icon:       child.Icon,
				OrderNum:   child.OrderNum,
				IsHidden:   child.IsHidden,
				CreateTime: child.CreateTime,
				UpdateTime: child.UpdateTime,
			}
			if child.ParentID != nil {
				parentID := *child.ParentID
				childDTO.ParentID = &parentID
			}
			menuDTOs = append(menuDTOs, childDTO)
		}
	}

	return menuDTOs, nil
}

// ListAllMenus 后台管理获取所有菜单(扁平列表)
func (s *MenuService) ListAllMenus(ctx context.Context) ([]dto.MenuDTO, error) {
	var menus []model.Menu

	err := s.db.WithContext(ctx).
		Order("order_num ASC").
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
			Component:  m.Component,
			Icon:       m.Icon,
			OrderNum:   m.OrderNum,
			IsHidden:   m.IsHidden,
			CreateTime: m.CreateTime,
			UpdateTime: m.UpdateTime,
		}
		if m.ParentID != nil {
			parentID := *m.ParentID
			list[i].ParentID = &parentID
		}
	}
	return list, nil
}

// ===== 内部方法 =====

// buildMenuTree 构建菜单树（后台管理用，对标Java listMenus）
func (s *MenuService) buildMenuTree(menus []model.Menu) []dto.MenuTreeDTO {
	menuMap := make(map[uint]*dto.MenuTreeDTO)
	var roots []dto.MenuTreeDTO

	for _, m := range menus {
		dto := dto.MenuTreeDTO{
			ID:         m.ID,
			Name:       m.Name,
			Label:      m.Name,
			Path:       m.Path,
			Component:  m.Component,
			Icon:       m.Icon,
			IsHidden:   m.IsHidden,
			OrderNum:   m.OrderNum,
			Children:   []dto.MenuTreeDTO{},
			CreateTime: m.CreateTime,
			UpdateTime: m.UpdateTime,
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

// buildUserMenuTree 构建用户菜单树（前端路由用，对标Java convertUserMenuList）
// Java逻辑：对于没有子菜单的一级菜单，将其转换为父级(component=Layout)，并添加虚拟子菜单
func (s *MenuService) buildUserMenuTree(menus []model.Menu) []dto.MenuTreeDTO {
	// 1. 分离一级菜单和子菜单
	type menuWithChildren struct {
		menu     model.Menu
		children []model.Menu
	}

	menuMap := make(map[uint]*menuWithChildren)
	var rootMenus []model.Menu

	// 初始化所有一级菜单
	for _, m := range menus {
		if m.ParentID == nil || *m.ParentID == 0 {
			rootMenus = append(rootMenus, m)
			menuMap[m.ID] = &menuWithChildren{menu: m}
		}
	}

	// 收集子菜单
	for _, m := range menus {
		if m.ParentID != nil && *m.ParentID > 0 {
			if parent, exists := menuMap[*m.ParentID]; exists {
				parent.children = append(parent.children, m)
			}
		}
	}

	// 2. 构建结果（对标Java convertUserMenuList逻辑）
	var result []dto.MenuTreeDTO

	for _, root := range rootMenus {
		wrapper := menuMap[root.ID]

		var parentMenu dto.MenuTreeDTO
		var children []dto.MenuTreeDTO

		if len(wrapper.children) > 0 {
			// 有子菜单：父菜单保持原样，子菜单正常添加
			parentMenu = dto.MenuTreeDTO{
				ID:         root.ID,
				Name:       root.Name,
				Label:      root.Name,
				Path:       root.Path,
				Component:  root.Component,
				Icon:       root.Icon,
				IsHidden:   root.IsHidden,
				OrderNum:   root.OrderNum,
				Children:   []dto.MenuTreeDTO{},
				CreateTime: root.CreateTime,
				UpdateTime: root.UpdateTime,
			}

			for _, child := range wrapper.children {
				children = append(children, dto.MenuTreeDTO{
					ID:         child.ID,
					Name:       child.Name,
					Label:      child.Name,
					Path:       child.Path,
					Component:  child.Component,
					Icon:       child.Icon,
					IsHidden:   child.IsHidden,
					OrderNum:   child.OrderNum,
					Children:   []dto.MenuTreeDTO{},
					CreateTime: child.CreateTime,
					UpdateTime: child.UpdateTime,
				})
			}
		} else {
			// 没有子菜单：对标Java第159-167行逻辑
			// 父菜单：path=原path, component=Layout
			parentMenu = dto.MenuTreeDTO{
				ID:         root.ID,
				Name:       root.Name,
				Label:      root.Name,
				Path:       root.Path,
				Component:  "Layout",
				Icon:       root.Icon,
				IsHidden:   root.IsHidden,
				OrderNum:   root.OrderNum,
				Children:   []dto.MenuTreeDTO{},
				CreateTime: root.CreateTime,
				UpdateTime: root.UpdateTime,
			}

			// 虚拟子菜单：path="", component=原component
			children = append(children, dto.MenuTreeDTO{
				Name:      root.Name,
				Label:     root.Name,
				Path:      "",
				Component: root.Component,
				Icon:      root.Icon,
				IsHidden:  root.IsHidden,
				Children:  []dto.MenuTreeDTO{},
			})
		}

		parentMenu.Children = children
		result = append(result, parentMenu)
	}

	return result
}

// convertUserMenuList 转换用户菜单列表（对标Java MenuServiceImpl.convertUserMenuList）
// 对于没有子菜单的菜单项，将其包装为父节点(component=Layout)，并添加虚拟子菜单
func (s *MenuService) convertUserMenuList(menus []model.Menu) []dto.MenuTreeDTO {
	// 1. 按 parent_id 分组
	type menuGroup struct {
		menu     model.Menu
		children []model.Menu
	}

	menuMap := make(map[uint]*menuGroup)
	var rootMenus []model.Menu

	// 初始化所有一级菜单（parent_id 为 nil 或 0）
	for _, m := range menus {
		if m.ParentID == nil || *m.ParentID == 0 {
			rootMenus = append(rootMenus, m)
			menuMap[m.ID] = &menuGroup{menu: m}
		}
	}

	// 收集子菜单
	for _, m := range menus {
		if m.ParentID != nil && *m.ParentID > 0 {
			if parent, exists := menuMap[*m.ParentID]; exists {
				parent.children = append(parent.children, m)
			}
		}
	}

	// 2. 构建结果（对标Java第144-173行逻辑）
	var result []dto.MenuTreeDTO

	for _, root := range rootMenus {
		wrapper := menuMap[root.ID]

		var parentMenu dto.MenuTreeDTO
		var children []dto.MenuTreeDTO

		if len(wrapper.children) > 0 {
			// 有子菜单：父菜单保持原样
			parentMenu = dto.MenuTreeDTO{
				ID:         root.ID,
				Name:       root.Name,
				Label:      root.Name,
				Path:       root.Path,
				Component:  root.Component, // 保留原始component
				Icon:       root.Icon,
				IsHidden:   root.IsHidden,
				OrderNum:   root.OrderNum,
				Children:   []dto.MenuTreeDTO{},
				CreateTime: root.CreateTime,
				UpdateTime: root.UpdateTime,
			}

			// 添加子菜单
			for _, child := range wrapper.children {
				children = append(children, dto.MenuTreeDTO{
					ID:         child.ID,
					Name:       child.Name,
					Label:      child.Name,
					Path:       child.Path,
					Component:  child.Component, // 保留原始component
					Icon:       child.Icon,
					IsHidden:   child.IsHidden,
					OrderNum:   child.OrderNum,
					Children:   []dto.MenuTreeDTO{},
					CreateTime: child.CreateTime,
					UpdateTime: child.UpdateTime,
				})
			}
		} else {
			// 没有子菜单：对标Java第159-167行逻辑
			// 父菜单：path=原path, component=Layout, name不设置（避免与子路由名称冲突）
			parentMenu = dto.MenuTreeDTO{
				ID: root.ID,
				// Name不设置  -- 对标Java第160-161行，只设置path和component
				Path:       root.Path,
				Component:  "Layout", // Java用COMPONENT常量，值为"Layout"
				Icon:       root.Icon,
				IsHidden:   root.IsHidden,
				OrderNum:   root.OrderNum,
				Children:   []dto.MenuTreeDTO{},
				CreateTime: root.CreateTime,
				UpdateTime: root.UpdateTime,
			}

			// 虚拟子菜单：path="", component=原component, name=原name
			// 特殊处理：如果是Talk.vue，使用:talkId路径以支持编辑功能
			childPath := ""
			if strings.HasSuffix(root.Component, "/Talk.vue") || strings.HasSuffix(root.Component, "\\Talk.vue") {
				childPath = ":talkId"
			}

			children = append(children, dto.MenuTreeDTO{
				Path:      childPath,
				Name:      root.Name,      // 只有子菜单设置name
				Label:     root.Name,      // label与name保持一致
				Component: root.Component, // 保留原始component（如 /home/Home.vue）
				Icon:      root.Icon,
				IsHidden:  root.IsHidden, // 使用原始int8值
				Children:  []dto.MenuTreeDTO{},
			})
		}

		parentMenu.Children = children
		result = append(result, parentMenu)
	}

	return result
}
