package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
)

//func (filterSet FilterSet) Names() []string {
//	keys := make([]string, len(filterSet))
//
//	var i uint64
//	for key, _ := range filterSet {
//		keys[i] = key
//		i++
//	}
//
//	return keys
//}

type Filter struct {
	Commands FilterOps `yaml:"commands,flow"`
	RuleSet  RuleSet   `yaml:"rules"`
}
type FilterOps map[string]interface{}
type RuleSet []Rule
type Rule map[string][]map[string]interface{}

func GetUnsortedMsgs(srv *server.Connection, mailbox string, withoutFlags []string) ([]*server.Message, error) {
	return srv.SearchAndFetch(mailbox, nil, withoutFlags)
}

func EvaluateFilterSetsOnMsgs(srv *server.Connection, inputMailbox string, inputWithoutFlags []string, fallbackMailbox string, filterSet map[string]Filter) error {

	var remainingMsgs []*server.Message
	msgs, err := GetUnsortedMsgs(srv, inputMailbox, inputWithoutFlags)

	for _, msg := range msgs {
		var matched bool

		log.Infow("Found new message in input mailbox to sort", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId)

		log.Debugw("Starting to filter message", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId)
		for filterName, filterConfig := range filterSet {
			log.Debugw(fmt.Sprintf("Evaluate filter %q against message headers", filterName), "uid", msg.RawMessage.Uid, "ruleSet", filterConfig.RuleSet)
			matched, err = ParseRuleSet(filterConfig.RuleSet, msg.Headers)

			if err != nil {
				return err
			}

			if !matched {
				continue
			}

			log.Infow("IT'S A MATCH! Apply commands to message via IMAP..", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "cmd", filterConfig.Commands)
			err = RunCommands(srv, inputMailbox, msg.RawMessage.Uid, filterConfig.Commands)
			if err != nil {
				log.Errorw("Failed to run command on matched message", err, "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "cmd", filterConfig.Commands)
				return err
			}

			break
		}

		if !matched {
			log.Debugw("No filter matched to this message, scheduling fallback action (flag/move)", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "headers", msg.Headers)
			remainingMsgs = append(remainingMsgs, msg)
		}
	}

	for _, msg := range remainingMsgs {
		if fallbackMailbox == inputMailbox || fallbackMailbox == "" {
			log.Infow("No filter matched to this message. Flagging the message now.", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "flags", []interface{}{server.FlaggedFlag})
			if err := srv.SetFlags(inputMailbox, []uint32{msg.RawMessage.Uid}, "+FLAGS", []interface{}{server.FlaggedFlag}, false); err != nil {
				return err
			}
		} else {
			log.Infow("No filter matched to this message. Moving it to the fallback mailbox now.", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "mailbox", fallbackMailbox)
			if err := srv.Move([]uint32{msg.RawMessage.Uid}, inputMailbox, fallbackMailbox); err != nil {
				return err
			}
		}
	}

	return nil
}
