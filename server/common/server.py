import logging
import signal
from common.common_socket import CommonSocket, STATUS_ERR, STATUS_OK
from struct import unpack, pack, calcsize
from common.utils import *

OK_RESPONSE_CODE = 0
NO_DRAW_YET_CODE = 1
RESPONSE_CODE_SIZE = 2
BET_MESSAGE_SIZE = 128
ACTION_INFO_MSG_SIZE = 5
BET_CODE = "B"
END_CODE = "E"
WINNERS_CODE = "W"

STRING_ENCODING = 'utf-8'

class Server:
    def __init__(self, port, listen_backlog, count_agencies):
        # Initialize server socket
        self._server_socket = CommonSocket()
        self._server_socket.bind_and_listen('', port, listen_backlog)
        self._keep_running = True
        signal.signal(signal.SIGTERM, self.sigterm_handler)
        
        self._agency_finished = {}
        for i in range(1, int(count_agencies) + 1):
            self._agency_finished [i] = False

        self._winners = None

        

    def sigterm_handler(self, _signo, _stack_frame):
        logging.info('action: sigterm_received')
        self._keep_running = False
        

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._keep_running:
            logging.info('action: accept_connections | result: in_progress')
            client_sock, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            self.handle_client_connection(client_sock)

        self._server_socket.close()
        logging.info(f'action: close_server_socket | result: success')

    def handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        status, msg, addr = client_sock.receive(ACTION_INFO_MSG_SIZE)
        if status == STATUS_ERR:
            logging.error("action: receive_message | result: fail")
            client_sock.close()
            return

        action_code, agency, batch_size = unpack('!cHH',msg)
        action_code = action_code.decode(STRING_ENCODING)
        

        if action_code == BET_CODE:
            logging.info(f'action: action_info_msg_received | result: success | type: receive bets batch | batch size: {batch_size} | agency: {agency}')
            self.receive_bet_batch(client_sock, agency, batch_size)
        elif action_code == END_CODE:
            logging.info(f'action: action_info_msg_received | result: success | type: notify that all of the agency\'s bets have been received | agency: {agency}')
            self.receive_end_code(client_sock, agency)
        elif action_code == WINNERS_CODE:
            logging.info(f'action: action_info_msg_received | result: success | type: winners query | agency: {agency}')
            self.winners_query(client_sock, agency)
        else:
            logging.error(f'action: action_info_msg_received | result: fail | reason: invalid action code ({action_code})')
            client_sock.close()
            return   

    def receive_bet_batch(self, client_sock, agency, batch_size):
        status, msg, addr = client_sock.receive(BET_MESSAGE_SIZE * batch_size)
        if status == STATUS_ERR:
            logging.error("action: receive_message | result: fail")
            client_sock.close()
            return

        bets = []
        for i in range(batch_size):                    
            first_byte = i * BET_MESSAGE_SIZE
            last_byte = (i+1) * BET_MESSAGE_SIZE

            msg_code, agency, first_name_size, first_name, last_name_size, last_name, document, birthdate, number = unpack('!cHH53sH52sI10sH',msg[first_byte:last_byte])
            msg_code = msg_code.decode(STRING_ENCODING)
            first_name = first_name[0:first_name_size].decode(STRING_ENCODING)
            last_name = last_name[0:last_name_size].decode(STRING_ENCODING)
            birthdate = birthdate.decode(STRING_ENCODING)

            bet = Bet(agency, first_name, last_name, document, birthdate, number)
            bets.append(bet)

        store_bets(bets)
        logging.info(f'action: apuestas_almacenadas | result: success | size: {batch_size}')

        response_code = (OK_RESPONSE_CODE).to_bytes(RESPONSE_CODE_SIZE, byteorder='big', signed=True)
        status = client_sock.send(response_code, RESPONSE_CODE_SIZE)
        if status == STATUS_ERR:
            logging.error("action: send_message | result: fail | ip: {addr[0]}")
            client_sock.close()
            return

        client_sock.close()
        logging.info(f'action: close_client_socket | result: success | ip: {addr[0]}')
            

            
    def receive_end_code(self, client_sock, agency):
        # Evito buscar los ganadores 2 veces
        if self._winners:
            client_sock.close()
            logging.info(f'action: close_client_socket | result: success')
            return

        self._agency_finished[agency] = True

        # Si ya terminaron todas las agencias
        if all(self._agency_finished.values()):
            self._winners = []
            for bet in load_bets():
                if has_won(bet):
                    self._winners.append(bet)

            logging.info('action: sorteo | result: success')

        client_sock.close()
        logging.info(f'action: close_client_socket | result: success')


    def winners_query(self, client_sock, agency):
        # Si todavia no estan los ganadores, devuelvo que aun no se hizo el sorteo
        if not self._winners:
            response_code = (NO_DRAW_YET_CODE).to_bytes(RESPONSE_CODE_SIZE, byteorder='big', signed=True)
            status = client_sock.send(response_code, RESPONSE_CODE_SIZE)
            if status == STATUS_ERR:
                logging.error("action: send_message | result: fail")
                client_sock.close()
                logging.info(f'action: close_client_socket | result: success')
            
            return

        response_code = OK_RESPONSE_CODE

        agency_winners = []

        for winner in self._winners:
            if winner.agency == agency:
                agency_winners.append(winner.number)
                
        count_winners = len(agency_winners)
        msg_format = "!hH" + count_winners * "H"
        msg = pack(msg_format, response_code, count_winners, *agency_winners)
        msg_size = calcsize(msg_format)
        status = client_sock.send(msg, msg_size)
        if status == STATUS_ERR:
                logging.error("action: send_message | result: fail")

        client_sock.close()
        logging.info(f'action: close_client_socket | result: success')
        
        
        
           