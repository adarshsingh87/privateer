package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {

	resp, err := http.Get("http://localhost:9117/api/v2.0/indexers/test:passed/results?apikey=m2td3d4z7xykqyb3end2gvvvydsyi1b4&Query=deamon+slayer+s03e05&Tracker%5B%5D=1337x&Tracker%5B%5D=bitsearch&Tracker%5B%5D=eztv&Tracker%5B%5D=gamestorrents&Tracker%5B%5D=internetarchive&Tracker%5B%5D=itorrent&Tracker%5B%5D=kickasstorrents-to&Tracker%5B%5D=kickasstorrents-ws&Tracker%5B%5D=limetorrents&Tracker%5B%5D=moviesdvdr&Tracker%5B%5D=nyaasi&Tracker%5B%5D=pctorrent&Tracker%5B%5D=rutor&Tracker%5B%5D=rutracker-ru&Tracker%5B%5D=solidtorrents&Tracker%5B%5D=torrentfunk&Tracker%5B%5D=yts")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result JackettResponseReq
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	fmt.Println(PrettyPrint(result))

}

type JackettResponseReq struct {
	Results []struct {
		Tracker      string    `json:"Tracker"`
		CategoryDesc string    `json:"CategoryDesc"`
		Title        string    `json:"Title"`
		Link         string    `json:"Link"`
		PublishDate  time.Time `json:"PublishDate"`
		Size         int       `json:"Size"`
		Seeders      int       `json:"Seeders"`
		Peers        int       `json:"Peers"`
		MagnetURI    string    `json:"MagnetUri"`
	} `json:"Results"`
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
