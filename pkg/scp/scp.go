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

package scp

import (
	"errors"
	"fmt"
	"io/ioutil"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"sigs.k8s.io/cluster-api-provider-digitalocean/pkg/ssh"

	"github.com/pkg/sftp"
)

// Client represents SCP client.
type Client struct {
	client *ssh.Client
}

func NewSCPClient(sshClient *ssh.Client) *Client {
	return &Client{
		client: sshClient,
	}
}

// ReadBytes reads from remote location.
func (cl *Client) ReadBytes(remotePath string) ([]byte, error) {
	if cl.client.Conn == nil {
		return nil, errors.New("connection not established")
	}

	c, err := sftp.NewClient(cl.client.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := c.Close(); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to close ssh connection: %v", err))
		}
	}()

	r, err := c.Open(remotePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to close ssh connection: %v", err))
		}
	}()

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// WriteBytes writes to remote location.
func (cl *Client) WriteBytes(remotePath string, content []byte) error {
	if cl.client.Conn == nil {
		return errors.New("connection not established")
	}

	c, err := sftp.NewClient(cl.client.Conn)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Close(); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to close ssh connection: %v", err))
		}
	}()

	f, err := c.Create(remotePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to close ssh connection: %v", err))
		}
	}()

	_, err = f.Write(content)
	if err != nil {
		return err
	}

	return nil
}
