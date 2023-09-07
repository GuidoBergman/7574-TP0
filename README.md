# Ejercicio 2

Para resolver este punto se utilizó un Volume Mount

Para verificar el correcto funcionamiento de esta solución, se puede ejecutar los comandos:

```bash
make docker-compose-up
make docker-compose-down
make docker-compose-up
```

modificando los archivos de configuración antes del último up. Se podrá verificiar que todas las capas se construyen del caché, pero la configuración de los clientes y el servidor cambió.
