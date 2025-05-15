package ssh

import (
	"monclissh/internal/config"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	Client   *ssh.Client
	Config   *ssh.ClientConfig
	Hostname string
}

func NewSSHClient(hostname, user, password string) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	client, err := ssh.Dial("tcp", hostname, config)
	if err != nil {
		return nil, err
	}
	return &SSHClient{Client: client, Config: config, Hostname: hostname}, nil
}

func NewSSHClientFromConfig(hostConfig config.Host) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User:            hostConfig.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(hostConfig.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	client, err := ssh.Dial("tcp", hostConfig.Hostname, config)
	if err != nil {
		return nil, err
	}
	return &SSHClient{Client: client, Config: config, Hostname: hostConfig.Hostname}, nil
}

func (s *SSHClient) ExecuteCommand(command string) (string, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.Output(command)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (s *SSHClient) Close() error {
	return s.Client.Close()
}
