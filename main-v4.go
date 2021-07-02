package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LINK_STATUS_START = iota // 0
	LINK_STATUS_SUCC         // 1
	LINK_STATUS_ERROR        // 2
)

const (
	READ_INIT  = iota // 0
	READ_START        // 1
	READ_END          // 2
)

type IpMng struct {
	todo     int
	windowch chan int
	okch     chan int
	done     int
	succ     int
	retryNum int
	timeout  int
	status   int // 0 syn ;1 ack;2 serverfaild
	RTT      float64
	ip       string
	client   *http.Client
	m        sync.Mutex
}

type ReqMng struct {
	ipNum   int
	reqNum  int
	todo    int
	retrych chan int
	url     string
	ips     []string
	ipsMng  []IpMng
	m       sync.Mutex
}

func main() {
	var reqs ReqMng
	var done, succ int

	reqs.init()
	reqs.gorun()

	log.Println("\n\n\n")
	for _, v := range reqs.ipsMng {
		log.Println("--------------main", v.ip, "\t", "done:", v.done, ",succ:", v.succ, ",RTT:", v.RTT)
		done += v.done
		succ += v.succ
	}
	log.Printf("--------------main  total=%d,succ=%d\n", done, succ)
}

func (r *ReqMng) init() {
	var ipsMng = []IpMng{}

	// best value
	ReqNum := 100
	httptimeout := 20

	r.url = "ebay.com"
	r.reqNum = ReqNum
	r.todo = r.reqNum

	ips, err := r.getips()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(ips)
		r.ips = ips
	}

	r.ipNum = len(ips)
	window := 10 // ReqNum / (r.IpNum + r.IpNum)
	r.retrych = make(chan int, r.reqNum)

	for _, v := range r.ips {
		ipm := IpMng{todo: 0, retryNum: 1, timeout: httptimeout, ip: v}
		ipm.windowch = make(chan int, window)
		ipm.okch = make(chan int, 1)

		if r.todo > 0 {
			ipm.todo = 1
			r.todo--
		}
		ipsMng = append(ipsMng, ipm)
	}

	r.ipsMng = ipsMng

	for _, v := range r.ipsMng {
		log.Printf("init ip %s,\t ipm=%#v\n", v.ip, v)
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
		start := READ_INIT // init
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			if start == READ_START && line == "\n" {
				start = READ_END // end
				break
			}
			if start == READ_INIT && line == ";; ANSWER SECTION:\n" {
				start = READ_START // start
				continue
			}
			if start == READ_START {
				rec := strings.Split(line[:len(line)-1], "\t")
				ip := rec[len(rec)-1]
				err := CheckIp(ip)
				if err == nil {
					ips = append(ips, ip)
				}
			}
		}
	}

	// debug ip
	ips = append(ips, "11.0.0.0")
	// ips = append(ips, "12.0.0.0")
	return
}

func (im *IpMng) getToken(r *ReqMng) bool {
	if im.status == LINK_STATUS_ERROR {
		return false
	}
	if im.status == LINK_STATUS_START {
		<-im.okch
	}
	im.todo = r.getNum()
	if im.todo > 0 {
		return true
	}

	return false
}

func (r *ReqMng) getNum() int {
	if r.todo <= 0 {
		log.Println("getNum 01 r.todo <= 0 wait <-r.retrych")
		<-r.retrych
		log.Println("getNum 02")
	}
	if r.todo > 0 {
		r.m.Lock()
		num := 1
		r.todo -= num
		r.m.Unlock()
		return num
	}
	return 0
}
func (r *ReqMng) gorun() {
	wg := sync.WaitGroup{}
	wg.Add(r.reqNum)
	for k, _ := range r.ipsMng {
		go func(v *IpMng) {
			ok := true
			for {
				if ok {
					v.windowch <- 1
					go func(v *IpMng) {
						var (
							i     = 0
							retry = false
							err   error
						)
						t1 := time.Now()
						for i = 0; i < v.retryNum; i++ {
							retry, err = GetUrl(v)
							if err != nil {
								log.Println("go run GetUrl err,", err)
							}
							if retry == false {
								<-v.windowch
								v.m.Lock()
								v.succ++
								v.m.Unlock()
								wg.Done()
								break
							} else {
								time.Sleep(time.Duration(2^i) * time.Second)
							}
						}

						v.m.Lock()
						if retry && i == v.retryNum {
							v.status = LINK_STATUS_ERROR
							r.m.Lock()
							log.Println("v.Status = 2 gorun  FOR", v.ip)
							select {
							case val := <-v.windowch:
								log.Println("v.Status = 2 del from v.windowch", v.ip)
								log.Println("status 2 ,val", val)
								if val > 0 {
									r.todo++
									log.Println("v.Status = 2 gorun 9 r.todo", r.todo, v.ip)
								}
							default:
							}
							r.m.Unlock()
							r.retrych <- 1
						} else if v.status == LINK_STATUS_START {
							t2 := time.Now()
							t3 := t2.Sub(t1).Seconds()
							log.Println("gorun 5,v.Status == 0=>1", v.ip, "RTT", t3)
							v.RTT = t3
							v.status = LINK_STATUS_SUCC
							v.okch <- 1
						}
						v.done++
						v.m.Unlock()
					}(v)
				}
				ok = v.getToken(r)
			}
		}(&r.ipsMng[k])
	}
	wg.Wait()
	// close(r.retrych)
	// time.Sleep(2 * time.Second)
}

func CheckIp(ip string) error {
	arr := strings.Split(ip, ".")
	if len(arr) == 4 {
		for _, val := range arr {
			_, err := GetIpNum(val)
			if err != nil {
				log.Println("invalid ip", arr)
				return fmt.Errorf("check ip error,%v\n", arr)
			}
		}
		return nil
	}
	return fmt.Errorf("check ip error,%v\n", arr)
}

func GetIpNum(ip string) (int, error) {
	ipnum, err := strconv.Atoi(ip)
	if err != nil || ipnum > 255 || ipnum < 0 {
		log.Println("invalid ip", ip)
		return 0, err
	}
	return ipnum, nil
}

func GetUrl(ipm *IpMng) (retry bool, err error) {
	ipm.m.Lock()
	if ipm.client == nil {
		// req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: time.Duration(ipm.timeout) * time.Second}
		ipm.client = client
	}
	req, err := http.NewRequest("GET", "http://"+ipm.ip, nil)
	if err != nil {
		log.Println(err)
		return false, err
	}
	ipm.m.Unlock()
	t1 := time.Now()
	resp, err := ipm.client.Do(req)
	t2 := time.Now()
	t3 := t2.Sub(t1).Seconds()
	if err != nil {
		log.Println("client.Do ", ipm.ip, " err need retry!!timeout =", ipm.timeout, "usetime", t3)
		return true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 500 {
		log.Println("need retry!! resp.StatusCode", resp.StatusCode)
		return true, fmt.Errorf("resp.StatusCode =%d,error", resp.StatusCode)
	}

	return
}
