# Sistema de Logging - GoDisk 2.0

## Arquitectura

El sistema de logging está diseñado para registrar todas las operaciones importantes tanto en el backend como visualizarlas en el frontend.

### Backend (Go)

#### Componentes

1. **Logger Centralizado** (`internal/logger/logger.go`)
   - Logger thread-safe con almacenamiento en memoria
   - Soporte para 4 niveles: DEBUG, INFO, WARN, ERROR
   - Buffer circular en memoria (configurable, default: 1000 entradas)
   - Persistencia opcional en archivo
   - Output a stdout configurable

2. **Middleware HTTP** (`cmd/server/middleware.go`)
   - Registra automáticamente todas las requests HTTP
   - Captura método, path, query params, IP
   - Mide duración de cada request
   - Registra status code y bytes enviados
   - Niveles automáticos según status:
     - 200-399: INFO
     - 400-499: WARN
     - 500+: ERROR

3. **Endpoints de Logs**
   - `GET /api/logs` - Obtener logs
     - Query params: `level` (DEBUG|INFO|WARN|ERROR), `limit` (número)
   - `POST /api/logs/clear` - Limpiar logs en memoria
   - `GET /api/logs/stats` - Estadísticas de logs

### Frontend (React)

#### Componentes

1. **LogViewer** (`components/LogViewer.tsx`)
   - Visualización en tabla de todos los logs
   - Filtros por nivel y límite
   - Auto-refresh opcional (2 segundos)
   - Estadísticas en tiempo real
   - Visualización de contexto JSON
   - Códigos de color por nivel:
     - DEBUG: Gris
     - INFO: Azul
     - WARN: Amarillo
     - ERROR: Rojo

2. **API Client** (`lib/api.ts`)
   - `getLogs(level?, limit?)` - Obtener logs filtrados
   - `clearLogs()` - Limpiar logs
   - `getLogStats()` - Obtener estadísticas

## Configuración

### Backend

Variables de entorno:

```bash
# Archivo de log persistente (opcional)
export LOG_FILE="godisk.log"

# Puerto del servidor
export PORT="8080"

# CORS origin
export ALLOW_ORIGIN="*"
```

Configuración programática:

```go
// En main.go
logger.Init(logFile, maxEntries, toStdout)
// Ejemplo: logger.Init("app.log", 1000, true)
```

### Frontend

No requiere configuración adicional. El componente LogViewer se puede usar directamente:

```tsx
import { LogViewer } from '@/components/LogViewer'

function MyPage() {
  return <LogViewer />
}
```

## Uso

### Logging en Backend

```go
import "MIA_2S2025_P2_201905884/internal/logger"

// Logs simples
logger.Info("Server started")
logger.Error("Failed to connect")

// Logs con contexto
logger.Info("User logged in", map[string]interface{}{
  "user_id": 123,
  "ip": "192.168.1.1",
})

logger.Error("Database error", map[string]interface{}{
  "error": err.Error(),
  "query": "SELECT * FROM users",
})
```

### Visualización en Frontend

1. Navega a `/logs` en la interfaz web
2. Usa los filtros para ver logs específicos:
   - **Nivel**: ALL, DEBUG, INFO, WARN, ERROR
   - **Límite**: 50, 100, 200, 500 entradas
3. Activa **Auto-refresh** para actualización en tiempo real
4. Haz clic en "Ver contexto" para ver detalles JSON
5. Usa "Limpiar" para borrar logs en memoria

## Características

### ✅ Implementadas

- [x] Logger centralizado thread-safe
- [x] Niveles de log (DEBUG, INFO, WARN, ERROR)
- [x] Almacenamiento en memoria con buffer circular
- [x] Persistencia en archivo opcional
- [x] Middleware HTTP automático
- [x] Endpoints REST para logs
- [x] Visualización web con filtros
- [x] Auto-refresh en tiempo real
- [x] Estadísticas por nivel
- [x] Contexto JSON por entrada
- [x] Códigos de color por severidad

### 🎯 Flujo de Logs

```
1. Evento en Backend
   ↓
2. Logger registra (memoria + archivo)
   ↓
3. Frontend solicita logs vía API
   ↓
4. Backend envía logs filtrados
   ↓
5. Frontend renderiza en tabla
```

### 📊 Ejemplo de Logs

**Backend:**
```json
{
  "timestamp": "2025-01-10T15:30:45Z",
  "level": "INFO",
  "message": "HTTP Response",
  "context": {
    "method": "POST",
    "path": "/api/cmd/run",
    "status": 200,
    "duration": "45ms",
    "bytes": 256
  }
}
```

**Frontend:**
| Timestamp | Nivel | Mensaje | Contexto |
|-----------|-------|---------|----------|
| 2025-01-10 15:30:45 | INFO | HTTP Response | [Ver contexto] |

## Troubleshooting

**Logs no aparecen en el frontend:**
- Verifica que el backend esté corriendo
- Verifica CORS (debe permitir el origen del frontend)
- Revisa la consola del navegador por errores

**Archivo de log no se crea:**
- Verifica permisos de escritura en el directorio
- Verifica que `LOG_FILE` esté configurado correctamente
- Revisa logs de stdout para errores

**Performance con muchos logs:**
- Reduce el `maxEntries` al inicializar el logger
- Usa filtros y límites en el frontend
- Considera limpiar logs periódicamente

## Best Practices

1. **Usa el nivel apropiado:**
   - DEBUG: Información de desarrollo
   - INFO: Operaciones normales
   - WARN: Situaciones inusuales pero manejables
   - ERROR: Errores que requieren atención

2. **Incluye contexto relevante:**
   ```go
   logger.Error("Failed to save user", map[string]interface{}{
     "user_id": user.ID,
     "error": err.Error(),
     "timestamp": time.Now(),
   })
   ```

3. **No loguees información sensible:**
   - Evita passwords, tokens, datos personales
   - Sanitiza datos antes de loguear

4. **Monitorea el tamaño del buffer:**
   - El buffer circular sobrescribe logs antiguos
   - Ajusta `maxEntries` según tus necesidades
   - Usa archivo de log para persistencia de largo plazo
