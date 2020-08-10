# Pingr
<img src="ui/assets/gopher.png" alt="pingr gopher" width="100"/>

Pingr is a service that monitors the health of other services. It can send you an email or post-hook whenever a monitored service is behaving unexpectedly. 

## Supported methods

The methods are divided into two different categories, poll and push. Pingr can poll services and ask them how they are doing.
Push methods are the opposite, Pingr sets up a unique endpoint for each service to send http requests to and thereby tell Pingr how they are doing.

**Poll methods**
+ HTTP
+ Prometheus
+ TLS
+ DNS
+ Ping
+ SSH
+ TCP

**Push methods**
+ HTTP
+ Prometheus

## Running on local
* Setup the `docker-compose.yml` file, see `docker-compose-example.yml`for an example.
* `docker compose up`

## Production
**Docker swarm example**

* Create a small node on a VPS provider such as Digital Ocean or Linode 
* Point a DNS `A` record to the IP of the node
* Install docker on the node
* Run the following (with the correct credentials)

```bash 
mkdir -p /root/pingr
mkdir -p /root/pingr-tls
echo '
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


