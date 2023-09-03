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
const ACTION_INFO_MSG_SIZE int = 5
const BET_CODE string = "B"


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

		
		for notEOF && (batchSize < c.config.MaxBatchSize){
			if errScan := scanner.Err(); errScan != nil {
				log.Errorf("action: parse_data | result: fail | error: %s", errScan)
				c.conn.close()
				return
			}
			
	
			betSplited := strings.Split(scanner.Text(), ",")  
			firstName := betSplited[0]
			lastName := betSplited[1]
			document, err := strconv.Atoi(betSplited[2])
			if err != nil {
				log.Errorf("action: read_document | result: fail | client_id: %v | document %s | error: %s",
					c.config.ID,
					betSplited[2],
					err,
				)
				c.conn.close()
				return
			}
			birthdate := betSplited[3]
			number, err := strconv.Atoi(betSplited[4])
			if err != nil {
				log.Errorf("action: read_number | result: fail | client_id: %v | number %s | error: %s",
					c.config.ID,
					betSplited[4],
					err,
				)
				c.conn.close()
				return
			}
	
	
			bet := NewBet(
				c.config.ID,
				firstName,
				lastName,
				int(document),
				birthdate,
				int(number),
			)
			betBuffer, bufferLen := bet.Serialize()

			totalBufferLen += bufferLen
			batchSize ++
			buffer = append(buffer, betBuffer...)

			notEOF = scanner.Scan()
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


		// Send action information message (message_code, batch size, agency_number)
		actionInfoBuffer := make([]byte, ACTION_INFO_MSG_SIZE)
		betCode := []byte(BET_CODE)
		copy(actionInfoBuffer, betCode)
		binary.BigEndian.PutUint16(actionInfoBuffer[1:], uint16(batchSize))
		binary.BigEndian.PutUint16(actionInfoBuffer[3:], uint16(c.config.ID))
		c.conn.send(actionInfoBuffer, ACTION_INFO_MSG_SIZE)	

		c.conn.send(buffer, totalBufferLen)	
		batchSize = 0
		totalBufferLen = 0
		buffer = make([]byte, 0)

		response_bytes, err := c.conn.receive(REPONSE_CODE_SIZE)
		response_code := int16(binary.BigEndian.Uint16(response_bytes))

		
		c.conn.close()

		if err != nil || response_code != RESPONSE_CODE_OK {
			log.Errorf("action: apuesta_enviada | result: fail | err: %s | response_code: %v", err, response_code)
			return
		}
		log.Infof("action: apuesta_enviada | result: success")
        

		if !notEOF{
			log.Infof("action: read_bets_file | result: success")
			return
		}

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

		notEOF = scanner.Scan()
	}


	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
