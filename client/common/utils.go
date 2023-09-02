package common

import (
	"encoding/binary"
)

type Bet struct {
	Agency		  int
	FirstName     string
	LastName	  string
	Document      int
	Birthdate     string
	Number        int	
}

const TOTAL_LENGTH int = 128
const BET_CODE string = "B"

func NewBet(agency int, firstName string, lastName string, document int, birthdate string, number int) *Bet {
        /*
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        */
		b := new(Bet)
			
		b.Agency = agency
		b.FirstName = firstName
		b.LastName = lastName
		b.Document = document
		b.Birthdate = birthdate
		b.Number = number
		
		return b
}

func (b *Bet) Serialize() ([]byte, int) {
	buffer := make([]byte, TOTAL_LENGTH)
	bet_code := []byte(BET_CODE)
	copy(buffer, bet_code)

	binary.BigEndian.PutUint16(buffer[1:], uint16(b.Agency))

	binary.BigEndian.PutUint16(buffer[3:], uint16(len(b.FirstName)))
	first_name := []byte(b.FirstName)
	copy(buffer[5:],first_name)

	binary.BigEndian.PutUint16(buffer[58:],  uint16(len(b.LastName)))
	last_name := []byte(b.LastName)
	copy(buffer[60:], last_name)


	binary.BigEndian.PutUint32(buffer[112:],  uint32(b.Document))

	birthdate := []byte(b.Birthdate)
	copy(buffer[116:], birthdate)

	binary.BigEndian.PutUint16(buffer[126:],  uint16(b.Number))

	return buffer, TOTAL_LENGTH
}