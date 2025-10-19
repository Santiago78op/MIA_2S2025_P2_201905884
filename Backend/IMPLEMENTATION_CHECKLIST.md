# 📋 Checklist de Implementación - Proyecto 2

Este documento lista todas las tareas pendientes para completar la implementación del Proyecto 2.

---

## 🎯 1. COMANDOS NUEVOS - Parsers y Lógica

### 1.1 Comandos de Disco (command/disk/)

#### ✅ Ya implementados:
- `MKDISK`, `RMDISK`, `FDISK`, `MOUNT`, `MOUNTED`

#### 📝 Por implementar:

**A. FDISK -add** (en `service.go`)
- [ ] Agregar lógica para `-add=±size` en FDISK
- [ ] Validar que hay espacio libre después de la partición
- [ ] Si size es negativo, validar que no se reduzca más allá del espacio usado
- [ ] Actualizar MBR/EBR con el nuevo tamaño

**B. FDISK -delete** (en `service.go`)
- [ ] Agregar lógica para `-delete=fast` (solo marca libre en MBR/EBR)
- [ ] Agregar lógica para `-delete=full` (llena con `\0` toda la región)
- [ ] Validar reglas de primarias/extendida/lógicas

**C. UNMOUNT** (crear parser en `parser.go`, lógica en `service.go`)
- [ ] Parsear comando `unmount -id=XXXX`
- [ ] Desmontar partición del mount store
- [ ] Resetear correlativo a 0 en la partición
- [ ] Cerrar sesión si está activa

---

### 1.2 Comandos de Filesystem (command/fs/)

#### ✅ Ya implementados:
- `MKFS`, `MKDIR`, `MKFILE`, `CAT`

#### 📝 Por implementar:

**A. REMOVE** (crear parser en `parser.go`, usar `recovery_loss.go`)
- [ ] Parsear `remove -path=/ruta`
- [ ] Validar permisos de escritura del usuario sobre el archivo/directorio
- [ ] Para directorios: validar permisos recursivos antes de borrar
- [ ] Eliminar entrada del directorio padre
- [ ] Liberar inodo y bloques en bitmaps
- [ ] Registrar en journal si es EXT3

**B. EDIT** (crear parser en `parser.go`)
- [ ] Parsear `edit -path=/archivo -contenido=/ruta/local`
- [ ] Validar permisos de lectura + escritura sobre el archivo
- [ ] Leer contenido del archivo local
- [ ] Reemplazar contenido del archivo en el FS
- [ ] Actualizar tamaño del inodo
- [ ] Registrar en journal si es EXT3

**C. RENAME** (crear parser en `parser.go`)
- [ ] Parsear `rename -path=/archivo -name=nuevo_nombre`
- [ ] Validar permisos de escritura en el directorio padre
- [ ] Verificar que no existe archivo con el nuevo nombre
- [ ] Actualizar entrada en el directorio padre
- [ ] Registrar en journal si es EXT3

**D. COPY** (crear parser en `parser.go`)
- [ ] Parsear `copy -path=/origen -destino=/destino`
- [ ] Validar permisos de lectura en origen
- [ ] Validar permisos de escritura en destino
- [ ] Copiar recursivamente si es directorio
- [ ] Crear nuevo inodo y bloques
- [ ] Registrar en journal si es EXT3

**E. MOVE** (crear parser en `parser.go`)
- [ ] Parsear `move -path=/origen -destino=/destino`
- [ ] Validar permisos de escritura en origen y destino
- [ ] Mover (cambiar parent del inodo, actualizar entradas)
- [ ] NO copiar bloques (solo cambiar referencias)
- [ ] Registrar en journal si es EXT3

**F. FIND** (crear parser en `parser.go`)
- [ ] Parsear `find -path=/dir -name=patron`
- [ ] Implementar búsqueda recursiva con wildcards `?` y `*`
- [ ] Validar permisos de lectura en cada directorio visitado
- [ ] Devolver lista de rutas que coinciden con el patrón

**G. CHMOD** (ya creado en `chmod_chown.go`, falta integrar)
- [ ] Parsear `chmod -path=/ruta -ugo=777 [-r]`
- [ ] Implementar en el runner
- [ ] Validar que solo root o el propietario puede cambiar permisos

**H. CHOWN** (ya creado en `chmod_chown.go`, falta integrar)
- [ ] Parsear `chown -path=/ruta -user=usuario [-r]`
- [ ] Implementar en el runner
- [ ] Validar que solo root puede cambiar propietario

