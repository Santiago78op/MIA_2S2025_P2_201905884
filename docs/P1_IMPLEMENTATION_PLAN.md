# Plan de Implementación P1 → P2
**Carnet:** 201905884
**Estudiante:** Santiago Julian Barrera Reyes

---

## 🎯 Estrategia: P1 Primero, P2 Después

Dado que P1 y P2 tienen **requisitos incompatibles** en algunos aspectos (especialmente formato de IDs), la estrategia es:

1. ✅ **Implementar P1 COMPLETO** para pasar el calificador
2. ✅ **Luego extender a P2** sin romper P1

---

## ✅ Progreso Actual

### Completado
- ✅ **Sistema de IDs P1** (`841A`, `842A`, `841B` en lugar de `vd84`)
  - Modificado: `internal/commands/mount_index.go`
  - Modificado: `internal/commands/handlers.go`
  - Formato: `<Carnet><Correlativo><LetraDisco>`

- ✅ **Infraestructura base:**
  - Disk Manager (MBR/EBR correctos)
  - EXT2 con `users.txt`
  - EXT3 con journal
  - Compilación exitosa

### Pendiente para P1

#### 1. Comandos Faltantes (CRÍTICO)
```bash
# Gestión de discos
rmdisk -path=<path>

# Sesión
login -user=<user> -pass=<pass> -id=<id>
logout

# Grupos
mkgrp -name=<name>
rmgrp -name=<name>

# Usuarios
mkusr -user=<user> -pass=<pass> -grp=<grp>
rmusr -user=<user>
chgrp -user=<user> -grp=<grp>

# Archivos
cat -file1=<path>

# Reportes
rep -id=<id> -path=<output> -name=<type>
# Types: disk, mbr, inode, block, bm_inode, bm_block, sb, file, ls, tree
```

#### 2. Mensajes de Error Exactos (CRÍTICO)
Deben ser **EXACTAMENTE** estos textos:

```go
const (
    ErrParams             = "ERROR PARAMETROS"
    ErrPathNotFound       = "ERROR RUTA NO ENCONTRADA"
    ErrAlreadyExists      = "ERROR YA EXISTE ..."
    ErrPartitionLimit     = "ERROR LIMITE PARTICION"
    ErrNoSession          = "ERROR NO HAY SESION INICIADA"
    ErrSessionExists      = "ERROR SESION INICIADA"
    ErrNoSpace            = "ERROR FALTA ESPACIO"
    ErrAlreadyMounted     = "ERROR PARTICION YA MONTADA"
    ErrPartitionNotFound  = "ERROR PARTICION NO EXISTE"
    ErrIDNotFound         = "ERROR ID NO ENCONTRADO"
    ErrNoParentFolders    = "ERROR NO EXISTEN LAS CARPETAS PADRES"
    ErrNegative           = "ERROR NEGATIVO"
    ErrPathDoesNotExist   = "ERROR NO EXISTE RUTA"
)
```

#### 3. Parámetro mkfs -type
```bash
# P1 usa:
mkfs -type=full -id=841A

# Implementar:
- Si -type=full → EXT2 completo
- Mantener compatibilidad con -fs=2fs|3fs para P2
```

---

## 📋 Plan de Implementación Detallado

### Fase 1: Comandos Críticos P1 (1-2 días)

#### 1.1 RMDISK
**Ubicación:** `internal/commands/rmdisk.go`

```go
type RmdiskCommand struct {
    BaseCommand
    Path string
}

func (c *RmdiskCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
    // Validar que el disco existe
    if _, err := os.Stat(c.Path); err != nil {
        return "", errors.New("ERROR DISCO NO EXISTE")
    }

    // Eliminar archivo
    if err := os.Remove(c.Path); err != nil {
        return "", err
    }

    return fmt.Sprintf("rmdisk OK path=%s", c.Path), nil
}
```

#### 1.2 LOGIN/LOGOUT
**Ubicación:** `internal/fs/auth/session.go`

```go
type SessionManager struct {
    mu      sync.RWMutex
    current *Session // Solo una sesión activa
}

type Session struct {
    User      string
    MountID   string
    Timestamp time.Time
}

func (sm *SessionManager) Login(user, pass, mountID string) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    if sm.current != nil {
        return errors.New("ERROR SESION INICIADA")
    }

    // Validar en /users.txt
    if !sm.validateUser(mountID, user, pass) {
        return errors.New("ERROR CREDENCIALES INVALIDAS")
    }

    sm.current = &Session{User: user, MountID: mountID}
    return nil
}
```

#### 1.3 MKGRP/RMGRP/MKUSR/RMUSR/CHGRP
**Ubicación:** `internal/fs/ext2/users.go`

```go
// Formato /users.txt:
// 1,G,root
// 1,U,root,root,123

type UserManager struct{}

func (um *UserManager) AddGroup(name string) error {
    // Leer /users.txt
    // Validar que no exista: ERROR YA EXISTE EL GRUPO
    // Agregar línea: <id>,G,<name>
    // Escribir /users.txt
}

func (um *UserManager) AddUser(user, pass, grp string) error {
    // Validar grupo existe: ERROR GRUPO NO EXISTE
    // Validar usuario no existe: ERROR EL USUARIO YA EXISTE
    // Agregar línea: <gid>,U,<user>,<grp>,<pass>
}
```

#### 1.4 CAT
**Ubicación:** `internal/commands/cat.go`

