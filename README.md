# <img src="ui/assets/gopher.png" alt="pingr gopher" width="100"/> Pingr 

Pingr is a service which monitors the health of other services through various different types of tests. It can send you an email or post-hook whenever a monitored service is behaving unexpectedly. 

### Supported test methods

The test methods are divided into two different categories, poll and push. Pingr can poll services and ask them how they are doing.
Push methods are the opposite, Pingr sets up a unique endpoint for each service to send http requests to and thereby tell Pingr how they are doing.

**Poll methods**
+ DNS
    + IP-Address/CNAME/TXT/MX/NS
+ HTTP
    + GET/POST/PUT/HEAD/DELETE
    + Request
        + Header
        + Body
    + Expected response
        + Status code
        + Header
        + Body
+ Ping
+ Prometheus
    + GAUGE/COUNTER
+ SSH
    + Username/Password
    + Username/Key
+ TCP
+ TLS/SSL

**Push methods**
+ HTTP
+ Prometheus
    + GAUGE/COUNTER

### General test settings
 + Hostname/Domain/Url - where test will poll against (poll tests)
 + Interval - duration between each test (poll tests)
 + Timeout - max allowed duration for test / max duration between each push (poll tests / push tests)


### Actions upon unexpected test response
When a test fails X amounts of times consecutively you can choose to send an email or post-hook to inform of the test failure. If the email should work properly the SMTP server and credentials has to be defined correctly in the `docker-compose.yml` file.
Whenever a test fails an incident will be created and stored, regardless of someone being contacted or not. The incident can be seen in the UI.

### Misc functionality
+ View average response times
+ View test logs
+ Pause test


## Running on local
* Setup the `docker-compose.yml` file, see `docker-compose-example.yml`for an example.
* `$ docker compose up`

## Production
**Docker swarm example**

* Create a small node on a VPS provider such as Digital Ocean or Linode 
* Point a DNS `A` record to the IP of the node
* Install docker on the node
* Run the following (with the correct credentials)

```bash 
$ mkdir -p /root/pingr-tls
$ touch /root/pingr.sqlite
$ echo '
version: "3.0"
services:
  pingr:
    image: modfin/pingrd:latest
    sysctls:
      - net.ipv4.ping_group_range=0 2147483647
    environment:
      AUTO_TLS: "true"
      AUTO_TLS_EMAIL: "mail@example.com"
      AUTO_TLS_DIR: "/pingr-tls"
      AUTO_TLS_DOMAINS: "pingr.domain.com"
      HTTP_PORT: "80"
      HTTPS_PORT: "443"
      BASE_URL: "https://pingr.domain.com"
      BASIC_AUTH_USER: "username"
      BASIC_AUTH_PASS: "password"
      SMTP_HOST: "smtp.example.com"
      SMTP_USERNAME: "example@smtp-server.com"
      SMTP_PASSWORD: "smtpPassword"
      AES_KEY: "a5148c8353eb2078eff44dfa8c3c890444418cf88627e17132a8d5d44335788a" ## Generate using 'openssl rand -hex 32'
      SQLITE_PATH: "/pingr.sqlite"
      SQLITE_MIGRATE: "false"
    ports:
    - "80:80"
    - "443:443"
    volumes:
    - /root/pingr.sqlite:/pingr.sqlite
    - /root/pingr-tls:/pingr-tls

' > docker-compose.yml

$ docker stack deploy -c docker-compose.yml pingr

```


