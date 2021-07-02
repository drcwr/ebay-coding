## Email
Hi Wurui,

 

This is the follow up coding problem for the interview,   please provide the github link for the code within one week.

 

Thanks

Shone


--------------------
## Problem

Description Suppose there is a web service, for example, the URL is http://www.ebay.com.

And behind the DNS, there are multiple servers providing the service, and each server has a different IP.

Please write a client program to query the IPs behind the DNS, and send 100 requests to different IPs according to some client-side loadbalancing strategy in order to achieve better performance and lower lantency.

There are the following assumptions:

1. The IP address behind the DNS may change.

2. The IP address may locate in a different geographic location, so the response time for the same request may be different.

3. The IP address may be unreachable due to system outage and other reasons.

4. If the return code is between [200-500], it is regarded as a success, if timeout or 500+ error occurs, regarded as a failure.

Please consider the following requirements:

1. Resolve the IP addresses corresponding to the DNS.

2. Requests should be evenly distributed to different IPs behind the DNS.

3. Requests should be retried upon failures.

4. The processing of the requests should be as efficient as possible.

Hint:

You could use following command can query the IPs behind ebay.com.

dig ebay.com

; <<>> DiG 9.10.6 <<>> ebay.com

;; global options: +cmd

;; Got answer:

;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 14924

;; flags: qr rd ra; QUERY: 1, ANSWER: 6, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:

; EDNS: version: 0, flags:; udp: 4000

;; QUESTION SECTION:

;ebay.com. IN A

;; ANSWER SECTION:

ebay.com. 10 IN A 209.140.148.143

ebay.com. 10 IN A 64.4.253.77

ebay.com. 10 IN A 66.211.172.37

ebay.com. 10 IN A 216.113.179.53

ebay.com. 10 IN A 66.211.175.229

ebay.com. 10 IN A 216.113.181.253

Evaluation:

1. The program could be run directly from local environment.

2. Good performance and robustness of the program is a plus.


------------------
## 思路
1. 查询 DNS ，获取域名对应的ip列表
2. 根据ip个数，调度并发发送请求
3. 判断请求返回
4. 返回失败，进行重试
5. 完成该轮请求，统计结果


--------------------------------

go run main-v4.go 
2021/07/02 14:35:41 [66.211.175.229 216.113.181.253 216.113.179.53 209.140.148.143 64.4.253.77 66.211.172.37 11.0.0.0]
2021/07/02 14:35:41 init ip 66.211.175.229,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c000), okch:(chan int)(0xc0000740e0), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"66.211.175.229", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:41 init ip 216.113.181.253,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c0b0), okch:(chan int)(0xc000074150), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"216.113.181.253", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:41 init ip 216.113.179.53,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c160), okch:(chan int)(0xc0000741c0), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"216.113.179.53", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:41 init ip 209.140.148.143,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c210), okch:(chan int)(0xc000074230), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"209.140.148.143", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:41 init ip 64.4.253.77,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c2c0), okch:(chan int)(0xc0000742a0), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"64.4.253.77", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:41 init ip 66.211.172.37,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c370), okch:(chan int)(0xc000074310), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"66.211.172.37", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:41 init ip 11.0.0.0,	 ipm=main.IpMng{todo:1, windowch:(chan int)(0xc00016c420), okch:(chan int)(0xc000074380), done:0, succ:0, retryNum:1, timeout:20, status:0, RTT:0, ip:"11.0.0.0", client:(*http.Client)(nil), m:sync.Mutex{state:0, sema:0x0}}
2021/07/02 14:35:42 gorun 5,v.Status == 0=>1 66.211.175.229 RTT 0.951031608
2021/07/02 14:35:42 gorun 5,v.Status == 0=>1 216.113.179.53 RTT 0.953126281
2021/07/02 14:35:42 gorun 5,v.Status == 0=>1 209.140.148.143 RTT 1.04709118
2021/07/02 14:35:42 gorun 5,v.Status == 0=>1 66.211.172.37 RTT 1.087046
2021/07/02 14:35:42 gorun 5,v.Status == 0=>1 64.4.253.77 RTT 1.155081043
2021/07/02 14:35:42 gorun 5,v.Status == 0=>1 216.113.181.253 RTT 1.155128192
2021/07/02 14:35:44 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:35:44 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:35:44 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:35:44 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:35:44 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:35:45 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:36:01 client.Do  11.0.0.0  err need retry!!timeout = 20 usetime 20.009784241
2021/07/02 14:36:01 go run GetUrl err, Get "http://11.0.0.0": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
2021/07/02 14:36:03 v.Status = 2 gorun  FOR 11.0.0.0
2021/07/02 14:36:03 v.Status = 2 del from v.windowch 11.0.0.0
2021/07/02 14:36:03 status 2 ,val 1
2021/07/02 14:36:03 v.Status = 2 gorun 9 r.todo 1 11.0.0.0
2021/07/02 14:36:03 getNum 02
2021/07/02 14:36:03 getNum 01 r.todo <= 0 wait <-r.retrych
2021/07/02 14:36:04 



2021/07/02 14:36:04 --------------main 66.211.175.229 	 done: 22 ,succ: 22 ,RTT: 0.951031608
2021/07/02 14:36:04 --------------main 216.113.181.253 	 done: 13 ,succ: 13 ,RTT: 1.155128192
2021/07/02 14:36:04 --------------main 216.113.179.53 	 done: 22 ,succ: 22 ,RTT: 0.953126281
2021/07/02 14:36:04 --------------main 209.140.148.143 	 done: 17 ,succ: 17 ,RTT: 1.04709118
2021/07/02 14:36:04 --------------main 64.4.253.77 	 done: 13 ,succ: 13 ,RTT: 1.155081043
2021/07/02 14:36:04 --------------main 66.211.172.37 	 done: 13 ,succ: 13 ,RTT: 1.087046
2021/07/02 14:36:04 --------------main 11.0.0.0 	 done: 1 ,succ: 0 ,RTT: 0
2021/07/02 14:36:04 --------------main  total=101,succ=100
