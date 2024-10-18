package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/fatih/color"
)

//go:embed creds.json
var credsData []byte

const (
	usersCount   = 3 // + 1 (admin)
	objectsCount = 4
)

func printPrivsMatrix(privs map[string][]Priv) {
	fmt.Printf("%10s", "")
	for j := 0; j < objectsCount; j++ {
		fmt.Printf("%8d", j+1)
	}
	fmt.Println()

	for i := 0; i < usersCount+1; i++ {
		uname := fmt.Sprintf("user_%d", i+1)
		if i == usersCount {
			uname = "admin"
		}
		fmt.Printf("%10s", uname)
		for j := 0; j < objectsCount; j++ {
			p := privs[uname][j]
			fmt.Printf("%5s (%d)", p, p)
		}
		fmt.Println()
	}
}

func main() {
	creds := map[string]string{}
	if err := json.Unmarshal(credsData, &creds); err != nil {
		panic(fmt.Errorf("unmarshaling creds: %w", err))
	}

	privs := map[string][]Priv{}
	for i := 0; i < usersCount; i++ {
		uname := fmt.Sprintf("user_%d", i+1)
		privs[uname] = make([]Priv, 0, objectsCount)
		for j := 0; j < objectsCount; j++ {
			privs[uname] = append(privs[uname], Priv(rand.Intn(8)))
		}
	}

	privs["admin"] = []Priv{}
	for j := 0; j < objectsCount; j++ {
		privs["admin"] = append(privs["admin"], Priv(7))
	}

	for {
		printPrivsMatrix(privs)
		fmt.Println("Enter creds in format \"{username} {password}\"")
		var user, pwd string
		fmt.Scanf("%s %s\n", &user, &pwd)

		if rightPwd, ok := creds[user]; !ok || rightPwd != pwd {
			color.Red("Bad credentials!")
			continue
		}
		for {
			fmt.Println("Enter action (R|W|G) or empty string to exit")

			var action string
			fmt.Scanln(&action)

			if action == "" {
				break
			}

			fmt.Println("Enter object number:")
			var objNumber int
			_, err := fmt.Scan(&objNumber)
			if err != nil || objNumber <= 0 || objNumber > objectsCount {
				color.Red("Bad object number: %s\n", err)
				continue
			}
			objNumber--

			switch action {
			default:
				color.Red("Unknown action")
				continue
			case "R":
				if !privs[user][objNumber].canRead() {
					color.Red("Access denied")
					continue
				}
				color.Green("Access granted")
				data, err := os.ReadFile(fmt.Sprintf("%d.txt", objNumber))
				if err != nil {
					data = []byte{}
				}
				fmt.Println(string(data))
			case "W":
				if !privs[user][objNumber].canWrite() {
					color.Red("Access denied")
					continue
				}
				color.Green("Access granted")
				fmt.Print("Enter data: ")
				var data string
				fmt.Scan(&data)
				os.WriteFile(fmt.Sprintf("%d.txt", objNumber), []byte(data), os.ModePerm)
			case "G":
				if !privs[user][objNumber].canGrant() {
					color.Red("Access denied")
					continue
				}
				var priv, dstUser string
				fmt.Println("Enter priviliege (R|W|G)")
				fmt.Scanln(&priv)
				fmt.Println("Enter username to transfer priv to")
				fmt.Scanln(&dstUser)
				fmt.Printf("|%s/%s|\n", priv, dstUser)

				if _, found := privs[dstUser]; !found {
					color.Red("Unknown dst user")
					continue
				}

				switch priv {
				default:
					color.Red("Unknown priv")
				case "R":
					if !privs[user][objNumber].canRead() {
						color.Red("Access denied")
						continue
					}
					privs[dstUser][objNumber] = privs[dstUser][objNumber] | privRead
					color.Green("Priv granted")
				case "W":
					if !privs[user][objNumber].canWrite() {
						color.Red("Access denied")
						continue
					}
					privs[dstUser][objNumber] = privs[dstUser][objNumber] | privWrite
					color.Green("Priv granted")
				case "G":
					privs[dstUser][objNumber] = privs[dstUser][objNumber] | privGrant
					color.Green("Priv granted")
				}
			}
		}
	}
}
