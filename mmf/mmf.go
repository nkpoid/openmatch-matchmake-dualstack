package mmf

import (
	"errors"

	"github.com/google/uuid"
	om "open-match.dev/open-match/pkg/pb"
)

const (
	MatchFunctionName = "dualstack-demo-match-function"
	V4Tag             = "v4"
	V6Tag             = "v6"
)

var FailedMatchMakeErr = errors.New("failed to create match")

func extractTicketsWithTag(tickets []*om.Ticket, tag string) []*om.Ticket {
	contains := func(s []string, e string) bool {
		for _, v := range s {
			if e == v {
				return true
			}
		}
		return false
	}

	var ret []*om.Ticket

	for _, ticket := range tickets {
		ticket := ticket
		if contains(ticket.SearchFields.GetTags(), tag) {
			ret = append(ret, ticket)
		}
	}

	return ret
}

// makeMatch は，Ticketの配列を受け取り，1v1のマッチを最大1つ作成します．
// マッチメイキングに失敗した場合，FailedMatchMakeErrを返却します．
func makeMatch(tickets []*om.Ticket, profile *om.MatchProfile) (*om.Match, error) {
	generateMatch := func(ts []*om.Ticket) *om.Match {
		return &om.Match{
			MatchId:       uuid.NewString(),
			MatchProfile:  profile.Name,
			MatchFunction: MatchFunctionName,
			Tickets:       ts,
		}
	}

	// 最初に，v6ユーザーの抽出を試みる
	v6Tickets := extractTicketsWithTag(tickets, V6Tag)
	if len(v6Tickets) >= 2 {
		return generateMatch(v6Tickets[:2]), nil
	}

	// 次に，v4ユーザー(含v6 dualstack)の抽出を試みます．
	v4Tickets := extractTicketsWithTag(tickets, V4Tag)
	if len(v4Tickets) >= 2 {
		return generateMatch(v4Tickets[:2]), nil
	}

	return nil, FailedMatchMakeErr
}

// MakeMatches は，チケットプールを受け取り，メイキングしたマッチのリストを返却します．
// 1件もマッチング成功しなかった場合，FailedMatchMakeErrを返却します．
func MakeMatches(ticketPools map[string][]*om.Ticket, profile *om.MatchProfile) ([]*om.Match, error) {
	var matches []*om.Match

	for _, tickets := range ticketPools {
		// TODO: poolを可能な限り使い切るようにする
		match, err := makeMatch(tickets, profile)
		if err != nil {
			continue
		}
		matches = append(matches, match)
	}

	if len(matches) == 0 {
		return nil, FailedMatchMakeErr
	}

	return matches, nil
}
