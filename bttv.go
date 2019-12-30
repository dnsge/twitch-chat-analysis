package main

import (
	"strings"
)

const (
	BttvAPIBase = "https://api.betterttv.net/2"
)

type bttvEmoteResp struct {
	Emotes []BttvEmote `json:"emotes"`
}

type BttvEmote struct {
	ID   string `json:"id"`
	Name string `json:"code"`
}

func (e *BttvEmote) Identifier() string {
	return "bttv:" + e.Name
}

func GetBttvRoomEmotes(roomName string) ([]BttvEmote, error) {
	globalEmotes := &bttvEmoteResp{}
	err := getJSON(BttvAPIBase+"/emotes", globalEmotes)
	if err != nil {
		return nil, err
	}

	roomEmotes := &bttvEmoteResp{}
	err = getJSON(BttvAPIBase+"/channels/"+strings.ToLower(roomName), roomEmotes)
	if err != nil {
		return nil, err
	}

	return append(globalEmotes.Emotes, roomEmotes.Emotes...), nil
}
