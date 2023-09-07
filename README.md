**Guido Bergman**

**Padrón: 10430**

# Entrega final

En esta rama se encuentran las implementaciones correspondientes a todos los ejercicios, a excepción de algunas modificaciones menores que se hicieron para que el TP siga funcionando. Además en este README se explica la implementación de cada uno de los ejercicios. La solución para cada uno de los ejercicios por separado se encuentra en su rama correspondiente.

# Ejercicio 1

Para resolver este ejercicio se agregó un nuevo cliente en el archivo docker-compose-dev.yaml

# Ejercicio 1.1

Para este ejercicio se creó el *script* createDockerCompose.py. Para ejecutarlo se utiliza el comando 
```bash
python createDockerCompose.py CANT_CLIENTES
```
por ejemplo
```bash
python createDockerCompose.py 5
```
creará 5 clientes.

El *script* escribe el contenido de distintas variables en el archivo docker-compose-dev.yaml.

# Ejercicio 2

Para resolver este punto se utilizó un Volume Mount

Para verificar el correcto funcionamiento de esta solución, se puede ejecutar los comandos:

```bash
make docker-compose-up
make docker-compose-down
make docker-compose-up
```

modificando los archivos de configuración antes del último up. Se podrá verificar que todas las capas se construyen del caché, pero la configuración de los clientes y el servidor cambió.


# Ejercicio 3

En este ejercicio se agregó una nueva imagen de docker, que contiene el *script* correspondiente al healt check. Este script envía un mensaje, cuyo texto recibe en un archivo de configuración, al echo server. Luego verifica que la respuesta obtenida sea igual al mensaje enviado. Esto se repite cada una cantidad de tiempo que se recibe también en la configuración. En el docker-compose-dev.yaml de la rama correspondiente al ejercicio se levanta la imagen creada para este punto.

# Ejercicio 4

Para implementar este punto se agregaron en el servidor y el cliente handlers para la señal SIGTERM. El del servidor cierra el socket aceptador, lo que tiene como consecuencia que se salga del *loop* en el que se manejan y aceptan los clientes. Del lado del cliente, cuando llega la señal se cierra el socket y se sale. En implementaciones para puntos posteriores, incluida la que está en se encuentra en la rama EntregaFinal, se aprovechó la función de golang `defer` para cerrar los recursos del cliente.

# Ejercicio 5

Para resolver este ejercicio se agregaron en el docker-compose variables de entorno, que contienen, para cada cliente, los datos de la apuesta que va a enviar. Además se desarrollo un protocolo binario de mensajes, en el cual el mensaje para enviar una apuesta consta de 128 bytes y contiene, en el siguiente orden:
- Un byte con un caracter para identificar de tipo del mensaje, que será una 'B'
- Un entero sin signo de 2 bytes para indicar el número de la agencia
- Un entero sin signo de 2 bytes para indicar la longitud del primer nombre
- 53 bytes que contendrán un string con el primer nombre endodeado usando el estándar UTF-8 y a continuación un padding de ceros
- Un entero sin signo de 2 bytes para indicar la longitud del apellido
- 52 bytes que contendrán el apellido, en el mismo formato que el primer nombre
- Un entero sin signo de 4 bytes con el número de documento
- 10 bytes con un string endodeado usando el estándar UTF-8 que contiene la fecha de nacimiento en formato 'YYYY-MM-DD'
- Un entero sin signo de 2 bytes para indicar el número por el cual se apostó



El servidor responderá un entero con signo de 2 bytes con el código de respuesta 0 en caso de haber podido procesar la apuesta correctamente.


# Ejercicio 6

Para este ejercicio se agrego un `unzip` al *Dockerfile* del cliente para poder leer los archivos en CSV y una variable de entorno en su archivo de configuración que contiene el tamaño máximo que pueden tener los batches a enviar.

Además, se extendió el protocolo del punto anterior, agregando un mensaje de información de la acción que se quiere realizar. Este mensaje ocupa 5 bytes y contiene:
- Un byte con un caracter que representa el código de la acción que se quiere realizar, en este caso enviar un batch usa la letra "B"
- Un entero sin signo de 2 bytes que indica el número de la agencia que está mandando las apuestas, que servirá también para el punto siguiente
- Un entero sin signo de 2 bytes que indica el tamaño del batch a enviar

A continuación de este mensaje, el cliente envía todo el batch de apuestas siguiendo el protocolo del punto anterior. El servidor responderá 0 en un entero con signo de 2 bytes si todo el batch se procesó correctamenete.

# Ejercicio 7

Para implementar este punto se agregaron 2 nuevos códigos de acción que el cliente puede enviar en el mensaje de información de acción:
- "E": se usa para indicar que el cliente terminó de enviar todas las apuestas
- "W": se usa para que el cliente consulte los ganadores
En ambos casos el mensaje a enviar contendrá un padding de ceros, para llenar el espacio del tamaño del batch, dado que en este caso no es necesario enviarlo.

Para responder la consulta de los ganadores, el servidor enviará un mensaje que contiene un entero con signo de 2 bytes. Este entero será 0 si el sorteo ya se realizó o -1 en caso de que el sorteo aún no se haya realizado. Si el sorteo ya se realizó, el servidor envía además un mensaje que contiene primero la cantidad de ganadores de la agencia que realizó la consulta y a continuación los números ganadores de la misma. Ambos datos se envian como enteros sin signo de 2 bytes.

Para realizar el sorteo, el servidor guarda un diccionario que utiliza para saber que agencias ya enviaron el mensaje para indicar que terminaron de enviar sus apuestas. El sorteo se realiza una vez que todas las agencias enviaron este mensaje.

Del lado del cliente, para esperar la realización del sorteo se implementó un *exponential backoff*, que termina cuando el servidor responde al mensaje consultando por los resultados con el código 0, que indica que el sorteo ya se realizó. 

# Ejercicio 8

Para agregar paralelismo al servidor se utilizó la biblioteca *Multiprocessing* de *Python*. A partir de la misma, se crea un nuevo proceso por cada agencia que el socket del servidor acepta. Dado que el servidor guarda las apuestas en un solo archivo .csv, se identificó que el acceso al mismo debería realizarse de forma secuencial. Para esto, se utilizó un lock que se adquiere antes de guardar o leer apuestas de ese archivo y se libera luego. Además de dicho lock, los procesos de cada agencia comparten:
- Un diccionario que se utilza para determinar si todas las agencias terminaron de enviar sus apuestas o no
- Una lista de ganadores que se llena cuando se realiza el sorteo
- Una variable de tipo entero, que se usa como booleano para verificar si el sorteo ya fue realizado o no


El proceso correspondiente a cada agencia en el servidor terminará una vez que se hayan enviado a la agencia las apuestas de la misma que ganaron el sorteo. La otra forma de que esos procesos terminen es cuando el servidor recibe la señal SIGTERM. Cuando esto ocurre, el proceso padre que creó a los procesos de las agencias les enviará esta señal usando la función *terminate* de *Multiprocessing*. Al recibir la señal, los procesos de las agencias cerrarán el socket que utilizan para comunicarse con el cliente y saldrán.
