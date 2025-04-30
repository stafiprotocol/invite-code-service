package main

import (
	"invite-code-service/cmd"
)

// @title point API
// @version 1.0
// @description  point api document. Error Codes: 80001 Invalid parameters; 80002 Internal server error; 80003 User already bound; 80004 Invite code already bound; 80005 Signature verification failed; 80006 Task verification failed; 80007 Invite code does not exist; 80008 Invite code type mismatch; 80009 Invite codes not enough

// @BasePath /api
func main() {
	cmd.Execute()
}