**I. LOSS** (ya creado en `recovery_loss.go`, falta integrar)
- [ ] Parsear `loss -id=XXXX`
- [ ] Implementar en el runner
- [ ] Validar que es EXT3

**J. RECOVERY** (ya creado en `recovery_loss.go`, falta integrar)
- [ ] Parsear `recovery -id=XXXX`
- [ ] Implementar en el runner
- [ ] Validar que es EXT3

---

### 1.3 MKFS con soporte EXT3

**En `command/fs/parser.go` y `service.go`:**
- [ ] Modificar parser para aceptar `-fs=2fs|3fs`
- [ ] Si es `3fs`, llamar a `ComputeLayoutExt3()` en lugar de `ComputeLayout()`
- [ ] Inicializar journal (50 entradas vacías con OpNone)
- [ ] Escribir SuperBlockExt3 con campos de journal

---

## 🗄️ 2. STORAGE - Implementación de Interfaces

### 2.1 storage/adapters/fs_adapter.go

**Métodos nuevos a implementar:**

```go
// Nuevas operaciones P2
func (a *FsAdapter) Remove(id string, path []string, uid int, gid int) error
func (a *FsAdapter) Edit(id string, path []string, contentHostPath string, uid int, gid int) error
func (a *FsAdapter) Rename(id string, path []string, newName string, uid int, gid int) error
func (a *FsAdapter) Copy(id string, srcPath []string, destPath []string, uid int, gid int) error
func (a *FsAdapter) Move(id string, srcPath []string, destPath []string, uid int, gid int) error
func (a *FsAdapter) Find(id string, basePath []string, pattern string, uid int, gid int) ([]string, error)

// Permisos
func (a *FsAdapter) Chmod(id string, path []string, ugo [3]byte, recursive bool, uid int, gid int) error
func (a *FsAdapter) Chown(id string, path []string, user string, recursive bool, uid int, gid int) error

// Journal
func (a *FsAdapter) JournalAppend(id string, entry models.JournalEntry) error
func (a *FsAdapter) JournalList(id string) ([]models.JournalEntry, error)
func (a *FsAdapter) JournalClear(id string) error

// Recovery y Loss
func (a *FsAdapter) WipeDataAreas(id string) error
func (a *FsAdapter) Recovery(id string) error
```

**Tareas:**
- [ ] Implementar todos los métodos delegando a `FileFsRepository`

---

### 2.2 storage/diskio/file_repo.go

**Métodos a agregar en FileFsRepository:**

```go
// EXT3 Support
func (r *FileFsRepository) MkfsWithType(id string, fsType int32) error
func (r *FileFsRepository) ReadSuperExt3(id string) (models.SuperBlockExt3, error)
func (r *FileFsRepository) WriteSuperExt3(id string, sb models.SuperBlockExt3) error

// Operaciones de archivos
func (r *FileFsRepository) Remove(id string, path []string) error
func (r *FileFsRepository) Edit(id string, path []string, contentHostPath string) error
func (r *FileFsRepository) Rename(id string, path []string, newName string) error
func (r *FileFsRepository) Copy(id string, srcPath, destPath []string) error
func (r *FileFsRepository) Move(id string, srcPath, destPath []string) error
func (r *FileFsRepository) Find(id string, basePath []string, pattern string) ([]string, error)

// Permisos
func (r *FileFsRepository) Chmod(id string, path []string, ugo [3]byte, recursive bool) error
func (r *FileFsRepository) Chown(id string, path []string, user string, recursive bool) error

// Journal
func (r *FileFsRepository) JournalAppend(id string, entry models.JournalEntry) error
func (r *FileFsRepository) JournalList(id string) ([]models.JournalEntry, error)
func (r *FileFsRepository) JournalClear(id string) error

// Recovery
func (r *FileFsRepository) WipeDataAreas(id string) error
func (r *FileFsRepository) Recovery(id string) error
```

**Tareas detalladas:**

#### A. MKFS con EXT3
- [ ] Detectar `-fs=3fs` y llamar a `utils.ComputeLayoutExt3()`
- [ ] Escribir SuperBlockExt3 con offsets de journal
- [ ] Inicializar journal (50 entradas vacías)
- [ ] Inicializar bitmaps, inodos, bloques igual que EXT2

