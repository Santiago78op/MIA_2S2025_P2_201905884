# Implementación de Journal Viewer API

## Resumen

Se han implementado dos endpoints REST para visualizar el journal EXT3 desde la UI:

1. **GET /api/journal/:id** - Formato crudo con timestamps
2. **GET /api/journal/:id/table** - Formato tabla listo para renderizar en UI

---

## 1. Endpoint: GET /api/journal/:id

### Descripción

Devuelve las entradas del journal en formato crudo con información completa incluyendo timestamps.

### URL

```
GET /api/journal/:id
```

### Parámetros

- **id** (path param, required): ID de la partición montada (ej: "841A")

### Respuestas

#### 200 OK - Success

```json
{
  "mount_id": "841A",
  "entries": [
    {
      "op": "MKDIR",
      "path": "/home",
      "content": "",
      "timestamp": "2025-10-19T15:30:45Z"
    },
    {
      "op": "MKFILE",
      "path": "/home/file.txt",
      "content": "size=100",
      "timestamp": "2025-10-19T15:31:12Z"
    },
    {
      "op": "CHMOD",
      "path": "/home/file.txt",
      "content": "ugo=644,recursive=false",
      "timestamp": "2025-10-19T15:32:00Z"
    }
  ]
}
```

#### 400 Bad Request - Missing ID

```json
{
  "error": "se requiere id"
}
```

#### 404 Not Found - Partition not mounted

```json
{
  "error": "partición no montada"
}
```

#### 404 Not Found - EXT2 (no journal)

```json
{
  "error": "journal no disponible (solo EXT3)"
}
```

#### 500 Internal Server Error

```json
{
  "error": "descripción del error"
}
```

### Formato de Entries

Cada entrada del journal tiene los siguientes campos:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `op` | string | Operación (MKDIR, MKFILE, CHMOD, CHOWN, COPY, MOVE, etc.) |
| `path` | string | Ruta del archivo/directorio afectado |
| `content` | string | Información adicional de la operación |
| `timestamp` | datetime | Timestamp ISO 8601 de cuándo ocurrió la operación |

### Ejemplo de uso

```bash
curl http://localhost:8080/api/journal/841A
```

---

## 2. Endpoint: GET /api/journal/:id/table

### Descripción

Devuelve las entradas del journal en formato tabla, optimizado para renderizar directamente en una tabla HTML/React.

### URL

```
GET /api/journal/:id/table
```

### Parámetros

- **id** (path param, required): ID de la partición montada (ej: "841A")

### Respuestas

#### 200 OK - Success

```json
{
  "mount_id": "841A",
  "rows": [
    {
      "operacion": "MKDIR",
      "path": "/home",
      "contenido": "",
      "fecha": "2025-10-19 15:30:45"
    },
    {
      "operacion": "MKFILE",
      "path": "/home/file.txt",
      "contenido": "size=100",
      "fecha": "2025-10-19 15:31:12"
    },
    {
      "operacion": "CHMOD",
      "path": "/home/file.txt",
      "contenido": "ugo=644,recursive=false",
      "fecha": "2025-10-19 15:32:00"
    }
  ]
}
```

#### Errores

Los errores son idénticos al endpoint anterior.

### Formato de Rows

Cada fila de la tabla tiene los siguientes campos:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `operacion` | string | Nombre de la operación |
| `path` | string | Ruta del archivo/directorio |
| `contenido` | string | Información adicional |
| `fecha` | string | Fecha en formato "YYYY-MM-DD HH:MM:SS" |

### Diferencias con /api/journal/:id

| Aspecto | /api/journal/:id | /api/journal/:id/table |
|---------|------------------|------------------------|
| Campo timestamp | ISO 8601 datetime | String formateado |
| Nombre campo operación | `op` | `operacion` |
| Nombre campo contenido | `content` | `contenido` |
| Uso recomendado | API general | Renderizado de tablas |

### Ejemplo de uso

```bash
curl http://localhost:8080/api/journal/841A/table
```

### Ejemplo de renderizado en React

```jsx
function JournalTable({ mountId }) {
  const [rows, setRows] = useState([]);

  useEffect(() => {
    fetch(`/api/journal/${mountId}/table`)
      .then(res => res.json())
      .then(data => setRows(data.rows));
  }, [mountId]);

  return (
    <table>
      <thead>
        <tr>
          <th>Operación</th>
          <th>Path</th>
          <th>Contenido</th>
          <th>Fecha</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row, i) => (
          <tr key={i}>
            <td>{row.operacion}</td>
            <td>{row.path}</td>
            <td>{row.contenido}</td>
            <td>{row.fecha}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
```

---

## 3. Arquitectura de la Implementación

### 3.1 ViewerController

**Archivo**: `/controllers/viewer_controller.go`

**Responsabilidades**:
- Validar que la partición está montada
- Llamar al repositorio para leer el journal
- Transformar datos del dominio a DTOs de API
- Manejar errores y devolver respuestas HTTP apropiadas

