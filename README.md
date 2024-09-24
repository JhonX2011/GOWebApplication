# GOWebApplication

### Descripción
Este proyecto es una implementación básica de una aplicación web en Go utilizando el router Chi. El objetivo principal es proporcionar una estructura sencilla para crear un servidor HTTP que responda a solicitudes en diferentes rutas.

### Estructura del Código
El código se encuentra en el paquete api y se basa en la biblioteca net/http de Go, junto con el paquete *github.com/JhonX2011/GOWebApplication/api/web* para el enrutamiento y *github.com/JhonX2011/GOWebApplication/api/utils/logger* para la gestión de logs.

### Componentes Clave
Application: La estructura principal que contiene el router, el logger y la dirección del servidor.

NewWebApplication(): Función que inicializa una nueva instancia de la aplicación. Configura el puerto del servidor, crea un listener y establece el logger.

Run(): Método que inicia el servidor HTTP y define los tiempos de espera para las conexiones.

defaultRoutes(): Método que define las rutas predeterminadas de la aplicación. Actualmente, incluye una ruta /ping que devuelve un JSON con el mensaje "pong".

### Uso
Instalación: Asegúrate de tener Go instalado en tu máquina. Luego, clona este repositorio y navega a la carpeta del proyecto.

Copiar código
```
git clone <URL_DEL_REPOSITORIO>

cd GOWebApplication
```

Configuración: Puedes establecer el puerto en el que deseas que se ejecute la aplicación mediante la variable de entorno PORT. Si no se establece, la aplicación se ejecutará en el puerto por defecto 8080.


```
export PORT=8000  # Cambia 8000 al puerto deseado
```

Ejecución: Compila y ejecuta la aplicación.

```
go run main.go
```
Prueba: Accede a http://localhost:8080/ping en tu navegador o utiliza curl para verificar que la aplicación está funcionando.

```
curl http://localhost:8080/ping
```
La respuesta debería ser:

json
```
{
"message": "pong"
}
```

### Logs
La aplicación utiliza un logger que imprimirá mensajes en la consola, incluyendo errores y el estado de la ejecución.
