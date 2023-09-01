import socket

STATUS_OK = 0
STATUS_ERR = -1

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

    def connect(self, host, port):
        self._socket.connect((host, port))

    def receive(self, size):
        received_bytes = 0
        chunks = []
        while received_bytes < size:
            chunk = self._socket.recv(size - received_bytes)
            if chunk == b'':
                return STATUS_ERR, None, None
            
            chunks.append(chunk)
            received_bytes += len(chunk)

        buffer = b''.join(chunks)
        
        addr = self._socket.getpeername()
        return STATUS_OK, buffer, addr

    def send(self, buffer, size):
        sent_bytes = 0
        while sent_bytes < size:
            sent = self._socket.send(buffer)
            if sent == 0:
                return STATUS_ERR, None 

            sent_bytes += sent

        return STATUS_OK    

    def close(self):
        """
        Close connecton of the server socket
        """
        self._socket.close()