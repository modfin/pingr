# Pingr
<img src="https://storage.googleapis.com/gopherizeme.appspot.com/gophers/62d0c3d5f52dbc9c803ea7bfaa2829d75a2f8fa2.png" alt="pingr gopher" width="100"/>

Pingr is a service that monitors the health of other services. It can send you an email or post-hook whenever a monitored service is behaving unexpectedly. 

## Supported methods

The methods are divided into two different categories, poll and push. Pingr can poll services and ask them how they are doing.
Push methods are the opposite, Pingr sets up a unique endpoint for each service to send http requests to and thereby tell Pingr how they are doing.

#### Poll methods
+ HTTP
+ Prometheus
+ TLS
+ DNS
+ Ping
+ SSH
+ TCP

#### Push methods
+ HTTP
+ Prometheus
