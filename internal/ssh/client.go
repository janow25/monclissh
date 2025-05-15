package ssh

import (
    "golang.org/x/crypto/ssh"
    "time"
)

type SSHClient struct {
    Client *ssh.Client
    Config *ssh.ClientConfig
    Hostname string
}

func NewSSHClient(hostname string, user string, password string) (*SSHClient, error) {
    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        Timeout:         5 * time.Second,
    }

    client, err := ssh.Dial("tcp", hostname, config)
    if err != nil {
        return nil, err
    }

    return &SSHClient{
        Client: client,
        Config: config,
        Hostname: hostname,
    }, nil
}

func (s *SSHClient) ExecuteCommand(command string) (string, error) {
    session, err := s.Client.NewSession()
    if err != nil {
        return "", err
    }
    defer session.Close()

    var output []byte
    output, err = session.Output(command)
    if err != nil {
        return "", err
    }

    return string(output), nil
}

func (s *SSHClient) Close() error {
    return s.Client.Close()
}