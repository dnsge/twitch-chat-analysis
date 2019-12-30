package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func collectStatsForTime(stream StreamInfo, t time.Duration, printEmotes bool, unique bool) TimedStats {
	messageChan := make(chan *ChatMessage)

	emoteCountStats := make(EmoteStats)
	messageCountStats := make(MessageStats)

	go connectToChat(messageChan, stream.StreamerName)

	messageCount := 0
	emoteCount := 0
	for start := time.Now(); time.Since(start) < t; {
		msg := <-messageChan
		ids := msg.ExtractEmoteIdentifiers(stream.CustomEmotes, unique)

		messageCount++

		// No Emotes?
		if len(ids) == 0 {
			continue
		}

		if printEmotes {
			fmt.Println(ids)
		}

		emoteCount += len(ids)

		binName := strconv.Itoa(msg.subMonths())
		messageCountStats[binName] += 1
		for _, emote := range ids {
			bin := emoteCountStats[binName]
			if bin == nil {
				bin = make(map[string]int)
				emoteCountStats[binName] = bin
			}
			bin[emote] += 1
		}
	}

	return TimedStats{
		EmoteStats:   emoteCountStats,
		MessageStats: messageCountStats,
		EmoteCount:   emoteCount,
		MessageCount: messageCount,
		Time:         t / time.Second,
	}
}

func trimUntilChar(str string, char rune) string {
	idx := strings.IndexRune(str, char)
	if idx == -1 {
		return str
	}
	return str[idx+1:]

}

func analyzeSpecificEmote(stats TimedStats, emoteIdentifier string) {
	keys := make([]int, 0, len(stats.MessageStats))
	for k := range stats.MessageStats {
		v, e := strconv.Atoi(k)
		if e != nil {
			panic(e)

		}
		keys = append(keys, v)
	}
	sort.Ints(keys)

	fmt.Printf("Months Subscribed vs emote \"%s\" usage:\n", trimUntilChar(emoteIdentifier, ':'))
	for iBin := range keys {
		bin := strconv.Itoa(iBin)
		emoteStat, anyOfEmote := stats.EmoteStats[bin][emoteIdentifier]

		messageStat := stats.MessageStats[bin]

		var rate float64
		if !anyOfEmote {
			rate = 0
		} else {
			if messageStat == 0 {
				rate = 0
			} else {
				rate = float64(emoteStat) / float64(messageStat) * 1000
			}
		}
		fmt.Printf("%2d months subscribed: %6.3f per 1000 messages (%d)\n", iBin, rate, messageStat)
	}
}

func interactiveEmoteStats(stream StreamInfo, stats TimedStats) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nCollected %d messages with %d total emotes counted\n", stats.MessageCount, stats.EmoteCount)

	for {
		fmt.Print("\nAnalyze Emote Usage: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			break
		}

		id, err := stream.CustomEmotes.FindIdentifierByName(text, true)
		if err != nil {
			panic(err)
		}
		fmt.Println("Found emote id " + id)

		analyzeSpecificEmote(stats, id)
	}
}

func saveStatistics(stats TimedStats, filename string) {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, data, 0o644)
	if err != nil {
		panic(err)
	}
}

func loadStatistics(filename string) (TimedStats, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return TimedStats{}, err
	}

	s := TimedStats{}
	err = json.Unmarshal(data, &s)

	if err != nil {
		return TimedStats{}, err
	}
	return s, nil
}

type ExecMode struct {
	Cmd     string
	Input   string
	Output  string
	Channel string
	Seconds int64
	Print   bool
	Unique  bool
}

func Execute() {

	channel := flag.String("c", "", "Twitch.tv channel name")
	input := flag.String("i", "", "Input file")
	output := flag.String("o", "", "Output file")
	printEmotes := flag.Bool("print", false, "Whether to print chat emotes")
	unique := flag.Bool("unique", false, "Only count emote once per message")
	t := flag.Int64("t", 0, "Time to run")

	flag.Parse()

	exec := &ExecMode{
		Cmd:     flag.Arg(0),
		Input:   *input,
		Output:  *output,
		Channel: *channel,
		Seconds: *t,
		Print:   *printEmotes,
		Unique:  *unique,
	}

	switch exec.Cmd {
	case "collect":
		fail := false
		if exec.Channel == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Channel (c) argument is required for collection")
			fail = true
		}
		if exec.Output == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Output (o) argument is required for collection")
			fail = true
		}
		if exec.Seconds < 1 {
			_, _ = fmt.Fprintln(os.Stderr, "Time (t) argument is required for collection & must be greater than 0")
			fail = true
		}

		if fail {
			os.Exit(1)
		}

		fmt.Println("Loading Channel Info...")
		info, err := NewStreamInfo(exec.Channel)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Failed to load channel info: "+err.Error())
			os.Exit(3)
		}

		fmt.Println("Collecting Data...")
		stats := collectStatsForTime(*info, time.Second*time.Duration(exec.Seconds), exec.Print, exec.Unique)
		saveStatistics(stats, exec.Output)

	case "analyze":
		fail := false
		if exec.Channel == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Channel (c) argument is required for analyze")
			fail = true
		}
		if exec.Input == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Input (i) argument is required for analyze")
			fail = true
		}

		if fail {
			os.Exit(1)
		}

		fmt.Println("Loading Channel Info...")
		info, err := NewStreamInfo(exec.Channel)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Failed to load channel info: "+err.Error())
			os.Exit(3)
		}

		fmt.Println("Loading File Data...")
		stats, err := loadStatistics(exec.Input)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Failed to read input file: "+err.Error())
			os.Exit(2)
		}
		interactiveEmoteStats(*info, stats)
	case "interactive":
		reader := bufio.NewReader(os.Stdin)

		if exec.Channel == "" {
			fmt.Print("Enter chat room name: ")
			cn, _ := reader.ReadString('\n')
			exec.Channel = strings.TrimSuffix(cn, "\n")
		}

		fmt.Println("Loading Channel Info...")
		info, err := NewStreamInfo(exec.Channel)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Failed to load channel info: "+err.Error())
			os.Exit(3)
		}

		if exec.Seconds == 0 {
			fmt.Print("\nSeconds to run collection: ")
			tm, _ := reader.ReadString('\n')
			tm = strings.TrimSuffix(tm, "\n")
			parsed, err := strconv.ParseInt(tm, 10, 64)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Invalid time: Not an integer")
				os.Exit(1)
			}
			exec.Seconds = parsed
		}

		fmt.Println("Collecting Data...")
		stats := collectStatsForTime(*info, time.Second*time.Duration(exec.Seconds), exec.Print, exec.Unique)
		interactiveEmoteStats(*info, stats)
	default:
		fmt.Println("Usage: collect, analyze, or interactive")
	}
}
