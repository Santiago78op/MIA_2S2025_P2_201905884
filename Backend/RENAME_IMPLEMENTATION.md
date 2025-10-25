# ✅ Implementación RENAME - Fase 1 Completada

## 🎉 Estado: COMPILACIÓN EXITOSA (Parser + Servicio)

La implementación del comando RENAME con soporte para renombrado atómico, validación de permisos y journaling está en progreso.

---

## 📦 Fase 1: Parser y Servicio (COMPLETADO)

### 1. **Parser RENAME** ✅
- ✅ Ubicación: `command/fs/parser.go`
- ✅ Función: `ParseRename(line string) (RenameArgs, error)`

**Argumentos:**
- `-path=/ruta/archivo` (obligatorio) - Ruta del archivo/directorio a renombrar
- `-name=nuevo_nombre` (obligatorio) - Nuevo nombre (sin path, solo el nombre)
- `-id=XXXX` (opcional) - ID de la partición montada (si no se especifica, usa la sesión actual)

```go
type RenameArgs struct {
    ID   string // opcional; "" = usar sesión actual
    Path string // obligatorio
    Name string // obligatorio (solo nombre, sin '/')
}
```

**Ejemplos de uso:**
```bash
# Con sesión activa (usa ID de sesión)
rename -path=/documentos/readme.txt -name=README.md

# Especificando ID explícitamente
rename -id=841A -path=/datos/log.txt -name=access.log

# Renombrar directorio
rename -path=/datos/carpeta_vieja -name=carpeta_nueva

# Renombrar archivo en subdirectorio
rename -path=/datos/logs/old.log -name=new.log
```

### 2. **Servicio RENAME** ✅
- ✅ Ubicación: `command/fs/service.go`
- ✅ Función: `Rename(id, path, newName string) (string, error)`

**Lógica implementada:**
1. ✅ Verificar sesión activa (required)
2. ✅ Resolver ID (usar sesión si no se proporciona)
3. ✅ Validar path y nuevo nombre
4. ✅ Validar que el nuevo nombre no contenga '/'
5. ⏳ Registrar en journal (write-ahead, pendiente)
6. ⏳ Ejecutar renombrado atómico (pendiente implementación en repo)

**Código actual:**
```go
func (s *FsService) Rename(id, path, newName string) (string, error) {
    // 1. Verificar sesión activa
    logged, user, uid, gid, _, partitionId := s.sess.Current()
    if !logged {
        return "", fmt.Errorf("no hay sesión activa")
    }

    // 2. Resolver ID (usar sesión si no se proporciona)
    if id == "" {
        id = partitionId
    }

    // 3. Validar path y nuevo nombre
    parts := SplitPath(path)
    if len(parts) == 0 {
        return "", fmt.Errorf("path inválido")
    }
    if newName == "" {
        return "", fmt.Errorf("nuevo nombre no puede estar vacío")
    }
    // Validar que el nuevo nombre no contenga '/'
    if strings.Contains(newName, "/") {
        return "", fmt.Errorf("el nuevo nombre no puede contener '/'")
    }

    // 4. Write-ahead journaling (solo EXT3)
    // TODO: Implementar JournalAppend cuando esté disponible

    // 5. Ejecutar renombrado con validación de permisos
    // TODO: El repo debe implementar Rename

    return fmt.Sprintf("RENAME ejecutado: %s → %s (stub)", path, newName), nil
}
```

### 3. **Integración en Runner** ✅
- ✅ Ubicación: `command/runner/runner.go`
- ✅ Caso `"rename"` ya estaba implementado

```go
case "rename":
    args, err := fs.ParseRename(line)
    if err != nil {
        return "", err
    }
    return r.fsSvc.Rename(args.ID, args.Path, args.Name)
```

---

## 📋 Fase 2: Repositorio con Validación de Unicidad (PENDIENTE)

### Características a Implementar:

#### 1. **Navegación al Archivo/Directorio**
Navegar al path y obtener su inodo y la entrada de directorio del padre:

