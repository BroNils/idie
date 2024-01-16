package requester

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

type Requester struct {
	//public
	TimeOut time.Duration
}

type Option func(*Requester)

const (
	// timeout is the time to wait for a response.
	pingTimeOut = 5 * time.Minute
)

func NewRequester(fields ...Option) *Requester {
	r := &Requester{
		TimeOut: pingTimeOut,
	}

	for _, field := range fields {
		field(r)
	}

	return r
}

func WithTimeOut(timeOut time.Duration) Option {
	return func(r *Requester) {
		r.TimeOut = timeOut
	}
}

// ping ip address with defined port is open or not
// best for local network
func (r *Requester) PingTcp(ip string, port int) bool {
	address := net.JoinHostPort(ip, strconv.Itoa(port)) //fmt.Sprintf("%s:%d", ip, port)

	conn, err := net.DialTimeout("tcp", address, pingTimeOut)
	if err != nil {
		//r.FileLogger.Printf("Error connecting to %s: %s\n", address, err)
		return false
	}
	if conn == nil {
		return false
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			//r.FileLogger.Printf("Error closing connection to %s: %s\n", address, err)
		}
	}(conn)

	//r.FileLogger.Printf("%s is open\n", address)
	return true
}

// ping syn
func (r *Requester) NmapSyn(ip string, port int) (ipRet string, portRet int, isOpen bool, protocol string, serviceName string) {
	if r.TimeOut < 5*time.Minute {
		r.TimeOut = pingTimeOut
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.TimeOut)
	defer cancel()

	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 -sS ipv4`,
	// with a 5-minute timeout.
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithSYNScan(),
		nmap.WithTargets(ip),
		nmap.WithPorts(strconv.Itoa(port)),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to create nmap scanner: %v", err))
	}

	result, warnings, err := scanner.Run()
	if len(*warnings) > 0 {
		// log.Printf("run finished with warnings: %s\n", *warnings) // Warnings are non-critical errors from nmap.
	}
	if err != nil {
		panic(fmt.Sprintf("unable to run nmap scan: %v", err))
	}

	ipRet = ip
	portRet = port
	isOpen = false
	protocol = ""
	serviceName = ""

	if len(result.Hosts) < 1 {
		return
	}

	host := result.Hosts[0]
	if len(host.Ports) == 0 || len(host.Addresses) == 0 {
		return
	}

	scanResult := host.Ports[0]
	if scanResult.State.String() == "open" {
		isOpen = true
	}
	protocol = scanResult.Protocol
	serviceName = scanResult.Service.Name

	return
}
