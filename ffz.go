package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	FfzApiBase = "https://api.frankerfacez.com/v1"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

type SetList map[string]FfzSet

type FfzGlobal struct {
	DefaultSets []int   `json:"default_sets"`
	Sets        SetList `json:"sets"`
}

type FfzRoom struct {
	RoomInfo FfzRoomInfo `json:"room"`
	Sets     SetList     `json:"sets"`
}

type FfzRoomInfo struct {
	ID          int    `json:"_id"`
	DisplayName string `json:"display_name"`
	Set         int    `json:"set"`
}

type FfzSet struct {
	Emoticons []FfzEmote `json:"emoticons"`
}

type FfzEmote struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (e *FfzEmote) Identifier() string {
	return "ffz:" + e.Name
}

func getFfzRoomEmotes(roomName string) []FfzEmote {
	globalEmotes := &FfzGlobal{}
	err := getJson(FfzApiBase+"/set/global", globalEmotes)
	if err != nil {
		panic(err)
	}

	roomEmotes := &FfzRoom{}
	err = getJson(FfzApiBase+"/room/"+strings.ToLower(roomName), roomEmotes)
	if err != nil {
		panic(err)
	}

	globalSets := make([]FfzSet, len(globalEmotes.DefaultSets))
	totalCount := 0
	for i, v := range globalEmotes.DefaultSets {
		globalSets[i] = globalEmotes.Sets[strconv.Itoa(v)]
		totalCount += len(globalSets[i].Emoticons)
	}

	roomSet := roomEmotes.Sets[strconv.Itoa(roomEmotes.RoomInfo.Set)]
	totalCount = len(roomSet.Emoticons) + totalCount

	allEmotes := make([]FfzEmote, totalCount)

	l := 0
	for _, gSet := range globalSets {
		for _, gEmote := range gSet.Emoticons {
			allEmotes[l] = gEmote
			l++
		}
	}

	for _, rEmote := range roomSet.Emoticons {
		allEmotes[l] = rEmote
		l++
	}

	return allEmotes
}
