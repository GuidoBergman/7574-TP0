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
	"math"
	"fmt"

	log "github.com/sirupsen/logrus"
)

const REPONSE_CODE_SIZE int = 2
const COUNT_WINNERS_SIZE int = 2
const WINNER_SIZE int = 2
const RESPONSE_CODE_OK int16 = 0
const NO_DRAW_YET_CODE int16 = 1
const ACTION_INFO_MSG_SIZE int = 5
const BET_CODE string = "B"
const END_CODE string = "E"
const WINNERS_CODE string = "W"
const WINNER_WAIT_BASE float64 = 2
const WINNER_WAIT_MIN_EXPONENT float64 = 1
const WINNER_WAIT_MAX_EXPONENT float64 = 7


// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            int
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	DataPath	  string
	MaxBatchSize  int
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




func (c *Client) StartClientLoop() {
	err := c.conn.createClientSocket(c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return 
	}
	defer c.conn.close()


	sigterm := make(chan os.Signal, 1) 
	signal.Notify(sigterm, syscall.SIGTERM)

	f, err := os.Open(c.config.DataPath)
    if err != nil {
        log.Errorf("action: open_data_file | result: fail | error: %s",
			err,
		)
    }
	defer c.CloseFile(f)
	scanner := bufio.NewScanner(f)
	

loop:

	// Send messages if the loopLapse threshold has not been surpassed
	for ; ; {
		select {
		case <-sigterm:
	        log.Infof("action: sigterm_received | client_id: %v",
                c.config.ID,
            )
	        return
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
			break loop
		}
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	
	err = c.SendEnd()
	if err != nil{
		return
	}

	c.QueryWinners()
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
		// Send action information message (message_code, batch size, agency_number)
		actionInfoBuffer := make([]byte, ACTION_INFO_MSG_SIZE)
		betCode := []byte(BET_CODE)
		copy(actionInfoBuffer, betCode)
		binary.BigEndian.PutUint16(actionInfoBuffer[1:], uint16(c.config.ID))
		binary.BigEndian.PutUint16(actionInfoBuffer[3:], uint16(batchSize))
				
		err := c.conn.send(actionInfoBuffer, ACTION_INFO_MSG_SIZE)	
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | err: %s", err)
			return err
		}

		err = c.conn.send(buffer, totalBufferLen)	
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | err: %s", err)
			return err
		}

		responseBytes, err := c.conn.receive(REPONSE_CODE_SIZE)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | err: %s", err)
			return err
		}
		responseCode := int16(binary.BigEndian.Uint16(responseBytes))
		if responseCode != RESPONSE_CODE_OK{
			log.Errorf("action: apuesta_enviada | result: fail | response_code: %v", responseCode)
			return fmt.Errorf("action: apuesta_enviada | result: fail | response_code: %v", responseCode)
		}
		log.Infof("action: apuesta_enviada | result: success")
		
		return nil
}

func (c *Client) SendEnd() error{
	// Send action information message to indicate that the agency has sent all it's bets
	actionInfoBuffer := make([]byte, ACTION_INFO_MSG_SIZE)
	endCode := []byte(END_CODE)
	copy(actionInfoBuffer, endCode)
	binary.BigEndian.PutUint16(actionInfoBuffer[1:], uint16(c.config.ID))
	err := c.conn.send(actionInfoBuffer, ACTION_INFO_MSG_SIZE)	
	if err != nil {
		log.Errorf("action: send_end_msg | result: fail | error: %s", err)
		return err
	}
	log.Infof("action: send_end_msg | result: success")

	return nil
}

func (c *Client) QueryWinners() error{
	exponent := WINNER_WAIT_MIN_EXPONENT
	responseCode := NO_DRAW_YET_CODE

	for responseCode != RESPONSE_CODE_OK{
		// Send action information message to query the agency's winners
		log.Infof("action: consulta_ganadores | result: in progress")
		actionInfoBuffer := make([]byte, ACTION_INFO_MSG_SIZE)
		winCode := []byte(WINNERS_CODE)
		copy(actionInfoBuffer, winCode)
		binary.BigEndian.PutUint16(actionInfoBuffer[1:], uint16(c.config.ID))
		err := c.conn.send(actionInfoBuffer, ACTION_INFO_MSG_SIZE)	
		if err != nil{
			log.Errorf("action: consulta_ganadores | stage: enviar la consulta | result: fail | error: %s", err)
			return err
		}

		responseBytes, err:= c.conn.receive(REPONSE_CODE_SIZE)
		if err != nil{
			log.Errorf("action: consulta_ganadores | stage: recibir la respuesta | result: fail | error: %s", err)
			return err
		}
		responseCode = int16(binary.BigEndian.Uint16(responseBytes))

		if responseCode == NO_DRAW_YET_CODE{
			waitTime := math.Pow(WINNER_WAIT_BASE, exponent)
			log.Infof("action: consulta_ganadores | result: aún no se hizo el sorteo")
			time.Sleep(time.Duration(waitTime) * time.Second)
			if exponent < WINNER_WAIT_MAX_EXPONENT{
				exponent ++
			}
		}
		
	}

	responseBytes, err := c.conn.receive(COUNT_WINNERS_SIZE)
	if err != nil{
		log.Errorf("action: consulta_ganadores | stage: recibir la respuesta | result: fail | error: %s", err)
		return err
	}
	countWinners := binary.BigEndian.Uint16(responseBytes)

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v.", countWinners)

	var i uint16 = 0
	for ;i  < countWinners; i++ {
		responseBytes, err := c.conn.receive(WINNER_SIZE)
		if err != nil{
			log.Errorf("action: consulta_ganadores | stage: recibir la respuesta | result: fail | error: %s", err)
			return err
		}
		winner := binary.BigEndian.Uint16(responseBytes)
		log.Infof("action: consulta_ganadores | result: success | Número ganador: %v.", winner)
	}

	return nil
}


func (c *Client) CloseFile(f *os.File){
	f.Close()
	log.Infof("action: close_data_file | result: success")
}