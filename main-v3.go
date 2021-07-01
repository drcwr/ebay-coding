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

type IpMng struct {
        // doing    int
        todo     int
        window   chan int
        ok chan int
        done     int
        succ     int
        retryNum int
        timeout  int
        goNum    int
        Status   int // 0 syn ;1 ack;2 serverfaild
        ip       string
        // ch       chan int
        m sync.Mutex
}

type ReqMng struct {
        IpNum   int
        ReqNum  int
        todo    int
        retrych chan int
        url     string
        Ips     []string
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
                log.Println("--------------main", v.ip, "\t", "done", v.done, "succ", v.succ)
                done += v.done
                succ += v.succ
        }
        log.Printf("--------------main  total=%d,succ=%d\n", done, succ)
}

func (r *ReqMng) init() {
        var ipsMng = []IpMng{}

        r.url = "ebay.com"
        r.ReqNum = 100
        r.todo = r.ReqNum

        ips, err := r.getips()
        if err != nil {
                log.Fatal(err)
        } else {
                log.Println(ips)
                r.Ips = ips
        }

        r.IpNum = len(ips)
        r.retrych = make(chan int,1)

        for _, v := range r.Ips {
                // ipm := IpMng{todo: 0, window: r.ReqNum / r.IpNum, retryNum: 1, timeout: 10, ip: v}
                ipm := IpMng{todo: 0, retryNum: 1, timeout: 10, ip: v}
                ipm.window = make(chan int, 10)
                ipm.ok = make(chan int,1)
                if r.todo > 0 {
                        ipm.todo = 1
                        // ipm.window <- 1
                        r.todo--
                }
                ipsMng = append(ipsMng, ipm)
        }

        r.ipsMng = ipsMng

        for _, v := range r.ipsMng {
                log.Printf("init ip %s,\t todo=%d\n", v.ip, v.todo)
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
        return
}

func (im *IpMng) getToken(r *ReqMng) bool {
        // log.Println("getToken 00",im.ip)
        if im.Status == 0 {
                // log.Println("getToken 01",im.ip)
                <- im.ok
                // return false
        }
                // log.Println("getToken Lock-----------------------------Lock", im.ip)
                // im.m.Lock()
                // log.Println("getToken Lock-----------------------------unLock", im.ip)
                // im.m.Unlock()
                // log.Println("getToken 03",im.ip)
        if im.Status == 2 {
                return false
        }
        im.todo = r.getNum()
        if im.todo > 0 {
                // log.Println("getToken 1",im.ip)
                return true
        }
        // log.Println("getToken 2")
        // log.Println("getToken 3")

        return false
}

func (r *ReqMng) getNum() int {
        // log.Println("getNum 00")
        if r.todo <= 0 {
                log.Println("getNum 01")
                <-r.retrych
                log.Println("getNum 02")
        }
        if r.todo > 0 {
                r.m.Lock()
                // log.Println("---------getNum r.todo", r.todo)
                num := 1//(r.todo + r.IpNum - 1) / r.IpNum
                r.todo -= num
                // log.Println("---------getNum r.todo", r.todo, "num", num)
                r.m.Unlock()
                return num
        }
        return 0
}
func (r *ReqMng) gorun() {
        wg := sync.WaitGroup{}
        wg.Add(r.ReqNum)
        for k, _ := range r.ipsMng {
                go func(v *IpMng) {
                        ok := true // v.getToken(r)
                        for {
                                // log.Printf("gorun start %s %v %d\n", v.ip, ok, v.todo)
                                if ok {
                                        todo := true
                                        // log.Println("gorun 1", v.ip)
                                        // v.m.Lock()
                                        // log.Printf("gorun %s %v %d lock\n", v.ip, ok, v.todo)
                                        // if v.todo > 0 {
                                        //      v.todo--
                                        //      todo = true
                                        // }
                                        // v.m.Unlock()
                                        if todo {
                                                // log.Println("gorun 2", v.ip)
                                                v.window <- 1
                                                go func(v *IpMng) {
                                                        i := 0
                                                        retry := false
                                                        var err error
                                                        for i = 0; i < v.retryNum; i++ {
                                                                // log.Println("gorun 3", v.ip)
                                                                retry, err = geturl(v)
                                                                // log.Println("gorun 4", v.ip)
                                                                if err != nil {
                                                                        log.Println("go run geturl err,", err)
                                                                }
                                                                // log.Println("gorun 5", v.ip)
                                                                if retry == false {
                                                                // log.Println("gorun 6", v.ip)
                                                                <-v.window
                                                                // log.Println("gorun 7", v.ip)
                                                                // log.Println("Lock-----------------------------Lock 1", v.ip)
                                                                v.m.Lock()
                                                                // log.Println("Lock-----------------------------Lock 2", v.ip)
                                                                v.succ++
                                                                // log.Println("Lock-----------------------------unLock", v.ip)
                                                                v.m.Unlock()
                                                                wg.Done()
                                                                        break
                                                                } else {
                                                                        time.Sleep(time.Duration(2^i) * time.Second)
                                                                }
                                                                log.Println("gorun geturl retry", v.ip)
                                                        }
                                                        // log.Println("gorun 8", v.ip)

                                                        v.m.Lock()
                                                        if retry && i == v.retryNum {
                                                                // todo may del the ip from ip list
                                                                log.Println("gorun need retry 4", v.ip)

                                                                v.Status = 2
                                                                close(v.window)
                                                                r.m.Lock()
                                                                log.Println("gorun 8", v.ip)
                                                                for {
                                                                        log.Println("gorun 9", v.ip)
                                                                        val,ok := <- v.window
                                                                                log.Println("status 2 ,val,ok",val,ok)
                                                                                if val > 0 {
                                                                                        r.todo++
                                                                                } else if ok == false {
                                                                                        break
                                                                                }
                                                                        }
                                                                        r.m.Unlock()
                                                                        r.retrych <- 1


                                                                log.Println("!!!! retry all error,i=", i)
                                                        } else if v.Status == 0 {
                                                                log.Println("gorun 5", v.ip)
                                                                v.Status = 1
                                                                v.ok <- 1
                                                        }
                                                        v.done++
                                                        v.m.Unlock()
                                                }(v)
                                        }
                                }
                                // log.Println("gorun 6", v.ip)

                                if v.Status == 2 {
                                // log.Println("gorun 7", v.ip)
                                r.m.Lock()
                                // log.Println("gorun 8", v.ip)
                                for {
                                        // log.Println("gorun 9", v.ip)
                                        val,ok := <- v.window
                                                log.Println("status 2 ,val,ok",val,ok)
                                                if val > 0 {
                                                        r.todo++
                                                } else if ok == false {
                                                        break
                                                }
                                        }
                                        r.m.Unlock()
                                        log.Println("!!!!!!!!!!!!!!1exit go",v.ip)
                                        break
                                }
                                // log.Println("gorun 99", v.ip)

                                ok = v.getToken(r)
                        }
                }(&r.ipsMng[k])
        }
        wg.Wait()
        // close(r.retrych)
        // time.Sleep(2 * time.Second)
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
        t2 := time.Now()
        t3 := t2.Sub(t1).Seconds()
        // log.Println("client.Do  usetime", t3, "----------", ipm.ip)
        if err != nil {
                log.Println("client.Do  err need retry!!timeout =", ipm.timeout, "usetime", t3)
                return true, err
        }
        defer resp.Body.Close()

        // log.Println("geturl ", ipm.ip, ",\tresp.StatusCode", resp.StatusCode)
        if resp.StatusCode < 200 || resp.StatusCode > 500 {
                log.Println("need retry!! resp.StatusCode", resp.StatusCode)
                return true, err
        }

        return
}
