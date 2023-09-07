**Guido Bergman**

**Padrón: 10430**

# Entrega final

En esta rama se encuentran las implementaciones correspondientes a todos los ejercicios, a excepción de algunas modificaciones menores que se hicieron para que el TP siga funcionando. Ademas en este README se explica la solución de cada uno de los ejercicios.

La solución para cada uno de los ejercicios por separado se encuentra en su rama correspondiente.


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