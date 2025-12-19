package model

import (
	"fmt"
	"time"

	"github.com/zhoudm1743/torm/db"
)

// SoftDeleteConfig 软删除配置
type SoftDeleteConfig struct {
	Enabled      bool   // 是否启用软删除
	DeletedField string // 软删除字段名
}

// DefaultSoftDeleteConfig 默认软删除配置
func DefaultSoftDeleteConfig() SoftDeleteConfig {
	return SoftDeleteConfig{
		Enabled:      false,
		DeletedField: "deleted_at",
	}
}

// SoftDeleteScope 软删除查询范围
type SoftDeleteScope int

const (
	// ScopeWithoutTrashed 不包含软删除记录（默认）
	ScopeWithoutTrashed SoftDeleteScope = iota
	// ScopeWithTrashed 包含软删除记录
	ScopeWithTrashed
	// ScopeOnlyTrashed 只包含软删除记录
	ScopeOnlyTrashed
)

// SoftDeleteQuery 软删除查询构建器
// 包装 QueryBuilder 以支持软删除范围控制和链式查询
type SoftDeleteQuery struct {
	model   *BaseModel
	scope   SoftDeleteScope
	builder *db.QueryBuilder // 缓存的查询构建器
	err     error            // 构建过程中的错误
}

// NewSoftDeleteQuery 创建软删除查询
func NewSoftDeleteQuery(model *BaseModel) *SoftDeleteQuery {
	return &SoftDeleteQuery{
		model: model,
		scope: ScopeWithoutTrashed,
	}
}

// WithTrashed 包含软删除记录
func (sdq *SoftDeleteQuery) WithTrashed() *SoftDeleteQuery {
	sdq.scope = ScopeWithTrashed
	return sdq
}

// OnlyTrashed 只查询软删除记录
func (sdq *SoftDeleteQuery) OnlyTrashed() *SoftDeleteQuery {
	sdq.scope = ScopeOnlyTrashed
	return sdq
}

// getBuilder 获取或创建查询构建器（内部方法）
func (sdq *SoftDeleteQuery) getBuilder() (*db.QueryBuilder, error) {
	if sdq.err != nil {
		return nil, sdq.err
	}
	if sdq.builder != nil {
		return sdq.builder, nil
	}

	// 使用 RawQuery 绕过自动软删除过滤，由 SoftDeleteQuery 自己控制过滤条件
	query, err := sdq.model.RawQuery()
	if err != nil {
		sdq.err = err
		return nil, err
	}

	// 根据范围添加条件
	if sdq.model.config.SoftDeletes {
		deletedAtField := sdq.model.config.DeletedAtCol
		switch sdq.scope {
		case ScopeWithoutTrashed:
			query = query.WhereNull(deletedAtField)
		case ScopeOnlyTrashed:
			query = query.WhereNotNull(deletedAtField)
		case ScopeWithTrashed:
			// 不添加任何条件，返回所有记录（包括软删除的）
		}
	}

	sdq.builder = query
	return query, nil
}

// Query 获取查询构建器（公开方法，用于高级用法）
func (sdq *SoftDeleteQuery) Query() (*db.QueryBuilder, error) {
	return sdq.getBuilder()
}

// Get 获取记录
func (sdq *SoftDeleteQuery) Get() ([]map[string]interface{}, error) {
	query, err := sdq.getBuilder()
	if err != nil {
		return nil, err
	}
	return query.Get()
}

// First 获取第一条记录
func (sdq *SoftDeleteQuery) First() (map[string]interface{}, error) {
	query, err := sdq.getBuilder()
	if err != nil {
		return nil, err
	}
	return query.First()
}

// Count 计数
func (sdq *SoftDeleteQuery) Count() (int64, error) {
	query, err := sdq.getBuilder()
	if err != nil {
		return 0, err
	}
	return query.Count()
}

// Where 添加条件
func (sdq *SoftDeleteQuery) Where(args ...interface{}) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.Where(args...)
	return sdq
}

// OrWhere 添加OR条件
func (sdq *SoftDeleteQuery) OrWhere(args ...interface{}) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.OrWhere(args...)
	return sdq
}

// WhereIn 字段值在数组中
func (sdq *SoftDeleteQuery) WhereIn(field string, values []interface{}) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.WhereIn(field, values)
	return sdq
}

// WhereNotIn 字段值不在数组中
func (sdq *SoftDeleteQuery) WhereNotIn(field string, values []interface{}) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.WhereNotIn(field, values)
	return sdq
}

// OrderBy 排序
func (sdq *SoftDeleteQuery) OrderBy(column, direction string) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.OrderBy(column, direction)
	return sdq
}

// Limit 限制记录数
func (sdq *SoftDeleteQuery) Limit(limit int) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.Limit(limit)
	return sdq
}

