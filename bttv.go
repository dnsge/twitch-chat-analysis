package main

import (
	"strings"
)

const (
	BttvApiBase = "https://api.betterttv.net/2"
)

type BttvEmoteResp struct {
	Emotes []BttvEmote `json:"emotes"`
}

type BttvEmote struct {
	ID   string `json:"id"`
	Name string `json:"code"`
}

func (e *BttvEmote) Identifier() string {
	return "bttv:" + e.Name
}

func getBttvRoomEmotes(roomName string) ([]BttvEmote, error) {
	globalEmotes := &BttvEmoteResp{}
	err := getJson(BttvApiBase+"/emotes", globalEmotes)
	if err != nil {
		return nil, err
	}

	roomEmotes := &BttvEmoteResp{}
	err = getJson(BttvApiBase+"/channels/"+strings.ToLower(roomName), roomEmotes)
	if err != nil {
		return nil, err
	}

	return append(globalEmotes.Emotes, roomEmotes.Emotes...), nil
}
