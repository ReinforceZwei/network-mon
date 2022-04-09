package ssh

import "golang.org/x/crypto/ssh"

type SshConnection struct {
	client *ssh.Client
}

func Connect(host, user, password string) (*SshConnection, error) {
	client, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}
	return &SshConnection{
		client: client,
	}, nil
}

func (c *SshConnection) Execute(command string) error {
	s, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	return s.Run(command)
}
