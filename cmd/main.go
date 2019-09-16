package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
	
	bl3 "github.com/matt1484/bl3_auto_vip"
)

func printError(err error) {
	fmt.Println("failed!")
	fmt.Print("Had error: ")
	fmt.Println(err)
	exit()
}

func exit() {
	fmt.Print("Exiting in ")
	for i := 5; i > 0; i-- {
		fmt.Print(strconv.Itoa(i) + " ")
		time.Sleep(time.Second)
	}
	fmt.Println("")
}

func main() {
	var username string
	var password string
	flag.StringVar(&username, "e", "", "Email")
	flag.StringVar(&username, "email", "", "Email")
	flag.StringVar(&password, "p", "", "Password")
	flag.StringVar(&password, "password", "", "Password")
	flag.Parse()

	if username == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter username (email): ")
		bytes, _, _ := reader.ReadLine()
		username = string(bytes)
	}
	if password == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter password        : ")
		bytes, _, _ := reader.ReadLine()
		// bytes, _ := terminal.ReadPassword(0)
		// fmt.Println("")
		password = string(bytes)
	}

	fmt.Print("Setting up . . . . . ")
	client, err := bl3.NewBl3Client()
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	fmt.Print("Logging in as '" + username + "' . . . . . ")
	err = client.Login(username, password)
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	fmt.Print("Getting previously redeemed codes . . . . . ")
	redeemedCodes, err := client.GetRedeemedVipCodeMap()
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	fmt.Print("Getting new codes . . . . . ")
	allCodes, err := bl3.GetFullVipCodeMap()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("success!")

	newCodes := allCodes.Diff(redeemedCodes)
	codeCount := 0
	for _, codes := range newCodes {
		codeCount += len(codes)
	}
	if codeCount == 0 {
		fmt.Println("No new codes at this time. Try again later.")
		exit()
		return
	}

	fmt.Print("Getting code redemption URLs . . . . . ")
	codeUrlMap := bl3.GetVipCodeUrlMap()
	fmt.Println("success!")

	for codeType, codes := range newCodes {
		if len(codes) < 1 {
			continue
		}

		fmt.Print("Setting up codes of type '" + codeType + "' . . . . . ")
		codeUrl, found := codeUrlMap[codeType]
		if !found {
			fmt.Println("invalid! Moving on.")
			continue
		}
		fmt.Println("success!")

		for code := range codes {
			fmt.Print("Trying '" + codeType + "' code '" + code + "' . . . . . ")

			res, err := client.PostJson(codeUrl, map[string]string{
				"code": code,
			})
			if err != nil {
				fmt.Println("failed! Moving on.")
				continue
			}

			resJson, err := res.BodyAsJson()
			if err != nil {
				fmt.Println("failed! Moving on.")
				continue
			}

			exception := resJson.Find("exception.model")
			success := resJson.Reset().Find("message")
			if exception != nil {
				fmt.Println(exception)
			}
			if success != nil {
				fmt.Println(success)
			}
		}
	}
	exit()
}