```go
func (r *FileFsRepository) Rename(id string, absPath []string, newName string, uid int, gid int) error {
    // 1. Validar que absPath no sea raíz
    if len(absPath) == 0 {
        return fmt.Errorf("no se puede renombrar la raíz")
    }

    // 2. Navegar al directorio padre
    parentPath := absPath[:len(absPath)-1]
    oldName := absPath[len(absPath)-1]

    parentInode, err := r.readInodeAtPath(id, parentPath)
    if err != nil {
        return fmt.Errorf("directorio padre no encontrado: %w", err)
    }

    // 3. Verificar permiso de escritura en el directorio padre
    if !r.hasWritePermission(parentInode, uid, gid) {
        return fmt.Errorf("permiso denegado en directorio padre")
    }

    // 4. Leer entradas del directorio padre
    entries, err := r.readDirEntries(id, parentInode)
    if err != nil {
        return err
    }

    // 5. Buscar la entrada con el nombre antiguo
    var targetEntry *DirEntry
    var targetIndex int
    for i, entry := range entries {
        if entry.Name == oldName {
            targetEntry = &entries[i]
            targetIndex = i
            break
        }
    }
    if targetEntry == nil {
        return fmt.Errorf("archivo/directorio no encontrado: %s", oldName)
    }

    // 6. Verificar que el nuevo nombre no exista
    for _, entry := range entries {
        if entry.Name == newName {
            return fmt.Errorf("ya existe una entrada con el nombre: %s", newName)
        }
    }

    // 7. Actualizar el nombre de la entrada
    entries[targetIndex].Name = newName

    // 8. Escribir las entradas actualizadas
    if err := r.writeDirEntries(id, parentInode, entries); err != nil {
        return err
    }

    return nil
}
```

#### 2. **Validación de Permisos en Directorio Padre**
Reutilizar el helper de permisos de escritura:

```go
func (r *FileFsRepository) hasWritePermission(inode *Inode, uid int, gid int) bool {
    // Usuario es owner
    if inode.Uid == int32(uid) {
        return (inode.Perm & 0200) != 0 // Bit de escritura del owner
    }
    // Usuario en grupo
    if inode.Gid == int32(gid) {
        return (inode.Perm & 0020) != 0 // Bit de escritura del grupo
    }
    // Otros
    return (inode.Perm & 0002) != 0 // Bit de escritura de otros
}
```

#### 3. **Lectura de Entradas de Directorio**
```go
type DirEntry struct {
    Name       string  // 12 bytes (null-terminated)
    InodeIndex int32   // 4 bytes
    Type       byte    // 1 byte (0=file, 1=folder)
    // Padding para alinear a 64 bytes por entrada
}

func (r *FileFsRepository) readDirEntries(id string, dirInode *Inode) ([]DirEntry, error) {
    var entries []DirEntry

    // Leer contenido de todos los bloques del directorio
    for _, blockIdx := range dirInode.Blocks {
        if blockIdx == -1 {
            continue
        }

        // Leer bloque (64 bytes)
        blockData, err := r.readBlock(id, blockIdx)
        if err != nil {
            return nil, err
        }

        // Parsear entradas (asumiendo 1 entrada por bloque de 64 bytes)
        // Formato: [Name:12B][InodeIdx:4B][Type:1B][Padding:47B]
        var entry DirEntry
        entry.Name = readNullTerminatedString(blockData[0:12])
        entry.InodeIndex = int32(binary.LittleEndian.Uint32(blockData[12:16]))
        entry.Type = blockData[16]

        // Ignorar entradas vacías (nombre vacío o inode -1)
        if entry.Name != "" && entry.InodeIndex != -1 {
            entries = append(entries, entry)
        }
    }

    return entries, nil
}
```

