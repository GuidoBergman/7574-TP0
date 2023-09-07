**Guido Bergman**

**Padrón: 10430**

# Entrega final

En esta rama se encuentran las implementaciones correspondientes a todos los ejercicios, a excepción de algunas modificaciones menores que se hicieron para que el TP siga funcionando. Ademas en este README se explica la solución de cada uno de los ejercicios.

La solución para cada uno de los ejercicios por separado se encuentra en su rama correspondiente.

# Ejercicio 1

Para resolver este ejercicio se agrego un nuevo cleinte en el archivo docker-compose-dev.yaml

# Ejercicio 1.1



# Ejercicio 2

Para resolver este punto se utilizó un Volume Mount

Para verificar el correcto funcionamiento de esta solución, se puede ejecutar los comandos:

```bash
make docker-compose-up
make docker-compose-down
make docker-compose-up
```

modificando los archivos de configuración antes del último up. Se podrá verificiar que todas las capas se construyen del caché, pero la configuración de los clientes y el servidor cambió.


# Ejercicio 3

En este ejercicio se agrego una nueva imagen de docker, que contiene el *script* correspondiente al healt check. Este script envia un mensaje, cuyo texto recibe en un archivo de configuración, al echo server y verifica que la respuesta obtenida sea igual al mensaje enviado. Esto se repite cada una cantidad de tiempo que se recibe también en la configuración. En el docker-compose-dev.yaml de la rama correspondiente al ejercicio se levanta la imagen creada para este punto.

# Ejercicio 4

# Ejercicio 5

Para resolver este ejercicio se agregaron en el docker-compose variables de entorno, que contienen para cada cliente los datos de la apuesta que va a enviar. Además se desarrollo un protocolo binario de mensajes de con longitud fija. Se tomo está decisión para simplificar la interpretación de los mensajes

El mensaje para enviar una apuesta consta de 128 bytes y contiene, en el siguiente orden:
- Un byte con un caracter para identificar de tipo del mensaje, que será una 'B'
- Un entero sin signo de 2 bytes para indicar el número de la agencia
- Un entero sin signo de 2 bytes para indicar la longitud del primer nombre
- 53 bytes que contendrán un string con el primer nombre endodeado usando el estandár UTF-8 y a continuación un padding de ceros
- Un entero sin signo de 2 bytes para indicar la longitud del apellido
- 52 bytes que contendrán el apellido, en el mismo formato que el primer nombre
- 10 bytes con un string endodeado usando el estandár UTF-8 que contiene la fecha de nacimento en formato 'YYYY-MM-DD'
- Un entero sin signo de 2 bytes para indicar el número por el cual se aposto



El servidor respondera un entero con signo de 2 bytes con el código de respuesta 0 en caso de haber podido procesar la apuesta correctamente.


# Ejercicio 6

Para este ejercicio se agrego un `unzip` al *Dockerfile* del cliente para poder leer los archivos en CSV y una variable de entrono en su archivo de configuración que contiene el tamaño de los batches a enviar.

Además se extendió el protocolo del punto anterior, agregando 

# Ejercicio 7


# Ejercicio 8

Para agregar paralelismo al servidor se utilizó la biblioteca *Multiprocessing* de *Python*. A partir de la misma, se crea un nuevo proceso por cada cada agencia que el socket del servidor acepta. Dado que el servidor guarda las apuestas en un solo archivo .csv, se identifico que el acceso al mismo debería realizarse de forma secuencial. Para esto, se utilizó un lock que se adquiere antes de guardar o leer apuestas de ese archivo y se libera luego. Además de dicho lock, los procesos de cada agencia comparten:
- Un diccionario que se utilza para determinar si todas las agencias terminaron de enviar sus apuestas o no
- Una lista de ganadores que se llena cuando se realiza el sorteo
- Una variable de tipo enterno, que se usa como booleano para verificar si el sorteo ya fue realizado o no


El proceso correspondiente a cada agencia en el servidor terminará una vez que se hayan enviado a la agencia las apuestas de la misma que ganaron el sorteo. La otra forma de que esos procesos terminen es cuando el servidor recibe la señal SIGTERM. Cuando esto ocurre, el proceso padre que creó a los procesos de las agencias les enviará esta señal usando la función *terminate* de *Multiprocessing*. Al recibir la señal, los procesos de las agencias cerraran el socket que utilizan para comunicarse con el cliente y saldrán.
