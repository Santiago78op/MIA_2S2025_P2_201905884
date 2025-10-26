# MIGRACION a EXT3 + JOURNALING (P2)

1. Estrategia de migración sin riesgo
2. Cambios puntuales en paquetes (árbol y contratos)
3. Nuevas estructuras (EXT3 + Journal)
4. Semántica exacta de cada comando nuevo
5. API para el frontend (login, visualizador, journaling)
6. Pruebas y guardrails
7. Snippets de código (Go) listos para pegar

---

# 1) Estrategia de migración sin riesgo

* **No toques P1.** Todo lo nuevo entra como “v2” paralelo o adiciones **atrás de interfaces** ya existentes en `core/ports`.
* **Feature flag**: añade `config.EnableExt3` y `config.EnableJournal`.
* **Versiona por estrategia** (no por tipos): EXT2 y EXT3 comparten `FsRepository`, cambian únicamente las implementaciones y el cálculo de estructuras.
* **Journal write-ahead**: envolver operaciones mutantes con `JournalTx` (begin → append → apply → commit/rollback).
* **Recovery/Loss**: módulos independientes que consumen `FsRepository` y `JournalReader`.

---

# 2) Estructura del Proyecto Backend

```
Backend/
├─ cmd/
│  └─ server/
│     └─ main.go              # Punto de entrada del servidor
├─ command/
│  ├─ disk/
│  │  ├─ parser.go            # Parseo de comandos MKDISK, FDISK, MOUNT
│  │  ├─ service.go           # Lógica de negocio: MKDISK, FDISK, MOUNT + UNMOUNT
│  │  └─ validator.go         # Validaciones de parámetros
│  ├─ fs/
│  │  ├─ parser.go            # Parseo de comandos FS
│  │  ├─ service.go           # Lógica: MKFS, MKFILE, MKDIR, CAT, REMOVE, EDIT, RENAME, COPY, MOVE, FIND
│  │  └─ pathwalk.go          # Navegación de rutas en el sistema de archivos
│  ├─ users/
│  │  ├─ parser.go            # Parseo de LOGIN, LOGOUT, MKGRP, RMGRP, MKUSR, RMUSR
│  │  └─ service.go           # Gestión de usuarios y grupos
│  ├─ reports/
│  │  ├─ parser.go            # Parseo de comandos REP
│  │  └─ service.go           # Generación de reportes (MBR, DISK, INODE, etc.)
│  └─ runner/
│     └─ runner.go            # Orquestador de comandos y scripts
├─ core/
│  ├─ models/
│  │  ├─ ext2.go              # Estructuras EXT2: Superblock, Inode, Block, etc.
│  │  ├─ mbr.go               # Estructuras MBR, EBR, Partition
│  │  ├─ types.go             # Tipos comunes, enums, constantes
│  │  └─ users.go             # Estructuras de usuarios y grupos
│  ├─ ports/
│  │  ├─ disk_repository.go   # Interface para operaciones de disco
│  │  ├─ fs_repository.go     # Interface para operaciones del filesystem
│  │  ├─ mount_store.go       # Interface para gestión de montajes
│  │  ├─ report_generator.go  # Interface para generación de reportes
│  │  └─ session_store.go     # Interface para sesiones de usuario
│  └─ services/
│     ├─ disk_service.go      # Servicio de disco (usa DiskRepository)
│     ├─ fs_service.go        # Servicio de filesystem (usa FsRepository)
│     └─ user_service.go      # Servicio de usuarios (usa SessionStore)
├─ storage/
│  ├─ adapters/
│  │  ├─ disk_adapter.go      # Adaptador para DiskRepository
│  │  ├─ fs_adapter.go        # Adaptador para FsRepository
│  │  ├─ mount_adapter.go     # Adaptador para MountStore
│  │  ├─ report_generator.go  # Implementación de generación de reportes
│  │  ├─ session_adapter.go   # Adaptador para SessionStore
│  │  └─ users_adapter.go     # Adaptador para operaciones de usuarios
│  ├─ diskio/
│  │  ├─ diskio.go            # Operaciones de I/O de bajo nivel
│  │  └─ file_repo.go         # Repositorio de archivos (implementa FsRepository)
│  ├─ graphviz/
│  │  └─ dot.go               # Generación de gráficos con Graphviz
│  ├─ mounts/
│  │  └─ state.go             # Estado global de particiones montadas
│  └─ session/
│     └─ memory.go            # Almacenamiento en memoria de sesiones
├─ controllers/
│  ├─ commands_controller.go  # POST /api/commands - Ejecución de comandos individuales
│  ├─ health_controller.go    # GET /api/health - Health check
│  ├─ reports_controller.go   # GET /api/reports/:name - Descarga de reportes
│  └─ script_controller.go    # POST /api/script - Ejecución de scripts .smia
├─ middleware/
│  ├─ errors.go               # Middleware de manejo de errores
│  └─ logging.go              # Middleware de logging
├─ router/
│  └─ router.go               # Configuración de rutas HTTP (Gin)
├─ config/
│  └─ config.go               # Configuración global de la aplicación
├─ pkg/
│  └─ logger/
│     └─ logger.go            # Logger personalizado
├─ utils/
│  ├─ bytes.go                # Utilidades para manipulación de bytes
│  ├─ calc.go                 # Cálculos de tamaños y estructuras FS
│  ├─ errors.go               # Definición de errores personalizados
│  ├─ permissions.go          # Helpers para permisos UGO
│  └─ smia.go                 # Parser y utilidades para archivos .smia
├─ tests/
│  ├─ disk_service_test.go    # Tests para comandos de disco
│  ├─ e2e_script_test.go      # Tests end-to-end con scripts
│  ├─ fs_service_test.go      # Tests para comandos de filesystem
│  ├─ mount_id_test.go        # Tests para IDs de montaje
│  └─ runner.go               # Utilidades para ejecutar tests
├─ bin/                       # Binarios compilados
├─ Discos/                    # Archivos .mia (discos virtuales)
├─ Reports/                   # Reportes generados (.txt, .svg, .png, etc.)
├─ CONT/                      # Scripts .smia de prueba
├─ go.mod                     # Definición de módulo Go
├─ go.sum                     # Checksums de dependencias
└─ debug_session.go           # Script de debug
```