#### B. Remove
- [ ] Navegar al padre usando `pathwalk.go`
- [ ] Verificar permisos de escritura
- [ ] Leer inodo del archivo/directorio
- [ ] Si es directorio, validar permisos recursivos
- [ ] Borrar entrada del directorio padre
- [ ] Liberar inodo en bitmap
- [ ] Liberar bloques en bitmap
- [ ] Actualizar contadores en SuperBlock

#### C. Edit
- [ ] Navegar al archivo
- [ ] Verificar permisos de lectura + escritura
- [ ] Leer contenido del archivo local (host)
- [ ] Liberar bloques viejos
- [ ] Asignar nuevos bloques
- [ ] Escribir nuevo contenido
- [ ] Actualizar i_size en el inodo

#### D. Rename
- [ ] Navegar al padre
- [ ] Verificar que el usuario tiene permiso de escritura en el padre
- [ ] Buscar entrada con el nombre actual
- [ ] Verificar que no existe entrada con el nuevo nombre
- [ ] Cambiar b_name en la entrada del directorio

#### E. Copy
- [ ] Validar permisos de lectura en origen
- [ ] Crear nuevo inodo en destino
- [ ] Copiar permisos y propietario
- [ ] Si es archivo: copiar bloques
- [ ] Si es directorio: copiar recursivamente

#### F. Move
- [ ] Validar permisos de escritura en origen y destino
- [ ] Borrar entrada del directorio origen
- [ ] Crear entrada en directorio destino con el mismo inodo
- [ ] NO copiar bloques (solo mover referencia)

#### G. Find con wildcards
- [ ] Implementar matching de patrones `?` (1 char) y `*` (≥1 char)
- [ ] Recorrer recursivamente desde basePath
- [ ] Validar permisos de lectura en cada directorio
- [ ] Retornar lista de rutas que coinciden

#### H. Chmod
- [ ] Navegar al archivo/directorio
- [ ] Verificar que el usuario es root o propietario
- [ ] Actualizar i_perm en el inodo
- [ ] Si recursive, aplicar a todo el subárbol

#### I. Chown
- [ ] Navegar al archivo/directorio
- [ ] Verificar que el usuario es root
- [ ] Actualizar i_uid en el inodo
- [ ] Si recursive, aplicar a todo el subárbol

#### J. Journal - Append/List/Clear
- [ ] Leer SuperBlock para obtener journal_start
- [ ] Usar `storage/journal/FileJournal` para operaciones
- [ ] JournalAppend: agregar entrada al final
- [ ] JournalList: leer todas las entradas no vacías
- [ ] JournalClear: llenar con OpNone

#### K. WipeDataAreas (Loss)
- [ ] Leer SuperBlock
- [ ] Llenar con `\0` los siguientes rangos:
  - Bitmap de inodos
  - Bitmap de bloques
  - Área de inodos
  - Área de bloques
- [ ] NO tocar SuperBlock ni Journal

#### L. Recovery
- [ ] Leer Journal usando JournalList()
- [ ] Ordenar entradas por timestamp
- [ ] Re-aplicar cada entrada llamando a las funciones correspondientes
- [ ] Manejar errores (decidir si continuar o abortar)

---

## 🎮 3. COMMAND RUNNER - Integración

### 3.1 command/runner/runner.go

**Agregar casos al switch en Run():**

```go
case "remove":
    args, err := fs.ParseRemove(line)
    if err != nil { return "", err }
    return r.fsSvc.Remove(args.ID, args.Path)

case "edit":
    args, err := fs.ParseEdit(line)
    if err != nil { return "", err }
    return r.fsSvc.Edit(args.ID, args.Path, args.Content)

case "rename":
    args, err := fs.ParseRename(line)
    if err != nil { return "", err }
    return r.fsSvc.Rename(args.ID, args.Path, args.Name)

case "copy":
    args, err := fs.ParseCopy(line)
    if err != nil { return "", err }
    return r.fsSvc.Copy(args.ID, args.SrcPath, args.DestPath)

case "move":
    args, err := fs.ParseMove(line)
    if err != nil { return "", err }
    return r.fsSvc.Move(args.ID, args.SrcPath, args.DestPath)

case "find":
    args, err := fs.ParseFind(line)
    if err != nil { return "", err }
    return r.fsSvc.Find(args.ID, args.Path, args.Name)

case "chmod":
    args, err := fs.ParseChmod(line)
    if err != nil { return "", err }
    return r.fsSvc.Chmod(args.ID, args.Path, args.Ugo, args.Recursive)

case "chown":
    args, err := fs.ParseChown(line)
    if err != nil { return "", err }
    return r.fsSvc.Chown(args.ID, args.Path, args.User, args.Recursive)

case "loss":
    args, err := fs.ParseLoss(line)
    if err != nil { return "", err }
    return r.fsSvc.Loss(args.ID)

case "recovery":
    args, err := fs.ParseRecovery(line)
    if err != nil { return "", err }
    return r.fsSvc.Recovery(args.ID)

case "unmount":
    args, err := disk.ParseUnmount(line)
    if err != nil { return "", err }
    return r.diskSvc.Unmount(args.ID)
```