// Offset 偏移量
func (sdq *SoftDeleteQuery) Offset(offset int) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.Offset(offset)
	return sdq
}

// Select 选择字段
func (sdq *SoftDeleteQuery) Select(fields ...interface{}) *SoftDeleteQuery {
	query, err := sdq.getBuilder()
	if err != nil {
		return sdq
	}
	sdq.builder = query.Select(fields...)
	return sdq
}

// SoftDeleteTrait 软删除特性 - 可以嵌入到模型中
type SoftDeleteTrait struct {
	DeletedAt *time.Time `json:"deleted_at" torm:"soft_delete"`
}

// IsSoftDeleted 检查是否已软删除
func (sdt *SoftDeleteTrait) IsSoftDeleted() bool {
	return sdt.DeletedAt != nil && !sdt.DeletedAt.IsZero()
}

// MarkAsDeleted 标记为已删除
func (sdt *SoftDeleteTrait) MarkAsDeleted() {
	now := time.Now()
	sdt.DeletedAt = &now
}

// MarkAsRestored 标记为已恢复
func (sdt *SoftDeleteTrait) MarkAsRestored() {
	sdt.DeletedAt = nil
}

// WithTrashed 创建包含软删除记录的查询
func (m *BaseModel) WithTrashed() *SoftDeleteQuery {
	return NewSoftDeleteQuery(m).WithTrashed()
}

// OnlyTrashed 创建只包含软删除记录的查询
func (m *BaseModel) OnlyTrashed() *SoftDeleteQuery {
	return NewSoftDeleteQuery(m).OnlyTrashed()
}

// Restore 恢复软删除的记录（已在base_model.go中实现，这里提供别名）
// func (m *BaseModel) Restore() error - 已在base_model.go中实现

// ForceDelete 强制删除（已在base_model.go中实现）
// func (m *BaseModel) ForceDelete() error - 已在base_model.go中实现

// RestoreByPK 根据主键恢复软删除的记录
func (m *BaseModel) RestoreByPK(pk interface{}) error {
	if !m.config.SoftDeletes {
		return fmt.Errorf("该模型未启用软删除")
	}

	// 使用 RawQuery 绕过软删除过滤，否则无法找到已软删除的记录
	query, err := m.RawQuery()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		m.config.DeletedAtCol: nil,
	}

	affected, err := query.Where(m.config.PrimaryKey, "=", pk).Update(data)
	if err != nil {
		return fmt.Errorf("恢复记录失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("没有找到要恢复的记录")
	}

	return nil
}

// ForceDeleteByPK 根据主键强制删除记录
func (m *BaseModel) ForceDeleteByPK(pk interface{}) error {
	// 使用 RawQuery 绕过软删除过滤，确保可以删除已软删除的记录
	query, err := m.RawQuery()
	if err != nil {
		return err
	}

	affected, err := query.Where(m.config.PrimaryKey, "=", pk).Delete()
	if err != nil {
		return fmt.Errorf("强制删除失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("没有找到要删除的记录")
	}

	return nil
}

// SoftDeleteByPK 根据主键软删除记录
func (m *BaseModel) SoftDeleteByPK(pk interface{}) error {
	if !m.config.SoftDeletes {
		return fmt.Errorf("该模型未启用软删除")
	}

	query, err := m.Query()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		m.config.DeletedAtCol: time.Now(),
	}

	affected, err := query.Where(m.config.PrimaryKey, "=", pk).Update(data)
	if err != nil {
		return fmt.Errorf("软删除失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("没有找到要删除的记录")
	}

	return nil
}

// TrashedCount 获取软删除记录数量
func (m *BaseModel) TrashedCount() (int64, error) {
	if !m.config.SoftDeletes {
		return 0, fmt.Errorf("该模型未启用软删除")
	}

	return m.OnlyTrashed().Count()
}

// RestoreAll 恢复所有软删除的记录
func (m *BaseModel) RestoreAll() (int64, error) {
	if !m.config.SoftDeletes {
		return 0, fmt.Errorf("该模型未启用软删除")
	}

	// 使用 RawQuery 绕过软删除过滤，否则无法找到已软删除的记录
	query, err := m.RawQuery()
	if err != nil {
		return 0, err
	}

	data := map[string]interface{}{
		m.config.DeletedAtCol: nil,
	}

	return query.WhereNotNull(m.config.DeletedAtCol).Update(data)
}

// ForceDeleteAll 强制删除所有软删除的记录
func (m *BaseModel) ForceDeleteAll() (int64, error) {
	if !m.config.SoftDeletes {
		return 0, fmt.Errorf("该模型未启用软删除")
	}

	// 使用 RawQuery 绕过软删除过滤，确保可以删除已软删除的记录
	query, err := m.RawQuery()
	if err != nil {
		return 0, err
	}

	return query.WhereNotNull(m.config.DeletedAtCol).Delete()
}
