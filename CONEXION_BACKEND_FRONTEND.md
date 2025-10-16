# Conexión Backend-Frontend MIA

Este documento describe cómo está configurada la conexión entre el backend (Go/Gin) y el frontend (React/Vite).

## Arquitectura

```
Frontend (React + Vite)          Backend (Go + Gin)
Port: 5173                       Port: 8080
┌─────────────────┐             ┌─────────────────┐
│                 │   Proxy     │                 │
│   Vite Dev      │────────────▶│   Gin Router    │
│   Server        │             │                 │
│                 │             │   CORS: *       │
└─────────────────┘             └─────────────────┘
```

## Configuración

### Backend (Puerto 8080)

**Ubicación:** `Backend/cmd/server/main.go`

El backend utiliza Gin y expone los siguientes endpoints:

#### Endpoints disponibles:

1. **Health Check**
   - `GET /health`
   - Retorna el estado del servidor y configuración

2. **Comandos individuales**
   - `POST /api/commands`
   - Body: `{ "input": "mkdisk -size=50 -unit=M -path=\"/tmp/Disco1.mia\"" }`
   - Ejecuta un comando individual

3. **Scripts SMIA**
   - `POST /api/script`
   - Body: `{ "script": "...", "stopOnError": true }`
   - Ejecuta un script .smia completo

4. **Reportes**
   - `POST /api/reports`
   - Body: `{ "name": "mbr", "id": "841A", "out": "/path/output.jpg", "extra": "" }`
   - Genera reportes (mbr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls)

5. **Archivos estáticos de reportes**
   - `GET /reports/static/<archivo>`
   - Sirve las imágenes y archivos generados

#### CORS configurado en `Backend/router/router.go:15-28`:
```go
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET,POST,PUT,DELETE,OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With
```

#### Variables de entorno (Backend/.env):
```env
PORT=8080
DISKS_PATH=./Discos
REPORTS_PATH=./Reports
CARNET_LAST_TWO=84
DEBUG=true
```

### Frontend (Puerto 5173)

**Ubicación:** `Frontend/vite.config.js`

Vite está configurado con un proxy que redirige todas las peticiones a `/api`, `/health` y `/reports` al backend en `localhost:8080`.

#### Configuración del proxy:
```javascript
server: {
  port: 5173,
  proxy: {
    '/api': { target: 'http://localhost:8080', changeOrigin: true },
    '/health': { target: 'http://localhost:8080', changeOrigin: true },
    '/reports': { target: 'http://localhost:8080', changeOrigin: true }
  }
}
```

#### Variables de entorno (Frontend/.env):
```env
VITE_API_URL=http://localhost:8080
VITE_PORT=5173
```

### API Client (Frontend/src/lib/api.js)

El módulo API exporta funciones para comunicarse con el backend:

```javascript
// Health check
await API.health()

// Login
await API.login(id, user, pass)

// Logout
await API.logout()

// Ejecutar comando individual
await API.run("mkdisk -size=50 -unit=M -path=\"/tmp/Disco1.mia\"")

// Ejecutar script .smia
await API.runScript(scriptContent, stopOnError = true)

// Listar discos
await API.disks()

// Obtener particiones de un disco
await API.partitions(diskPath)

// Listar archivos en un directorio
await API.list(id, path)

// Leer contenido de un archivo
await API.file(id, path)

// Obtener información de journaling
await API.journaling(id)

// Generar reporte
await API.genReport(name, id, out, extra = '')
```

## Cómo ejecutar

### 1. Iniciar el Backend

```bash
cd Backend
go run cmd/server/main.go
```

El servidor estará disponible en `http://localhost:8080`

### 2. Iniciar el Frontend

```bash
cd Frontend
npm install  # Solo la primera vez
npm run dev
```

El frontend estará disponible en `http://localhost:5173`

## Flujo de una petición

1. El usuario interactúa con la UI (React)
2. La UI llama a una función de `api.js` (ej: `API.run(command)`)
3. `api.js` hace un `fetch` a `/api/commands`
4. El proxy de Vite intercepta la petición y la redirige a `http://localhost:8080/api/commands`
5. El backend (Gin) recibe la petición en el router
6. El controller procesa la petición y llama al servicio correspondiente
7. El servicio ejecuta la lógica de negocio
8. La respuesta viaja de vuelta al frontend
9. La UI muestra el resultado al usuario

## Estructura del Backend

```
Backend/
├─ cmd/server/main.go           # Punto de entrada, DI wiring
├─ config/config.go             # Configuración (puertos, paths)
├─ router/router.go             # Gin router, middlewares, CORS
├─ controllers/                 # Controladores HTTP
│  ├─ commands_controller.go    # POST /api/commands
│  ├─ script_controller.go      # POST /api/script
│  └─ reports_controller.go     # POST /api/reports
├─ command/                     # Lógica de aplicación (casos de uso)
│  ├─ disk/                     # mkdisk, rmdisk, fdisk, mount, mounted
│  ├─ fs/                       # mkfs, mkdir, mkfile, cat
│  ├─ users/                    # login, logout, mkgrp, rmgrp, mkusr, rmusr
│  └─ reports/                  # reportes (mbr, disk, tree, etc.)
├─ core/                        # Dominio (modelos + interfaces)
│  ├─ models/                   # MBR, Partition, EBR, SuperBlock, Inode
│  └─ ports/                    # Interfaces (hexagonal ports)
└─ storage/                     # Adaptadores de infraestructura
   ├─ diskio/                   # I/O seguro de discos
   ├─ mounts/                   # Estado de montajes
   ├─ session/                  # Sesión en memoria
   └─ graphviz/                 # Generación de reportes visuales
```

## Manejo de errores

El backend retorna errores en formato JSON:
```json
{
  "message": "descripción del error"
}
```

El frontend captura estos errores en las funciones de `api.js` y los propaga como excepciones JavaScript.

## Notas importantes

1. **CORS:** El backend permite peticiones desde cualquier origen (`Access-Control-Allow-Origin: *`). En producción, esto debería limitarse al dominio del frontend.

2. **Proxy de desarrollo:** El proxy de Vite solo funciona en desarrollo. En producción, el frontend debe apuntar directamente a la URL del backend.

3. **Rutas estáticas:** Los reportes generados se sirven desde `/reports/static/`. Asegúrate de que la carpeta `Reports` existe en el backend.

4. **Autenticación:** Los endpoints de login/logout están definidos en el frontend pero puede que necesiten implementación en el backend si no están ya.

5. **Variables de entorno:** Copia `.env.example` a `.env` y ajusta los valores según tu entorno.

## Testing

### Test del backend
```bash
cd Backend
curl http://localhost:8080/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "disks_path": "./Discos",
  "reports_path": "./Reports",
  "debug": true
}
```

### Test del frontend
Abre `http://localhost:5173` en tu navegador y verifica que la UI se carga correctamente.

## Troubleshooting

### Error: "Connection refused"
- Verifica que el backend esté corriendo en el puerto 8080
- Verifica que no haya otro proceso usando el puerto

### Error: "CORS policy"
- Verifica que el middleware CORS esté configurado en `Backend/router/router.go`
- Verifica que el backend esté permitiendo el origen correcto

### Error: "404 Not Found"
- Verifica que la ruta esté correctamente registrada en `Backend/router/router.go`
- Verifica que el proxy de Vite esté configurado correctamente

### Error: "Cannot read property of undefined"
- Verifica que el backend esté retornando la estructura de datos esperada
- Verifica que el frontend esté parseando correctamente la respuesta