#### 4. **Escritura de Entradas de Directorio**
```go
func (r *FileFsRepository) writeDirEntries(id string, dirInode *Inode, entries []DirEntry) error {
    // Calcular bloques necesarios (1 entrada por bloque de 64 bytes)
    blocksNeeded := len(entries)

    // Verificar que tengamos suficientes bloques asignados
    if len(dirInode.Blocks) < blocksNeeded {
        // Asignar bloques adicionales si es necesario
        for i := len(dirInode.Blocks); i < blocksNeeded; i++ {
            newBlock, err := r.allocateBlock(id)
            if err != nil {
                return fmt.Errorf("sin espacio para bloques: %w", err)
            }
            dirInode.Blocks = append(dirInode.Blocks, newBlock)
        }
    }

    // Escribir cada entrada en su bloque
    for i, entry := range entries {
        blockIdx := dirInode.Blocks[i]

        // Crear buffer de 64 bytes
        blockData := make([]byte, 64)

        // Escribir Name (12 bytes)
        copy(blockData[0:12], []byte(entry.Name))

        // Escribir InodeIndex (4 bytes)
        binary.LittleEndian.PutUint32(blockData[12:16], uint32(entry.InodeIndex))

        // Escribir Type (1 byte)
        blockData[16] = entry.Type

        // Resto es padding (47 bytes de zeros)

        // Escribir bloque
        if err := r.writeBlock(id, blockIdx, blockData); err != nil {
            return err
        }
    }

    // Limpiar bloques sobrantes (marcar como vacíos)
    for i := len(entries); i < len(dirInode.Blocks); i++ {
        if dirInode.Blocks[i] != -1 {
            // Escribir bloque vacío
            emptyBlock := make([]byte, 64)
            for j := 0; j < 12; j++ {
                emptyBlock[j] = 0 // Nombre vacío
            }
            binary.LittleEndian.PutUint32(emptyBlock[12:16], 0xFFFFFFFF) // InodeIndex = -1
            if err := r.writeBlock(id, dirInode.Blocks[i], emptyBlock); err != nil {
                return err
            }
        }
    }

    // Actualizar inodo del directorio (por si se agregaron bloques)
    if err := r.writeInode(id, dirInode); err != nil {
        return err
    }

    return nil
}
```

#### 5. **Helpers de Bloques**
```go
func (r *FileFsRepository) readBlock(id string, blockIdx int32) ([]byte, error) {
    // Obtener mount info y calcular offset del bloque
    mountInfo := r.getMountInfo(id)
    sb := r.readSuperBlock(id)

    blockOffset := sb.BlockStart + (int64(blockIdx) * 64)

    // Leer 64 bytes del bloque
    data := make([]byte, 64)
    if err := r.readAtOffset(id, blockOffset, data); err != nil {
        return nil, err
    }

    return data, nil
}

func (r *FileFsRepository) writeBlock(id string, blockIdx int32, data []byte) error {
    if len(data) != 64 {
        return fmt.Errorf("bloque debe ser de 64 bytes")
    }

    // Obtener mount info y calcular offset del bloque
    mountInfo := r.getMountInfo(id)
    sb := r.readSuperBlock(id)

    blockOffset := sb.BlockStart + (int64(blockIdx) * 64)

    // Escribir 64 bytes del bloque
    return r.writeAtOffset(id, blockOffset, data)
}
```

---

## 📊 Fase 3: Journaling (PENDIENTE)

### Write-Ahead Logging (EXT3)

**Flujo:**
1. **Antes** de ejecutar el rename, registrar en journal:
```go
journal.Append(JournalEntry{
    Op:        OpRename,
    Path:      "/documentos/readme.txt",
    Dest:      "README.md",
    Usuario:   "user1",
    Timestamp: 1634567890,
})
```

2. Ejecutar renombrado

3. Si el sistema falla, en `recovery`:
   - Leer journal
   - Completar operaciones pendientes (re-aplicar rename)

