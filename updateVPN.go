package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const dnsURL = "https://github.com/felixonmars/dnsmasq-china-list/raw/master/accelerated-domains.china.conf"
const ipURL = "https://raw.githubusercontent.com/mayaxcn/china-ip-list/master/chnroute.txt"
const fileName = "accelerated-domains.china.conf"
const ipBypassName = "shadowsocks-ignore.list"
const customIPBypassName = "custom-ip-bypass"
const customDomainName = "accelerated-domains.custom.conf"
const dns = "223.5.5.5"

func ifErrPanic(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func readLine(reader io.Reader) (line string, err error) {
	var currentChar string = ""
	if reader == nil {
		return "", errors.New("reader is nil")
	}
	buff := make([]byte, 1)
	for currentChar != "\n" && currentChar != "\r" {
		totalRead, readErr := reader.Read(buff)
		err = readErr
		if totalRead > 0 && err == nil {
			currentChar = string(buff[:])
			line += currentChar
		}
		if err != nil {
			if err == io.EOF {
				if currentChar != "\n" && currentChar != "\r" {
					line += "\n"
				}
				break
			}
			line = ""
			break
		}
	}
	if line == "\n" || line == "\r" {
		line = ""
	}
	return line, err
}

func updateIP(customIPFile *string) {
	if err := os.Remove(ipBypassName); err != nil {
		fmt.Println(err)
	}
	fi, err := os.Create(ipBypassName)
	ifErrPanic(err)
	fmt.Println("file created")
	defer fi.Close()
	rf, err := os.Open(*customIPFile)
	if err == nil {
		defer rf.Close()
		for {
			line, err := readLine(rf)
			if len(line) > 0 {
				if _, writeErr := fi.Write([]byte(line)); writeErr != nil {
					fmt.Println(writeErr)
					break
				}
			}
			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
		}
		fmt.Println("Copied bytes: ")
	} else {
		fmt.Println(err)
	}
	fmt.Println("Reading IP URL")
	repsonse, err := http.Get(ipURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("URL read")
	defer repsonse.Body.Close()
	buff := make([]byte, 4)
	for {
		totalBytes, err := repsonse.Body.Read(buff)
		if totalBytes > 0 && err == nil {
			if _, writeErr := fi.Write(buff); writeErr != nil {
				ifErrPanic(writeErr)
			}
		}
		if err != nil {
			if err != io.EOF {
				ifErrPanic(err)
			}
			break
		}
	}
}

func updateDNS(customDNSFile *string, dnsStr *string) {
	fmt.Println("Reading from URL")
	repsonse, err := http.Get(dnsURL)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("URL read")
	defer repsonse.Body.Close()
	ifErrPanic(err)
	if err := os.Remove(fileName); err != nil {
		fmt.Println(err)
	}
	fi, err := os.Create(fileName)
	ifErrPanic(err)
	fmt.Println("file created")
	defer fi.Close()
	rf, err := os.Open(*customDNSFile)
	if err == nil {
		defer rf.Close()
		for {
			line, err := readLine(rf)
			if len(line) > 0 {
				if _, writeErr := fi.Write([]byte(line)); writeErr != nil {
					ifErrPanic(writeErr)
				}
			}
			if err != nil {
				if err != io.EOF {
					ifErrPanic(err)
				}
				break
			}
		}
	} else {
		fmt.Println(err)
	}
	fmt.Println("Writing file")
	for {
		line, err := readLine(repsonse.Body)
		if len(line) > 0 {
			strs := strings.Split(line, "/")
			if strs[0] != "server=" {
				fmt.Println("Invalid domain listing")
				fmt.Println(line)
			} else {
				strs[2] = *dnsStr + "\n"
				if _, writeErr := fi.Write([]byte(strings.Join(strs, "/"))); writeErr != nil {
					ifErrPanic(writeErr)
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				ifErrPanic(err)
			}
			break
		}
	}
}

func main() {
	var customIP string
	var customDomain string
	var customDNS string
	flag.StringVar(&customDomain, "customDomains", customDomainName, "File containing custom domain to by pass")
	flag.StringVar(&customIP, "customIPs", customIPBypassName, "File containing custom IPs to by pass")
	flag.StringVar(&customDNS, "customDNS", dns, "Custom DNS to use")
	flag.Parse()
	fmt.Println("Custom Domain", customDomain)
	fmt.Println("Custom IP", customIP)
	fmt.Println("Custom DNS", customDNS)
	updateIP(&customIP)
	updateDNS(&customDomain, &customDNS)
}
