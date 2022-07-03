package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
)

var (
	file        string
	threadnum   int
	formatMode  int
	wformatMode int
	port        int
)

func init() {
	flag.StringVar(&file, "f", "ip.txt", "Name of file to load")
	flag.IntVar(&threadnum, "tn", 1000, "thread number")
	flag.IntVar(&formatMode, "fm", 2, "format mode")
	flag.IntVar(&wformatMode, "wfm", 2, "write format mode")
	flag.IntVar(&port, "p", 25565, "check port")
	flag.Parse()

	versionfile, _ := os.Stat("version")
	if versionfile == nil {
		err := os.Mkdir("version", 0777)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	ip, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(fmt.Sprintf("%s is none.", file))
		return
	}
	iplist := strings.Split(string(ip), "\n")

	wg := &sync.WaitGroup{}
	ch := make(chan struct{}, threadnum)
	done := 0
	for _, ip := range iplist {
		ip := strings.ReplaceAll(ip, "\r", "")
		ch <- struct{}{}
		wg.Add(1)
		go func(ip string) {
			pinger := mcpinger.New(ip, uint16(port), mcpinger.McPingerOption(mcpinger.WithTimeout(3*time.Second)))
			info, err := pinger.Ping()
			if err == nil {
				var players []string
				for _, player := range info.Players.Sample {
					players = append(players, player.Name)
				}
				format1 := fmt.Sprintf("=================================================\nip: %s\nVERSION: %s\nONLINE: %d/%d\nPLAYERS: %s\nMOTD: %s", ip, info.Version.Name, info.Players.Online, info.Players.Max, players, info.Description.Text)
				format2 := fmt.Sprintf("%s | %s | %s | %d/%d | %s", ip, strings.ReplaceAll(info.Description.Text, "\n", " "), info.Version.Name, info.Players.Online, info.Players.Max, players)

				all, err := os.OpenFile("all.txt", os.O_APPEND|os.O_CREATE, 0664)
				if err != nil {
					fmt.Println(err)
				}

				version, err := os.OpenFile("version/"+info.Version.Name+".txt", os.O_APPEND|os.O_CREATE, 0664)
				if err != nil {
					fmt.Println(err)
				}

				if info.Players.Online > 0 {
					player, err := os.OpenFile("player.txt", os.O_APPEND|os.O_CREATE, 0664)
					if err != nil {
						fmt.Println(err)
					}
					switch wformatMode {
					case 1:
						player.WriteString(format1 + "\n")

					case 2:
						player.WriteString(format2 + "\n")
					}
				}

				switch wformatMode {
				case 1:
					all.WriteString(format1 + "\n")
					version.WriteString(format1 + "\n")

				case 2:
					all.WriteString(format2 + "\n")
					version.WriteString(format2 + "\n")
				}

				switch formatMode {
				case 1:
					fmt.Println(format1)
				case 2:
					fmt.Println(format2)
				}
			}
			<-ch
			done++
			fmt.Print(fmt.Sprintf("%d/%d \r", done, len(iplist)))
			wg.Done()
		}(ip)
	}
	wg.Wait()
}
