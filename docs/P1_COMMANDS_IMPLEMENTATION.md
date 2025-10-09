# P1 Commands Implementation Status
**Carnet:** 201905884
**Estudiante:** Santiago Julian Barrera Reyes
**Fecha:** 2025-10-08

---

## ✅ Implementación Completada

### 1. Comandos P1 Implementados

Todos los comandos P1 han sido agregados al sistema con tipos, parsers y handlers:

#### Gestión de Discos
- ✅ **rmdisk** - Eliminar disco
  - Ubicación: `internal/commands/types.go:447-457`
  - Handler: `internal/commands/handlers.go:367-379`
  - Parser: `internal/commands/parser.go:363-368`

#### Sesión
- ✅ **login** - Iniciar sesión
  - Ubicación: `internal/commands/types.go:459-478`
  - Handler: `internal/commands/handlers.go:381-393`
  - Parser: `internal/commands/parser.go:370-377`

- ✅ **logout** - Cerrar sesión
  - Ubicación: `internal/commands/types.go:480-488`
  - Handler: `internal/commands/handlers.go:395-405`
  - Parser: `internal/commands/parser.go:379-383`

#### Gestión de Grupos
- ✅ **mkgrp** - Crear grupo
  - Ubicación: `internal/commands/types.go:490-501`
  - Handler: `internal/commands/handlers.go:408-426`
  - Parser: `internal/commands/parser.go:385-390`

- ✅ **rmgrp** - Eliminar grupo
  - Ubicación: `internal/commands/types.go:503-514`
  - Handler: `internal/commands/handlers.go:428-446`
  - Parser: `internal/commands/parser.go:392-397`

#### Gestión de Usuarios
- ✅ **mkusr** - Crear usuario
  - Ubicación: `internal/commands/types.go:516-535`
  - Handler: `internal/commands/handlers.go:448-465`
  - Parser: `internal/commands/parser.go:399-406`

- ✅ **rmusr** - Eliminar usuario
  - Ubicación: `internal/commands/types.go:537-548`
  - Handler: `internal/commands/handlers.go:467-485`
  - Parser: `internal/commands/parser.go:408-413`

- ✅ **chgrp** - Cambiar grupo de usuario
  - Ubicación: `internal/commands/types.go:550-565`
  - Handler: `internal/commands/handlers.go:487-505`
  - Parser: `internal/commands/parser.go:415-421`

#### Archivos
- ✅ **cat** - Leer archivo
  - Ubicación: `internal/commands/types.go:567-578`
  - Handler: `internal/commands/handlers.go:508-527`
  - Parser: `internal/commands/parser.go:423-428`

#### Reportes
- ✅ **rep** - Generar reportes
  - Ubicación: `internal/commands/rep.go` (ya existía)
  - Tipos soportados: disk, mbr, inode, block, bm_inode, bm_block, sb, file, ls, tree

---

## 🏗️ Infraestructura Agregada

### 1. Session Manager
**Ubicación:** `internal/auth/session.go`

```go
type SessionManager struct {
    mu      sync.RWMutex
    current *Session // Solo una sesión activa (P1)
    fs      fs.FS
}
```

**Funcionalidad:**
- `IsActive()` - Verifica si hay sesión activa
- `Login(ctx, user, pass, mountID)` - Inicia sesión
- `Logout()` - Cierra sesión
- `CurrentUser()` - Retorna usuario actual
- `CurrentMountID()` - Retorna ID de montaje actual

### 2. Reports Generator
**Ubicación:** `internal/reports/reports.go`

```go
type Generator interface {
    GenerateDiskReport(ctx, diskPath, outputPath) (string, error)
    GenerateMBRReport(ctx, diskPath, outputPath) (string, error)
    GenerateInodeReport(ctx, h, outputPath) (string, error)
    // ... más reportes
}
```

**Estado:** Stubs implementados, retornan "no implementado"

### 3. User/Group Management en FS
**Ubicación:** `internal/fs/fs.go:33-38`

Agregadas interfaces al FS:
```go
AddGroup(ctx, h, name) error
RemoveGroup(ctx, h, name) error
AddUser(ctx, h, user, pass, group) error
RemoveUser(ctx, h, user) error
ChangeUserGroup(ctx, h, user, group) error
```

**Stubs creados:**
- `internal/fs/ext2/users_stub.go` - Implementación EXT2
- `internal/fs/ext3/users_stub.go` - Implementación EXT3

---

## 🔄 Formato de IDs de Montaje P1

**Implementación:** `internal/commands/mount_index.go:85-105`

### Algoritmo
```go
func GenerateID(diskPath string) string {
    // Asignar letra al disco (A, B, C...)
    if _, ok := m.diskLetter[diskPath]; !ok {
        nextLetter := rune('A' + len(m.diskLetter))
        m.diskLetter[diskPath] = nextLetter
    }

    // Incrementar correlativo para este disco
    m.diskSeq[diskPath]++

    // Formato: <84><correlativo><letra>
    return fmt.Sprintf("%s%d%c", "84", m.diskSeq[diskPath], m.diskLetter[diskPath])
}
```

### Ejemplos de IDs Generados
- Primer montaje en Disco1 → **841A**
- Segundo montaje en Disco1 → **842A**
- Tercer montaje en Disco1 → **843A**
- Primer montaje en Disco3 → **841B**
- Segundo montaje en Disco3 → **842B**

---

## 📝 Mensajes de Uso

Todos los comandos P1 tienen mensajes de uso actualizados en `internal/commands/parser.go:442-487`:

