package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	ips, err := getips()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(ips)
	}

	gorun(ips)

}

var total = 20
var m sync.Mutex

func gettoken() bool {
	m.Lock()
	defer m.Unlock()
	if total > 0 {
		total--
		return true
	}

	return false
}

func gorun(ips []string) {
	wg := sync.WaitGroup{}
	wg.Add(total)
	for _, v := range ips {
		go func(v string) {
			for {
				ok := gettoken()
				if ok {
					log.Println(v)
					geturl(v)
					// time.Sleep(2 * time.Second)
					wg.Done()
				} else {
					break
				}
			}
		}(v)
	}
	wg.Wait()
	time.Sleep(2 * time.Second)
}

func getips() (ips []string, err error) {
	cmd := exec.Command("dig", "ebay.com")
	if stdout, err := cmd.StdoutPipe(); err != nil {
		log.Fatal(err)
		return ips, err
	} else {
		defer stdout.Close()
		if err = cmd.Start(); err != nil {
			log.Fatal(err)
			return ips, err
		}
		// if opBytes, err := ioutil.ReadAll(stdout); err != nil {
		// 	log.Fatal(err)
		// 	return opBytes, err
		// } else {
		// 	// log.Println(string(opBytes))
		// 	return opBytes, nil
		// }

		reader := bufio.NewReader(stdout)
		start := 0 // init
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			// fmt.Printf(line)
			if start == 1 && line == "\n" {
				// fmt.Printf("end")
				start = 2 // end
				break
			}
			if start == 0 && line == ";; ANSWER SECTION:\n" {
				start = 1 // start
				continue
			}
			if start == 1 {
				rec := strings.Split(line[:len(line)-1], "\t")
				// fmt.Println(rec[len(rec)-1])
				ip := rec[len(rec)-1]
				err := checkip(ip)
				if err == nil {
					ips = append(ips, ip)
				}
			}
		}
	}

	ips = append(ips, "11.0.0.0")
	return
}

func checkip(ip string) error {
	return nil
}

// func getips2(input []byte) (ips []string, err error) {

// 	ret, err := rundig()
// 	if err != nil {
// 		log.Fatal(err)
// 	} else {
// 		log.Println(string(ret))
// 	}

// 	ips = append(ips, "0.0.0.0")
// 	return
// }

// func rundig() ([]byte, error) {
// 	var opBytes []byte
// 	cmd := exec.Command("dig", "ebay.com")
// 	if stdout, err := cmd.StdoutPipe(); err != nil {
// 		log.Fatal(err)
// 		return opBytes, err
// 	} else {
// 		defer stdout.Close()
// 		if err := cmd.Start(); err != nil {
// 			log.Fatal(err)
// 			return opBytes, err
// 		}
// 		if opBytes, err := ioutil.ReadAll(stdout); err != nil {
// 			log.Fatal(err)
// 			return opBytes, err
// 		} else {
// 			// log.Println(string(opBytes))
// 			return opBytes, nil
// 		}
// 	}
// }

func geturl(ip string) {
	req, err := http.NewRequest("GET", "http://"+ip, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("client.Do", err)
		log.Println("need retry!!")
		return
	}
	defer resp.Body.Close()

	log.Println("resp.StatusCode", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode > 500 {
		log.Println("need retry!!")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll", err)
		return
	}
	_ = body
	// log.Println(string(body))

}