**Tareas:**
- [ ] Implementar todos los parsers faltantes
- [ ] Crear métodos correspondientes en FsService
- [ ] Agregar casos al switch

---

## 🎛️ 4. SERVICES - Lógica de Negocio

### 4.1 command/fs/service.go (FsService)

**Métodos a agregar:**

```go
func (s *FsService) Remove(id string, path string) (string, error)
func (s *FsService) Edit(id string, path string, content string) (string, error)
func (s *FsService) Rename(id string, path string, newName string) (string, error)
func (s *FsService) Copy(id string, src string, dest string) (string, error)
func (s *FsService) Move(id string, src string, dest string) (string, error)
func (s *FsService) Find(id string, path string, pattern string) (string, error)
func (s *FsService) Chmod(id string, path string, ugo string, recursive bool) (string, error)
func (s *FsService) Chown(id string, path string, user string, recursive bool) (string, error)
func (s *FsService) Loss(id string) (string, error)
func (s *FsService) Recovery(id string) (string, error)
```

**Cada método debe:**
- [ ] Validar sesión activa
- [ ] Parsear path a []string
- [ ] Llamar al repositorio
- [ ] Si es EXT3, registrar en journal (excepto Loss/Recovery)
- [ ] Retornar mensaje de éxito

---

### 4.2 command/disk/service.go (DiskService)

**Métodos a agregar/modificar:**

```go
func (s *DiskService) FDiskAdd(path string, name string, add int, unit string) (string, error)
func (s *DiskService) FDiskDelete(path string, name string, deleteType string) (string, error)
func (s *DiskService) Unmount(id string) (string, error)
```

**Tareas:**
- [ ] Implementar FDISK -add
- [ ] Implementar FDISK -delete (fast/full)
- [ ] Implementar UNMOUNT

---

## 🌐 5. MAIN - Wiring de Dependencias

### 5.1 cmd/server/main.go

**Agregar ViewerController:**

```go
// Después de reportSvc
viewerCtrl := controllers.NewViewerController(
    fsRepoAdapter,
    mountStoreAdapter,
    sessionAdapter,
)

// Modificar SetupRouter
r := router.SetupRouter(cfg, cs, ss, rs, viewerCtrl)
```

**Tareas:**
- [ ] Instanciar ViewerController
- [ ] Pasar como parámetro a SetupRouter

---

## 📊 6. REPORTES - Extensión (Opcional)

### 6.1 Reporte de Journal

**En `command/reports/service.go`:**

```go
case "journal":
    return r.generator.GenerateJournal(id, out)
```

**En `storage/adapters/report_generator.go`:**

```go
func (g *ReportGenerator) GenerateJournal(id string, out string) (string, error) {
    entries, err := g.fsRepo.JournalList(id)
    // Generar tabla HTML/DOT con las entradas
    // ...
}
```

**Tareas:**
- [ ] Agregar caso "journal" al switch de reportes
- [ ] Implementar GenerateJournal en el adaptador
- [ ] Generar salida en formato tabla (HTML, DOT, etc.)

---

## 🧪 7. TESTS - Validación

### 7.1 Tests Unitarios

**En `tests/fs_service_test.go`:**

```go
func TestRemove(t *testing.T)
func TestEdit(t *testing.T)
func TestRename(t *testing.T)
func TestCopy(t *testing.T)
func TestMove(t *testing.T)
func TestFind(t *testing.T)
func TestChmod(t *testing.T)
func TestChown(t *testing.T)
func TestLossRecovery(t *testing.T)
```

**Tareas:**
- [ ] Crear tests para cada comando nuevo
- [ ] Validar que journal se registra correctamente en EXT3
- [ ] Validar recovery después de loss

---

### 7.2 Test E2E

