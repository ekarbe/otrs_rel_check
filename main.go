package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type otrs struct {
	version string
	date    string
}

var help = flag.Bool("help", false, "prints the help")
var major = flag.Int64("major", 0, "defines the version to be checked. defaults to all versions.")
var t = flag.Int("t", 0, "defines the time of a new version")

func init() {
	flag.Parse()
}

func main() {
	if *help {
		PrintHelp()
	}
	resp, _ := http.Get("https://ftp.otrs.org/pub/otrs/")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	entries := strings.Split(string(body), "</tr>")
	version := regexp.MustCompile("([0-9]+\\.[0-9]+\\.[0-9]+)")
	date := regexp.MustCompile("([0-9])([0-9]*-[0-9]*-[0-9]*) [0-9]*:[0-9]*")
	releases := make(map[string]otrs)
	for i := range entries {
		o := otrs{}
		o.version = version.FindString(entries[i])
		o.date = date.FindString(entries[i])
		if o.version != "" && o.date != "" {
			temp := string(o.version[0])
			majorVersion, _ := strconv.ParseInt(temp, 10, 64)
			if *major == 0 || majorVersion == *major {
				releases[o.version] = o
			}
		}
	}
	if *t != 0 {
		timeWindow := time.Now().AddDate(0, 0, -*t)
		for key, otrs := range releases {
			layout := "2006-01-02 15:04"
			parseTime, _ := time.Parse(layout, otrs.date)
			if parseTime.Sub(timeWindow) < 0 {
				delete(releases, key)
			}
		}
	}
	fmt.Println(releases)
}

//PrintHelp ...
func PrintHelp() {
	print("main help\n[Option] [Description]\n-version | defines the version to be checked\n")
}
