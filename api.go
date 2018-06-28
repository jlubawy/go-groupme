// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package groupme implements access to the GroupMe public API.

See the API documentation: https://dev.groupme.com/docs/v3.
*/
package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// BaseURL is the base URL of which all API endpoint are built from.
const BaseURL = "https://api.groupme.com/v3"

// Client is the interface that implements the Do method for making API requests.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type client struct {
	ctx         context.Context
	client      *http.Client
	accessToken string
}

// NewClient creates a client with the given context and access token.
func NewClient(ctx context.Context, accessToken string) Client {
	if ctx == nil {
		ctx = context.Background()
	}
	return &client{
		ctx:         ctx,
		client:      http.DefaultClient,
		accessToken: accessToken,
	}
}

// Do makes an API request correctly setting the 'Content-Type' header to
// 'application/json' and the 'token' URL parameter.
func (c *client) Do(req *http.Request) (resp *http.Response, err error) {
	// Set the client context
	req = req.WithContext(c.ctx)

	// Set the access token URL parameter
	params := req.URL.Query()
	params.Set("token", c.accessToken)
	req.URL.RawQuery = params.Encode()

	// Set the content-type header
	req.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err = c.client.Do(req)
	if err != nil {
		return
	}

	// Check for any errors
	if resp.StatusCode >= 400 {
		var apiErr Error
		err = json.NewDecoder(resp.Body).Decode(&apiErr)
		if err != nil {
			return
		}
		resp.Body.Close()
		err = apiErr
	}

	return
}

// An Error is an API error message.
type Error struct {
	Meta struct {
		Code   int      `json:"code"`
		Errors []string `json:"errors"`
	} `json:"meta"`
	Response struct{} `json:"response"`
}

func (err Error) Error() string {
	return fmt.Sprintf("%+v", err.Meta.Errors)
}

// GroupsService implements all the methods needed to access the groups endpoints.
type GroupsService interface {
	Index(options *GroupsIndexOptions) (groups []Group, err error)
	Show(id string) (group Group, err error)
	Former() (groups []Group, err error)
	Create(g *Group) (group Group, err error)
	Update(id string, g *Group) (group Group, err error)
	Destroy(id string) (err error)
	Join(id string, shareToken string) (group Group, err error)
	Rejoin(id string) (group Group, err error)
	// TODO(jlubawy): implement ChangeOwners
}

type groupsService struct {
	client Client
}

func NewGroupsService(client Client) GroupsService {
	return &groupsService{
		client: client,
	}
}

// A GroupsIndexOptions sets all the options for a groups index request.
type GroupsIndexOptions struct {
	// Offset is the page offset to start the index request at. It starts at zero
	// unlike the 'page' parameter. If set to zero no parameter is sent in the
	// request and the server default value is used.
	Offset int

	// Limit limits the number of groups returned by the index request. If set
	// to zero no parameter is sent in the request and the server default value
	// is used.
	Limit int

	// Omit is a slice of strings sent as a comma-separated string in the request.
	Omit []string
}

// Index lists the authenticated user's active groups.
func (s *groupsService) Index(options *GroupsIndexOptions) (groups []Group, err error) {
	if options == nil {
		options = new(GroupsIndexOptions)
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, BaseURL+"/groups", nil)
	if err != nil {
		return
	}

	params := req.URL.Query()
	if options.Offset != 0 {
		params.Set("page", strconv.Itoa(options.Offset+1))
	}
	if options.Limit != 0 {
		params.Set("per_page", strconv.Itoa(options.Limit))
	}
	if len(options.Omit) > 0 {
		params.Set("omit", strings.Join(options.Omit, ","))
	}
	req.URL.RawQuery = params.Encode()

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Groups []Group `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	groups = respEnv.Groups
	return
}

// Former list any groups you have left but can rejoin.
func (s *groupsService) Former() (groups []Group, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, BaseURL+"/groups/former", nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Groups []Group `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	groups = respEnv.Groups
	return
}

// Show retrieves a specific group from the given ID.
func (s *groupsService) Show(id string) (group Group, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, BaseURL+fmt.Sprintf("/groups/%s", id), nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Group Group `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	group = respEnv.Group
	return
}

// Create creates a new group. See the API documentation for what fields are
// required.
func (s *groupsService) Create(g *Group) (group Group, err error) {
	if g.Name == "" {
		err = fmt.Errorf("GroupsService.Create: group name is required")
		return
	}
	if len(g.Name) > 140 {
		err = fmt.Errorf("GroupsService.Create: group name length maximum is 140 characters")
		return
	}
	if len(g.Description) > 255 {
		err = fmt.Errorf("GroupsService.Create: group description length maximum is 255 characters")
		return
	}

	reqBuf := &bytes.Buffer{}
	err = json.NewEncoder(reqBuf).Encode(g)
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+"/groups", reqBuf)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Group Group `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	group = respEnv.Group
	return
}

// Update updates a group with the given ID.
func (s *groupsService) Update(id string, g *Group) (group Group, err error) {
	if g.Name == "" {
		err = fmt.Errorf("GroupsService.Update: group name is required")
		return
	}
	if len(g.Name) > 140 {
		err = fmt.Errorf("GroupsService.Update: group name length maximum is 140 characters")
		return
	}
	if len(g.Description) > 255 {
		err = fmt.Errorf("GroupsService.Update: group description length maximum is 255 characters")
		return
	}

	reqBuf := &bytes.Buffer{}
	err = json.NewEncoder(reqBuf).Encode(g)
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+fmt.Sprintf("/groups/%s/update", id), reqBuf)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Group Group `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	group = respEnv.Group
	return
}

// Destroy disbands a group. It is only available to the group creator.
func (s *groupsService) Destroy(id string) (err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+fmt.Sprintf("/groups/%s/destroy", id), nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

// Join joins a shared group.
func (s *groupsService) Join(id string, shareToken string) (group Group, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+fmt.Sprintf("/groups/%s/join/%s", id, shareToken), nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Response struct {
			Group Group `json:"group"`
		} `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	group = respEnv.Response.Group
	return
}