```go
usageMap := map[CommandName]string{
    CmdRmdisk: "rmdisk -path <ruta>",
    CmdLogin:  "login -user <usuario> -pass <password> -id <id>",
    CmdLogout: "logout",
    CmdMkgrp:  "mkgrp -name <nombre>",
    CmdRmgrp:  "rmgrp -name <nombre>",
    CmdMkusr:  "mkusr -user <usuario> -pass <password> -grp <grupo>",
    CmdRmusr:  "rmusr -user <usuario>",
    CmdChgrp:  "chgrp -user <usuario> -grp <grupo>",
    CmdCat:    "cat -file1 <ruta>",
    CmdRep:    "rep -id <id> -path <output> -name <tipo> [-path_file_ls <ruta>]",
}
```

---

## 🔧 Inicialización del Servidor

**Ubicación:** `cmd/server/main.go:51-63`

```go
// Inicializar sesión y reportes para P1
session := auth.NewSessionManager(fs2)
reportGen := reports.NewSimpleGenerator()

adapter := &commands.Adapter{
    FS2:     fs2,
    FS3:     fs3,
    DM:      dm,
    Index:   idx,
    State:   meta,
    Session: session,  // ✅ Agregado
    Reports: reportGen, // ✅ Agregado
}
```

---

## 🚧 Trabajo Pendiente (Prioridad ALTA)

### 1. Estandarizar Mensajes de Error P1
**Estado:** Pendiente
**Ubicación:** Crear `internal/errors/p1_errors.go`

**Mensajes requeridos exactos:**
```go
var (
    ErrParams            = errors.New("ERROR PARAMETROS")
    ErrPathNotFound      = errors.New("ERROR RUTA NO ENCONTRADA")
    ErrAlreadyExists     = errors.New("ERROR YA EXISTE")
    ErrPartitionLimit    = errors.New("ERROR LIMITE PARTICION")
    ErrNoSession         = errors.New("ERROR NO HAY SESION INICIADA")
    ErrSessionExists     = errors.New("ERROR SESION INICIADA")
    ErrNoSpace           = errors.New("ERROR FALTA ESPACIO")
    ErrAlreadyMounted    = errors.New("ERROR PARTICION YA MONTADA")
    ErrPartitionNotFound = errors.New("ERROR PARTICION NO EXISTE")
    ErrIDNotFound        = errors.New("ERROR ID NO ENCONTRADO")
    ErrNoParentFolders   = errors.New("ERROR NO EXISTEN LAS CARPETAS PADRES")
    ErrNegative          = errors.New("ERROR NEGATIVO")
    ErrPathDoesNotExist  = errors.New("ERROR NO EXISTE RUTA")
)
```

**Archivos a modificar:**
- `internal/disk/*.go`
- `internal/commands/*.go`
- `internal/fs/ext2/*.go`
- `internal/fs/ext3/*.go`

### 2. Implementar User/Group Management
**Estado:** Stubs creados
**Prioridad:** ALTA

**Archivos a completar:**
- `internal/fs/ext2/users_stub.go` → Cambiar a `users.go`
- `internal/fs/ext3/users_stub.go` → Cambiar a `users.go`

**Funcionalidad requerida:**
1. Leer `/users.txt` del filesystem
2. Parsear formato: `<id>,G,<name>` y `<gid>,U,<user>,<group>,<pass>`
3. Validar operaciones (grupo existe, usuario no duplicado, etc.)
4. Escribir cambios de vuelta a `/users.txt`
5. Para EXT3: Registrar operaciones en journal

### 3. Implementar Generadores de Reportes
**Estado:** Stubs creados
**Prioridad:** ALTA

**Archivo:** `internal/reports/reports.go`

**Reportes a implementar:**
- `disk` - Estructura del disco
- `mbr` - Master Boot Record
- `inode` - Tabla de inodos
- `block` - Bloques de datos
- `bm_inode` - Bitmap de inodos
- `bm_block` - Bitmap de bloques
- `sb` - SuperBlock
- `file` - Contenido de archivo
- `ls` - Listado de directorio
- `tree` - Árbol completo del FS

**Formato:** DOT (Graphviz)

### 4. Agregar soporte para mkfs -type=full
**Estado:** Pendiente
**Ubicación:** `internal/commands/types.go` MkfsCommand

**Cambio requerido:**
```go
type MkfsCommand struct {
    BaseCommand
    ID     string
    FSKind string // 2fs|3fs
    Type   string // full (P1)
}

func (c *MkfsCommand) Validate() error {
    // ... validación existente ...

    // Normalizar: -type=full → -fs=2fs
    if c.Type == "full" {
        c.FSKind = "2fs"
    }

    // ... resto de validación ...
}
```

---

## ✅ Verificación de Compilación

```bash
$ go build -o bin/godisk ./cmd/server
# ✅ Sin errores
```

---

## 🧪 Próximos Pasos

1. **Crear `internal/errors/p1_errors.go`** con mensajes exactos
2. **Reemplazar** todos los errores en el código
3. **Implementar** user/group management en EXT2/EXT3
4. **Implementar** generadores de reportes DOT
5. **Agregar** soporte `-type=full` en mkfs
6. **Ejecutar** smoke test P1 completo
7. **Validar** con script del calificador

---

## 📊 Estadísticas

- **Comandos P1 agregados:** 10
- **Archivos creados:** 4
- **Archivos modificados:** 10
- **Líneas de código agregadas:** ~800
- **Estado de compilación:** ✅ EXITOSO

---

**Nota:** Los stubs permiten compilación exitosa. La funcionalidad completa requiere implementar los TODOs marcados en los archivos stub.
