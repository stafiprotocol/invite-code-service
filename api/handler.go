package api

import (
	"invite-code-service/pkg/db"
)

type Handler struct {
	db    *db.WrapDb
	cache map[string]uint64
}

func NewHandler(db *db.WrapDb, cache map[string]uint64) *Handler {
	return &Handler{db: db, cache: cache}
}

const (
	codeParamParseErr = "80001"
	codeInternalErr   = "80002"
)
