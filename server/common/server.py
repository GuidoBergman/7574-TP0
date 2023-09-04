import logging
import signal
from common.common_socket import CommonSocket, STATUS_ERR, STATUS_OK
from common.utils import *
from common.handler import Handler



class Server:
    def __init__(self, port, listen_backlog, count_agencies):
        # Initialize server socket
        self._server_socket = CommonSocket()
        self._server_socket.bind_and_listen('', port, listen_backlog)
        self._keep_running = True
        signal.signal(signal.SIGTERM, self.sigterm_handler)
        
        self._handler = Handler(count_agencies)

        

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
            self._handler.handle_client_connection(client_sock)

        self._server_socket.close()
        logging.info(f'action: close_server_socket | result: success')

    
        
        
        
           