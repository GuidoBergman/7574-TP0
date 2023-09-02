package common

import (
	"bufio"
	"net"
	"io"
)

type ClientSocket struct {
	conn   net.Conn
}

// CreateClientSocket Initializes client socket. In case of
// failure exit 1 is returned
func (c *ClientSocket) createClientSocket(serverAddress string) error {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *ClientSocket) send(buffer []byte, size int) error {
	w := bufio.NewWriter(c.conn)
	bytes_sent := 0
	for {
		n, _:= w.Write(buffer[bytes_sent:size])
		bytes_sent := bytes_sent + n
		if (bytes_sent < size) {
			break
		}
	}
	return w.Flush() 
}

func (c *ClientSocket) receive(size int) ([]byte, error) {
	buffer := make([]byte, size)
	reader := bufio.NewReader(c.conn)
	_, err := io.ReadFull(reader, buffer)
	return buffer, err
}

func (c *ClientSocket) close() {
	c.conn.Close()
}