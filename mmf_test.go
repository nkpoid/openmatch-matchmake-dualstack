package mmf

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	om "open-match.dev/open-match/pkg/pb"
)

func makeTicket(t *testing.T, tags ...string) *om.Ticket {
	t.Helper()

	return &om.Ticket{
		Id:           uuid.NewString(),
		SearchFields: &om.SearchFields{Tags: tags},
	}
}

func TestMakeMatches(t *testing.T) {
	profile := &om.MatchProfile{Name: "fake"}

	v4OnlyTicket1 := makeTicket(t, v4Tag)
	v4OnlyTicket2 := makeTicket(t, v4Tag)
	v6OnlyTicket := makeTicket(t, v6Tag)
	dualstackTicket1 := makeTicket(t, v4Tag, v6Tag)
	dualstackTicket2 := makeTicket(t, v4Tag, v6Tag)

	type testCase struct {
		tickets     []*om.Ticket
		wantMatches []*om.Match
		wantError   error
	}
	for name, tt := range map[string]testCase{
		"NG: v4とv6ユーザー同士ではマッチングしない": {
			tickets: []*om.Ticket{v4OnlyTicket1, v6OnlyTicket},
			wantError: FailedMatchMakeErr,
		},
		"OK: v4同士でマッチングができる": {
			tickets: []*om.Ticket{v4OnlyTicket1, v4OnlyTicket2},
			wantMatches: []*om.Match{
				{
					MatchProfile:  "fake",
					MatchFunction: MatchFunctionName,
					Tickets:       []*om.Ticket{v4OnlyTicket1, v4OnlyTicket2},
				},
			},
		},
		"OK: デュアルスタック同士でマッチングができる": {
			tickets: []*om.Ticket{dualstackTicket1, dualstackTicket2},
			wantMatches: []*om.Match{
				{
					MatchProfile:  "fake",
					MatchFunction: MatchFunctionName,
					Tickets:       []*om.Ticket{dualstackTicket1, dualstackTicket2},
				},
			},
		},
		"OK: v4とデュアルスタック同士でマッチングができる": {
			tickets: []*om.Ticket{dualstackTicket1, v4OnlyTicket1},
			wantMatches: []*om.Match{
				{
					MatchProfile:  "fake",
					MatchFunction: MatchFunctionName,
					Tickets:       []*om.Ticket{dualstackTicket1, v4OnlyTicket1},
				},
			},
		},
	} {
		tt := tt
		t.Run(name, func(t *testing.T) {
			pool := map[string][]*om.Ticket{
				"test-pool": tt.tickets,
			}
			matches, err := MakeMatches(pool, profile)
			if err != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("mismatch expected error. want: %v, got %v", tt.wantError, err)
				}
				return
			}

			opts := []cmp.Option{
				protocmp.Transform(),
				protocmp.IgnoreFields(&om.Match{},
					"match_id",
					"match_profile",
					"match_function",
					"tickets",
					"extensions",
				),
			}
			if diff := cmp.Diff(matches, tt.wantMatches, opts...); diff != "" {
				t.Errorf("mismatch. diff: %v", diff)
			}
		})
	}
}
