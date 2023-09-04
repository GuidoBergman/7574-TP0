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
	

loop:

	// Send messages if the loopLapse threshold has not been surpassed
	for timeout := time.After(c.config.LoopLapse); ; {
		select {
		case <-timeout:
	        log.Infof("action: timeout_detected | result: success | client_id: %v",
                c.config.ID,
            )
			break loop
		case <-sigterm:
	        log.Infof("action: sigterm_received | client_id: %v",
                c.config.ID,
            )
	        log.Infof("action: connection_closed | client_id: %v",
                c.config.ID,
            )
			break loop
		default:
		}

		
		buffer, totalBufferLen, batchSize, notEOF, errType, err := c.CreateBatch(scanner)
		if err != nil {
			log.Errorf("action: %s | result: fail | error: %s", errType, err)
			return
		}

		err = c.SendBatch(buffer, totalBufferLen, batchSize)
		if err != nil {
			return
		}
        

		if !notEOF{
			log.Infof("action: read_bets_file | result: success")
			return
		}

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}


func (c *Client) CreateBatch(scanner *bufio.Scanner) ([]byte, int, int, bool, string, error){
	notEOF := scanner.Scan()

	batchSize := 0
	totalBufferLen := 0
	buffer := make([]byte, 0)

	for notEOF && (batchSize < c.config.MaxBatchSize){
		if errScan := scanner.Err(); errScan != nil {
			return nil, 0, 0, false, "scanning_file", errScan
		}		

		betSplited := strings.Split(scanner.Text(), ",")  
		firstName := betSplited[0]
		lastName := betSplited[1]
		document, err := strconv.Atoi(betSplited[2])
		if err != nil {
			return nil, 0, 0, false, "parse_document", err
		}
		birthdate := betSplited[3]
		number, err := strconv.Atoi(betSplited[4])
		if err != nil {
			return nil, 0, 0, false, "parse_number", err
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

	return buffer, totalBufferLen, batchSize, notEOF, "", nil
}

func (c *Client) SendBatch(buffer []byte, totalBufferLen int, batchSize int) error{
		// Create the connection the server in every loop iteration. Send an
		err := c.conn.createClientSocket(c.config.ServerAddress)
		if err != nil {
			log.Errorf(
	        	"action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
		defer c.conn.close()


		// Send action information message (message_code, batch size, agency_number)
		actionInfoBuffer := make([]byte, ACTION_INFO_MSG_SIZE)
		betCode := []byte(BET_CODE)
		copy(actionInfoBuffer, betCode)
		binary.BigEndian.PutUint16(actionInfoBuffer[1:], uint16(batchSize))
		binary.BigEndian.PutUint16(actionInfoBuffer[3:], uint16(c.config.ID))
		c.conn.send(actionInfoBuffer, ACTION_INFO_MSG_SIZE)	

		c.conn.send(buffer, totalBufferLen)	

		response_bytes, err := c.conn.receive(REPONSE_CODE_SIZE)
		response_code := int16(binary.BigEndian.Uint16(response_bytes))
		
		if err != nil || response_code != RESPONSE_CODE_OK {
			log.Errorf("action: apuesta_enviada | result: fail | err: %s | response_code: %v", err, response_code)
			return err
		}
		log.Infof("action: apuesta_enviada | result: success")
		
		return nil
}