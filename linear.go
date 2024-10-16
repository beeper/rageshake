package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	linearTeamAndroid      = "39dc6884-3753-4b41-ad19-a166a0f2f51d"
	linearTeamiOS          = "6d2402bc-d4bc-4d3d-8f5e-96df51cafe22"
	linearTeamDesktop      = "4c83bd23-2236-40b5-a250-88bbc8cc446a"
	linearTeamBackend      = "a5b96b19-c49e-4f2a-8372-206eefeba471"
	linearTeamArchitecture = "fd2966d8-5d9a-445f-bbf9-b366b2e5951b"
	linearTeamProduct      = "bcc6420f-b63d-4e44-b7de-004ee8338d80"
	linearTeamBooperBugz   = "6f30428e-b614-4753-93da-d64107d6ff91"
	linearTeamBooperEng    = "f477325d-646a-49f4-9d69-bf0e81be71c9"
	linearTeamBleeper      = "7921fd20-54a4-4aab-acc2-a8d31128d71f"
	linearTeamA8cDesktop   = "bf34ce9c-1187-40f8-b749-5f60f6e52436"
	linearTeamA8CiOS       = "bb037da9-b240-4cf8-a47c-eceed138e45a"
)

const (
	problemSignIn         = "Trouble connecting Beeper to a chat network"
	problemSend           = "I can't send a message"
	problemReceiveAny     = "I can't receive any messages"
	problemReceiveCertain = "I can't receive certain messages"
	problemUI             = "Problem with app buttons/interface/text"
	problemEncryption     = "Encryption/decryption error"
	problemNotifications  = "Notifications problem"
	problemFeatureRequest = "Feature request"
	problemBridgeRequest  = "Bridge Request"
	problemOther          = "Other"
	problemSuggestion     = "Suggestion"
)

const (
	userPriorityLow    = "low"
	userPriorityMedium = "medium"
	userPriorityHigh   = "high"
)

const (
	labelRageshake     = "3fc786e7-b4f1-472e-8e27-4aa97c2eb27c"
	labelSupportReview = "f1d19cb7-0839-4349-aa9a-f5eaec84a3a2"
	labelNewUser       = "440de7a1-3082-4180-9990-19f0f5aa6efd" // onboarding-spring-2023
	labelInternalUser  = "f72b7249-393f-46a2-9e44-62f0300aed2e"
	labelNightlyUser   = "74d1438d-eb93-4352-8efe-c6f3b291874f" // external-alpha-tester
	labelBooperApp     = "67871d86-ef4b-45bc-9144-0368e70ec9bb" // booper-app
)

var appToTeamID = map[string]string{
	"beeper-android":     linearTeamAndroid,
	"booper":             linearTeamBooperEng,
	"beeper-ios":         linearTeamiOS,
	"beeper-desktop":     linearTeamDesktop,
	"bleeper":            linearTeamBleeper,
	"beeper-a8c-desktop": linearTeamA8cDesktop,
	"beeper-a8c-ios":     linearTeamA8CiOS,
}

var problemToLabelID = map[string]string{
	problemSend:           "02805b84-e966-49ee-8c8b-ac5b3350a9e4",
	problemReceiveAny:     "140462f3-1ef2-4bad-a540-b3ee38a6a654",
	problemReceiveCertain: "f3574891-6854-46ba-b82e-b695bcbdf613",
	problemSignIn:         "14ca00de-66e3-4b18-a855-6ff86841e0e6",
	problemUI:             "b58efad1-22ed-4c88-99b2-e7d99f9c8556",
	problemEncryption:     "e57ee874-2924-4eea-9c64-57d60c478653",
	problemNotifications:  "2f562c13-2a64-44f8-a580-dd175cc4b6f5",
	problemFeatureRequest: "32c7fb7d-a155-4857-9333-2c203e7b731f",
	problemBridgeRequest:  "eed94025-eae7-4e02-9abf-870519f7369b",
	problemOther:          "0b40c728-66af-4ca9-b1fb-62c0bcda81ba",
	problemSuggestion:     "5c03e405-dcfe-4472-97e4-65fd225dbc5b",
}

var userPriorityToLabelID = map[string]string{
	userPriorityLow:    "31f988df-5abf-465c-80ba-99c9de78e877",
	userPriorityMedium: "a0a54aaa-20a4-475f-9085-362c0de94ffa",
	userPriorityHigh:   "658296b7-5d59-4963-8dbd-0bd0e48d65c1",
}

var bridgeToLabelID = map[string]string{
	"android-sms":    "23b9c42b-eb91-424a-9810-181748f98543",
	"androidsms":     "23b9c42b-eb91-424a-9810-181748f98543",
	"discord":        "6ce5f0c2-13ec-48ca-b4d8-e3f170fcfb8c",
	"discordgo":      "6ce5f0c2-13ec-48ca-b4d8-e3f170fcfb8c",
	"facebook":       "076cce46-9efb-463d-9cce-3726945091d9",
	"googlechat":     "f2fcfb8e-15ba-41f0-bd7e-6080660aa4fc",
	"gmessages":      "475775b3-3f5b-4e20-bdc1-ac4107530c0d",
	"imessage-cloud": "10ac3928-b657-409d-a1eb-4f9ec7df870e",
	"imessagecloud":  "10ac3928-b657-409d-a1eb-4f9ec7df870e",
	"imessagego":     "57992d44-cef4-46f8-a23d-3bfff810fc42",
	"imessage":       "57992d44-cef4-46f8-a23d-3bfff810fc42",
	"instagram":      "e4b3fa54-c9da-462e-a680-6946fd5ba1fb",
	"linkedin":       "d0d8b87b-6058-4093-946a-b395f31aba1e",
	"signal":         "8ea186ae-c3da-4c57-b50e-b5b82d2c32f0",
	"slack":          "306ca483-10e8-4da3-b24b-e7696466e5a9",
	"slackgo":        "306ca483-10e8-4da3-b24b-e7696466e5a9",
	"telegram":       "95089bee-0341-4363-bdf0-d420c968bb73",
	"twitter":        "35f6be99-f9f0-480e-b3e9-be29e74fa8cf",
	"whatsapp":       "efd1d28a-5188-4ab3-9a27-51a63f9c7a16",
}

