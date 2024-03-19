//go:generate ast-gen-dyn gin -I Hello
package dyn

import (
	protoenum "github.com/things-go/protogen-saber/internal/protoenum"
)

type ListHelloRequest struct {
	Title   string
	Page    int64
	PerPage int64
}

type ListHelloResponse struct {
	Total   int64
	Page    int64
	PerPage int64
	List    []int64
}

/*
Hello 呀
*/
type Hello interface {
	// ListHello 获取你好列表
	ListHello(*protoenum.Enum) *ListHelloResponse
	// GetHello 获取你好
	GetHello(req *protoenum.Enum) *ListHelloResponse
}
