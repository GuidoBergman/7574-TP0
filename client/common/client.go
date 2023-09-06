package common

import (
	"time"
	"os"
    "os/signal"
	"syscall"
	"encoding/binary"

	log "github.com/sirupsen/logrus"
)

const REPONSE_CODE_SIZE int = 2
const RESPONSE_CODE_OK int16 = 0

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            int
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	FirstName     string
	LastName	  string
	Document      int
	Birthdate     string
	Number        int	
}


type Client struct {
	config ClientConfig
	conn   ClientSocket
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}



// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// autoincremental msgID to identify every message sent

	sigterm := make(chan os.Signal, 1) 
	signal.Notify(sigterm, syscall.SIGTERM)


	select {
	case <-sigterm:
	    log.Infof("action: sigterm_received | client_id: %v",
            c.config.ID,
        )
		c.conn.close()
		return
	default:
	}
	// Create the connection the server 
	err := c.conn.createClientSocket(c.config.ServerAddress)
	if err != nil {
		log.Errorf(
	    	"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	bet := NewBet(
		c.config.ID,
		c.config.FirstName,
		c.config.LastName,
		c.config.Document,
		c.config.Birthdate,
		c.config.Number,
	)
	buffer, buffer_len := bet.Serialize()
	c.conn.send(buffer, buffer_len)	
	response_bytes, err := c.conn.receive(REPONSE_CODE_SIZE)
	
	
	
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v",
		c.config.Document,
        c.config.Number,
    	)
		return
	}
	response_code := int16(binary.BigEndian.Uint16(response_bytes))
	if response_code != RESPONSE_CODE_OK{
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | response_code: %v",
		c.config.Document,
        c.config.Number,
		response_code,
    	)
		return
	}
	
	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		c.config.Document,
        c.config.Number,
    )

	c.conn.close()
}
