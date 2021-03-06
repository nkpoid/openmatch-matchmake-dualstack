package mmf_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	om "open-match.dev/open-match/pkg/pb"

	"github.com/nkpoid/openmatch-matchmake-dualstack/mmf"
)

const MatchProfileName = "fake"

func makeTicket(t *testing.T, tags ...string) *om.Ticket {
	t.Helper()

	return &om.Ticket{
		Id:           uuid.NewString(),
		SearchFields: &om.SearchFields{Tags: tags},
	}
}

func TestMakeMatches(t *testing.T) {
	profile := &om.MatchProfile{Name: MatchProfileName}

	v4OnlyTicket1 := makeTicket(t, mmf.V4Tag)
	v4OnlyTicket2 := makeTicket(t, mmf.V4Tag)
	v6OnlyTicket1 := makeTicket(t, mmf.V6Tag)
	v6OnlyTicket2 := makeTicket(t, mmf.V6Tag)
	dualstackTicket1 := makeTicket(t, mmf.V4Tag, mmf.V6Tag)
	dualstackTicket2 := makeTicket(t, mmf.V4Tag, mmf.V6Tag)

	type testCase struct {
		in          []*om.Ticket
		wantMatches []*om.Match
		wantError   error
	}
	for name, tt := range map[string]testCase{
		"NG: v4とv6ユーザー同士ではマッチングしない": {
			in:        []*om.Ticket{v4OnlyTicket1, v6OnlyTicket1},
			wantError: mmf.FailedMatchMakeErr,
		},
		"OK: v4同士でマッチングができる": {
			in: []*om.Ticket{v4OnlyTicket1, v4OnlyTicket2},
			wantMatches: []*om.Match{
				{
					MatchProfile:  MatchProfileName,
					MatchFunction: mmf.MatchFunctionName,
					Tickets:       []*om.Ticket{v4OnlyTicket1, v4OnlyTicket2},
				},
			},
		},
		"OK: デュアルスタック同士でマッチングができる": {
			in: []*om.Ticket{dualstackTicket1, dualstackTicket2},
			wantMatches: []*om.Match{
				{
					MatchProfile:  MatchProfileName,
					MatchFunction: mmf.MatchFunctionName,
					Tickets:       []*om.Ticket{dualstackTicket1, dualstackTicket2},
				},
			},
		},
		"OK: v4とデュアルスタックでマッチングができる": {
			in: []*om.Ticket{dualstackTicket1, v4OnlyTicket1},
			wantMatches: []*om.Match{
				{
					MatchProfile:  MatchProfileName,
					MatchFunction: mmf.MatchFunctionName,
					Tickets:       []*om.Ticket{dualstackTicket1, v4OnlyTicket1},
				},
			},
		},
		"OK: v6同士でマッチングができる": {
			in: []*om.Ticket{v6OnlyTicket1, v6OnlyTicket2},
			wantMatches: []*om.Match{
				{
					MatchProfile:  MatchProfileName,
					MatchFunction: mmf.MatchFunctionName,
					Tickets:       []*om.Ticket{v6OnlyTicket1, v6OnlyTicket2},
				},
			},
		},
		"OK: v6とデュアルスタックでマッチングができる": {
			in: []*om.Ticket{dualstackTicket1, v6OnlyTicket1},
			wantMatches: []*om.Match{
				{
					MatchProfile:  MatchProfileName,
					MatchFunction: mmf.MatchFunctionName,
					Tickets:       []*om.Ticket{dualstackTicket1, v6OnlyTicket1},
				},
			},
		},
	} {
		tt := tt
		t.Run(name, func(t *testing.T) {
			pool := map[string][]*om.Ticket{
				"test-pool": tt.in,
			}
			matches, err := mmf.MakeMatches(pool, profile)
			if err != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("mismatch expected error. want: %v, got %v", tt.wantError, err)
				}
				return
			}

			opts := []cmp.Option{
				protocmp.Transform(),
				protocmp.IgnoreFields(&om.Match{}, "match_id"),
			}
			if diff := cmp.Diff(matches, tt.wantMatches, opts...); diff != "" {
				t.Errorf("mismatch. diff: %v", diff)
			}
		})
	}
}