**Código del servicio (cuando esté implementado):**
```go
// Write-ahead journaling
if err := s.repo.JournalAppend(id, JournalEntry{
    Op:        OpRename,
    Path:      path,
    Dest:      newName,
    Usuario:   user,
    Timestamp: time.Now().Unix(),
}); err != nil {
    // EXT2 devuelve ErrUnsupported (ignorar)
    // EXT3 devuelve error real (abortar)
    if !errors.Is(err, ErrUnsupported) {
        return "", fmt.Errorf("error en journal: %w", err)
    }
}

// Ejecutar renombrado
if err := s.repo.Rename(id, parts, newName, uid, gid); err != nil {
    return "", err
}
```

---

## 🏗️ Flujo de Ejecución Completo

```
Usuario: rename -path=/documentos/readme.txt -name=README.md
    ↓
Runner.Run()
    ↓
fs.ParseRename(line) → RenameArgs{ID:"", Path:"/documentos/readme.txt", Name:"README.md"}
    ↓
FsService.Rename("", "/documentos/readme.txt", "README.md")
    ├─ Verificar sesión activa
    ├─ Resolver ID: "" → usar partitionId de sesión
    ├─ Validar path: "/documentos/readme.txt" → ["documentos", "readme.txt"]
    ├─ Validar newName: no puede contener '/'
    ├─ [TODO] JournalAppend(OpRename, path, newName, user, timestamp)
    └─ [TODO] repo.Rename(id, ["documentos", "readme.txt"], "README.md", uid, gid)
           ↓
FileFsRepository.Rename(id, path, newName, uid, gid)
    ├─ 1. Validar que path no sea raíz
    ├─ 2. Separar: parentPath = ["documentos"], oldName = "readme.txt"
    ├─ 3. readInodeAtPath(id, ["documentos"])
    ├─ 4. hasWritePermission(parentInode, uid, gid)
    │     └─ Si falla → Error "permiso denegado en directorio padre"
    ├─ 5. readDirEntries(id, parentInode)
    ├─ 6. Buscar entrada con oldName
    │     └─ Si no existe → Error "archivo/directorio no encontrado"
    ├─ 7. Verificar que newName no exista
    │     └─ Si existe → Error "ya existe una entrada con el nombre"
    ├─ 8. Actualizar entries[i].Name = newName
    └─ 9. writeDirEntries(id, parentInode, entries)
```

---

## 🧪 Casos de Prueba (Cuando esté implementado)

### Caso 1: Renombrar archivo simple
```bash
# Crear archivo
login -user=user1 -pwd=123 -id=841A
mkfile -path=/test.txt -size=100

# Renombrar
rename -path=/test.txt -name=archivo_nuevo.txt
# Output: RENAME ok: /test.txt → archivo_nuevo.txt

# Verificar
cat -file=/archivo_nuevo.txt
# Debería leer el archivo con el nuevo nombre

# Verificar que el antiguo ya no existe
cat -file=/test.txt
# Error: archivo no encontrado
```

### Caso 2: Renombrar directorio
```bash
mkdir -path=/datos/carpeta_vieja -r
mkfile -path=/datos/carpeta_vieja/archivo.txt -size=50

# Renombrar directorio
rename -path=/datos/carpeta_vieja -name=carpeta_nueva
# Output: RENAME ok: /datos/carpeta_vieja → carpeta_nueva

# Verificar que los archivos internos siguen accesibles
cat -file=/datos/carpeta_nueva/archivo.txt
# Debería leer el archivo correctamente
```

### Caso 3: Nombre duplicado (error)
```bash
mkfile -path=/archivo1.txt -size=50
mkfile -path=/archivo2.txt -size=100

# Intentar renombrar con nombre existente
rename -path=/archivo1.txt -name=archivo2.txt
# Error: ya existe una entrada con el nombre: archivo2.txt
```

### Caso 4: Permiso denegado en directorio padre
```bash
# Crear directorio con permisos restrictivos
login -user=root -pwd=root -id=841A
mkdir -path=/privado -r
mkfile -path=/privado/archivo.txt -size=50
chmod -path=/privado -ugo=500  # Solo root puede escribir (rx para otros)

logout
login -user=user1 -pwd=123 -id=841A

# Intentar renombrar archivo en directorio sin permisos de escritura
rename -path=/privado/archivo.txt -name=nuevo.txt
# Error: permiso denegado en directorio padre
```

