package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IpMng struct {
	totalNum int
	count    int
	succ     int
	retryNum int
	timeout  int
	ip       string
	m        sync.Mutex
}

type ReqMng struct {
	IpNum  int
	ReqNum int
	url    string
	Ips    []string
	ipsMng []IpMng
}

func main() {
	var reqs ReqMng
	var count, succ int

	reqs.init()
	reqs.gorun()

	log.Println("\n\n\n")
	for _, v := range reqs.ipsMng {
		log.Println("--------------main", v.ip, "\t", "conut", v.count, "succ", v.succ)
		count += v.count
		succ += v.succ
	}
	log.Printf("--------------main  total=%d,succ=%d\n", count, succ)
}

func (r *ReqMng) init() {
	var ipsMng = []IpMng{}

	r.url = "ebay.com"
	r.ReqNum = 100

	ips, err := r.getips()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(ips)
		r.Ips = ips
	}

	r.IpNum = len(ips)

	for k, v := range r.Ips {
		ipm := IpMng{totalNum: r.ReqNum / r.IpNum, retryNum: 5, timeout: 10 + k, ip: v}
		ipsMng = append(ipsMng, ipm)
	}
	ipsMng[0].totalNum = r.ReqNum - ipsMng[0].totalNum*(r.IpNum-1)

	r.ipsMng = ipsMng

	for _, v := range r.ipsMng {
		log.Printf("init ip %s,\t total=%d\n", v.ip, v.totalNum)
	}
}

func (r *ReqMng) getips() (ips []string, err error) {
	cmd := exec.Command("dig", r.url)
	if stdout, err := cmd.StdoutPipe(); err != nil {
		log.Fatal(err)
		return ips, err
	} else {
		defer stdout.Close()
		if err = cmd.Start(); err != nil {
			log.Fatal(err)
			return ips, err
		}

		reader := bufio.NewReader(stdout)
		start := 0 // init
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			if start == 1 && line == "\n" {
				start = 2 // end
				break
			}
			if start == 0 && line == ";; ANSWER SECTION:\n" {
				start = 1 // start
				continue
			}
			if start == 1 {
				rec := strings.Split(line[:len(line)-1], "\t")
				ip := rec[len(rec)-1]
				err := checkip(ip)
				if err == nil {
					ips = append(ips, ip)
				}
			}
		}
	}

	// debug ip
	ips = append(ips, "11.0.0.0")
	ips = append(ips, "12.0.0.0")
	ips = append(ips, "13.0.0.0")
	ips = append(ips, "14.0.0.0")
	ips = append(ips, "15.0.0.0")
	ips = append(ips, "16.0.0.0")
	return
}

func (im *IpMng) getToken(r *ReqMng) bool {
	im.m.Lock()
	defer im.m.Unlock()
	if im.totalNum > 0 {
		im.totalNum--
		return true
	} else {
		im.totalNum = r.getNum()
		if im.totalNum > 0 {
			im.totalNum--
			return true
		}
	}

	return false
}

func (r *ReqMng) getNum() int {
	maxId := -1
	maxNum := 0
	for k, v := range r.ipsMng {
		if v.totalNum > maxNum {
			maxId = k
			maxNum = v.totalNum
		}
	}
	if maxId > -1 {
		r.ipsMng[maxId].m.Lock()
		defer r.ipsMng[maxId].m.Unlock()
		num := (r.ipsMng[maxId].totalNum + 1) / 2
		r.ipsMng[maxId].totalNum -= num
		return num
	}
	return 0
}
func (r *ReqMng) gorun() {
	wg := sync.WaitGroup{}
	wg.Add(r.ReqNum)
	for k, _ := range r.ipsMng {
		go func(v *IpMng) {
			for {
				ok := v.getToken(r)
				if ok {
					i := 0
					for i = 0; i < v.retryNum; i++ {
						retry, err := geturl(v)
						if err != nil {
							log.Println("go run geturl err,", err)
						}
						if retry == false {
							v.succ++
							break
						}
						time.Sleep(time.Duration(2^i) * time.Second)
					}
					if i == 5 {
						// todo may del the ip from ip list
						log.Println("!!!! retry all error,i=", i)
					}
					v.count++
					wg.Done()
				} else {
					log.Println("--------------loop end", v.ip, "\t", "conut", v.count, "succ", v.succ)
					break
				}
			}
		}(&r.ipsMng[k])
	}
	wg.Wait()

	time.Sleep(2 * time.Second)
}

func checkip(ip string) error {
	arr := strings.Split(ip, ".")
	if len(arr) == 4 {
		for _, val := range arr {
			_, err := getipnum(val)
			if err != nil {
				log.Println("not ip", arr)
				return fmt.Errorf("check ip error,%v\n", arr)
			}
		}
		return nil
	}
	return fmt.Errorf("check ip error,%v\n", arr)
}

func getipnum(ip string) (int, error) {
	ipnum, err := strconv.Atoi(ip)
	if err != nil || ipnum > 255 || ipnum < 0 {
		log.Println("not ip", ip)
		return 0, err
	}
	return ipnum, nil
}

func geturl(ipm *IpMng) (retry bool, err error) {
	req, err := http.NewRequest("GET", "http://"+ipm.ip, nil)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// req.Header.Set("Content-Type", "application/json")
	t1 := time.Now()
	client := &http.Client{Timeout: time.Duration(ipm.timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t2 := time.Now()
		t3 := t2.Sub(t1).Seconds()
		log.Println("client.Do  err need retry!!timeout =", ipm.timeout, "usetime", t3)
		return true, err
	}
	defer resp.Body.Close()

	log.Println("geturl ", ipm.ip, ",\tresp.StatusCode", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode > 500 {
		log.Println("need retry!!")
		return true, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll", err)
		return false, err
	}
	_ = body
	return
}
