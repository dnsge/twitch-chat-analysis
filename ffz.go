package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	FfzAPIBase = "https://api.frankerfacez.com/v1"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

type setList map[string]ffzSet

type ffzGlobal struct {
	DefaultSets []int   `json:"default_sets"`
	Sets        setList `json:"sets"`
}

type ffzRoom struct {
	RoomInfo ffzRoomInfo `json:"room"`
	Sets     setList     `json:"sets"`
}

type ffzRoomInfo struct {
	ID          int    `json:"_id"`
	DisplayName string `json:"display_name"`
	Set         int    `json:"set"`
}

type ffzSet struct {
	Emoticons []FfzEmote `json:"emoticons"`
}

type FfzEmote struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (e *FfzEmote) Identifier() string {
	return "ffz:" + e.Name
}

func GetFfzRoomEmotes(roomName string) ([]FfzEmote, error) {
	globalEmotes := &ffzGlobal{}
	err := getJSON(FfzAPIBase+"/set/global", globalEmotes)
	if err != nil {
		return nil, err
	}

	roomEmotes := &ffzRoom{}
	err = getJSON(FfzAPIBase+"/room/"+strings.ToLower(roomName), roomEmotes)
	if err != nil {
		return nil, err
	}

	globalSets := make([]ffzSet, len(globalEmotes.DefaultSets))
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

	return allEmotes, nil
}
