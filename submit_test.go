package main

import "testing"

type usernameOutput struct {
	username   string
	adminLink  string
	isVerified bool
}

var usernamePayloadTests = []struct {
	name string
	in   map[string]string
	out  usernameOutput
}{
	{
		name: "unverified",
		in: map[string]string{
			"unverified_user_id": "heeheehooha",
		},
		out: usernameOutput{
			username:   "[unverified] heeheehooha",
			isVerified: false,
		},
	},
	{
		name: "verified overrides unverified",
		in: map[string]string{
			"unverified_user_id": "heeheehooha",
			"user_id":            "@bender:beeper.com",
		},
		out: usernameOutput{
			username:   "bender",
			adminLink:  "https://admin.beeper.com/user/bender",
			isVerified: true,
		},
	},
	{
		in: map[string]string{
			"user_id": "@bender:beeper-staging.com",
		},
		out: usernameOutput{
			username:   "bender:beeper-staging.com",
			adminLink:  "https://admin.beeper-staging.com/user/bender",
			isVerified: true,
		},
	},
	{
		in: map[string]string{
			"user_id": "@bender:beeper-anywhere.com",
		},
		out: usernameOutput{
			username:   "bender:beeper-anywhere.com",
			adminLink:  "https://admin.beeper.com/user/bender:beeper-anywhere.com",
			isVerified: true,
		},
	},
}

func TestGetUsernameFromPayload(t *testing.T) {
	for _, tt := range usernamePayloadTests {
		t.Run(tt.name, func(t *testing.T) {
			username, adminLink, isVerified := getUsernameFromPayload(parsedPayload{
				Data: tt.in,
			})
			out := usernameOutput{username, adminLink, isVerified}
			if out != tt.out {
				t.Errorf("got %+v, want %+v", out, tt.out)
			}
		})
	}
}
