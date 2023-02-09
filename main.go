package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type skinData struct {
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
	data, err := request(url)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	_, filename := filepath.Split(data.Path)
	downloadURL := os.Getenv("PUBLIC") + data.Path
	if err := download(filename, downloadURL); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	localMD5 := getMD5(filename)
	if localMD5 == "none" {
		//download and run on create cmd`s
		execCommands(data.OnCreate)
	} else if localMD5 != data.MD5 {
		//download and run on update cmd`s
		execCommands(data.OnUpdate)
	}

	vol := getVolume(data.Volume, data.Type)

	newVars := fmt.Sprintf("RES=%s\nANGLE=%s\nVOL=%s", data.Resolution, data.Angle, vol)
	if checkVars(newVars) {
		writeToFile(newVars)
	}
}

func request(url string) (skinData, error) {
	var data skinData
	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return data, nil
}

func download(filename, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("status code : %d", response.StatusCode)
	}
	//Create an empty file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func execCommands(str string) {
	commands := strings.Split(str, "\n")
	for _, command := range commands {
		args := strings.Split(command, " ")
		name := args[0]
		log.Println("exec ", name, args[1:])
		output, err := exec.Command(name, args[1:]...).Output()
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(string(output))
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
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "None"
	}
	line = string(file)[0:17]

	return line
}

// check vars file if not exists create and put vars in there
// if exists check vars if they not equals rewrite them
// ignore otherwise
func getVolume(vol, skin string) string {
	if skin != "h" && skin != "v" {
		return vol
	}

	layout := "15"
	t := time.Now().Format(layout)

	check, _ := time.Parse(layout, t)
	start, _ := time.Parse(layout, "22:00")
	end, _ := time.Parse(layout, "07:00")

	if inTimeSpan(start, end, check) {
		vol = "-6000"
	}
	return vol
}

func checkVars(newVars string) bool {
	content, err := os.ReadFile("vars")
	if err != nil {
		return true
	}
	oldVars := string(content)

	return oldVars != newVars
}

func writeToFile(newVars string) error {
	vars, err := os.Create("vars")
	if err != nil {
		return err
	}
	defer vars.Close()
	_, err = vars.WriteString(newVars)
	return err
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
