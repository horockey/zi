package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	kb "github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/samber/lo"
)

//go:embed phrases.txt
var phrasesData []byte

//go:embed settings.json
var settingsData []byte

type Config struct {
	LogoutTimeoutSec int `json:"logout_timeout_sec"`
	EpsMs            int `json:"eps_ms"`
	Users            map[string]struct {
		DurMs int `json:"dur_ms"`
		DevMs int `json:"dev_ms"`
	} `json:"users"`
}

func main() {
	cfg := Config{}
	if err := json.Unmarshal(settingsData, &cfg); err != nil {
		panic(fmt.Errorf("unmarshaing settings data: %w", err))
	}

	phrases := lo.Map(
		bytes.Split(phrasesData, []byte("\n")),
		func(el []byte, _ int) string { return string(el) },
	)

	fmt.Println("Enter username:")
	uname := ""
	fmt.Scanln(&uname)
	ucfg, ok := cfg.Users[uname]
	if !ok {
		color.Red("No such user")
		return
	}

	phrase := phrases[rand.Intn(len(phrases))]
	color.Blue("%s\n", phrase)

	strokes := make([]time.Duration, 0, len(phrase)-1)
	var lastStrokeTsMs int64 = 0
	for i := 0; i < len(phrase); i++ {
		ch, key, err := kb.GetSingleKey()
		if err != nil {
			panic(fmt.Errorf("getting key from KB: %w", err))
		}
		if ch == 0 && key == kb.KeySpace {
			ch = ' '
		}
		fmt.Print(string(ch))
		if byte(ch) != phrase[i] {
			color.Red("Typing error")
			return
		}
		now := time.Now().UnixMilli()
		if lastStrokeTsMs > 0 {
			strokes = append(strokes, time.Duration(now-lastStrokeTsMs)*time.Millisecond)
		}
		lastStrokeTsMs = now
	}
	fmt.Println()

	meanPerSymbolMs := lo.Mean(strokes).Milliseconds()
	if math.Abs(float64(meanPerSymbolMs)-float64(ucfg.DurMs)) > float64(cfg.EpsMs) {
		color.Red("Bad mean stroke time: %dms", meanPerSymbolMs)
		return
	}

	deviation := lo.Reduce(
		strokes,
		func(eps time.Duration, stroke time.Duration, _ int) time.Duration {
			return eps + time.Duration(stroke.Milliseconds()-meanPerSymbolMs).Abs()*time.Millisecond
		},
		0,
	) / time.Duration(len(strokes))
	if math.Abs(float64(deviation.Milliseconds())-float64(ucfg.DevMs)) > float64(cfg.EpsMs) {
		color.Red("Bad deviation: %dms", deviation)
		return
	}

	color.Green("Logged in")

	strokeCh, err := kb.GetKeys(10_000)
	if err != nil {
		panic(fmt.Errorf("creating strokes channel: %w", err))
	}

	logoutDur := time.Duration(cfg.LogoutTimeoutSec) * time.Second
	ticker := time.NewTicker(logoutDur)
	for {
		select {
		case ev := <-strokeCh:
			ticker.Reset(logoutDur)
			fmt.Print(string(ev.Rune))
		case <-ticker.C:
			color.Red("Timed out")
			return
		}
	}
}
