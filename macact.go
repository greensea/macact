package main

import (
    "fmt"
    "regexp"
    "io/ioutil"
    "strings"
    "os"
    "time"
    "os/exec"
)

var IP string = ""
var COMMAND string = ""
var MAC string = ""
var IPCh = make(chan bool, 1)

var MAC2IPReg = regexp.MustCompile("[ ]+")

func main() {
    if len(os.Args) < 3 {
        fmt.Println("用法: <%s> <MAC 地址> <命令>");
        fmt.Println("命令中的 “%h” 会被替换成 MAC 对应的 IP 地址");
        os.Exit(0)
    }
    
    MAC = os.Args[1]
    COMMAND = os.Args[2]
    
    go MACWatcher()
    go CommandHandler()
    select{}
}

func MACWatcher() {
    for {
        CurrentIP := MAC2IP(MAC)

        if CurrentIP != IP {
            fmt.Printf("MAC (%s) 的 IP 地址变化了: '%s' -> '%s'\n", MAC, IP, CurrentIP)
            IP = CurrentIP
            IPCh <- true
        }
        
        time.Sleep(30 * time.Second)
    }
}

func CommandHandler() {
    select {
        case <- time.NewTicker(60 * time.Second).C:
            fmt.Printf("尚未查询到 %s 的 IP 地址\n", MAC)
        case <- IPCh:
            break
    }
    
    c := SpawnCommand()
    
    for {
        select {    
            case <- IPCh:
                /// 停止现有的进程，重新启动进程
                c.Process.Kill()
                c = SpawnCommand()
            case <- time.NewTicker(1 * time.Second).C:                
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
    for {
        fmt.Println(args)
        err := c.Start()
        if err != nil {
            fmt.Printf("启动命令失败，将重试: %s")
            time.Sleep(1 * time.Second)
            continue
        }
        go c.Wait()
        break;
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

