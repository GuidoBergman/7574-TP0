import logging
import signal
from common.common_socket import CommonSocket, STATUS_ERR, STATUS_OK
from struct import unpack
from common.utils import *

OK_RESPONSE_CODE = 0
RESPONSE_CODE_SIZE = 2
BET_MESSAGE_SIZE = 128
ACTION_INFO_MSG_SIZE = 5

STRING_ENCODING = 'utf-8'

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = CommonSocket()
        self._server_socket.bind_and_listen('', port, listen_backlog)
        self._keep_running = True
        signal.signal(signal.SIGTERM, self.sigterm_handler)
        

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

        action_code, batch_size, agency = unpack('!cHH',msg)
        action_code = action_code.decode(STRING_ENCODING)
        logging.info(f'action: action_info_msg_received | result: success | action code: {action_code} | batch size: {batch_size} | agency {agency}.')


        status, msg, addr = client_sock.receive(BET_MESSAGE_SIZE)
        if status == STATUS_ERR:
            logging.error("action: receive_message | result: fail")
            client_sock.close()
            return

        msg_code, agency, first_name_size, first_name, last_name_size, last_name, document, birthdate, number = unpack('!cHH53sH52sI10sH',msg)
        msg_code = msg_code.decode(STRING_ENCODING)
        first_name = first_name[0:first_name_size].decode(STRING_ENCODING)
        last_name = last_name[0:last_name_size].decode(STRING_ENCODING)
        birthdate = birthdate.decode(STRING_ENCODING)

        bet = Bet(agency, first_name, last_name, document, birthdate, number)
        store_bets([bet])
        logging.info(f'action: apuesta_almacenada | result: success | dni: {document} | numero: {number}.')

        response_code = (OK_RESPONSE_CODE).to_bytes(RESPONSE_CODE_SIZE, byteorder='big', signed=True)
        status = client_sock.send(response_code, RESPONSE_CODE_SIZE)
        if status == STATUS_ERR:
            logging.error("action: send_message | result: fail | ip: {addr[0]}")
            client_sock.close()
            return

        client_sock.close()
        logging.info(f'action: close_client_socket | result: success | ip: {addr[0]}')
            

            


        