# ✅ Plan de Implementación P1→P2 (Carnet 84)
**Basado en el script del calificador P1**

---

## 0) Parser y Contratos de Error (aplica a TODO)

### Reglas del Parser
* Acepta **parámetros en cualquier orden**; `-path` puede ir sin o con comillas.
* Normaliza **case-insensitive** de comando y flags (`Mkdisk`, `mkDisk`, `-UNIT`…).
* Si un flag es desconocido (ej. `-param=x`, `-tamaño=`), responde con:
  ```
  ERROR PARAMETROS
  ```

### Mensajes de Error Exactos (CRÍTICO)

```go
// internal/errors/p1_errors.go
package errors

import "errors"

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
    ErrDiskNotExist      = errors.New("ERROR DISCO NO EXISTE")
    ErrGroupExists       = errors.New("ERROR YA EXISTE EL GRUPO")
    ErrUserExists        = errors.New("ERROR EL USUARIO YA EXISTE")
    ErrGroupNotExist     = errors.New("ERROR GRUPO NO EXISTE")
)
```

### Validador Central

```go
// internal/commands/parse/validate.go
package parse

func RequireAllowedFlags(got map[string]string, allowed ...string) error {
    allowedSet := make(map[string]bool)
    for _, a := range allowed {
        allowedSet[strings.ToLower(a)] = true
    }

    for key := range got {
        if !allowedSet[strings.ToLower(key)] {
            return ErrParams
        }
    }
    return nil
}
```

---

## 1) MKDISK / RMDISK

### MKDISK Requisitos
* `-size` (int), `-unit` = `B|K|M` (default M), `-fit` = `FF|BF|WF` (default FF), `-path` (ruta absoluta)
* Debe **crear carpetas** intermedias si no existen (`os.MkdirAll`)

### Errores
* `mkdisk -param=x ...` → **ERROR PARAMETROS**
* `mkdisk -tamaño=...` (flag mal escrito) → **ERROR PARAMETROS**

### RMDISK
* Si no existe: **ERROR DISCO NO EXISTE**
* Si existe: borrar y confirmar OK

### Archivos a Modificar
- `internal/disk/manager.go::Mkdisk`
- `internal/disk/io.go`
- `internal/commands/mkdisk.go`
- `internal/commands/rmdisk.go`
- `internal/commands/parser.go`

---

## 2) FDISK (ADD) — Primarias / Extendida / Lógicas

### Reglas Críticas
* **4 primarias** máx en el MBR
* **1 extendida** máx
* Lógicas dentro de la extendida con **EBRs encadenados**
* `-unit` = `b|k|m` (default k o m)
* `-fit` respeta `BF`, `FF`, `WF`

### Errores
* Falta de espacio: **ERROR FALTA ESPACIO**
* 5ta primaria: **ERROR LIMITE PARTICION**

### Implementación Sugerida

```go
func (m *Manager) AddPartition(path, name string, size int64, unit Unit, ptype PType, fit Fit) error
func (m *Manager) AddLogical(path, extName, logName string, size int64, unit Unit, fit Fit) error
```

### Archivos a Modificar
- `internal/disk/mbr.go` (layout, huecos libres)
- `internal/disk/ebr.go` (insertar EBR y encadenado)
- `internal/disk/alloc.go` (aplicar fits)
- `internal/commands/fdisk.go`

---

## 3) MOUNT / MOUNTED — **Formato de ID Exacto**

### Patrón de IDs (CRÍTICO)

```
<ID> = <TermCarnet><CorrelativoPorDisco><LetraDisco>
TermCarnet = "84"
CorrelativoPorDisco = 1..n (por cada disco)
LetraDisco = A para Disco1, B para Disco3, C para Disco5
```

**Ejemplos:**
- Disco1: `841A`, `842A`, `843A`
- Disco3: `841B`, `842B`
- Disco5: `841C`, `842C`

### Errores
* Partición ya montada: **ERROR PARTICION YA MONTADA**
* Nombre no existe: **ERROR PARTICION NO EXISTE**

### Implementación (YA HECHA ✅)

```go
// internal/commands/mount_index.go
const carnetSuffix = "84"

func (m *memoryIndex) GenerateID(diskPath string) string {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, ok := m.diskLetter[diskPath]; !ok {
        nextLetter := rune('A' + len(m.diskLetter))
        m.diskLetter[diskPath] = nextLetter
    }

    m.diskSeq[diskPath]++

    return fmt.Sprintf("%s%d%c", carnetSuffix, m.diskSeq[diskPath], m.diskLetter[diskPath])
}
```

---

## 4) REP (Reportes) — Nombres y Comportamiento

### Comando
```bash
rep -id=841A -name=disk|mbr|inode|block|bm_inode|bm_block|sb|file|ls|tree -path=...
```

### Errores
* ID no existe: **ERROR ID NO ENCONTRADO**