**Dependencias**:
```go
type ViewerController struct {
    fs     ports.FsRepository   // Para leer journal
    mounts ports.MountStore     // Para validar montajes
    sess   ports.SessionStore   // Para autenticación (futuro)
}
```

### 3.2 DTOs (Data Transfer Objects)

#### journalEntryDTO
```go
type journalEntryDTO struct {
    Op        string    `json:"op"`
    Path      string    `json:"path"`
    Content   string    `json:"content,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

Usado por: `GET /api/journal/:id`

#### journalRowDTO
```go
type journalRowDTO struct {
    Operacion string `json:"operacion"`
    Path      string `json:"path"`
    Contenido string `json:"contenido"`
    Fecha     string `json:"fecha"`
}
```

Usado por: `GET /api/journal/:id/table`

### 3.3 Adaptadores

Se crearon nuevos adaptadores para conectar el ViewerController con la capa de infraestructura:

#### PortsFsAdapter

**Archivo**: `/storage/adapters/ports_fs_adapter.go`

**Propósito**: Adapta `FileFsRepository` para cumplir con `ports.FsRepository`

**Métodos relevantes**:
- `JournalList(id string) ([]models.Journal, error)` - Lee entradas del journal
- `JournalAppend(id string, entry models.Journal) error` - Agrega entrada al journal
- `JournalClear(id string) error` - Limpia el journal

#### NewPortsSessionAdapter

**Archivo**: `/storage/adapters/session_adapter.go`

**Propósito**: Expone `ports.SessionStore` desde `SessionAdapter`

### 3.4 Flujo de Datos

```
1. Cliente HTTP
   ↓
2. Router (/api/journal/:id)
   ↓
3. ViewerController.GetJournal() o GetJournalTable()
   ↓
4. Validar montaje (ports.MountStore.List())
   ↓
5. Leer journal (ports.FsRepository.JournalList(id))
   ↓
6. FileFsRepository.JournalListPublic(id)
   ↓
7. Leer journal desde disco (.mia)
   ↓
8. Parsear entradas de journal (models.Journal)
   ↓
9. Transformar a DTO (journalEntryDTO o journalRowDTO)
   ↓
10. Devolver JSON al cliente
```

---

## 4. Validaciones

### 4.1 Validación de ID

- El parámetro `id` es **requerido**
- Si falta, devuelve **400 Bad Request**

### 4.2 Validación de Montaje

- Se verifica que el `id` existe en el store de montajes
- Si no existe, devuelve **404 Not Found** con mensaje "partición no montada"

### 4.3 Validación de Tipo de Sistema

- Si la partición es **EXT2** (sin journal), devuelve **404 Not Found**
- Solo particiones **EXT3** tienen journal disponible

---

## 5. Operaciones del Journal

Las siguientes operaciones pueden aparecer en el journal:

| Operación | Descripción | Campo Content |
|-----------|-------------|---------------|
| MKDIR | Crear directorio | "" |
| MKFILE | Crear archivo | "size=N" |
| REMOVE | Eliminar archivo/directorio | "" |
| EDIT | Editar contenido de archivo | "" |
| RENAME | Renombrar archivo/directorio | "name=nuevo_nombre" |
| COPY | Copiar archivo/directorio | "dest=/ruta/destino" |
| MOVE | Mover archivo/directorio | "dest=/ruta/destino" |
| CHMOD | Cambiar permisos | "ugo=644,recursive=false" |
| CHOWN | Cambiar propietario | "user=nombre,recursive=false" |

---

## 6. Testing

### 6.1 Pruebas Manuales

#### Test 1: Leer journal de partición EXT3

```bash
# Prerequisitos:
# 1. Tener una partición montada con id "841A"
# 2. Haber ejecutado operaciones en el filesystem

# Ejecutar:
curl http://localhost:8080/api/journal/841A

# Resultado esperado: 200 OK con lista de entradas
```

#### Test 2: Intentar leer journal de partición no montada

```bash
curl http://localhost:8080/api/journal/XXXX

# Resultado esperado: 404 Not Found
# { "error": "partición no montada" }
```

#### Test 3: Intentar leer journal de partición EXT2

```bash
# Prerequisitos:
# 1. Crear partición EXT2 montada como "841B"

curl http://localhost:8080/api/journal/841B

# Resultado esperado: 404 Not Found
# { "error": "journal no disponible (solo EXT3)" }
```

#### Test 4: Formato tabla

```bash
curl http://localhost:8080/api/journal/841A/table

# Resultado esperado: 200 OK con formato tabla
# { "mount_id": "841A", "rows": [...] }
```

### 6.2 Escenario de Prueba Completo