// Rejoin rejoins a group. It only works if you previously left the group.
func (s *groupsService) Rejoin(id string) (group Group, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+"/groups/join", nil)
	if err != nil {
		return
	}

	params := req.URL.Query()
	params.Set("group_id", id)
	req.URL.RawQuery = params.Encode()

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Response struct {
			Group Group `json:"group"`
		} `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	group = respEnv.Response.Group
	return
}

// MembersService implements all the methods needed to access the members endpoints.
type MembersService interface {
	// TODO(jlubawy): implement the following
	// Add
	// AddResults
	// Remove
	// Update
}

type membersService struct {
	client Client
}

func NewMembersService(client Client) MembersService {
	return &membersService{
		client: client,
	}
}

// MessagesService implements all the methods needed to access the messages endpoints.
type MessagesService interface {
	// TODO(jlubawy): implement the following
	Index(groupID string, options *MessagesIndexOptions) (messages []Message, err error)
	// Create
}

type messagesService struct {
	client Client
}

func NewMessagesService(client Client) MessagesService {
	return &messagesService{
		client: client,
	}
}

type MessagesIndexOptions struct {
	// Returns messages created before the given message ID.
	BeforeID string

	// Returns most recent messages created after the given message ID
	SinceID string

	// Returns messages created immediately after the given message ID
	AfterID string

	// Number of messages returned. Default is 20. Max is 100.
	Limit int
}

func (s *messagesService) Index(groupID string, options *MessagesIndexOptions) (messages []Message, err error) {
	if options == nil {
		options = new(MessagesIndexOptions)
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, BaseURL+fmt.Sprintf("/groups/%s/messages", groupID), nil)
	if err != nil {
		return
	}

	params := req.URL.Query()
	if options.BeforeID != "" {
		params.Set("before_id", options.BeforeID)
	}
	if options.SinceID != "" {
		params.Set("since_id", options.SinceID)
	}
	if options.AfterID != "" {
		params.Set("after_id", options.AfterID)
	}
	if options.Limit != 0 {
		if options.Limit > 100 {
			err = fmt.Errorf("MessagesService.Index: page limit maximum is 100")
			return
		}
		params.Set("limit", strconv.Itoa(options.Limit))
	}
	req.URL.RawQuery = params.Encode()

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respEnv struct {
		Response struct {
			Count    int       `json:"count"`
			Messages []Message `json:"messages"`
		} `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respEnv)
	if err != nil {
		return
	}
	messages = respEnv.Response.Messages
	return
}

// ChatsService implements all the methods needed to access the chats endpoints.
type ChatsService interface {
	// TODO(jlubawy): implement the following
	// Index
}

type chatsService struct {
	client Client
}

func NewChatsService(client Client) ChatsService {
	return &chatsService{
		client: client,
	}
}

// DirectMessagesService implements all the methods needed to access the direct
// messages endpoints.
type DirectMessagesService interface {
	// TODO(jlubawy): implement the following
	// Index
	// Create
}

type directMessagesService struct {
	client Client
}

func NewDirectMessagesService(client Client) DirectMessagesService {
	return &directMessagesService{
		client: client,
	}
}

// LikesService implements all the methods needed to access the likes endpoints.
type LikesService interface {
	Create(conversationID, messageID string) (err error)
	Destroy(conversationID, messageID string) (err error)
}

type likesService struct {
	client Client
}

func NewLikesService(client Client) LikesService {
	return &likesService{
		client: client,
	}
}

func (s *likesService) Create(conversationID, messageID string) (err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+fmt.Sprintf("/messages/%s/%s/like", conversationID, messageID), nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

func (s *likesService) Destroy(conversationID, messageID string) (err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, BaseURL+fmt.Sprintf("/messages/%s/%s/unlike", conversationID, messageID), nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return
}

// LeaderboardService implements all the methods needed to access the leaderboard
// endpoints.
type LeaderboardService interface {
	// TODO(jlubawy): implement the following
	// Index
	// MyLikes
	// MyHits
}

type leaderboardService struct {
	client Client
}

func NewLeaderboardService(client Client) LeaderboardService {
	return &leaderboardService{
		client: client,
	}
}

// BotsService implements all the methods needed to access the bots endpoints.
type BotsService interface {
	// TODO(jlubawy): implement the following
	// Create
	// PostMessage
	// Index
	// Destroy
}

type botsService struct {
	client Client
}

func NewBotsService(client Client) BotsService {
	return &botsService{
		client: client,
	}
}

// UsersService implements all the methods needed to access the users endpoints.
type UsersService interface {
	// TODO(jlubawy): implement the following
	// Me
	// Update
}

type usersService struct {
	client Client
}

func NewUsersService(client Client) UsersService {
	return &usersService{
		client: client,
	}
}

// SmsService implements all the methods needed to access the SMS endpoints.
type SmsService interface {
	// TODO(jlubawy): implement the following
	// Create
	// Delete
}

type smsService struct {
	client Client
}

func NewSmsService(client Client) SmsService {
	return &smsService{
		client: client,
	}
}

// BlocksService implements all the methods needed to access the blocks endpoints.
type BlocksService interface {
	// TODO(jlubawy): implement the following
	// Index
	// BlockBetween
	// CreateBlock
	// Unblock
}

type blocksService struct {
	client Client
}

func NewBlocksService(client Client) BlocksService {
	return &blocksService{
		client: client,
	}
}
