package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/samber/lo"

	_ "embed"

	"github.com/tadvi/winc"
)

//go:embed creds.json
var credsData []byte

//go:embed locks.json
var locksData []byte

const (
	xivl = 10
	yivl = 10

	canvasSize = 5
	pbSize     = 50
)

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var (
	lastBtnRow   = -1
	lastBtnCol   = -1
	failedInputs = 0

	path = Path{}
)

func main() {
	patterns := map[string]Path{}
	if err := json.Unmarshal(credsData, &patterns); err != nil {
		panic(fmt.Errorf("loading creds: %w", err))
	}

	locks := []Lock{}
	if err := json.Unmarshal(locksData, &locks); err != nil {
		panic(fmt.Errorf("loading creds: %w", err))
	}

	mainForm := winc.NewForm(nil)
	mainForm.SetSize(800, 600)
	mainForm.SetText("Lab 7")

	loginLabel := winc.NewLabel(mainForm)
	loginLabel.SetText("Enter login")
	loginLabel.SetPos(
		xivl,
		yivl,
	)

	loginTextBox := winc.NewEdit(mainForm)
	loginTextBox.SetSize(200, loginTextBox.Height())
	loginTextBox.SetPos(
		xivl,
		loginLabel.ClientHeight()+yivl,
	)

	y0 := loginLabel.ClientHeight() + loginTextBox.ClientHeight() + 2*yivl
	cvs := make([][]*winc.PushButton, 0, canvasSize)
	for i := 0; i < canvasSize; i++ {
		row := make([]*winc.PushButton, canvasSize)
		cvs = append(cvs, row)
		for j := 0; j < len(row); j++ {
			btn := winc.NewPushButton(mainForm)
			btn.SetText(fmt.Sprintf("%d/%d", i, j))
			btn.SetSize(pbSize, pbSize)
			btn.SetPos((pbSize+xivl)*j, y0+(pbSize+yivl)*i)
			btn.OnClick().Bind(func(arg *winc.Event) {
				if lastBtnRow == -1 || lastBtnCol == -1 {
					lastBtnRow = i
					lastBtnCol = j
					path = append(path, Coord{i, j})
					cvs[i][j].SetText("***")
					return
				}
				switch {
				case lastBtnRow != i && lastBtnCol != j:
					return
				case lastBtnRow != i:
					r1, r2 := min(lastBtnRow, i), max(lastBtnRow, i)
					for r := r1; r <= r2; r++ {
						cvs[r][j].SetText("*")
						if r == r1 {
							continue
						}
						path = append(path, Coord{r, j})
					}
					cvs[i][j].SetText("***")
					lastBtnRow = i
					lastBtnCol = j
				case lastBtnCol != j:
					c1, c2 := min(lastBtnCol, j), max(lastBtnCol, j)
					for c := c1; c <= c2; c++ {
						cvs[i][c].SetText("*")
						if c == c1 {
							continue
						}
						path = append(path, Coord{i, c})
					}
					cvs[i][j].SetText("***")
					lastBtnRow = i
					lastBtnCol = j
				}
			})
			row[j] = btn
		}
	}

	loginButton := winc.NewPushButton(mainForm)
	loginButton.SetText("LOGIN")
	loginButton.SetSize(200, 50)
	loginButton.SetPos(xivl, mainForm.ClientHeight()-loginButton.ClientHeight()-yivl)
	loginButton.OnClick().Bind(func(arg *winc.Event) {
		defer func() {
			path = Path{}
			lastBtnRow = -1
			lastBtnCol = -1
			for i := 0; i < len(cvs); i++ {
				for j := 0; j < len(cvs[i]); j++ {
					cvs[i][j].SetText(fmt.Sprintf("%d/%d", i, j))
				}
			}
		}()

		pstr := strings.Join(lo.Map(path, func(el Coord, _ int) string { return fmt.Sprintf("{%d, %d}", el.X, el.Y) }), ", ")
		fmt.Println(pstr)

		uname := loginTextBox.Text()
		p, found := patterns[uname]
		if !found {
			color.Red("No such user")
			return
		}
		activeLocks := lo.Filter(locks, func(el Lock, _ int) bool {
			if el.User != uname || el.UntilUnix < time.Now().Unix() {
				return false
			}
			return true
		})
		if len(activeLocks) > 0 {
			color.Red("User %s is blocked until %s", uname, time.Unix(activeLocks[0].UntilUnix, 0).Format(time.DateTime))
			return
		}
		if !p.Matches(path) {
			color.Red("Bad pattern")
			failedInputs++
			if failedInputs >= 3 {
				locks = append(locks, Lock{User: uname, UntilUnix: time.Now().Unix() + 15})
				color.Red("User %s blocked for 15s", uname)
			}
			return
		}
		color.Green("Access granted")
		failedInputs = 0
	})

	registerButton := winc.NewPushButton(mainForm)
	registerButton.SetText("REGISTER")
	registerButton.SetSize(200, 50)
	registerButton.SetPos(xivl*2+loginButton.ClientWidth(), mainForm.ClientHeight()-loginButton.ClientHeight()-yivl)
	registerButton.OnClick().Bind(func(arg *winc.Event) {
		defer func() {
			path = Path{}
			lastBtnRow = -1
			lastBtnCol = -1
			for i := 0; i < len(cvs); i++ {
				for j := 0; j < len(cvs[i]); j++ {
					cvs[i][j].SetText(fmt.Sprintf("%d/%d", i, j))
				}
			}
		}()

		uname := loginTextBox.Text()
		if uname == "" {
			return
		}

		patterns[uname] = path
		data, err := json.MarshalIndent(patterns, "", "\t")
		if err != nil {
			panic(fmt.Errorf("marshaling patterns: %w", err))
		}
		if err := os.WriteFile("creds.json", data, os.ModePerm); err != nil {
			panic(fmt.Errorf("writing creds to file: %w", err))
		}

		pstr := strings.Join(lo.Map(path, func(el Coord, _ int) string { return fmt.Sprintf("{%d, %d}", el.X, el.Y) }), ", ")
		fmt.Printf("User %s regitered with pattern %s\n", uname, pstr)
	})

	mainForm.OnClose().Bind(func(arg *winc.Event) {
		data, err := json.MarshalIndent(locks, "", "\t")
		if err != nil {
			panic(fmt.Errorf("marshaling locks: %w", err))
		}
		if err := os.WriteFile("locks.json", data, os.ModePerm); err != nil {
			panic(fmt.Errorf("writing file: %w", err))
		}
		winc.Exit()
	})
	mainForm.Center()
	mainForm.Show()
	winc.RunMainLoop()
}