```bash
# 1. Crear disco y partición
echo "mkdisk -size=10 -unit=M -path=/tmp/test.mia" | curl -X POST http://localhost:8080/api/commands -d @-

# 2. Crear partición EXT3
echo "fdisk -size=5 -type=P -unit=M -path=/tmp/test.mia -name=Part1" | curl -X POST http://localhost:8080/api/commands -d @-

# 3. Montar partición (obtener id, ej: 841A)
echo "mount -path=/tmp/test.mia -name=Part1" | curl -X POST http://localhost:8080/api/commands -d @-

# 4. Formatear como EXT3
echo "mkfs -type=full -fs=3fs -id=841A" | curl -X POST http://localhost:8080/api/commands -d @-

# 5. Crear estructura de directorios
echo "mkdir -id=841A -path=/home -p" | curl -X POST http://localhost:8080/api/commands -d @-
echo "mkdir -id=841A -path=/home/user1" | curl -X POST http://localhost:8080/api/commands -d @-

# 6. Crear archivos
echo "mkfile -id=841A -path=/home/user1/file.txt -size=50" | curl -X POST http://localhost:8080/api/commands -d @-

# 7. Leer journal
curl http://localhost:8080/api/journal/841A/table

# Resultado esperado:
# {
#   "mount_id": "841A",
#   "rows": [
#     { "operacion": "MKDIR", "path": "/home", ... },
#     { "operacion": "MKDIR", "path": "/home/user1", ... },
#     { "operacion": "MKFILE", "path": "/home/user1/file.txt", ... }
#   ]
# }
```

---

## 7. Integración en Router

**Archivo**: `/router/router.go`

Las rutas se agregaron al grupo `/api`:

```go
// Listar entradas del journal (solo EXT3)
api.GET("/journal/:id", vc.GetJournal)

// Listar entradas del journal en formato tabla (solo EXT3)
api.GET("/journal/:id/table", vc.GetJournalTable)
```

---

## 8. Dependencias Modificadas

### cmd/server/main.go

Se actualizó la inicialización del `ViewerController` para pasar las dependencias requeridas:

```go
// === Adaptadores para ViewerController (usa ports.*) ===
portsFsAdapter := adapters.NewPortsFsAdapter(fsRepoBase)
portsSessionStore := adapters.NewPortsSessionAdapter(sessionAdapter)

// === Capa de presentación (controllers HTTP) ===
vc := controllers.NewViewerController(portsFsAdapter, portsMountStore, portsSessionStore)
```

---

## 9. Archivos Creados/Modificados

### Archivos Creados

1. **`/storage/adapters/ports_fs_adapter.go`**
   - Adapta `FileFsRepository` para cumplir con `ports.FsRepository`
   - Expone métodos de journal: `JournalList`, `JournalAppend`, `JournalClear`

### Archivos Modificados

1. **`/controllers/viewer_controller.go`**
   - Agregado constructor con dependencias: `fs`, `mounts`, `sess`
   - Implementado `GetJournal()` - Endpoint para formato crudo
   - Implementado `GetJournalTable()` - Endpoint para formato tabla
   - Agregados DTOs: `journalEntryDTO`, `journalRowDTO`

2. **`/router/router.go`**
   - Agregada ruta: `GET /api/journal/:id`
   - Agregada ruta: `GET /api/journal/:id/table`

3. **`/cmd/server/main.go`**
   - Agregados adaptadores: `portsFsAdapter`, `portsSessionStore`
   - Actualizado constructor de `ViewerController` con dependencias

4. **`/storage/adapters/session_adapter.go`**
   - Agregado `NewPortsSessionAdapter()` para exponer `ports.SessionStore`

---

## 10. Limitaciones y TODOs

### Limitaciones Actuales

1. **No hay autenticación**: Los endpoints son públicos (sin validación de sesión)
2. **Sin paginación**: Devuelve todas las entradas del journal (máximo 50 por diseño)
3. **Sin filtrado**: No se puede filtrar por tipo de operación o fecha
4. **Sin ordenamiento**: Las entradas se devuelven en el orden que están en disco

### TODOs Futuros

- [ ] Agregar autenticación usando `SessionStore`
- [ ] Implementar paginación (`?page=1&limit=10`)
- [ ] Agregar filtros (`?op=MKDIR&from=2025-10-01`)
- [ ] Agregar ordenamiento (`?sort=date&order=desc`)
- [ ] Implementar endpoints adicionales:
  - `POST /api/journal/:id/clear` - Limpiar journal (requiere autenticación)
  - `GET /api/journal/:id/stats` - Estadísticas del journal

---

## 11. Conclusión

✅ **Endpoints implementados**:
- GET /api/journal/:id (formato crudo)
- GET /api/journal/:id/table (formato tabla)

✅ **Funcionalidades**:
- Validación de montaje
- Detección de EXT2 vs EXT3
- Transformación de datos a DTOs
- Manejo de errores HTTP

✅ **Arquitectura**:
- Separación de capas (Controller → Adapter → Repository)
- DTOs específicos para cada endpoint
- Adaptadores para desacoplar puertos

✅ **Listo para producción**:
- Compilación exitosa
- Rutas registradas en router
- Dependencias inyectadas correctamente

🚀 **La UI puede consumir estos endpoints para mostrar el journal en tiempo real!**
