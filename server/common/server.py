import logging
import signal
import sys
from common.common_socket import CommonSocket, STATUS_ERR, STATUS_OK

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
            if self._keep_running:
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
        status, msg, addr = client_sock.receive(10)
        if status == STATUS_ERR:
            logging.error("action: receive_message | result: fail")
            client_sock.close()
            return

        msg = msg.rstrip().decode('utf-8')
        logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
        msg = "{}\n".format(msg).encode('utf-8')
        status = client_sock.send(msg, len(msg))
        if status == STATUS_ERR:
            logging.error("action: send_message | result: fail | ip: {addr[0]}")
            client_sock.close()
            return

        client_sock.close()
        logging.info(f'action: close_client_socket | result: success | ip: {addr[0]}')
            

            


        