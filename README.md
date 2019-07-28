# Distributed Group Chat System

## Overview
This is a [course project](https://courses.engr.illinois.edu/ece428/sp2019/mps/mp1.html) of CS425 in UIUC. 

A collection of nodes will each run a program that communicats with other nodes' programs over the network.  
The addresses of the chat participants are **hard-coded** in this project, so you may **NOT** run this program on your local environment, we suggest you modify the hosts (*var Addrs* in *server.go*) to be a list of your own hostnames.
## Running Instruction
```
$go build ./  
$./distributedChatystem [username] [port] [number]  
```

* **username** represents the name of the chat participant
* **port** represents the port number to listen to for new connections
* **number** represents the number of people in the group

***

## Design

<p align="center">

  <img width="600" src="https://ws3.sinaimg.cn/large/006tNc79ly1g04eoi1yigj30qo0k0jsc.jpg">
  
</p>
### Marshalled Message Format
messages over the network follow this format:  
```
"[timeStamp]#[message typed from user]"
```
### Reliable Delivery

**Reliable TCP:** In this project, direct TCP connections are established between hosts, which can provide reliable communication via IP network. And also, once the connection was broken, the alive paricipant could immediately "monitor" that. -- This mechanism works like heartbeat protocol. 

**B-multicast:** Sending message is implemented via B-multicast, the one-to-all transmission among all clients.   


### Causal ordering  
Using vector timestamp algorithm to implement causal ordering. To avoid the blocked holdback problem (if one or some nodes failed), we check and update the local holdback queue periodically. 
![](http://ww4.sinaimg.cn/large/006tNc79ly1g5fngq0gk6j30q60hidjb.jpg)

* whenever a node receives a message, it deals with the timestamp received along with the input message from other nodes.
* whenever a node received a message:   
  with **"correct"** timestamp, the local timestamp would be updated.  
  with **"more updated"** timestamp, the message would be stored locally in a holdback queue, until the "correct" timestamp arrives!
* what if any nodes **FAILED**? the holdback queue would be blocked...   
we check the holdback queue every 1 second, if msg has already stayed for more than 2s, we will release the msg at once.


***

## Demo  
Until All participants entered the room, prints out "READY"
![](https://ws1.sinaimg.cn/large/006tNc79ly1g04ermqgm2j323m0jw7j2.jpg)
All participants would receive the message in Causal Order
![](https://ws1.sinaimg.cn/large/006tNc79ly1g04esvut0sj31gk0u0txo.jpg)
![](https://ws1.sinaimg.cn/large/006tNc79ly1g04et29kmuj31r20k2qjq.jpg)
Once someone left the room, prints out xx has left. And the rest of users can still continue the chat
![](https://ws2.sinaimg.cn/large/006tNc79ly1g04et7q6taj31en0u0e7h.jpg)
 



