import logging
import signal
import sys
from common.common_socket import CommonSocket

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = CommonSocket()
        self._server_socket.bind_and_listen('', port, listen_backlog)
        self._keep_running = True
        

    def sigterm_handler(self, _signo, _stack_frame):
        logging.info('SIGTERM received')
        self._keep_running = False
        

    def run(self):
        signal.signal(signal.SIGTERM, self.sigterm_handler)
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

    def handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg, addr = client_sock.receive(1024)
            msg = msg.rstrip().decode('utf-8')
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            client_sock.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()


        