type GraphQLRequest struct {
	Token     string                 `json:"-"`
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLError struct {
	Message    string `json:"message"`
	Extensions struct {
		UserPresentableMessage string `json:"userPresentableMessage"`
	}
}

type GraphQLResponse struct {
	Errors []GraphQLError
	Data   json.RawMessage
}

type CreateIssueResponse struct {
	IssueCreate struct {
		Success bool
		Issue   struct {
			ID         string
			Title      string
			Identifier string
			URL        string
		}
	}
}

const mutationCreateIssue = `
mutation CreateIssue($input: IssueCreateInput!) {
    issueCreate(input: $input) {
        success
        issue {
            id
            title
            identifier
            url
        }
    }
}
`

type UserEmail struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetUserEmailsResponse struct {
	Users struct {
		Nodes []UserEmail `json:"nodes"`
	} `json:"users"`
}

// This will start missing users if we have more than 250 active linear accounts
const queryGetUserEmails = `
query {
	users(first: 250) {
		nodes {
			id
			name
			email
		}
	}
}
`

const queryFindUserByEmail = `
query FindUserByEmail($filter: UserFilter!) {
	users(filter: $filter) {
		nodes {
			id
			name
			email
		}
	}
}
`

var emailToLinearIDCache = make(map[string]string)

func isValidEmailLocalpart(localpart string) bool {
	for _, ch := range localpart {
		if (ch < 'a' || ch > 'z') && ch != '.' && ch != '-' {
			return false
		}
	}
	return true
}

func getLinearID(ctx context.Context, email, token string) string {
	if len(email) > 100 {
		return ""
	}
	// Ensure there's only one @ and the domain is beeper.com
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] != "beeper.com" {
		return ""
	}
	// Remove anything after a +
	parts = strings.Split(parts[0], "+")
	if !isValidEmailLocalpart(parts[0]) {
		return ""
	}
	email = fmt.Sprintf("%s@beeper.com", parts[0])
	userID, ok := emailToLinearIDCache[email]
	if ok {
		return userID
	}
	log := zerolog.Ctx(ctx).With().
		Str("email", email).
		Str("user_id", userID).
		Logger()

	log.Warn().Msg("Linear user ID for email is not cached, fetching from Linear")

	var userResp GetUserEmailsResponse
	err := LinearRequest(ctx, &GraphQLRequest{
		Token: token,
		Query: queryFindUserByEmail,
		Variables: map[string]any{
			"filter": map[string]any{"email": map[string]any{"eq": email}},
		},
	}, &userResp)
	if err != nil {
		log.Err(err).Msg("Error finding linear user ID for email")
		emailToLinearIDCache[email] = ""
		return ""
	}
	for _, user := range userResp.Users.Nodes {
		log.Info().
			Str("found_email", user.Email).
			Str("found_name", user.Name).
			Str("found_id", user.ID).
			Msg("Found linear user ID for email")
		emailToLinearIDCache[user.Email] = user.ID
	}
	return emailToLinearIDCache[email]
}

func fillEmailCache(ctx context.Context, token string) error {
	var userResp GetUserEmailsResponse
	err := LinearRequest(ctx, &GraphQLRequest{
		Token: token,
		Query: queryGetUserEmails,
	}, &userResp)
	if err != nil {
		return err
	}
	for _, user := range userResp.Users.Nodes {
		zerolog.Ctx(ctx).Info().
			Str("email", user.Email).
			Str("name", user.Name).
			Str("user_id", user.ID).
			Msg("Found linear user ID for email")
		emailToLinearIDCache[user.Email] = user.ID
	}
	return nil
}

func LinearRequest(ctx context.Context, payload *GraphQLRequest, into any) error {
	log := zerolog.Ctx(ctx).With().Str("action", "linear_request").Logger()

	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request JSON: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linear.app/graphql", &buf)
	if err != nil {
		return fmt.Errorf("failed to create GraphQL request: %w", err)
	}
	req.Header.Add("Authorization", payload.Token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send GraphQL request: %w", err)
	}
	defer resp.Body.Close()
	var respData GraphQLResponse
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		log.Error().Int("status_code", resp.StatusCode).Str("resp_data", string(data)).Msg("Got non-200 response")
	} else if json.Valid(data) {
		log.Info().RawJSON("resp_data", data).Msg("Received GraphQL response from Linear")
	} else {
		log.Warn().Str("resp_data_invalid", string(data)).Msg("Received non-JSON GraphQL response from Linear")
	}
	err = json.Unmarshal(data, &respData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response JSON (status %d): %w: %s", resp.StatusCode, err, data)
	}
	if len(respData.Errors) > 0 {
		if len(respData.Errors[0].Extensions.UserPresentableMessage) > 0 {
			return fmt.Errorf("GraphQL error: %s", respData.Errors[0].Extensions.UserPresentableMessage)
		}
		return fmt.Errorf("GraphQL error: %s", respData.Errors[0].Message)
	}
	if into != nil {
		err = json.Unmarshal(respData.Data, &into)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response data: %w", err)
		}
	}
	return nil
}
