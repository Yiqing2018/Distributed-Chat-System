package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	Addrs_flag = [10]bool{
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
	}

	Addrs = [10]string{
		"172.22.158.45",
		"172.22.94.54",
		"172.22.156.46",
		"172.22.158.46",
		"172.22.94.55",
		"172.22.156.47",
		"172.22.158.47",
		"172.22.94.56",
		"172.22.156.48",
		"172.22.158.48",
	}

	dial_conns = make(map[string]net.Conn)

	IP2Username = make(map[string]string)

	pNumber = 0

	localStamp []int

	holdback = make(map[string]string)
	stayLong = make(map[string]string)
)

// Helper function: get the IP address of the Server
func getServerAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("(getServerAddr) Failed to get the IP address: %v", err)
	}

	var return_ip string
	for _, address := range addrs {
		// Remove the Loopback Address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return_ip = ipnet.IP.String()
			}
		}
	}

	return return_ip
}

// Dial to other Servers
func Dial2Servers(port string, n int) {
	ip_server := getServerAddr()
	// fmt.Println("(Dial2Servers) The Local IP is: " + ip_server)
	// dial_conns := make(map[string]net.Conn)

	count := 0
	for {
		for index, ip_value := range Addrs {
			if Addrs_flag[index] == true {
				continue
			}

			if ip_value == ip_server {
				continue
			}

			dial_addr := ip_value + ":" + port
			dial_conn, err := net.Dial("tcp", dial_addr)
			if err == nil {
				Addrs_flag[index] = true
				count = count + 1
				dial_conns[dial_conn.RemoteAddr().String()] = dial_conn
				go Handler(dial_conn, &dial_conns, n)
				// fmt.Println("(Dial2Servers)" + ip_server + " Dial to IP: " + dial_addr + "--OK")
			}
		} // address loop
		if count == n-1 {
			//know all other servers' IP
			//go through Addrs_flag to know my own pNumber
			//pNumber := 0
			for i, val := range Addrs_flag {
				if Addrs[i] == ip_server {
					break
				}
				if val {
					//fmt.Println(Addrs[i])
					pNumber = pNumber + 1
				}
			}
			//fmt.Println("pNumber: ", pNumber)

			//initialize local_stamp
			localStamp = InitTimestamp(n, pNumber)
			//fmt.Println("# initialize local stamp: ", localStamp)
			go releaseHoldback(holdback, stayLong)

			break
		}
	} // dead loop
	// fmt.Println("(Dial2Servers)" + ip_server + " Dial to ALL IP" + "--OK")
	// fmt.Println("------")
	fmt.Println("READY")
	// fmt.Println("------")

}

func StartServer(port string, username string, n int) {
	// fmt.Println("------Where Amazing Happens------")

	/**
	 * Check the connection between Local Server and other Servers:
	 * 1. The Local Server should dial to the other Servers;
	 * 2. The other Servers will dial to the Local Server, and the Local Server should confirm this connection.
	 */

	// 1. Dial to other Servers
	go Dial2Servers(port, n)

	// 2. Check the connection with other Servers
	host := ":" + port
	// Setup the TCP Server
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		fmt.Println("(StartServer) ResolveTCPAddr Failed" + "--SAD")
		return
	}
	// Monitor the Port
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("(StartServer) ListenTCP Failed" + "--SAD")
		return
	}
	// Build the Connection Pool (the other Servers)
	conns := make(map[string]net.Conn)
	// Build the Message Channels (with the other Servers)
	//messageChan := make(chan string, 10)

	// Confirm the connections with other servers, and build the connections
	go BroadUsernames(&conns, username, n-1)
	go BroadMessages(&conns, username)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("(StartServer) Accept Failed" + "--SAD")
			continue
		}

		conns[conn.RemoteAddr().String()] = conn
		// fmt.Println(conns)
	}
}

func BroadMessages(conns *map[string]net.Conn, username string) {

	for {
		// fmt.Println("(BroadMessages) Please Write Your Msg:")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = input[:len(input)-1]

		if len(input) > 0 {
			realIput := "[" + username + "]: " + input
			// add local timestamp to message you want to deliver, formated like this:
			// 0,0,0,1#[username]: hello world
			msg := addTimestamp(localStamp, realIput)

			for key, conn := range *conns {
				//fmt.Println("connection is connected from ", key)
				_, err := conn.Write([]byte(msg))
				if err != nil {
					//fmt.Println("(BroadMessages) broad message to %s failed: %v\n", key, err)
					delete(*conns, key)
				}
			}
		} // if
	} // infinite for loop
}

func Handler(conn net.Conn, conns *map[string]net.Conn, n int) {
	// fmt.Println("(Handler) Connect from the Server ", conn.RemoteAddr().String())
	buf := make([]byte, 1024)

	length, err := conn.Read(buf)
	if err != nil {
		fmt.Println("(Handler) Read Client Username Failed--SAD")
		delete(*conns, conn.RemoteAddr().String())
		conn.Close()
	}
	recvStr := string(buf[0:length])
	// fmt.Println(recvStr)
	// add the username to the Server's Dictionary
	ip_a := strings.Split(conn.RemoteAddr().String(), ":")[0]
	IP2Username[ip_a] = recvStr
	for {
		length, err := conn.Read(buf)
		if err != nil {
			ip_left := strings.Split(conn.RemoteAddr().String(), ":")[0]
			username_left := IP2Username[ip_left]
			//fmt.Println("(Handler) Read Client Message Failed--SAD")
			fmt.Println(username_left + " has left")
			delete(*conns, conn.RemoteAddr().String())
			conn.Close()
			break
		}

		recvStr := string(buf[0:length])
		//recive a message, decide when to deliver it!
		//fmt.Println("received msg :", recvStr)

		handleMsg(recvStr, localStamp, n, holdback, stayLong)

	} // dead loop

}

func releaseHoldback(holdback map[string]string, stayLong map[string]string) {
	for {
		time.Sleep(1 * time.Second)
		layout := "2006-01-02 15:04:05"
		currentT := time.Now()
		for key, value := range holdback {
			whenEntered, err := time.Parse(layout, stayLong[key])
			if err != nil {
				fmt.Println("cannot parse time")
			}
			diff := currentT.Sub(whenEntered)
			dura := int64(diff / time.Second)
			if dura >= 2 {
				fmt.Println("#stay too long")
				_, ok1 := stayLong[key]
				if ok1 {
					delete(stayLong, key)
				}

				_, ok2 := holdback[key]
				if ok2 {
					delete(holdback, key)
				}

				fmt.Println(value)
			}

		}

	}

}

func BroadUsernames(conns *map[string]net.Conn, username string, ip_num int) {

	for {
		msg := username
		// fmt.Println(len(*conns))
		if len(*conns) == ip_num {
			// fmt.Println(len(*conns))
			for key, conn := range *conns {
				_, err := conn.Write([]byte(msg))
				if err != nil {
					fmt.Println("(BroadUsernames) broad Username to %s failed: %v\n", key, err)
					// delete(*conns, key)
				}
			}
			break
		} // if

		time.Sleep(1 * time.Second)
	} //dead loop

}