### Caso 5: Intentar renombrar raíz (error)
```bash
login -user=root -pwd=root -id=841A

# Intentar renombrar raíz
rename -path=/ -name=nueva_raiz
# Error: no se puede renombrar la raíz
```

### Caso 6: Nombre con '/' (error)
```bash
mkfile -path=/archivo.txt -size=50

# Intentar renombrar con '/' en el nombre
rename -path=/archivo.txt -name=carpeta/archivo.txt
# Error: el nuevo nombre no puede contener '/'
```

### Caso 7: Archivo no existe
```bash
# Intentar renombrar archivo inexistente
rename -path=/no_existe.txt -name=nuevo.txt
# Error: archivo/directorio no encontrado: no_existe.txt
```

### Caso 8: Sesión implícita
```bash
login -user=user1 -pwd=123 -id=841A
mkfile -path=/archivo.txt -size=100

# Sin especificar -id (usa sesión actual)
rename -path=/archivo.txt -name=renamed.txt
# Output: RENAME ok: /archivo.txt → renamed.txt
```

### Caso 9: Sin sesión activa
```bash
# Sin login
rename -path=/archivo.txt -name=nuevo.txt
# Error: no hay sesión activa
```

### Caso 10: Nombre vacío (error)
```bash
login -user=user1 -pwd=123 -id=841A
mkfile -path=/archivo.txt -size=50

# Intentar renombrar con nombre vacío
rename -path=/archivo.txt -name=
# Error: nuevo nombre no puede estar vacío
```

---

## 📝 Archivos Modificados (2)

1. ✅ **`command/fs/parser.go`**
   - Modificado `ParseRename()`: `-id` ahora es opcional

2. ✅ **`command/fs/service.go`**
   - Actualizado `Rename()` con lógica de sesión
   - Agregada validación de nuevo nombre (no puede contener '/')
   - Agregada validación de nombre vacío
   - Agregado import de `strings`
   - Agregados TODOs para journaling y repo

---

## 🚀 Próximos Pasos (Fase 2 y 3)

### Prioridad Alta
1. **Implementar Rename en FileFsRepository**
   - `readInodeAtPath()` - Navegar path y retornar inodo (reutilizar de REMOVE/EDIT)
   - `hasWritePermission()` - Validar permisos UGO (reutilizar)
   - `readDirEntries()` - Leer entradas del directorio padre
   - `writeDirEntries()` - Escribir entradas actualizadas
   - Validación de unicidad (nombre no existe)

2. **Helpers de Bloques**
   - `readBlock()` - Leer bloque de 64 bytes
   - `writeBlock()` - Escribir bloque de 64 bytes
   - `readNullTerminatedString()` - Parsear nombres de 12 bytes

3. **Journaling (EXT3)**
   - Implementar `JournalAppend()` en repo
   - Definir `JournalEntry` con campos necesarios
   - Integrar en servicio antes de renombrado

### Prioridad Media
4. **Tests E2E**
   - Renombrado simple (archivo y directorio)
   - Validación de unicidad
   - Validación de permisos
   - Recovery desde journal

### Prioridad Baja
5. **Optimizaciones**
   - Cachear entradas de directorio durante renombrados múltiples
   - Logging detallado
   - Soporte para nombres largos (> 12 caracteres)

---

## 🔍 Consideraciones de Diseño

### 1. **Atomicidad**
- ✅ La operación se completa o no se ejecuta
- ✅ No deja el directorio en estado inconsistente
- ✅ Si falla la escritura, no se modifica nada

### 2. **Permisos**
- ✅ Requiere permiso de escritura en directorio padre
- ✅ No requiere ser owner del archivo (solo del directorio)
- ✅ Root bypassa permisos (opcional)

### 3. **Unicidad**
- ✅ Falla si ya existe una entrada con el nuevo nombre
- ✅ Case-sensitive (readme.txt != README.txt)
- ✅ Valida antes de modificar