## Notas sobre la arquitectura

- **Arquitectura Hexagonal**: Separación clara entre dominio (`core`), casos de uso (`command`), y adaptadores (`storage`, `controllers`)
- **Interfaces (Ports)**: Contratos definidos en `core/ports` que desacoplan la lógica de negocio de la infraestructura
- **Comandos**: Cada módulo (`disk`, `fs`, `users`, `reports`) tiene su propio parser y servicio
- **Estado compartido**: `storage/mounts` y `storage/session` mantienen el estado global de montajes y sesiones activas

## Próximas extensiones (P2)

La estructura está preparada para integrar:
- **EXT3 + Journaling**: Nuevo archivo `core/models/ext3.go` y `core/models/journal.go`
- **Comandos adicionales**: CHMOD, CHOWN, RECOVERY, LOSS
- **API REST para visualizador**: Nuevo `controllers/viewer_controller.go`
- **Autenticación**: Endpoints en `/api/auth`

---

# 3) Estructuras nuevas clave

## 3.1 EXT3: cálculo y layout

* **Constante del Journal**: `const JournalEntries = 50`
* **Cálculo de n (EXT3)**:

  ```
  partition_size =
    sizeof(superblock) +
    n * sizeof(Journaling) +   // entradas de journal
    n +                        // bitmap inodos
    3*n +                      // bitmap bloques (3 tipos)
    n * sizeof(inodos) +
    3*n * sizeof(block)
  n = floor( ...despeje... )
  ```
* **SB**: igual a EXT2 + `s_journal_start`, `s_journal_count`, `s_fs_type=3`.

## 3.2 Journal

```go
// core/models/journal.go
type JournalOp uint8
const (
  OpMkFile JournalOp = iota
  OpMkDir
  OpRemove
  OpEdit
  OpRename
  OpCopy
  OpMove
  OpChmod
  OpChown
  // ... otros si lo ves útil
)

type JournalEntry struct {
  Op        JournalOp
  Path      [128]byte
  Dest      [128]byte // para move/copy/rename
  Ugo       [3]byte   // para chmod
  Usuario   [16]byte  // para chown
  Timestamp int64
  ContentOfs int64    // offset a staging de contenido si aplica (edit/mkfile)
}

type JournalInfo struct {
  Count int32   // usadas
  Max   int32   // 50
}
```

---

