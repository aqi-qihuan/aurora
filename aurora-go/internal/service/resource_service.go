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

// ResourceService 资源权限业务逻辑 (对标 Java ResourceServiceImpl)
type ResourceService struct {
	db *gorm.DB
}

func NewResourceService(db *gorm.DB) *ResourceService {
	return &ResourceService{db: db}
}

// CreateResource 创建资源权限
func (s *ResourceService) CreateResource(ctx context.Context, vo vo.ResourceVO) (*model.Resource, error) {
	resource := model.Resource{
		ResourceName:  vo.ResourceName,
		URL:           vo.URL,
		RequestMethod: vo.RequestMethod,
		ParentID:      vo.ParentID,
		IsAnonymous:   0, // 默认非匿名
	}

	if err := s.db.WithContext(ctx).Create(&resource).Error; err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}
	return &resource, nil
}

// UpdateResource 更新资源权限
func (s *ResourceService) UpdateResource(ctx context.Context, id uint, vo vo.ResourceVO) error {
	var resource model.Resource
	if err := s.db.WithContext(ctx).First(&resource, id).Error; err != nil {
		return errors.ErrResourceNotFound
	}

	updates := map[string]interface{}{
		"resource_name":  vo.ResourceName,
	}
	if vo.URL != "" {
		updates["url"] = vo.URL
	}
	if vo.RequestMethod != "" {
		updates["request_method"] = vo.RequestMethod
	}
	if vo.ParentID != nil {
		updates["parent_id"] = vo.ParentID
	}
	if vo.IsAnonymous != nil {
		updates["is_anonymous"] = *vo.IsAnonymous
	}
	return s.db.WithContext(ctx).Model(&resource).Updates(updates).Error
}

// DeleteResource 删除资源权限
func (s *ResourceService) DeleteResource(ctx context.Context, id uint) error {
	var resource model.Resource

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&resource, id).Error; err != nil {
			return errors.ErrResourceNotFound
		}

		// 清理角色关联
		tx.Exec("DELETE FROM t_role_resource WHERE resource_id = ?", id)

		if err := tx.Delete(&resource).Error; err != nil {
			return fmt.Errorf("删除资源失败: %w", err)
		}
		return nil
	})
}

// ListResources 获取所有资源权限列表(分页)
func (s *ResourceService) ListResources(ctx context.Context, cond dto.ConditionVO, page dto.PageVO) (*dto.PageResultDTO, error) {
	var resources []model.Resource
	var count int64

	baseQuery := s.db.WithContext(ctx).Model(&model.Resource{})

	if cond.Keywords != "" {
		baseQuery = baseQuery.Where("resource_name LIKE ? OR url LIKE ?", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
	}

	baseQuery.Count(&count)

	offset := page.GetOffset()
	err := baseQuery.
		Order("create_time DESC").
		Limit(page.PageSize).
		Offset(offset).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("查询资源列表失败: %w", err)
	}

	list := make([]dto.ResourceDTO, len(resources))
	for i, r := range resources {
		list[i] = dto.ResourceDTO{
			ID:            r.ID,
			ResourceName:  r.ResourceName,
			URL:           r.URL,
			RequestMethod: r.RequestMethod,
			ParentID:      r.ParentID,
			IsAnonymous:   r.IsAnonymous,
			CreateTime:    r.CreateTime,
			UpdateTime:    r.UpdateTime,
		}
	}

	return &dto.PageResultDTO{
		List:     list,
		Count:    count,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

// ListResourcesTree 获取资源树形列表(对标Java版 listResources)
// 前端期望: data.data 直接是树形数组
func (s *ResourceService) ListResourcesTree(ctx context.Context, cond dto.ConditionVO) ([]dto.ResourceDTO, error) {
	var resources []model.Resource

	query := s.db.WithContext(ctx).Order("id ASC")

	if cond.Keywords != "" {
		query = query.Where("resource_name LIKE ? OR url LIKE ?", "%"+cond.Keywords+"%", "%"+cond.Keywords+"%")
	}

	err := query.Find(&resources).Error
	if err != nil {
		return nil, fmt.Errorf("查询资源列表失败: %w", err)
	}

	// 构建树形结构 (对标Java listResourceModule + listResourceChildren)
	parents := make([]model.Resource, 0)
	childrenMap := make(map[uint][]model.Resource)

	for _, r := range resources {
		if r.ParentID == nil {
			parents = append(parents, r)
		} else {
			childrenMap[*r.ParentID] = append(childrenMap[*r.ParentID], r)
		}
	}

	// 转换为DTO并组装树
	result := make([]dto.ResourceDTO, 0, len(parents))
	for _, parent := range parents {
		parentDTO := dto.ResourceDTO{
			ID:            parent.ID,
			ResourceName:  parent.ResourceName,
			URL:           parent.URL,
			RequestMethod: parent.RequestMethod,
			ParentID:      parent.ParentID,
			IsAnonymous:   parent.IsAnonymous,
			CreateTime:    parent.CreateTime,
			UpdateTime:    parent.UpdateTime,
			Children:      make([]dto.ResourceDTO, 0),
		}

		if children, ok := childrenMap[parent.ID]; ok {
			for _, child := range children {
				childDTO := dto.ResourceDTO{
					ID:            child.ID,
					ResourceName:  child.ResourceName,
					URL:           child.URL,
					RequestMethod: child.RequestMethod,
					ParentID:      child.ParentID,
					IsAnonymous:   child.IsAnonymous,
					CreateTime:    child.CreateTime,
					UpdateTime:    child.UpdateTime,
				}
				parentDTO.Children = append(parentDTO.Children, childDTO)
			}
			delete(childrenMap, parent.ID)
		}
		result = append(result, parentDTO)
	}

	// 处理孤儿节点(父节点已删除的情况)
	for _, children := range childrenMap {
		for _, child := range children {
			result = append(result, dto.ResourceDTO{
				ID:            child.ID,
				ResourceName:  child.ResourceName,
				URL:           child.URL,
				RequestMethod: child.RequestMethod,
				ParentID:      child.ParentID,
				IsAnonymous:   child.IsAnonymous,
				CreateTime:    child.CreateTime,
				UpdateTime:    child.UpdateTime,
			})
		}
	}

	return result, nil
}

// ListResourceOptions 获取角色资源选项(用于角色授权下拉框，树形结构)
// 对标Java listResourceOption，返回 LabelOptionDTO 格式
func (s *ResourceService) ListResourceOptions(ctx context.Context) ([]dto.LabelOptionDTO, error) {
	var resources []model.Resource

	err := s.db.WithContext(ctx).
		Select("id", "resource_name", "parent_id").
		Order("id ASC").
		Find(&resources).Error
	if err != nil {
		return nil, fmt.Errorf("查询资源选项失败: %w", err)
	}

	// 构建树形结构
	parents := make([]model.Resource, 0)
	childrenMap := make(map[uint][]model.Resource)

	for _, r := range resources {
		if r.ParentID == nil {
			parents = append(parents, r)
		} else {
			childrenMap[*r.ParentID] = append(childrenMap[*r.ParentID], r)
		}
	}

	// 转换为LabelOptionDTO
	result := make([]dto.LabelOptionDTO, 0, len(parents))
	for _, parent := range parents {
		option := dto.LabelOptionDTO{
			ID:       parent.ID,
			Label:    parent.ResourceName,
			Children: make([]dto.LabelOptionDTO, 0),
		}

		if children, ok := childrenMap[parent.ID]; ok {
			for _, child := range children {
				option.Children = append(option.Children, dto.LabelOptionDTO{
					ID:    child.ID,
					Label: child.ResourceName,
				})
			}
			delete(childrenMap, parent.ID)
		}
		result = append(result, option)
	}

	return result, nil
}

// AssignResourceToRole 为角色分配资源权限
func (s *ResourceService) AssignResourceToRole(ctx context.Context, roleID uint, resourceIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先清除原有权限
		tx.Exec("DELETE FROM t_role_resource WHERE role_id = ?", roleID)

		// 分配新权限
		for _, resID := range resourceIDs {
			err := tx.Exec(
				"INSERT INTO t_role_resource(role_id, resource_id) VALUES (?, ?)",
				roleID, resID,
			).Error
			if err != nil {
				return fmt.Errorf("分配资源[%d]给角色[%d]失败: %w", resID, roleID, err)
			}
		}

		slog.Info("资源权限分配完成", "role_id", roleID, "resources", resourceIDs)
		return nil
	})
}