### Reportes a Generar (Formato DOT)
* `disk` - Imagen/resumen del disco (MBR + particiones/EBR)
* `mbr` - Detalle de MBR y EBRs
* `inode` - Tabla de inodos
* `block` - Bloques de datos
* `bm_inode` - Bitmap de inodos
* `bm_block` - Bitmap de bloques
* `sb` - SuperBlock
* `file` - Requiere `-path_file_ls=/ruta/al/archivo`
* `ls` - Listado de carpeta (`-path_file_ls=/ruta/al/directorio`)
* `tree` - Árbol del FS desde `/`

### Archivos a Modificar
- `internal/commands/rep.go`
- `pkg/reports/*.go` (todos los generadores DOT)

### API P2 (Compatibilidad)
```
GET /api/reports/{mbr|tree|journal}?id=...
```
Retorna `text/plain` con DOT

---

## 5) MKFS EXT2 — **Obligatorio Crear `/users.txt`**

### Validación del Script
```bash
mkfs -type=full -id=841A
```

### Requisitos CRÍTICOS
* Crear `/users.txt` con contenido base:
  ```
  1,G,root
  1,U,root,root,123
  ```
* Actualizar **SuperBlock**: inodos/bloques libres disminuyen
* Bitmaps: marcar inodo y bloque usados

### Archivos a Modificar
- `internal/fs/ext2/mkfs.go` (YA HECHO ✅)
- `internal/fs/ext2/persistence.go` (YA HECHO ✅)

---

## 6) Sesión y Administración (login/logout/mkgrp/…)

### Flujo del Script
```bash
login -user=root -pass=123 -id=841A
mkgrp -name=usuarios
mkusr -user=user1 -pass=pass1 -grp=usuarios
cat -file1=/users.txt
logout
```

### Errores
* Login repetido: **ERROR SESION INICIADA**
* Logout sin sesión: **ERROR NO HAY SESION INICIADA**
* Grupo duplicado: **ERROR YA EXISTE EL GRUPO**
* Usuario duplicado: **ERROR EL USUARIO YA EXISTE**
* Grupo no existe: **ERROR GRUPO NO EXISTE**

### Formato `/users.txt`
```
<status>,G,<nombre>           # Grupo
<status>,U,<user>,<grp>,<pass>  # Usuario
```
Status: `1` = activo, `0` = borrado lógico

### Archivos a Implementar
- `internal/fs/ext2/users.go` (reemplaza `users_stub.go`)
- `internal/fs/ext3/users.go` (reemplaza `users_stub.go`)
- `internal/auth/session.go` (validar credenciales)

---

## 7) MKDIR / MKFILE / CAT

### Reglas
* `mkdir -path=/bin` sin padres: **ERROR NO EXISTEN LAS CARPETAS PADRES**
* `mkdir -p -path=/home/archivos/...` crea recursivo
* `mkfile -r -path=...` crea padres
* `mkfile -size=-45` → **ERROR NEGATIVO**
* `mkfile` sin ruta válida → **ERROR NO EXISTE RUTA**
* `cat -file1=/path`: imprime contenido

### Archivos a Modificar
- `internal/fs/path/lookup.go`
- `internal/fs/ext2/write.go`
- `internal/fs/ext2/mkdir.go`
- `internal/fs/ext2/read.go`
- `internal/commands/mkdir.go`
- `internal/commands/mkfile.go`
- `internal/commands/cat.go`

---

## 8) Reportes de FS (DOT)

### Generadores a Implementar
```go
// pkg/reports/
func GenerateInode(ctx, h, path) string      // Tabla inodos
func GenerateBlock(ctx, h, path) string      // Bloques
func GenerateBitmapInode(ctx, h, path) string
func GenerateBitmapBlock(ctx, h, path) string
func GenerateSuperblock(ctx, h, path) string
func GenerateFile(ctx, h, filePath, outPath) string
func GenerateLS(ctx, h, dirPath, outPath) string
func GenerateTree(ctx, h, path) string
```

---

## 🔧 P2: EXT3 + Journal (Sin Romper P1)

### Comandos EXT3
```bash
mkfs -fs=3fs -id=841A
journaling -id=841A    # Lista operaciones
recovery -id=841A      # Reejecuta operaciones
loss -id=841A          # Limpia journal
```

### Cálculo de N para EXT3
```
n = floor((partSize - sizeof(SuperBlock) - sizeof(Journal)) /
          (1 + 3 + sizeof(Inode) + 3*blockSize))

Donde:
- sizeof(SuperBlock) = 512
- sizeof(Journal) = 50 * 64 = 3200
- sizeof(Inode) = 128
```

### Archivos a Implementar
- `internal/fs/ext3/mkfs.go`
- `internal/fs/ext3/journal.go`
- `internal/commands/journaling.go`
- `internal/commands/recovery.go`
- `internal/commands/loss.go`

---

## 🧪 Smoke Test P1 (smoke_84.sh)

