package common

import (
	"time"
	"os"
    "os/signal"
	"syscall"
	"encoding/binary"
	"bufio"
	"strings"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const REPONSE_CODE_SIZE int = 2
const RESPONSE_CODE_OK int16 = 0
const ACTION_INFO_MSG_LEN int = 2


// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            int
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	DataPath	  string
	MaxBatchSize  int
}

// Client Entity that encapsulates how
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

	f, err := os.Open(c.config.DataPath)
	defer f.Close()
    if err != nil {
        log.Fatalf("action: open_data_file | result: fail | error: %s",
			err,
		)
    }
	scanner := bufio.NewScanner(f)
	notEOF := scanner.Scan()

	batchSize := 0
	totalBufferLen := 0
	buffer := make([]byte, 0)


loop:

	// Send messages if the loopLapse threshold has not been surpassed
	for timeout := time.After(c.config.LoopLapse); ; {
		select {
		case <-timeout:
	        log.Infof("action: timeout_detected | result: success | client_id: %v",
                c.config.ID,
            )
			c.conn.close()
			break loop
		case <-sigterm:
	        log.Infof("action: sigterm_received | client_id: %v",
                c.config.ID,
            )
			c.conn.close()
	        log.Infof("action: connection_closed | client_id: %v",
                c.config.ID,
            )
			break loop
		default:
		}

		// Create the connection the server in every loop iteration. Send an
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
		response_code := int16(binary.BigEndian.Uint16(response_bytes))

		
		c.conn.close()

		if err != nil || response_code != RESPONSE_CODE_OK {
			log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v",
			c.config.Document,
            c.config.Number,
        	)
			return
		}
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			c.config.Document,
            c.config.Number,
        )

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}


	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
