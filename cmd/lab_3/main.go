package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/fatih/color"
)

//go:embed creds.json
var credsData []byte

const (
	usersCount   = 3
	objectsCount = 4
)

func main() {
	creds := map[string]string{}
	if err := json.Unmarshal(credsData, &creds); err != nil {
		panic(fmt.Errorf("unmarshaling creds: %w", err))
	}

	objects := make([]AccessLevel, 0, objectsCount)
	for i := 0; i < objectsCount; i++ {
		objects = append(objects, AccessLevel(rand.Intn(3)+1))
	}

	users := map[string]AccessLevel{}
	for i := 0; i < usersCount; i++ {
		uname := fmt.Sprintf("user_%d", i+1)
		users[uname] = AccessLevel(rand.Intn(3) + 1)
	}

	for {
		printAccessLevels(users, objects)

		fmt.Println("Enter creds in format \"{username} {password}\"")
		var user, pwd string
		fmt.Scanf("%s %s\n", &user, &pwd)

		if rightPwd, ok := creds[user]; !ok || rightPwd != pwd {
			color.Red("Bad credentials!")
			continue
		}
		for {
			fmt.Println("Enter object number to access to, or 0 to exit")
			var obj int
			fmt.Scanln(&obj)

			if obj < 0 || obj > objectsCount {
				color.Red("Bad object number")
				continue
			}
			if obj == 0 {
				break
			}

			if objects[obj-1] > users[user] {
				color.Red("Access denied")
				continue
			}

			color.Green("Access granted")
		}
	}
}

func printAccessLevels(users map[string]AccessLevel, objects []AccessLevel) {
	fmt.Printf("%10s", "Users:")
	for uname, al := range users {
		fmt.Printf("%10s %-15s", uname, "("+al.String()+")")
	}
	fmt.Println()
	fmt.Printf("%10s", "Objects:")
	for idx, al := range objects {
		fmt.Printf("%10d %-15s", idx+1, "("+al.String()+")")
	}
	fmt.Println()
}
