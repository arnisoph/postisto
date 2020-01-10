package filter

import (
	"github.com/arnisoph/postisto/pkg/server"
)

//type UnknownCommandTypeError struct {
//	typeName string
//}
//
//func (err *UnknownCommandTypeError) Error() string {
//	return fmt.Sprintf("Command type %q unknown", err.typeName)
//}
//
//type BadCommandTargetError struct {
//	targetName string
//}
//
//func (err *BadCommandTargetError) Error() string {
//	return fmt.Sprintf("Bad command target %q", err.targetName)
//}

func RunCommands(srv *server.Connection, from string, uid uint32, cmds FilterOps) error {
	var err error
	uids := []uint32{uid}

	if cmds["move"] != nil {
		if err := srv.Move(uids, from, cmds["move"].(string)); err != nil {
			return err
		}
	}

	to := from
	if cmds["move"] != nil {
		to = cmds["move"].(string)
	}

	if cmds["add_flags"] != nil {
		if err := srv.SetFlags(to, uids, "+FLAGS", cmds["add_flags"].([]interface{}), false); err != nil {
			return err
		}
	}

	if cmds["remove_flags"] != nil {
		if err := srv.SetFlags(to, uids, "-FLAGS", cmds["remove_flags"].([]interface{}), false); err != nil {
			return err
		}
	}

	if cmds["replace_all_flags"] != nil {
		if err := srv.SetFlags(to, uids, "FLAGS", cmds["replace_all_flags"].([]interface{}), false); err != nil {
			return err
		}
	}

	return err
}
