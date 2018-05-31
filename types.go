// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package groupme

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Attachment struct {
	Type string `json:"type"`

	// Image attachment fields.
	URL string `json:"url,omitempty"`

	// Location attachment fields.
	Lat  string `json:"lat,omitempty"`
	Lng  string `json:"lng,omitempty"`
	Name string `json:"name,omitempty"`

	// Mentions attachment fields.
	Loci    [][]int  `json:"loci,omitempty"`
	UserIDs []string `json:"user_ids,omitempty"`

	// Split attachment fields.
	Token string `json:"token,omitempty"`

	// Emoji attachment fields.
	Placeholder string    `json:"placeholder,omitempty"`
	Charmap     []Charmap `json:"charmap,omitempty"`
}

func (a Attachment) IsTypeImage() bool    { return a.Type == "image" }
func (a Attachment) IsTypeLocation() bool { return a.Type == "location" }
func (a Attachment) IsTypeMentions() bool { return a.Type == "mentions" }
func (a Attachment) IsTypeSplit() bool    { return a.Type == "split" }
func (a Attachment) IsTypeEmoji() bool    { return a.Type == "emoji" }

type Charmap []uint64

type Group struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Description   string   `json:"description"`
	ImageURL      string   `json:"image_url"`
	CreatorUserID string   `json:"creator_user_id"`
	CreatedAt     UnixTime `json:"created_at"`
	UpdatedAt     UnixTime `json:"updated_at"`
	Members       []Member `json:"members"`
	ShareURL      string   `json:"share_url"`
	Messages      Messages `json:"messages"`
}

type Member struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Muted    bool   `json:"muted"`
	ImageURL string `json:"image_url"`
}

type Message struct {
	ID          string       `json:"id"`
	SourceGUID  string       `json:"source_guid"`
	CreatedAt   UnixTime     `json:"created_at"`
	UserID      string       `json:"user_id"`
	GroupID     string       `json:"group_id"`
	Name        string       `json:"name"`
	AvatarURL   string       `json:"avatar_url"`
	Text        string       `json:"text"`
	System      bool         `json:"system"`
	FavoritedBy []string     `json:"favorited_by"`
	Attachments []Attachment `json:"attachments"`
}

type Messages struct {
	Count                uint64   `json:"count"`
	LastMessageID        string   `json:"last_message_id"`
	LastMessageCreatedAt UnixTime `json:"last_message_created_at"`
	Preview              Preview  `json:"preview"`
}

type Preview struct {
	Nickname    string       `json:"nickname"`
	Text        string       `json:"text"`
	ImageURL    string       `json:"image_url"`
	Attachments []Attachment `json:"attachments"`
}

type UnixTime struct {
	time.Time
}

var (
	_ json.Marshaler   = (*UnixTime)(nil)
	_ json.Unmarshaler = (*UnixTime)(nil)
)

func (t *UnixTime) MarshalJSON() (data []byte, err error) {
	data = []byte(fmt.Sprintf("%d", t.Unix()))
	return
}

func (t *UnixTime) UnmarshalJSON(data []byte) (err error) {
	var sec int64
	sec, err = strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return
	}
	(*t).Time = time.Unix(sec, 0)
	return
}
