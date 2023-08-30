import sys

FILEPATH='docker-compose-dev.yaml'



# Define the text of the Docker Compose file
textInitialConfig = """version: '3.9'
name: tp0
services:
"""

textServerConfig = """
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/congig.ini:ro
"""


textNetworkConfig = """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""


if len(sys.argv) == 1 or sys.argv[1] == '-h':
    print('Debe ingresar la cantidad de clientes que desea crear. Ej: el comando \'createDockerCompose.py 3\' creara 3 clientes.')
elif not sys.argv[1].isdigit() or int(sys.argv[1]) <= 0:
    print('La cantidad de clientes a crear debe ser un número mayor a cero')

else:
    countClients = int(sys.argv[1])
    
    textClientConfig = ""

    for i in range(1, countClients+ 1):
        textClientConfig += f"""
  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/build/config.yaml:ro
    """

    dockerComposeFile = open(FILEPATH, 'w')

    fileContent = [textInitialConfig, textServerConfig, 
               textClientConfig, textNetworkConfig]

    dockerComposeFile.writelines(fileContent)
 

    dockerComposeFile.close()