// GetResourcesByRole 获取角色的资源权限列表(用于RBAC中间件)
func (s *ResourceService) GetResourcesByRole(ctx context.Context, roleID uint) ([]model.Resource, error) {
	var resources []model.Resource

	err := s.db.WithContext(ctx).
		Distinct().
		Joins("JOIN t_role_resource ON t_role_resource.resource_id = t_resource.id").
		Where("t_role_resource.role_id = ?", roleID).
		Find(&resources).Error

	if err != nil {
		return nil, fmt.Errorf("查询角色资源失败: %w", err)
	}
	return resources, nil
}

// ListResourceRoles 查询所有资源-角色映射 (用于RBAC中间件权限检查)
// 对标Java RoleMapper.listResourceRoles()
func (s *ResourceService) ListResourceRoles(ctx context.Context) ([]dto.ResourceRoleDTO, error) {
	type ResourceRoleRow struct {
		ResourceID    uint   `gorm:"column:resource_id"`
		URL           string `gorm:"column:url"`
		RequestMethod string `gorm:"column:request_method"`
		RoleName      string `gorm:"column:role_name"`
	}

	var rows []ResourceRoleRow
	err := s.db.WithContext(ctx).
		Raw(`
			SELECT r.id AS resource_id, r.url, r.request_method, rl.role_name
			FROM t_resource r
			LEFT JOIN t_role_resource rr ON rr.resource_id = r.id
			LEFT JOIN t_role rl ON rl.id = rr.role_id
			ORDER BY r.id ASC
		`).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("查询资源角色映射失败: %w", err)
	}

	// 按 (URL, RequestMethod) 分组聚合角色列表
	groupMap := make(map[string]*dto.ResourceRoleDTO)
	for _, row := range rows {
		key := row.URL + ":" + row.RequestMethod
		if existing, ok := groupMap[key]; ok {
			existing.RoleList = append(existing.RoleList, row.RoleName)
		} else {
			groupMap[key] = &dto.ResourceRoleDTO{
				URL:           row.URL,
				RequestMethod: row.RequestMethod,
				RoleList:      []string{row.RoleName},
			}
		}
	}

	result := make([]dto.ResourceRoleDTO, 0, len(groupMap))
	for _, v := range groupMap {
		result = append(result, *v)
	}

	return result, nil
}