# 4) Semántica de comandos nuevos (backend)

### FDISK: `-add` / `-delete=[fast|full]`

* `-add` (±size con `-unit`): valida espacio libre posterior; si negativo, no puede quedar < tamaño usado real.
* `-delete=fast`: limpia entrada en MBR (o EBR) y marca libre.
* `-delete=full`: además **rellena a `\0`** la región.
* **Valida** reglas de primarias/extendida/lógicas.

### UNMOUNT `-id=XXXX`

* Desmonta por ID, **resetea correlativo a 0** en la partición y la quita del `mount_store`.

### MKFS

* `-fs=2fs|3fs` (por defecto EXT2). Con `-type=full` formatea todo, crea `users.txt`, inicializa bitmaps, SB y (si 3fs) el Journal.

### REMOVE `-path=...`

* Borra archivo o carpeta **solo si** el usuario tiene **escritura**.
* En carpetas, eliminación **atómica**: si hay un hijo sin permiso, **no** borra nada.

### EDIT `-path=... -contenido=/ruta/os`

* Reemplaza el contenido del archivo (requiere **lectura + escritura** sobre el archivo).

### RENAME `-path=... -name=nuevo`

* Cambia nombre (requiere **escritura** en el padre). Error si existe homónimo.

### COPY `-path=... -destino=...`

* Copia recursiva **solo** de elementos con **permiso de lectura**. Error si destino no existe o sin **escritura**.

### MOVE `-path=... -destino=...`

* **Misma partición**: solo re-enlaza (actualiza parent/entries). Requiere **escritura** en origen y destino.

### FIND `-path=dir -name=patrón`

* Búsqueda recursiva con `?` (1 char) y `*` (≥1 char). Requiere **lectura** donde inspecciona.

### CHOWN `-path=... [-r] -usuario=...`

* `root` puede todo; otros solo sobre **sus** archivos. Con `-r` aplica recursivo.

### CHMOD `-path=... -ugo=XYZ [-r]`

* Cambia permisos (octal U-G-O 0–7). Con `-r` aplica recursivo a **elementos del usuario actual**.

### LOSS `-id=...`

* “Daño simulado”: limpia a `\0` **bm inodos, bm bloques, área de inodos y área de bloques**.

### RECOVERY `-id=...`

* Restaura a estado **consistente previo al último formateo** re-aplicando **Journal + SB** (replay en orden de tiempo).

> Todo lo mutante (remove/edit/rename/copy/move/chmod/chown/mkfile/mkdir) **registra entrada de Journal** (solo EXT3).

---

# 5) API para el frontend (UI/Login/Viewer)

Mantén `/api/commands` para ejecutar scripts/CLI, y expón endpoints REST **solo lectura** para el explorador:

```
POST   /api/auth/login                 {user, pass, id}  -> 200 {token, user}
POST   /api/auth/logout                -> 204
GET    /api/disks                      -> lista con tamaño, fit, montajes
GET    /api/disks/:disk/partitions     -> lista particiones (tamaño, fit, estado)
GET    /api/fs/:id/tree?path=/         -> { entries: [{name,type,perm,owner,group,size,mtime}] }
GET    /api/fs/:id/file?path=/a.txt    -> { name, size, content }
GET    /api/journal/:id                -> [{op,path,dest,ugo,usuario,timestamp}]
```

* El **login** usa `UserService` + `SessionStore` (mismo que P1), solo que ahora viene de la UI.
* El **visualizador** nunca muta; todas las mutaciones van por `/api/commands`.

---

# 6) Pruebas y guardrails

* **Unit**:

  * `disk_service_test.go`: FDISK add/delete (fast/full), UNMOUNT.
  * `fs_service_test.go`: mkfs 2fs/3fs, chmod/chown, find (patrones).
  * `journal_test.go`: append/replay/rollback, recovery tras loss.
* **E2E**:

  * Script `.smia` con: mkdisk→fdisk→mount→mkfs(3fs)→login→mkdir/mkfile→edit→copy/move→chmod/chown→find→remove→journaling→loss→recovery.
* **Invariantes**:

  * Tamaño `.mia` **no cambia**.
  * Bitmaps/contadores en SB siempre consistentes.
  * En EXT3, cada mutación **apendea** journal **antes** de aplicar.

---

# 7) Snippets (Go) listos para pegar