```go
type CatCommand struct {
    BaseCommand
    File1 string // -file1=<path>
}

func (c *CatCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
    // Requiere sesión activa
    if !adapter.Session.IsActive() {
        return "", errors.New("ERROR NO HAY SESION INICIADA")
    }

    // Leer archivo del FS
    content, _, err := adapter.FS.ReadFile(ctx, handle, c.File1)
    if err != nil {
        return "", err
    }

    return string(content), nil
}
```

#### 1.5 REP
**Ubicación:** `internal/commands/rep.go`

```go
type RepCommand struct {
    BaseCommand
    ID          string // -id=841A
    Path        string // -path=/output/file.jpg
    Name        string // -name=mbr|disk|inode|...
    PathFileLS  string // -path_file_ls=/ruta (para file/ls)
}

func (c *RepCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
    // Validar ID existe
    if _, ok := adapter.Index.GetRef(c.ID); !ok {
        return "", errors.New("ERROR ID NO ENCONTRADO")
    }

    switch c.Name {
    case "disk":
        return generateDiskReport(c.ID, c.Path)
    case "mbr":
        return generateMBRReport(c.ID, c.Path)
    case "inode":
        return generateInodeReport(c.ID, c.Path)
    // ... etc
    }
}
```

### Fase 2: Mensajes de Error Estandarizados (1 día)

**Crear:** `internal/errors/p1_errors.go`

```go
package errors

import "errors"

// Errores P1 - TEXTOS EXACTOS requeridos por calificador
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

Luego **reemplazar** todos los errores en:
- `internal/disk/*.go`
- `internal/commands/*.go`
- `internal/fs/ext2/*.go`

### Fase 3: Parámetro mkfs -type (30 min)

**Modificar:** `internal/commands/types.go`

```go
type MkfsCommand struct {
    BaseCommand
    ID     string
    FSKind string // -fs=2fs|3fs (P2)
    Type   string // -type=full (P1)
}

func (c *MkfsCommand) Validate() error {
    if c.ID == "" {
        return ErrParams
    }

    // Normalizar: -type=full → -fs=2fs
    if c.Type == "full" {
        c.FSKind = "2fs"
    }

    kind := strings.ToLower(c.FSKind)
    if kind != "2fs" && kind != "3fs" {
        return ErrParams
    }

    return nil
}
```

### Fase 4: Smoke Test P1 (1 hora)

**Crear:** `tools/smoke_p1.sh`

```bash
#!/bin/bash
# Ejecuta el script EXACTO del calificador P1

API="http://localhost:8080/api/cmd/run"

run_cmd() {
    curl -sS -X POST "$API" \
        -H 'Content-Type: application/json' \
        -d "{\"line\": \"$1\"}" | jq -r '.output'
}

# Copiar EXACTAMENTE el script del calificador
run_cmd 'mkdisk -size=50 -unit=M -fit=FF -path=/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia'
run_cmd 'fdisk -type=P -unit=b -name=Part11 -size=10485760 -path=/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia -fit=BF'
# ... etc (todo el script)
```

---

## 🔄 Transición P1 → P2

Una vez que P1 pasa completamente:

### Cambios para P2 (SIN ROMPER P1):

1. **Mantener ambos formatos de ID:**
   ```go
   func (m *memoryIndex) GenerateID(diskPath string, version int) string {
       if version == 1 {
           // Formato P1: 841A
           return m.generateP1ID(diskPath)
       }
       // Formato P2: vd84
       return m.generateP2ID()
   }
   ```

2. **Agregar comandos P2:**
   - `journaling`, `recovery`, `loss` (ya están)
   - Endpoints HTTP para reportes DOT

3. **Frontend:**
   - Terminal web (ya existe)
   - Explorador de archivos
   - Visualización de reportes con Viz.js

---

## 📊 Checklist de Validación P1

### Pre-entrega
- [ ] Todos los comandos P1 funcionan
- [ ] Mensajes de error EXACTOS
- [ ] IDs formato `841A`, `842A`, `841B`
- [ ] Script del calificador pasa 100%
- [ ] `/users.txt` se crea automáticamente
- [ ] `login`/`logout` funciona
- [ ] Reportes generan archivos correctos

### Durante calificación
- [ ] `mkdisk` con errores devuelve "ERROR PARAMETROS"
- [ ] `fdisk` respeta límite de 4 primarias
- [ ] `mount` genera IDs correctos
- [ ] `mkfs -type=full` crea EXT2
- [ ] `cat /users.txt` muestra contenido
- [ ] Reportes MBR/Tree/etc funcionan

---

## 🚀 Cronograma Sugerido

| Día | Tarea | Horas |
|-----|-------|-------|
| 1 | Implementar rmdisk, login/logout | 4h |
| 1 | Implementar mkgrp/rmgrp/mkusr/rmusr | 4h |
| 2 | Implementar cat, rep (todos los tipos) | 6h |
| 2 | Estandarizar mensajes de error | 2h |
| 3 | Smoke test P1 completo | 2h |
| 3 | Ajustes finales y testing | 6h |

**Total:** ~24 horas de desarrollo

---

## 📝 Notas Importantes

1. **NO tocar lo que ya funciona:**
   - Disk Manager (MBR/EBR)
   - EXT2/EXT3 mkfs
   - Estructura de inodos/bloques

2. **Priorizar compatibilidad P1:**
   - Usar mensajes EXACTOS
   - IDs formato correcto
   - Todos los comandos del script

3. **Documentar cambios:**
   - Mantener IMPLEMENTATION_STATUS.md actualizado
   - Agregar ejemplos de uso
   - Registrar decisiones técnicas

---

**Siguiente paso:** Implementar `rmdisk` y sistema de sesión (login/logout)