```bash
#!/usr/bin/env bash
set -euo pipefail
API=${API:-http://localhost:8080}

post(){
    curl -sS -X POST "$API/api/cmd/run" \
         -H 'Content-Type: application/json' \
         -d "{\"line\":\"$*\"}"
    echo
}

# MKDISK
post 'mkdisk -size=50 -unit=M -fit=FF -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia"'
post 'mkdisk -size=13 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco3.mia"'
post 'mkdisk -size=20 -unit=M -fit=WF -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia"'

# FDISK Disco1 (4 primarias)
post 'fdisk -type=P -unit=b -name=Part11 -size=10485760 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia" -fit=BF'
post 'fdisk -type=P -unit=k -name=Part12 -size=10240 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia" -fit=BF'
post 'fdisk -type=P -unit=M -name=Part13 -size=10 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia" -fit=BF'
post 'fdisk -type=P -unit=b -name=Part14 -size=10485760 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia" -fit=BF'

# FDISK Disco3 (3 primarias)
post 'fdisk -type=P -unit=m -name=Part31 -size=4 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco3.mia"'
post 'fdisk -type=P -unit=m -name=Part32 -size=4 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco3.mia"'
post 'fdisk -type=P -unit=m -name=Part33 -size=1 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco3.mia"'

# FDISK Disco5 (E + L + P)
post 'fdisk -type=E -unit=k -name=Part51 -size=5120 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -fit=BF'
post 'fdisk -type=L -unit=k -name=Part52 -size=1024 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -fit=BF'
post 'fdisk -type=P -unit=k -name=Part53 -size=5120 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -fit=BF'
post 'fdisk -type=L -unit=k -name=Part54 -size=1024 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -fit=BF'
post 'fdisk -type=L -unit=k -name=Part55 -size=1024 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -fit=BF'
post 'fdisk -type=L -unit=k -name=Part56 -size=1024 -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -fit=BF'

# MOUNT - IDs con terminación 84
post 'mount -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia" -name=Part11' # -> 841A
post 'mount -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco1.mia" -name=Part12' # -> 842A
post 'mount -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco3.mia" -name=Part31' # -> 841B
post 'mount -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco3.mia" -name=Part32' # -> 842B
post 'mount -path="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos/Disco5.mia" -name=Part53' # -> 841C

post 'mounted'

# MKFS EXT2 (crea /users.txt)
post 'mkfs -type=full -id=841A'

# LOGIN/ADMIN
post 'login -user=root -pass=123 -id=841A'
post 'mkgrp -name=usuarios'
post 'mkusr -user=user1 -pass=pass1 -grp=usuarios'
post 'cat -file1=/users.txt'
post 'logout'
```

---

## 📋 Checklist de Implementación

### Fase 1: Errores y Validación (CRÍTICO)
- [ ] Crear `internal/errors/p1_errors.go` con mensajes exactos
- [ ] Crear `internal/commands/parse/validate.go`
- [ ] Reemplazar TODOS los errores en el código

### Fase 2: Comandos Core
- [ ] MKDISK: crear carpetas intermedias
- [ ] RMDISK: validar existencia
- [ ] FDISK: validar límites y espacios
- [ ] MOUNT: validar duplicados
- [ ] MKFS: soporte `-type=full`

### Fase 3: Administración
- [ ] Implementar `internal/fs/ext2/users.go`
- [ ] Login con validación de `/users.txt`
- [ ] mkgrp/rmgrp/mkusr/rmusr/chgrp

### Fase 4: Archivos
- [ ] MKDIR con validación de padres
- [ ] MKFILE con validación de tamaño
- [ ] CAT funcional

### Fase 5: Reportes
- [ ] Implementar todos los generadores DOT
- [ ] REP con todos los tipos

### Fase 6: Testing
- [ ] Crear `tools/smoke_84.sh`
- [ ] Ejecutar contra calificador P1
- [ ] Validar TODOS los IDs (841A, 842A, 841B...)

---

## 🧷 Resumen por Carpeta

### `internal/commands/`
- mkdisk.go, rmdisk.go, fdisk.go
- mount.go, mounted.go
- rep.go, mkfs.go
- login.go, logout.go
- mkgrp.go, rmgrp.go, mkusr.go, rmusr.go, chgrp.go
- mkdir.go, mkfile.go, cat.go
- adapter.go, parser.go
- `parse/validate.go` (NUEVO)

### `internal/disk/`
- manager.go, mbr.go, ebr.go
- alloc.go, io.go
- errors.go, mount_table.go

### `internal/fs/`
- fs.go
- `ext2/*.go` (mkfs, users, IO)
- `ext3/*.go` (mkfs, journal, users)

### `internal/errors/`
- `p1_errors.go` (NUEVO)

### `pkg/reports/`
- mbr.go, disk.go
- inode.go, block.go
- bm_inode.go, bm_block.go
- sb.go, file.go, ls.go, tree.go

### `cmd/server/`
- main.go (POST `/api/cmd/run` + GET `/api/reports/*`)

---

**¡Con esto pasarás el calificador P1 y tendrás base sólida para P2!**
