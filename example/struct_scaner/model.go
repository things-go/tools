package rapier

import (
	"database/sql"
	"time"
)

// DictItem 字典项
type DictItem struct {
	Id          int64          `gorm:"column:id;not null;autoIncrement:true;primaryKey;comment:字典项ID" json:"id,omitempty"`
	Key         string         `gorm:"column:key;type:varchar(50);not null;default:'';uniqueIndex:uk_key_value,priority:0;comment:字典key" json:"key,omitempty"`
	Name        sql.NullString `gorm:"column:name;type:varchar(50);not null;default:'';comment:字典项名称" json:"name,omitempty"`
	Value       float64        `gorm:"column:value;type:varchar(50);null;default:'';uniqueIndex:uk_key_value,priority:1;comment:字典项值" json:"value,omitempty"`
	Sort        uint32         `gorm:"column:sort;type:int(10) unsigned;not null;default:0;comment:序号" json:"sort,omitempty"`
	Remark      *string        `gorm:"column:remark;type:varchar(50);null;default:'';comment:备注" json:"remark,omitempty"`
	IgnoreField string         `gorm:"-"`
	IsEnabled   bool           `gorm:"column:is_enabled;type:tinyint(1);not null;default:0;comment:是否启用" json:"is_enabled,omitempty"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:datetime;not null;comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:datetime;not null;comment:更新时间" json:"updated_at,omitempty"`
}
