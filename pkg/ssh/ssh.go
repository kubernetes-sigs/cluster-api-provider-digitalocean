/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ssh

import (
	"fmt"
	"io/ioutil"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	gossh "golang.org/x/crypto/ssh"
)

// Client contains parameters for connection to the node.
type Client struct {
	// IP address or FQDN of the node.
	Address string

	// Port of the node's SSH server.
	Port string

	// ClientConfig is a basic Go SSH client needed to make SSH connection.
	// This is populated automatically from fields provided on Client creation time.
	ClientConfig *gossh.ClientConfig

	// Conn is connection to the remote SSH server.
	// Connection is made using the Connect function.
	Conn *gossh.Client
}

// NewClient returns a SSH client representation.
// TODO: This assumes the SSH key doesn't have password! This is same as for other upstream providers.
func NewClient(address, port, username, privateKeyPath string) (*Client, error) {
	pk, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	key := []byte(pk)
	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	s := &Client{
		Address: address,
		Port:    port,
		ClientConfig: &gossh.ClientConfig{
			User: username,
			Auth: []gossh.AuthMethod{
				gossh.PublicKeys(signer),
			},
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		},
		Conn: nil,
	}

	return s, nil
}

// Connect starts a headless connection against the node.
func (s *Client) Connect() error {
	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%s", s.Address, s.Port), s.ClientConfig)
	if err != nil {
		return err
	}

	s.Conn = conn
	return nil
}

// Execute executes command on the remote server and returns stdout and stderr output.
func (s *Client) Execute(cmd string) ([]byte, error) {
	if s.Conn == nil {
		return nil, fmt.Errorf("not connected to the server")
	}

	// Start interactive session.
	session, err := s.Conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := session.Close(); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to close ssh session: %v", err))
		}
	}()

	return session.CombinedOutput(cmd)
}

// Close closes the SSH connection.
func (s *Client) Close() error {
	if s.Conn == nil {
		return fmt.Errorf("connection not existing")
	}
	return s.Conn.Close()
}
