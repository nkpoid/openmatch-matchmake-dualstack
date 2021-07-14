package mmf

import (
	"github.com/google/uuid"
	"open-match.dev/open-match/pkg/pb"
)

const MatchFunctionName = "dualstack-demo-match-function"

func makeMatch(tickets []*pb.Ticket, profile *pb.MatchProfile) (*pb.Match, error) {
	return &pb.Match{
		MatchId:       uuid.NewString(),
		MatchProfile:  profile.Name,
		MatchFunction: MatchFunctionName,
		Tickets:       []*pb.Ticket{},
	}, nil
}

func MakeMatches(ticketPools map[string][]*pb.Ticket, profile *pb.MatchProfile) ([]*pb.Match, error) {
	var matches []*pb.Match

	for _, tickets := range ticketPools {
		// TODO: poolを可能な限り使い切るようにする
		match, err := makeMatch(tickets, profile)
		if err != nil {
			continue
		}
		matches = append(matches, match)
	}

	return matches, nil
}
