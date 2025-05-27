package main

import (
	"invite-code-service/cmd"
)

// @title invite code API
// @version 1.0
// @description  invite code api document.
// @description  Error Codes:
// @description  80001 Invalid parameters
// @description  80002 Internal server error
// @description  80003 User already bound
// @description  80004 Invite code already bound
// @description  80005 Signature verification failed
// @description  80006 Task verification failed
// @description  80007 Invite code does not exist
// @description  80008 Invite code type mismatch
// @description  80009 Invite codes not enough
// @description  80010 Discord already bound
// @BasePath /api
func main() {
	cmd.Execute()
}
