package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const pluginVersion = "Version 0.1\n"

type otrs struct {
	version string
	date    string
}

var h = flag.Bool("h", false, "prints the help")
var self = flag.Bool("V", false, "prints the version of the plugin")
var v = flag.Int64("v", 0, "defines the version to be checked. defaults to all versions.")
var t = flag.Int("t", 0, "defines the time of a new version")

func init() {
	flag.Parse()
}

func main() {
	if *h {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *self {
		print(pluginVersion)
		os.Exit(0)
	}
	body, err := GetBody()
	if err != nil {
		print(err)
	}
	releases, err := GetReleases(body)
	if err != nil {
		print(err)
	}
	if *t != 0 {
		releases, err = GetTimeWindowReleases(releases)
		if err != nil {
			print(err)
		}
	}
	fmt.Print(releases)
}

//GetBody ...
func GetBody() (string, error) {
	resp, err := http.Get("https://ftp.otrs.org/pub/otrs/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

//GetReleases ...
func GetReleases(body string) (map[string]otrs, error) {
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
			majorVersion, err := strconv.ParseInt(temp, 10, 64)
			if err != nil {
				return nil, err
			}
			if *v == 0 || majorVersion == *v {
				releases[o.version] = o
			}
		}
	}
	return releases, nil
}

//GetTimeWindowReleases ...
func GetTimeWindowReleases(releases map[string]otrs) (map[string]otrs, error) {
	timeWindow := time.Now().AddDate(0, 0, -*t)
	for key, otrs := range releases {
		layout := "2006-01-02 15:04"
		parseTime, err := time.Parse(layout, otrs.date)
		if err != nil {
			return releases, err
		}
		if parseTime.Sub(timeWindow) < 0 {
			delete(releases, key)
		}
	}
	return releases, nil
}
