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
	bytesSent := 0
	for bytesSent < size{
		n, err:= c.conn.Write(buffer[bytesSent:size])
		if err != nil{
			return err
		}
		bytesSent = bytesSent + n
	}
	return nil
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