package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"strings"
	"sync"
)

func sshCommand(hostname, port, username, password string, commands []string) {
	// 配置 SSH 客户端
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 不验证主机密钥（仅用于测试）
	}
	log.Printf("starting %v >>>>", hostname)
	// 连接到 SSH 服务器
	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", hostname, port), config)
	if err != nil {
		fmt.Printf("Failed to dial: %s\n", err)
		return
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create session: %s\n", err)
		return
	}
	defer session.Close()

	// 执行命令
	for _, command := range commands {
		// 创建新的会话
		session, err := client.NewSession()
		if err != nil {
			fmt.Println("Failed to create session:", err)
			return
		}
		defer session.Close()
		output, err := session.Output(command)
		if err != nil {
			fmt.Printf("Failed to run command '%s': %s\n", command, err)
			return
		}
		fmt.Printf("Output from command '%s':\n%s\n", command, output)
	}
	//fmt.Println(strings.Repeat("=", 40))
	fmt.Println("End of output")
}

func readSSHInfo(filePath string) ([][4]string, error) {
	var sshInfo [][4]string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			parts := strings.Split(line, "----")
			if len(parts) == 4 {
				sshInfo = append(sshInfo, [4]string{parts[0], parts[1], parts[2], parts[3]})
			} else {
				fmt.Printf("Invalid line in file: %s\n", line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sshInfo, nil
}

// readLines 读取指定文件的内容，并返回每行的切片
func readLines(filePath string) ([]string, error) {
	var lines []string

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close() // 确保文件在最后关闭

	// 创建一个扫描器逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()      // 获取当前行的内容
		lines = append(lines, line) // 将行添加到切片中
	}

	// 检查扫描过程中是否出现错误
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func main() {
	// SSH 信息文件路径
	sshInfoFile := "ssh_info.txt"
	commands, err := readLines("command.txt")
	if err != nil {
		fmt.Printf("Error reading commands: %s\n", err)
		return
	}

	// 读取 SSH 信息
	sshInfoList, err := readSSHInfo(sshInfoFile)
	if err != nil {
		fmt.Printf("Error reading SSH info: %s\n", err)
		return
	}

	// 批量登录并执行命令
	var wg sync.WaitGroup
	for _, info := range sshInfoList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sshCommand(info[0], info[1], info[2], info[3], commands)
		}()
	}
	wg.Wait()
}
