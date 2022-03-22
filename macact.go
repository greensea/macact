package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"
)

var IP string = ""
var COMMAND string = ""
var MAC string = ""
var IPCh = make(chan bool, 1)

var MAC2IPReg = regexp.MustCompile("[ ]+")

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	if len(os.Args) < 3 {
		fmt.Println("用法: <%s> <MAC 地址> <命令>")
		fmt.Println("命令中的 “%h” 会被替换成 MAC 对应的 IP 地址")
		os.Exit(0)
	}

	MAC = os.Args[1]
	COMMAND = os.Args[2]

	go MACWatcher()
	go CommandHandler()
	select {}
}

func MACWatcher() {
	for {
		CurrentIP := MAC2IP(MAC)

		if CurrentIP != IP {
			fmt.Printf("当前 MAC (%s) 的 IP 地址是: '%s' -> '%s'\n", MAC, IP, CurrentIP)
			IP = CurrentIP
			IPCh <- true
		}

		time.Sleep(30 * time.Second)
	}
}

func CommandHandler() {
	select {
	case <-time.NewTicker(60 * time.Second).C:
		fmt.Printf("尚未查询到 %s 的 IP 地址\n", MAC)
	case <-IPCh:
		break
	}

	c := SpawnCommand()

	t2 := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-IPCh:
			/// 停止现有的进程，重新启动进程
			fmt.Println("IP 地址改变了，杀死现有进程")
			c.Process.Kill()
			c = SpawnCommand()
		case <-t2.C:
			// fmt.Println("Tick 了一下")
			if c.ProcessState != nil {
				if c.ProcessState.Exited() == true {
					c = SpawnCommand()
				}
			}
		}
	}

}

func SpawnCommand() *exec.Cmd {
	rawcmd := strings.ReplaceAll(COMMAND, "%h", IP)
	var args []string
	for _, v := range strings.Split(rawcmd, " ") {
		if v == "" {
			continue
		}
		args = append(args, v)
	}

	c := exec.Command(args[0], args[1:]...)
	stderr, err := c.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	for {
		fmt.Println(args)
		err := c.Start()
		if err != nil {
			fmt.Printf("启动命令失败，将重试")
			time.Sleep(1 * time.Second)
			continue
		}

		go io.Copy(os.Stderr, stderr)
		go io.Copy(os.Stdout, stdout)
		go c.Wait()
		break
	}

	return c
}

func MAC2IP(mac string) string {
	raw, err := ioutil.ReadFile("/proc/net/arp")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(raw), "\n")

	for _, line := range lines {
		tokens := MAC2IPReg.Split(line, -1)
		if len(tokens) < 3 {
			continue
		}
		ip := tokens[0]
		MAC := tokens[3]

		if MAC == mac {
			return ip
		}
	}

	return ""
}
