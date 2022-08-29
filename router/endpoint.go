package router

import "github.com/ixtendio/gow/internal/path"

type endpoint struct {
	method   string
	rootPath *path.Element
	handler  Handler
}
