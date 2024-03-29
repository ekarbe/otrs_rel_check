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

var (
	//Version variable for build
	Version string
	//Build variable for build
	Build string
)

type packagePackage struct {
	name        string
	version     string
	releaseDate string
}

type otrsPackage struct {
	version     string
	releaseDate string
}

var helpFlag = flag.Bool("help", false, "Print verbose help information.")
var verbose1Flag = flag.Bool("v", false, "Verbose output level 1.")
var verbose2Flag = flag.Bool("vv", false, "Verbose output level 2.")
var verbose3Flag = flag.Bool("vvv", false, "Verbose output level 3.")
var versionFlag = flag.Bool("V", false, "Print the version.")
var packageVersion = flag.Int64("p", 0, "A major version of OTRS.")
var packageFlag = flag.Bool("P", false, "Check the packages.")
var releaseTime = flag.Int("t", 31, "The time in days where a release happened.")

var stateFlag = 0

func init() {
	flag.Parse()
}

func main() {
	fmt.Print(run())
	os.Exit(stateFlag)
}

func run() string {
	if *helpFlag {
		flag.PrintDefaults()
		os.Exit(0)
	} else if *versionFlag {
		fmt.Printf("Version %s on build %s\n", Version, Build)
		os.Exit(0)
	} else if *packageFlag {
		fmt.Printf(checkPackage())
		return ""
	}
	return checkRelease()
}

func checkPackage() string {
	body, err := getBody("https://ftp.otrs.org/pub/otrs/packages/")
	if err != nil {
		stateFlag = 3
		return err.Error()
	}
	packages, err := getPackages(body)
	if err != nil {
		stateFlag = 3
		return err.Error()
	}
	currentPackages, err := getTimeWindowPackages(packages)
	if err != nil {
		stateFlag = 3
		return err.Error()
	}
	releaseCount := len(currentPackages)
	if releaseCount == 0 {
		stateFlag = 0
		return "No releases available"
	}
	stateFlag = 2
	output := strconv.Itoa(releaseCount) + " release(s) available\n"
	for _, packagePackage := range packages {
		output += packagePackage.name + "" + packagePackage.version + " released on " + packagePackage.releaseDate + "\n"
	}
	return output
}

func checkRelease() string {
	body, err := getBody("https://ftp.otrs.org/pub/otrs/")
	if err != nil {
		stateFlag = 3
		return err.Error()
	}
	releases, err := getReleases(body)
	if err != nil {
		stateFlag = 3
		return err.Error()
	}
	currentReleases, err := getTimeWindowReleases(releases)
	if err != nil {
		stateFlag = 3
		return err.Error()
	}
	releaseCount := len(currentReleases)
	if releaseCount == 0 {
		stateFlag = 0
		return "No releases available"
	}
	stateFlag = 2
	output := strconv.Itoa(releaseCount) + " release(s) available\n"
	for _, otrsPackage := range releases {
		output += otrsPackage.version + " released on " + otrsPackage.releaseDate + "\n"
	}
	return output
}

func getBody(url string) (string, error) {
	resp, err := http.Get(url)
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

func getPackages(body string) (map[string]packagePackage, error) {
	entries := strings.Split(string(body), "</tr>")
	name := regexp.MustCompile("(([A-Z])\\w+)-")
	version := regexp.MustCompile("([0-9]+\\.[0-9]+\\.[0-9]+)")
	date := regexp.MustCompile("([0-9])([0-9]*-[0-9]*-[0-9]*) [0-9]*:[0-9]*")
	packages := make(map[string]packagePackage)
	for i := range entries {
		p := packagePackage{}
		p.name = name.FindString(entries[i])
		p.version = version.FindString(entries[i])
		p.releaseDate = date.FindString(entries[i])
		if p.name != "" && p.version != "" && p.releaseDate != "" {
			temp := string(p.version[0])
			majorVersion, err := strconv.ParseInt(temp, 10, 64)
			if err != nil {
				return nil, err
			}
			if *packageVersion == 0 || majorVersion == *packageVersion {
				packages[p.version] = p
			}
		}
	}
	return packages, nil
}

func getReleases(body string) (map[string]otrsPackage, error) {
	entries := strings.Split(string(body), "</tr>")
	version := regexp.MustCompile("([0-9]+\\.[0-9]+\\.[0-9]+)")
	date := regexp.MustCompile("([0-9])([0-9]*-[0-9]*-[0-9]*) [0-9]*:[0-9]*")
	releases := make(map[string]otrsPackage)
	for i := range entries {
		o := otrsPackage{}
		o.version = version.FindString(entries[i])
		o.releaseDate = date.FindString(entries[i])
		if o.version != "" && o.releaseDate != "" {
			temp := string(o.version[0])
			majorVersion, err := strconv.ParseInt(temp, 10, 64)
			if err != nil {
				return nil, err
			}
			if *packageVersion == 0 || majorVersion == *packageVersion {
				releases[o.version] = o
			}
		}
	}
	return releases, nil
}

func getTimeWindowPackages(packages map[string]packagePackage) (map[string]packagePackage, error) {
	timeWindow := time.Now().AddDate(0, 0, -*releaseTime)
	for key, packagePackage := range packages {
		layout := "2006-01-02 15:04"
		parseTime, err := time.Parse(layout, packagePackage.releaseDate)
		if err != nil {
			return packages, err
		}
		if parseTime.Sub(timeWindow) < 0 {
			delete(packages, key)
		}
	}
	return packages, nil
}

func getTimeWindowReleases(releases map[string]otrsPackage) (map[string]otrsPackage, error) {
	timeWindow := time.Now().AddDate(0, 0, -*releaseTime)
	for key, otrsPackage := range releases {
		layout := "2006-01-02 15:04"
		parseTime, err := time.Parse(layout, otrsPackage.releaseDate)
		if err != nil {
			return releases, err
		}
		if parseTime.Sub(timeWindow) < 0 {
			delete(releases, key)
		}
	}
	return releases, nil
}
