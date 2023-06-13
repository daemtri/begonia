package app

import (
	"context"
)

type contextKey struct{ name string }

var (
	objectContainerCtxKey = &contextKey{name: "object_contianer"}
)

type objectContainer struct {
	*moduleRuntime
}

func withObjectContainer(ctx context.Context, mr *moduleRuntime) context.Context {
	return context.WithValue(ctx, objectContainerCtxKey, &objectContainer{
		moduleRuntime: mr,
	})
}

func objectContainerFromCtx(ctx context.Context) *objectContainer {
	v := ctx.Value(objectContainerCtxKey)
	if v == nil {
		panic("no object container in context")
	}
	return v.(*objectContainer)
}
