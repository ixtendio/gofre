package gow

import (
	"context"
	"html/template"
)

type ctxKey int

const KeyValues ctxKey = 1

type CtxValues struct {
	ContextPath string
	Template    *template.Template
}

func GetContextPath(ctx context.Context) string {
	ctxVal := ctx.Value(KeyValues).(CtxValues)
	return ctxVal.ContextPath
}

func GetTemplate(ctx context.Context) *template.Template {
	ctxVal := ctx.Value(KeyValues).(CtxValues)
	return ctxVal.Template
}
