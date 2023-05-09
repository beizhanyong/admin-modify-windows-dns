package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}

func main() {
	// 检查操作系统类型
	if strings.ToUpper(os.Getenv("PROCESSOR_ARCHITECTURE")) != "AMD64" {
		fmt.Println("请在 64 位操作系统上运行此程序")
		os.Exit(1)
	}

	// 修改DNS服务器
	dnsServers := []string{"8.8.8.8", "4.4.4.4"}
	err := SetDNS(dnsServers)
	if err != nil {
		os.Exit(1)
	}

	// 刷新DNS缓存
	err = FlushDNS()
	if err != nil {
		os.Exit(1)
	}
	time.Sleep(time.Second * 5)
}

// SetDNS 设置DNS服务器
func SetDNS(servers []string) error {

	for _, conn := range GetActiveConnections() {
		var cmd *exec.Cmd
		var dns_name string
		for k, server := range servers {

			if k == 0 {
				cmd = exec.Command("netsh", "interface", "ip", "set", "dns", conn, "static", server, "primary")
				dns_name = "主"
			} else {
				cmd = exec.Command("netsh", "interface", "ip", "add", "dns", conn, server)
				dns_name = "备"
			}

			err := cmd.Run()

			if err != nil {
				fmt.Printf("%s 的%sDNS服务器设置失败：%s\n", conn, dns_name, err)
				return err
			}
			fmt.Printf("%s 的%sDNS服务器已设置为：%s\n", conn, dns_name, server)
		}

	}
	fmt.Println("DNS服务器设置成功")
	return nil
}

// FlushDNS 刷新DNS缓存
func FlushDNS() error {
	cmd := exec.Command("ipconfig", "/flushdns")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("刷新dns失败：%s\n", err)
		return err
	}
	fmt.Println("DNS缓存已刷新")
	return nil
}

// GetActiveConnections 获取当前活动的网络连接名称
func GetActiveConnections() []string {
	var connections []string
	cmd := exec.Command("netsh", "interface", "show", "interface")
	output, err := cmd.Output()
	if err != nil {
		return connections
	}

	cmdRe := ConvertByte2String(output, "GB18030")

	lines := strings.Split(string(cmdRe), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Connected") || strings.Contains(line, "已连接") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				connections = append(connections, fields[3])
			}
		}
	}
	return connections
}
