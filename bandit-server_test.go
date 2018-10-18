package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBandit(t *testing.T) {
	//sample size
	ss := 10000
	var a, b, rewa, rewb int

	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= ss; i++ {
		ret := GetReq()
		if ret == 1 {
			a++
		} else {
			b++
		}
		if i%10 == 0 {
			//send reward (10% ctr)
			random := rand.Intn(9)
			switch ret {
			case 1:
				if random < 7 {
					//log.Println(random, "var A win in 70%")
					PutReq("a")
					rewa++
				} else {
					//log.Println(random, "var A not win")
				}
			case 2:
				if random < 3 {
					//log.Println(random, "var B win in 30%")
					PutReq("b")
					rewb++
				} else {
					//log.Println(random, "var B not win")
				}
			}
		}
		//log.Println(ret)
	}
	log.Println("A:", a, "B:", b, "rewA:", rewa, "rewB:", rewb)
}

func GetReq() int {
	var content string
	response, err := http.Get("http://localhost:3000/ucb1?ab=a,b")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		content = string(contents)
	}
	if content == "{\"ab\":\"a\"}" {
		return 1
	}
	return 2
}

func PutReq(win string) {
	d := "ab=" + win
	url := "http://localhost:3000/ucb1"
	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, strings.NewReader(d))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.ContentLength = int64(len(d))
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		_, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println("The calculated length is:", len(string(contents)), "for the url:", url)
		//fmt.Println("   ", response.StatusCode)
		//hdr := response.Header
		//for key, value := range hdr {
		//fmt.Println("   ", key, ":", value)
		//}
		//fmt.Println(string(contents), win)
	}
}
