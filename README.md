# __PDN-WebRTC__

This repository contains the source code for a __1:N streaming server__ built with __WebRTC__.   
The server is designed to support __real-time video streaming and chat__ functionality, using a __peer-assisted CDN__ architecture to optimize network efficiency and reduce server load.

# Features

- __1:N Real-time Streaming__ : Supports one-to-many video streaming, enabling a single broadcaster to stream to multiple viewers simultaneously.
- __Real-time Chat__: Provides a chat feature integrated within the streaming session, allowing viewers to communicate in real-time.
- __Peer-assisted CDN__: Utilizes a Peer-assisted Content Delivery Network (CDN) to improve scalability and reduce bandwidth costs by allowing viewers to share video data among themselves.


# Milestone

## pdn server

### server logic
1. implement socketIO (m) 
2. refactor CMD logic (m)
3. introduce grpc signal-admin (L)
4. implement in-memory db (m)

### Auth 
1. Channel-key Authentication (m) 
2. Validate channel-key and channel-id (m)
3. set max-request-num in duration (m)

### Metric, Logging
1. Add prometheus component in pdn server (L) 
2. import prometheus (s)
3. Add business metric (s)
4. how much time consumed for some method or query (s)
5. how many users are in channel (s)
6. monitor server resource - cpu, memory.. (s)
7. Set prometheus component in admin server (L)
8. Add prometheus to docker-compose file (s)
9. Set grafana json (m)
10. Add Grafana to docker-compose file (s)
11. Add Logging db like loki (L)
12. channel and API key usage history (XL)

### WebRTC
1. test basic server -> peer1 -> peer2 (XL) 
2. Remove peer from channel (m) 
3. Peer health check (L) 
4. Super peer checking function (L)
5. Superpeer Candidate Management (m) 
6. Fault Recovery and Reconnection (XXL) 
6.1 handle super peer die (XL) 
6.2 Peer rebalancing (XL) 

### business
1. implement project and customer (L)
2. Set channel name, user name (m)
3. get userlist of a channel api (m)
4. channel list api (m)
5. evict user api (m)
6. user rename api (s)

### Storage 
1. select redis datastructure modeling (m)
2. make docker setting of redis (s)
3. Backup DB (L)
4. regular data transfer from Redis to the backup db (L)

### CI/CD
1. Add linter (m)
2. build pdn server docker image (s)
3. add docker compose file (s)
4. implement CI/CD pipeline (XL)
  - CI
    - Add Testing and linting in CI (s)
  - CD
    - dev server on homeserver (m)
    - run on AWS (L)
5. implement on k8s (XL)

### TEST
1. add unit test (XL) V
2. add benchmark (L)
3. massive user test (XXL)
4. check how many users can run on one server (XXL)
5. add k6 test (XL)
6. check test coverage (m)

### Document
1. Add design document of code architecture (s) V
2. whole architecture (s)
3. api document (m)

### SDK
1. Wrap JDK with an SDK (L)
2. release npm (s)

