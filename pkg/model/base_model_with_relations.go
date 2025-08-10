package model

// 关联关系方法扩展

// HasOne 定义一对一关联关系
func (m *BaseModel) HasOne(related interface{}, foreignKey, localKey string) *HasOne {
	return NewHasOne(m, related, foreignKey, localKey)
}

// HasMany 定义一对多关联关系
func (m *BaseModel) HasMany(related interface{}, foreignKey, localKey string) *HasMany {
	return NewHasMany(m, related, foreignKey, localKey)
}

// BelongsTo 定义多对一关联关系
func (m *BaseModel) BelongsTo(related interface{}, foreignKey, ownerKey string) *BelongsTo {
	return NewBelongsTo(m, related, foreignKey, ownerKey)
}

// BelongsToMany 定义多对多关联关系
func (m *BaseModel) BelongsToMany(related interface{}, pivotTable, pivotForeignKey, pivotRelatedKey, localKey, relatedKey string) *ManyToMany {
	return NewManyToMany(m, related, pivotTable, pivotForeignKey, pivotRelatedKey, localKey, relatedKey)
}
