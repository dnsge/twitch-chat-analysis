package main

import (
	"encoding/json"
	"errors"
	"github.com/gempir/go-twitch-irc"
	"strconv"
	"strings"
	"time"
)

func getJson(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	return json.NewDecoder(r.Body).Decode(target)
}

type ChatMessage struct {
	*twitch.PrivateMessage
}

type CustomEmotes struct {
	FfzEmotes  []FfzEmote  `json:"ffz"`
	BttvEmotes []BttvEmote `json:"bttv"`
}

func LoadCustomEmotes(roomName string) *CustomEmotes {
	return &CustomEmotes{
		FfzEmotes:  getFfzRoomEmotes(roomName),
		BttvEmotes: getBttvRoomEmotes(roomName),
	}
}

func (ce *CustomEmotes) IsEmote(word string) (bool, string) {
	for _, e := range ce.FfzEmotes {
		if word == e.Name {
			return true, e.Identifier()
		}
	}

	for _, e := range ce.BttvEmotes {
		if word == e.Name {
			return true, e.Identifier()
		}
	}

	return false, ""
}

func (m *ChatMessage) ExtractEmoteIdentifiers(ce *CustomEmotes) []string {
	var emotes []string

	for _, word := range strings.Split(m.Message, " ") {
		if isEmote, id := ce.IsEmote(word); isEmote {
			emotes = append(emotes, id)
		}
	}

	for _, emote := range m.Emotes {
		for i := 0; i < emote.Count; i++ {
			emotes = append(emotes, "ttv:"+emote.Name)
		}
	}

	return emotes
}

func (m *ChatMessage) isSubscribed() bool {
	_, hasSub := m.User.Badges["subscriber"]
	return hasSub
}

func (m *ChatMessage) hasPrime() bool {
	val, hasPremium := m.User.Badges["premium"]
	return hasPremium && val == 1
}

func (m *ChatMessage) subCategory() int {
	val, hasSub := m.User.Badges["subscriber"]
	if !hasSub {
		return 0
	} else {
		return val
	}
}

func (m *ChatMessage) subMonths() int {
	badgeInfo := m.Tags["badge-info"]

	for _, badge := range strings.Split(badgeInfo, ",") {
		pair := strings.SplitN(badge, "/", 2)
		if pair[0] == "subscriber" {
			val, g := strconv.Atoi(pair[1])
			if g != nil {
				return 0
			}
			return val
		}
	}

	return 0
}

func connectToChat(rChan chan *ChatMessage, username string) {
	client := twitch.NewClient("justinfan1", "oauth:1")

	client.OnPrivateMessage(func(pm twitch.PrivateMessage) {
		message := &ChatMessage{&pm}
		rChan <- message
	})

	defer func() {
		close(rChan)
		err := client.Disconnect()
		if err != nil {
			panic(err)
		}
	}()

	client.Join(username)
	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

type EmoteStats = map[string]map[string]int
type MessageStats = map[string]int

type TimedStats struct {
	EmoteStats   EmoteStats    `json:"emote_stats"`
	MessageStats MessageStats  `json:"message_stats"`
	EmoteCount   int           `json:"emote_count"`
	MessageCount int           `json:"message_count"`
	Time         time.Duration `json:"time"`
}

type StreamInfo struct {
	StreamerName string        `json:"name"`
	CustomEmotes *CustomEmotes `json:"emotes"`
}

func NewStreamInfo(streamerName string) StreamInfo {
	emotes := LoadCustomEmotes(streamerName)
	return StreamInfo{
		StreamerName: streamerName,
		CustomEmotes: emotes,
	}
}

func (ce *CustomEmotes) FindIdentifierByName(name string, defaultTtv bool) (string, error) {
	for _, v := range ce.FfzEmotes {
		if v.Name == name {
			return v.Identifier(), nil
		}
	}

	for _, v := range ce.BttvEmotes {
		if v.Name == name {
			return v.Identifier(), nil
		}
	}

	if defaultTtv {
		return "ttv:" + name, nil
	}
	return "", errors.New("unable to find specified emote")
}

