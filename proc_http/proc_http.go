package proc_http

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/thinkgos/astgo/proc"
)

// proc value
const (
	Identity = "http"

	AttributeName_Body         = "body"          // request body
	AttributeName_ResponseBody = "response_body" // response body
	AttributeRoute             = "route"         // http route
)

type HttpRule struct {
	// Rules http rules
	Rules []*Rule
	// Comments
	// if keep_proc is true, we will keep the proc,
	// otherwise we will remove it from comments.
	Comments []string
}

type Rule struct {
	Method       string // http method {get|put|post|delete|patch}
	Path         string // http path "/v1/hello"
	Body         string // http request body
	ResponseBody string // http response body
	Custom       bool   // http method and path is customize
}

// keep_proc 保留注解
// *HttpRule = nil means no rules
func Parse(comments []string, keep_proc bool) (*HttpRule, error) {
	hr := &HttpRule{
		Rules:    make([]*Rule, 0, 1), // in most cases only one rule
		Comments: comments,
	}
	if !keep_proc {
		hr.Comments = make([]string, 0, len(comments))
	}
	for _, comment := range comments {
		r, err := parserProc(strings.TrimPrefix(strings.TrimSpace(comment), "//"))
		if err != nil {
			return nil, err
		}
		if r != nil {
			hr.Rules = append(hr.Rules, r)
		}
		if r == nil && !keep_proc || keep_proc {
			hr.Comments = append(hr.Comments, comment)
		}
	}
	if len(hr.Rules) > 0 {
		return hr, nil
	}
	return nil, nil
}

func parserProc(comment string) (*Rule, error) {
	derive, err := proc.Match(comment)
	if err != nil || derive.Identity != Identity {
		return nil, nil
	}
	custom := false
	attrs := make(map[string]*proc.NameValue, len(derive.Attrs))
	for _, attr := range derive.Attrs {
		switch attr.Name {
		case AttributeName_Body,
			AttributeName_ResponseBody:
			attrs[attr.Name] = attr
		case "get", "put", "post", "delete", "patch", AttributeRoute:
			custom = attr.Name == AttributeRoute
			attrs[AttributeRoute] = attr
		}
	}
	r := &Rule{
		Method:       "",
		Path:         "",
		Body:         "",
		ResponseBody: "",
		Custom:       custom,
	}
	//* router
	attr, ok := attrs[AttributeRoute]
	if !ok {
		return nil, fmt.Errorf("")
	}
	if !r.Custom {
		r.Method = strings.ToUpper(attr.Name)
	} else {
		r.Method = attr.Name
	}
	if val, ok := attr.Value.(proc.String); !ok {
		return nil, fmt.Errorf("proc(http): the path type should be a string")
	} else {
		r.Path = val.Value
	}
	// body
	attr, ok = attrs[AttributeName_Body]
	if ok {
		if val, ok := attr.Value.(proc.String); !ok {
			return nil, fmt.Errorf("proc(http): the body value type should be a string")
		} else {
			r.Body = val.Value
		}
	}

	// response body
	attr, ok = attrs[AttributeName_ResponseBody]
	if ok {
		if val, ok := attr.Value.(proc.String); !ok {
			return nil, fmt.Errorf("proc(http): the response body value type should be a string")
		} else {
			r.ResponseBody = val.Value
		}
	}
	if !r.Custom &&
		slices.Contains([]string{http.MethodPost, http.MethodPut, http.MethodPatch}, r.Method) &&
		r.Body == "" {
		return nil, fmt.Errorf("%v %v is does not declare a body.", r.Method, r.Path)
	}
	return r, err

}