## 7.1 Extiende `core/ports/fs_repository.go`

```go
package ports

import "gondisk/core/models"

type FsRepository interface {
  // EXISTENTES (mkfs, read/write SB/Inodes/Blocks, etc.)

  // NUEVOS – permisos y gestión
  Chmod(path string, ugo [3]byte) error
  Chown(path string, user string, recursive bool) error
  Move(src, dst string) error
  Copy(src, dst string) error
  Remove(path string) error
  Edit(path string, content []byte) error

  // Búsqueda
  ScanByGlob(root string, pattern string) ([]string, error)

  // Journal (solo EXT3; en EXT2 retorna ErrUnsupported)
  JournalAppend(id string, e models.JournalEntry) error
  JournalList(id string) ([]models.JournalEntry, error)
  JournalClear(id string) error
}
```

## 7.2 Cálculo EXT3 (utils/calc.go)

```go
package utils

import "gondisk/core/models"

const JournalEntries = 50

func CalcNExt3(partSize int64, szSB, szInode, szBlock, szJournalEntry int64) int64 {
  // part = SB + n*Journal + n (bm inodos) + 3n (bm bloques) + n*inode + 3n*block
  // => n = floor((part - SB) / (Journal + 1 + 3 + szInode + 3*szBlock))
  denom := szJournalEntry + 1 + 3 + szInode + 3*szBlock
  if denom <= 0 { return 0 }
  n := (partSize - szSB) / denom
  if n < 0 { n = 0 }
  return n
}

func IsExt3(fsType int32) bool { return fsType == 3 }
```

## 7.3 Wrapper de transacciones con journal

```go
type JournalTx struct {
  fs   ports.FsRepository
  id   string
  ops  []models.JournalEntry
}

func (tx *JournalTx) Append(e models.JournalEntry) error {
  tx.ops = append(tx.ops, e)
  return tx.fs.JournalAppend(tx.id, e) // write-ahead
}

// Ejemplo en FsService.Remove:
func (s *FsService) Remove(id, path string) error {
  if s.isExt3(id) {
    tx := &JournalTx{fs: s.repo, id: id}
    _ = tx.Append(models.JournalEntry{Op: models.OpRemove, Path: toFixed(path)})
  }
  // permisos + borrado real (o rollback si falla)
  return s.repo.Remove(path)
}
```

## 7.4 Recovery y Loss (command/fs/recovery_loss.go)

```go
func (s *FsService) Loss(id string) error {
  // wipe: bm inodos, bm bloques, tabla de inodos, tabla de bloques (llenar con 0x00)
  return s.repo.WipeDataAreas(id)
}

func (s *FsService) Recovery(id string) error {
  // leer SB, verificar fsType=3, listar Journal, re-aplicar en orden
  entries, err := s.repo.JournalList(id)
  if err != nil { return err }
  for _, e := range entries {
    if err := s.applyJournalEntry(id, e); err != nil { /* decide: stop/continue */ }
  }
  return nil
}
```

## 7.5 Router (router/router.go)

```go
r.POST("/api/auth/login", auth.Login)
r.POST("/api/auth/logout", auth.Logout)
r.GET("/api/disks", viewer.ListDisks)
r.GET("/api/disks/:disk/partitions", viewer.ListPartitions)
r.GET("/api/fs/:id/tree", viewer.Tree)
r.GET("/api/fs/:id/file", viewer.File)
r.GET("/api/journal/:id", viewer.Journal)
```

---

# 8) Checklist de integración (paso a paso)

1. **Config**: agrega flags `EnableExt3`, `EnableJournal`.
2. **Models**: crea `core/models/ext3.go` y `journal.go`.
3. **Ports**: amplía `FsRepository` como arriba.
4. **Storage**: implementa métodos nuevos en `diskio/file_repo.go`; crea `storage/journal/file_journal.go`.
5. **Command/disk**: añade FDISK `add/delete` y `unmount`.
6. **Command/fs**: extiende `Mkfs(fs=2fs|3fs)`, implementa `remove/edit/rename/copy/move/find`, `chmod/chown`, `loss`, `recovery`.
7. **Controllers**: agrega `viewer_controller.go` y rutas nuevas.
8. **Tests**: unit + e2e script (incluye journaling/loss/recovery).
9. **No regresa nada** de P1: endpoints/contratos actuales siguen funcionando.

---
