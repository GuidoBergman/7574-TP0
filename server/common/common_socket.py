import socket


class CommonSocket:
    def __init__(self,sock=None):
        # sock param is used only to build the socket from one wich is already initialized
        if not sock:
            self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        else:
            self._socket = sock

    def bind_and_listen(self, host, port, listen_backlog):
        self._socket.bind((host, port))
        self._socket.listen(listen_backlog)

    def accept(self):
        """
        Accept new connections

        Function blocks until a connection is made.
        Then connection created is printed and returned
        """
        c, addr = self._socket.accept()
        return CommonSocket(c), addr

    def receive(self, size):
            msg = self._socket.recv(size)
            addr = self._socket.getpeername()
            return msg, addr

    def send(self, buffer):
        self._socket.send(buffer)

    def close(self):
        """
        Close connecton of the server socket
        """
        self._socket.close()