**En `tests/e2e_script_test.go`:**

```go
func TestP2FullWorkflow(t *testing.T) {
    script := `
    mkdisk -size=10 -unit=M -path=test.mia
    fdisk -size=5 -unit=M -path=test.mia -type=P -name=part1
    mount -path=test.mia -name=part1
    mkfs -id=XXXX -fs=3fs -type=full
    login -id=XXXX -user=root -pass=123
    mkdir -id=XXXX -path=/carpeta1
    mkfile -id=XXXX -path=/carpeta1/archivo.txt -size=10
    edit -id=XXXX -path=/carpeta1/archivo.txt -contenido=/tmp/test.txt
    copy -id=XXXX -path=/carpeta1 -destino=/carpeta2
    move -id=XXXX -path=/carpeta2 -destino=/carpeta3
    chmod -id=XXXX -path=/carpeta3 -ugo=777
    find -id=XXXX -path=/ -name=*.txt
    loss -id=XXXX
    recovery -id=XXXX
    `
    // Ejecutar y validar
}
```

**Tareas:**
- [ ] Crear script E2E completo con todos los comandos
- [ ] Validar que no hay errores
- [ ] Validar que recovery restaura correctamente

---

## 📦 8. UTILS - Helpers

### 8.1 utils/permissions.go

**Ya creado, agregar helpers:**

```go
func HasPermission(perm byte, owner bool, group bool, other bool, read bool, write bool, exec bool) bool
func ParseUGO(ugo string) ([3]byte, error)
func FormatUGO(ugo [3]byte) string
```

**Tareas:**
- [ ] Implementar validación de permisos UGO
- [ ] Helpers para conversión entre string y [3]byte

---

### 8.2 utils/calc.go

**Ya actualizado con:**
- ✅ CalcNExt3()
- ✅ ComputeLayoutExt3()

**Validar:**
- [ ] Fórmula correcta de n para EXT3
- [ ] Offsets correctos (SB → Journal → Bitmaps → Inodos → Bloques)

---

## 📝 9. DOCUMENTACIÓN

### 9.1 README.md

**Actualizar con:**
- [ ] Descripción de comandos nuevos
- [ ] Ejemplos de uso de EXT3 + Journal
- [ ] API REST endpoints
- [ ] Ejemplos de scripts .smia

---

### 9.2 API Documentation

**Crear `API.md` con:**
- [ ] Endpoints de autenticación (`/api/auth/login`, `/api/auth/logout`)
- [ ] Endpoints de visualizador (`/api/disks`, `/api/fs/:id/*`)
- [ ] Endpoints de journal (`/api/journal/:id`)
- [ ] Ejemplos de requests/responses

---

## ✅ RESUMEN DE PRIORIDADES

### 🔴 Alta Prioridad (Core P2)

1. **MKFS con EXT3** - Base para todo lo demás
2. **Journal** - Append/List/Clear en diskio
3. **REMOVE, EDIT, RENAME** - Comandos básicos de archivos
4. **CHMOD, CHOWN** - Permisos
5. **LOSS, RECOVERY** - Journaling funcional
6. **Integrar todo en runner.go** - Hacer comandos ejecutables

### 🟡 Media Prioridad

7. **COPY, MOVE, FIND** - Comandos avanzados
8. **FDISK -add/-delete** - Extensión de FDISK
9. **UNMOUNT** - Gestión de montajes
10. **ViewerController en main.go** - API REST funcional

### 🟢 Baja Prioridad (Extras)

11. **Reporte de Journal** - Visualización
12. **Tests E2E completos** - Validación exhaustiva
13. **Documentación API** - Para frontend

---

## 📊 Progreso Estimado

```
Total de tareas: ~80
Completadas: ~15 (estructura base)
Pendientes: ~65

Tiempo estimado: 20-30 horas
```

---

## 🚀 Orden Sugerido de Implementación

1. **Día 1**: MKFS EXT3 + Journal (Append/List/Clear)
2. **Día 2**: REMOVE + EDIT + registro en Journal
3. **Día 3**: RENAME + CHMOD + CHOWN
4. **Día 4**: LOSS + RECOVERY (validar que funciona)
5. **Día 5**: COPY + MOVE + FIND
6. **Día 6**: FDISK -add/-delete + UNMOUNT
7. **Día 7**: Integración en runner + ViewerController
8. **Día 8**: Tests E2E + Debugging

---

**Última actualización:** 2025-10-19
