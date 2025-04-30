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
	codeParamErr                  = "80001"
	codeInternalErr               = "80002"
	codeUserAlreadyBoundErr       = "80003"
	codeInviteCodeAlreadyBoundErr = "80004"
	codeUserSigVerifyErr          = "80005"
	codeUserTaskVerifyErr         = "80006"
	codeInviteCodeNotExistErr     = "80007"
	codeInviteCodeTypeNotMatchErr = "80008"
	codeInviteCodeNotEnoughErr    = "80009"
)
