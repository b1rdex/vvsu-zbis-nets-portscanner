package main

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	ip := prompt("Target ip", "127.0.0.1")
	from, _ := strconv.Atoi(prompt("Start port", "1024"))
	to, _ := strconv.Atoi(prompt("End port", "65535"))
	connections, _ := strconv.ParseInt(prompt("Connections", "256"), 10, 64)
	ps := &PortScanner{
		ip:   ip,
		lock: semaphore.NewWeighted(connections),
	}
	ps.Scan(from, to, 500*time.Millisecond)
}

func prompt(question string, def string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question + " [" + def + "]: ")
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\r\n", "", -1)
	text = strings.Replace(text, "\n", "", -1)
	if text == "" {
		return def
	}
	return text
}

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func ScanPort(ip string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		} else {
			//fmt.Println(port, "closed")
		}
		return
	}

	conn.Close()
	fmt.Println(port, "open")
}

func (ps *PortScanner) Scan(f, l int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for port := f; port <= l; port++ {
		wg.Add(1)
		ps.lock.Acquire(context.TODO(), 1)
		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}