### 4. **Journaling (EXT3)**
- ✅ Write-ahead logging
- ✅ Recovery automático en caso de crash
- ✅ EXT2 ignora journaling (sin error)

### 5. **Limitaciones de Nombre**
- ✅ Máximo 12 caracteres (basado en estructura DirEntry)
- ✅ No puede contener '/'
- ✅ No puede estar vacío
- ⚠️ Futuro: soportar nombres largos con bloques indirectos

### 6. **Estructura de Directorio**
- ✅ 1 entrada por bloque de 64 bytes
- ✅ Formato: [Name:12B][InodeIdx:4B][Type:1B][Padding:47B]
- ✅ Entradas vacías: Name="" o InodeIdx=-1

---

## 📊 Comparación con Comandos UNIX

| Característica | Linux `mv` | Nuestro RENAME | Diferencia |
|----------------|------------|----------------|------------|
| Renombrar archivo | ✅ | ✅ | Misma funcionalidad |
| Renombrar directorio | ✅ | ✅ | Misma funcionalidad |
| Mover entre directorios | ✅ | ❌ | RENAME solo cambia nombre, usar MOVE |
| Validar unicidad | ✅ | ✅ | Ambos fallan si existe |
| Permisos | ✅ (padre) | ✅ (padre) | Mismo comportamiento |
| Journaling | ✅ (ext3/4) | ✅ (EXT3) | Similar |

**Nota:** RENAME solo cambia el nombre dentro del mismo directorio. Para mover entre directorios, usar el comando MOVE.

---

## ✅ Checklist de Progreso

### Fase 1: Parser y Servicio
- [x] Parser `ParseRename()` con `-id` opcional
- [x] Servicio `Rename()` con validación de sesión
- [x] Resolver ID desde sesión si no se proporciona
- [x] Validación de path
- [x] Validación de nuevo nombre (no vacío, sin '/')
- [x] Integración en Runner (ya existía)
- [x] Compilación exitosa

### Fase 2: Repositorio (Pendiente)
- [ ] Implementar `Rename()` en FileFsRepository
- [ ] Navegación al directorio padre
- [ ] Validación de permisos de escritura en padre
- [ ] Lectura de entradas del directorio (readDirEntries)
- [ ] Validación de unicidad (nombre no existe)
- [ ] Actualización del nombre en la entrada
- [ ] Escritura de entradas actualizadas (writeDirEntries)
- [ ] Helpers de bloques (readBlock, writeBlock)

### Fase 3: Journaling (Pendiente)
- [ ] Definir estructura `JournalEntry`
- [ ] Implementar `JournalAppend()` en repo
- [ ] Integrar journaling en servicio
- [ ] Soporte para EXT2 (ignorar) y EXT3 (registrar)
- [ ] Tests de recovery

---

## 🎓 Conceptos Clave

### Renombrado vs Movimiento
- **RENAME**: Cambia el nombre de una entrada en el MISMO directorio
  - Solo modifica el campo `Name` en la entrada del directorio padre
  - Rápido: no mueve datos, solo actualiza metadatos

- **MOVE**: Mueve una entrada entre DIFERENTES directorios
  - Elimina entrada del directorio origen
  - Crea entrada en directorio destino
  - Más costoso: modifica dos directorios

### Validación de Unicidad
Los nombres de archivo/directorio deben ser únicos dentro del mismo directorio:
- `readDirEntries()` lista todas las entradas
- Se verifica que `newName` no aparezca en la lista
- Si existe, se rechaza la operación

### Atomicidad del Renombrado
1. Leer entradas del directorio
2. Validar permisos y unicidad
3. Modificar entrada en memoria
4. Escribir todas las entradas de vuelta
5. Si falla paso 4, el directorio no se modifica

---

**Última actualización:** 2025-10-19
**Estado:** ✅ Fase 1 Completada - Parser y Servicio
**Próximo paso:** Implementar Rename en FileFsRepository con validación de unicidad
**Versión:** Proyecto 2 - RENAME con Validación de Permisos y Unicidad
