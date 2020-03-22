
# Coding Exercise
For this exercise, we would like to see how you go about solving a rather straightforward coding challenge and
architecting for the future. One of our key values in how we develop new systems is to start with very simple
implementations and progressively make them more capable, scalable and reliable. And releasing them each step of
the way. As you work through this exercise it would be good to "release" frequent updates by pushing updates to a
shared git repo (we like to use Bitbucket's free private repos for this, but gitlab or github also work). It's up to you
how frequently you do this and what you decide to include in each push. Don't forget some unit tests (at least
something representative).
Here's what we would like you to build.
# URL lookup service
We have an HTTP proxy that is scanning traffic looking for malware URLs. Before allowing HTTP connections to be
made, this proxy asks a service that maintains several databases of malware URLs if the resource being requested is
known to contain malware.
Write a small web service, in the language/framework your choice, that responds to GET requests where the caller
passes in a URL and the service responds with some information about that URL. The GET requests look like this:
GET /urlinfo/1/{hostname_and_port}/{original_path_and_query_string}
The caller wants to know if it is safe to access that URL or not. As the implementer, you get to choose the response
format and structure. These lookups are blocking users from accessing the URL until the caller receives a response
from your service.
Give some thought to the following:

- The size of the URL list could grow infinitely, how might you scale this beyond the memory capacity of this
VM? Bonus if you implement this.
- The number of requests may exceed the capacity of this VM, how might you solve that? Bonus if you
implement this.
- What are some strategies you might use to update the service with new URLs? Updates may be as much as 5
thousand URLs a day with updates arriving every 10 minutes.
- Bonus points if you containerize the app.
---
# Design
#### The size of the URL list can grow indefinitly
I picked up [rqlite](https://github.com/rqlite/rqlite) and put it to work. The leader container is in charge of importing new rows in to the dataset using rqlite API. Worker containers subscribe to updates, but they do not take part in the raft. This keeps the write latency low. Worker containers can read the dataset directly from disk (rqlite flag -on-disk) using SQLite.
#### The number of requests may exceed the capacity of this VM
Each worker container seems to be able to do 16k req/s on the test machine. We can activate as many (N) worker containers as we need to service the requests. We can load balance the workers behind a proxy until we reach the proxy's req/s ceiling. Then we can scale out to M leader containers and proxy containers, each with N number of worker containers
#### What are some strategies you might use to update the service with new URLs?
The rqlite API provides everything we need to accomplish this, including updating the worker containers as new data comes in. We can insert 5000 new rows in about 0.8s on the test machine.
# How to build?
Start a fresh rqlite instance
```bash
rqlited -on-disk ~/rqlited &
```
Run import to get data from urlhaus
```bash
go run cmd/import/import.go
```
Start the server
```bash
go run cmd/lookup/lookup.go
```
In a web browser visit

http://localhost:8080/urlinfo/1/111.43.223.158:52727/Mozi.m

http://localhost:8080/urlinfo/1/Safe

The API is designed to be similar to rqlite API.
