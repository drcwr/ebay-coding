Hi Wurui,

 

This is the follow up coding problem for the interview,   please provide the github link for the code within one week.

 

Thanks

Shone


--------------------


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