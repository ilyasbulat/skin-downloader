package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type SkinData struct {
	ID         int    `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Path       string `json:"path,omitempty"`
	OnCreate   string `json:"on_create,omitempty"`
	OnUpdate   string `json:"oc_update,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	Angle      string `json:"angle,omitempty"`
	Volume     string `json:"volume,omitempty"`
	MD5        string `json:"md_5,omitempty"`
}

func main() {
	url := os.Getenv("URL") + "/skin?macaddress=" + getMac("wlan0")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var data SkinData
	json.Unmarshal(body, &data)
	splitPath := strings.Split(data.Path, "/")
	filename := splitPath[len(splitPath)-1]
	downloadURL := os.Getenv("PUBLIC") + data.Path
	localMD5 := getMD5(filename)

	if localMD5 == "none" {
		//download and run on create cmd`s
		downloadAndRun(filename, downloadURL, data.OnCreate)
	} else if localMD5 != data.MD5 {
		//download and run on update cmd`s
		downloadAndRun(filename, downloadURL, data.OnUpdate)
	}

	// check vars file if not exists create and put vars in there
	// if exists check vars if they not equals rewrite them
	// ignore otherwise
	newVars := fmt.Sprintf("RES=%s\nANGLE=%s\nVOL=%s", data.Resolution, data.Angle, data.Volume)
	if checkVars(newVars) {
		writeToFile(newVars)
	}
}

func checkVars(newVars string) bool {

	content, err := ioutil.ReadFile("vars")
	if err != nil {
		return true
	}
	oldVars := string(content)

	return oldVars != newVars

}

func writeToFile(newVars string) {
	vars, err := os.Create("vars")
	if err != nil {
		log.Println("trying to write vars: ", err.Error())
		os.Exit(1)
	}
	defer vars.Close()
	vars.WriteString(newVars)
}

func downloadAndRun(filename, url, rawCommands string) {
	response, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("status code : %d \n", response.StatusCode)
		os.Exit(1)
	}
	//Create an empty file
	file, err := os.Create(filename)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	defer file.Close()
	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	commands := strings.Split(rawCommands, "\n")
	for _, command := range commands {
		args := strings.Split(command, " ")
		name := args[0]
		log.Println("exec ", name, args[1:])
		output, err := exec.Command(name, args[1:]...).Output()
		log.Println(string(output))
		if err != nil {
			log.Println(err.Error())
		}
	}

}

func getMD5(filePath string) string {
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return "none"
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return "none"
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String

}

func getMac(itf string) string {

	if itf == "" {
		itf = "eth0"
	}
	fileName := fmt.Sprintf("/sys/class/net/%s/address", itf)

	var line string
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "None"
	}
	line = string(file)[0:17]

	return line
}
