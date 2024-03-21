//go:generate ast-gen-dyn gin -i Dict
package dyn

type DictEntry struct {
	// 字典ID
	// #[seaql(type="bigint(20) NOT NULL AUTO_INCREMENT")]
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// 关键字
	// #[seaql(type="varchar(50) NOT NULL DEFAULT ”")]
	Key string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	// 名称
	// #[seaql(type="varchar(50) NOT NULL DEFAULT ”")]
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// 备注
	// #[seaql(type="varchar(100) NOT NULL DEFAULT ”")]
	Remark string `protobuf:"bytes,4,opt,name=remark,proto3" json:"remark,omitempty"`
	// 创建时间
	// #[seaql(type="datetime NOT NULL")]
	CreatedAt int64 `protobuf:"varint,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// 更新时间
	// #[seaql(type="datetime NOT NULL")]
	UpdatedAt int64 `protobuf:"varint,6,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}
type ListDictRequest struct {
	// 关键字
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// 名称
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// @gotags: binding:"gt=0"
	Page int64 `protobuf:"varint,30,opt,name=page,proto3" json:"page,omitempty" binding:"gt=0"`
	// @gotags: binding:"min=1,max=500"
	PerPage int64 `protobuf:"varint,31,opt,name=per_page,json=perPage,proto3" json:"per_page,omitempty" binding:"min=1,max=500"`
}

type ListDictResponse struct {
	Total   int64        `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Page    int64        `protobuf:"varint,30,opt,name=page,proto3" json:"page,omitempty"`
	PerPage int64        `protobuf:"varint,31,opt,name=per_page,json=perPage,proto3" json:"per_page,omitempty"`
	List    []*DictEntry `protobuf:"bytes,32,rep,name=list,proto3" json:"list,omitempty"`
}

type GetDictRequest struct {
	// @gotags: binding:"required"
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty" binding:"required"`
}

type GetDictResponse struct {
	Dict *DictEntry `protobuf:"bytes,1,opt,name=dict,proto3" json:"dict,omitempty"`
}

type AddDictRequest struct {
	// @gotags: binding:"required"
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty" binding:"required"`
	// @gotags: binding:"required"
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty" binding:"required"`
	IsPin  bool   `protobuf:"varint,3,opt,name=is_pin,json=isPin,proto3" json:"is_pin,omitempty"`
	Remark string `protobuf:"bytes,4,opt,name=remark,proto3" json:"remark,omitempty"`
}

type UpdateDictRequest struct {
	// @gotags: binding:"gt=0"
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty" binding:"gt=0"`
	// @gotags: binding:"required"
	Key string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty" binding:"required"`
	// @gotags: binding:"required"
	Name   string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty" binding:"required"`
	Remark string `protobuf:"bytes,4,opt,name=remark,proto3" json:"remark,omitempty"`
}

type DeleteDictRequest struct {
	// @gotags: binding:"required"
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty" binding:"required"`
}

type Empty struct{}

// 字典服务
type Dict interface {
	// ListDict 获取字典列表
	// #[http(get="/v1/dicts")]
	ListDict(*ListDictRequest) *ListDictResponse
	// GetDict 获取字典
	// #[http(get="/v1/dicts/{id}")]
	GetDict(*GetDictRequest) *GetDictResponse
	// AddDict 增加字典
	// #[http(post="/v1/dicts", body="*")]
	AddDict(*AddDictRequest) *Empty
	// UpdateDict 更新字典
	// #[http(put="/v1/dicts/{id}",body="*")]
	UpdateDict(*UpdateDictRequest) *Empty
	// DeleteDict 删除字典
	// #[http(delete="/v1/dicts/{id}")]
	DeleteDict(*DeleteDictRequest) *Empty
}
