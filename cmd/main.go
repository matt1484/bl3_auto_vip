package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	
	"github.com/shibukawa/configdir"
	bl3 "github.com/matt1484/bl3_auto_vip"
)

// gross but effective for now
const version = "2.1"

var usernameHash string

func printError(err error) {
	fmt.Println("failed!")
	fmt.Print("Had error: ")
	fmt.Println(err)
}

func exit() {
	fmt.Print("Exiting in ")
	for i := 5; i > 0; i-- {
		fmt.Print(strconv.Itoa(i) + " ")
		time.Sleep(time.Second)
	}
	fmt.Println("")
}

func doVip(client *bl3.Bl3Client) {
	fmt.Print("Getting previously redeemed VIP codes . . . . . ")
	redeemedCodes, err := client.GetRedeemedVipCodeMap()
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	fmt.Print("Getting new VIP codes . . . . . ")
	allCodes, err := client.GetFullVipCodeMap()
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
		fmt.Println("No new VIP codes at this time. Try again later.")
	} else {
		for codeType, codes := range newCodes {
			if len(codes) < 1 {
				continue
			}

			fmt.Print("Setting up VIP codes of type '" + codeType + "' . . . . . ")
			_, found := client.Config.Vip.CodeTypeUrlMap[codeType]
			if !found {
				fmt.Println("invalid! Moving on.")
				continue
			}
			fmt.Println("success!")

			for code := range codes {
				fmt.Print("Trying '" + codeType + "' VIP code '" + code + "' . . . . . ")
				res, valid := client.RedeemVipCode(codeType, code)
				if !valid {
					fmt.Println("failed! Moving on.")
					continue
				}
				fmt.Println(res)
			}
		}
	}
}

func doShift(client *bl3.Bl3Client) {
	fmt.Print("Getting SHIFT platforms . . . . . ")
	platforms, err := client.GetShiftPlatforms()
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	configDirs := configdir.New("bl3-auto-vip", "bl3-auto-vip")
	configFilename := usernameHash + "-shift-codes.json"

	fmt.Print("Getting previously redeemed SHIFT codes . . . . . ")
	redeemedCodes := bl3.ShiftCodeMap{}
	folder := configDirs.QueryFolderContainsFile(configFilename)
	if folder != nil {
		data, err := folder.ReadFile(configFilename)
		if err == nil {
			json := bl3.JsonFromBytes(data)
			if json != nil {
				json.Out(&redeemedCodes)
				fmt.Println("success!")
			} else {
				fmt.Println("not found.")
			}
		} else {
			fmt.Println("not found.")
		}
	} else {
		fmt.Println("not found.")
	}

	fmt.Print("Getting new SHIFT codes . . . . . ")
	shiftCodes, err := client.GetFullShiftCodeList()
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	foundCodes := false
	for code, codePlatforms := range shiftCodes {
		for _, platform := range codePlatforms {
			if _, found := platforms[platform]; found {
				if !redeemedCodes.Contains(code, platform) {
					foundCodes = true
					fmt.Print("Trying '" + platform + "' SHIFT code '" + code + "' . . . . . ")
					err := client.RedeemShiftCode(code, platform)
					if err != nil {
						fmt.Println(err)
						if strings.Contains(strings.ToLower(err.Error()), "already") {
							redeemedCodes[code] = append(redeemedCodes[code], platform)
						}
					} else {
						redeemedCodes[code] = append(redeemedCodes[code], platform)
						fmt.Println("success!")
					}
				}
			}
		}
	}

	fmt.Println(redeemedCodes)
	if !foundCodes {
		fmt.Println("No new SHIFT codes at this time. Try again later.")
	} else {
		folders := configDirs.QueryFolders(configdir.System)
		data, err := json.Marshal(&redeemedCodes)
		if err == nil {
			folders[0].WriteFile(configFilename, data)
		}
	}
	
}

func main() {
	username := ""
	password := ""
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
		password = string(bytes)
	}

	hasher := md5.New()
    hasher.Write([]byte(username))
	usernameHash = hex.EncodeToString(hasher.Sum(nil))

	fmt.Print("Setting up . . . . . ")
	client, err := bl3.NewBl3Client()
	if err != nil {
		printError(err)
		return
	}

	fmt.Println("success!")

	if client.Config.Version != version {
		fmt.Println("Your version (" + version + ") is out of date. Please consider downloading the latest version (" + client.Config.Version + ") at https://github.com/matt1484/bl3_auto_vip/releases/latest")
	}

	fmt.Print("Logging in as '" + username + "' . . . . . ")
	err = client.Login(username, password)
	if err != nil {
		printError(err)
		return
	}
	fmt.Println("success!")

	doShift(client)
	doVip(client)
	exit()
}
