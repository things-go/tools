package gingo

import (
	"strings"

	"github.com/thinkgos/astgo/infra"
)

type ServiceDesc struct {
	Deprecated  bool   // deprecated or not
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/v1/helloworld.proto
	Comment     string // comment
	Methods     []*MethodDesc

	UseEncoding bool
}

type MethodDesc struct {
	Deprecated bool // deprecated or not
	// method
	Name    string // 方法名
	Num     int    // 方法号
	Request string // 请求结构
	Reply   string // 回复结构
	Comment string // 方法注释
	// http_rule
	Path         string // 路径
	Method       string // 方法
	HasVars      bool   // 是否有url参数
	HasBody      bool   // 是否有消息体
	Body         string // 请求消息体
	ResponseBody string // 回复消息体
}

// TransformPathParams 路由 {xx} --> :xx
func TransformPathParams(path string) string {
	paths := strings.Split(path, "/")
	for i, p := range paths {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") || strings.HasPrefix(p, ":") {
			paths[i] = ":" + p[1:len(p)-1]
		}
	}
	return strings.Join(paths, "/")
}

func BuildPathVars(path string) (res []string) {
	for _, v := range strings.Split(path, "/") {
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			res = append(res, strings.TrimSuffix(strings.TrimPrefix(v, "{"), "}"))
		}
	}
	return
}

func CamelCaseVars(s string) string {
	vars := make([]string, 0)
	subs := strings.Split(s, ".")
	for _, sub := range subs {
		vars = append(vars, infra.CamelCase(sub))
	}
	return strings.Join(vars, ".")
}
