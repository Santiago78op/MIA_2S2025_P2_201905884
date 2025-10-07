# GoDisk 2.0 - Setup Guide

## Arquitectura del Proyecto

```
MIA_2S2025_P2_201905884/
├── Backend/          # Servidor Go con API REST
│   ├── cmd/server/   # Punto de entrada del servidor
│   ├── internal/     # Lógica de negocio (disk, fs, commands)
│   └── pkg/reports/  # Generación de reportes DOT
└── Frontend/         # Aplicación React + TypeScript
    └── godisk-frontend/
```

## Requisitos

### Backend
- Go 1.21+
- Puerto 8080 disponible

### Frontend
- Node.js 18+
- npm o pnpm

## Instalación y Ejecución

### 1. Backend (API Go)

```bash
cd Backend

# Compilar
go build -o godisk-server ./cmd/server

# Ejecutar
./godisk-server

# El servidor estará corriendo en http://localhost:8080
```

Variables de entorno opcionales:
- `PORT`: Puerto del servidor (default: 8080)
- `ALLOW_ORIGIN`: CORS origin (default: *)
- `LOG_FILE`: Archivo de log persistente (default: godisk.log)

### 2. Frontend (React)

```bash
cd Frontend/godisk-frontend

# Instalar dependencias
npm install

# Configurar la URL del backend (opcional)
# Edita .env si el backend NO está en localhost:8080
echo "VITE_API_URL=http://localhost:8080" > .env

# Ejecutar en desarrollo
npm run dev

# El frontend estará en http://localhost:5173
```

## Uso de la Aplicación

### Terminal de Comandos
Ejecuta comandos individuales del sistema de archivos:
```
mkdisk -size=50 -unit=m -fit=ff -path="/tmp/Disco1.mia"
fdisk -path="/tmp/Disco1.mia" -mode=add -name=Part1 -size=10 -unit=m -type=p
mount -path="/tmp/Disco1.mia" -name=Part1
mkfs -id=<mount_id> -fs=2fs
```

### Ejecutor de Scripts
Carga archivos `.smia` con múltiples comandos y ejecútalos secuencialmente.

### Explorador de Discos
- Visualiza discos `.mia` disponibles
- Inspecciona particiones y montajes
- Ve información detallada del MBR

### Reportes (DOT)
Genera visualizaciones Graphviz de:
- **MBR**: Estructura del Master Boot Record
- **Disk**: Uso del espacio en disco
- **SuperBlock**: Metadata del filesystem
- **Tree**: Árbol de directorios (placeholder)
- **Journal**: Registro de operaciones EXT3 (placeholder)

### Logs del Sistema
Visualiza y gestiona logs en tiempo real:
- Filtrado por nivel (DEBUG, INFO, WARN, ERROR)
- Auto-refresh cada 2 segundos
- Estadísticas por nivel
- Visualización de contexto JSON
- Limpieza de logs
- Persistencia en archivo (backend)

## API Endpoints

### Comandos
- `POST /api/cmd/run` - Ejecutar comando individual
- `POST /api/cmd/execute` - Ejecutar comando con validación
- `POST /api/cmd/script` - Ejecutar script completo
- `POST /api/cmd/validate` - Validar sintaxis de comando

### Discos y Montajes
- `GET /api/disks?path=<path>` - Listar discos .mia
- `GET /api/disks/info?path=<path>` - Info detallada de un disco
- `GET /api/mounted` - Listar particiones montadas

### Reportes (retornan DOT)
- `GET /api/reports/mbr?id=<mount_id>` - Reporte MBR
- `GET /api/reports/disk?id=<mount_id>` - Reporte de uso
- `GET /api/reports/sb?id=<mount_id>` - Reporte superblock
- `GET /api/reports/tree?id=<mount_id>&path=/` - Árbol de archivos
- `GET /api/reports/journal?id=<mount_id>` - Journal EXT3

### Health
- `GET /healthz` - Health check
- `GET /api/version` - Información de versión

### Logs
- `GET /api/logs?level=<LEVEL>&limit=<N>` - Obtener logs
- `POST /api/logs/clear` - Limpiar logs en memoria
- `GET /api/logs/stats` - Estadísticas de logs

## Comandos Soportados

### Gestión de Discos
- `mkdisk` - Crear disco virtual
- `fdisk` - Gestionar particiones
- `mount` / `unmount` - Montar/desmontar particiones

### Filesystems
- `mkfs` - Crear filesystem (EXT2/EXT3)
- `mkdir`, `mkfile` - Crear directorios y archivos
- `remove`, `edit`, `rename`, `copy`, `move` - Operaciones de archivos
- `find`, `chown`, `chmod` - Búsqueda y permisos

### EXT3 (Journaling)
- `journaling` - Activar journaling
- `recovery` - Recuperar desde journal
- `loss` - Simular pérdida de datos

## Desarrollo

### Backend
```bash
cd Backend
go mod tidy              # Actualizar dependencias
go test ./...            # Ejecutar tests (si existen)
go build ./cmd/server    # Compilar
```

### Frontend
```bash
cd Frontend/godisk-frontend
npm run build            # Build de producción
npm run preview          # Preview del build
```

## Notas

- El backend usa CORS permisivo (`*`) por defecto
- Los reportes se generan en formato DOT y se renderizan con Viz.js en el frontend
- El sistema soporta EXT2 y EXT3 con journaling
- Los montajes se identifican con IDs tipo "vdXXXX" (hash del path+partición)

## Troubleshooting

**Error de CORS**: Verifica que el backend esté en el puerto correcto y que `VITE_API_URL` apunte a él.

**Error "mount not found"**: Asegúrate de montar la partición antes de ejecutar comandos del filesystem.

**No se generan reportes**: Verifica que la partición esté montada y que el ID sea correcto.
