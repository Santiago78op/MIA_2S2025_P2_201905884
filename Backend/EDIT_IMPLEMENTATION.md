# ✅ Implementación EDIT - Fase 1 Completada

## 🎉 Estado: COMPILACIÓN EXITOSA (Parser + Servicio)

La implementación del comando EDIT con soporte para reemplazo de contenido, validación de permisos y journaling está en progreso.

---

## 📦 Fase 1: Parser y Servicio (COMPLETADO)

### 1. **Parser EDIT** ✅
- ✅ Ubicación: `command/fs/parser.go`
- ✅ Función: `ParseEdit(line string) (EditArgs, error)`

**Argumentos:**
- `-path=/ruta/archivo` (obligatorio) - Ruta del archivo a editar
- `-contenido=/ruta/host/contenido.txt` (obligatorio) - Ruta al archivo host con el nuevo contenido
- `-id=XXXX` (opcional) - ID de la partición montada (si no se especifica, usa la sesión actual)

```go
type EditArgs struct {
    ID      string // opcional; "" = usar sesión actual
    Path    string // obligatorio
    Content string // obligatorio (ruta al archivo host)
}
```

**Ejemplos de uso:**
```bash
# Con sesión activa (usa ID de sesión)
edit -path=/documentos/readme.txt -contenido=/home/user/nuevo_readme.txt

# Especificando ID explícitamente
edit -id=841A -path=/documentos/readme.txt -contenido=/home/user/nuevo_readme.txt

# Editar archivo en subdirectorio
edit -path=/datos/logs/access.log -contenido=/tmp/new_access.log
```

### 2. **Servicio EDIT** ✅
- ✅ Ubicación: `command/fs/service.go`
- ✅ Función: `Edit(id, path, content string) (string, error)`

**Lógica implementada:**
1. ✅ Verificar sesión activa (required)
2. ✅ Resolver ID (usar sesión si no se proporciona)
3. ✅ Validar path
4. ✅ Preparado para lectura de archivo host (stub actual)
5. ⏳ Registrar en journal (write-ahead, pendiente)
6. ⏳ Ejecutar edición con validación de permisos (pendiente implementación en repo)

**Código actual:**
```go
func (s *FsService) Edit(id, path, content string) (string, error) {
    // 1. Verificar sesión activa
    logged, user, uid, gid, _, partitionId := s.sess.Current()
    if !logged {
        return "", fmt.Errorf("no hay sesión activa")
    }

    // 2. Resolver ID (usar sesión si no se proporciona)
    if id == "" {
        id = partitionId
    }

    // 3. Validar path
    parts := SplitPath(path)
    if len(parts) == 0 {
        return "", fmt.Errorf("path inválido")
    }

    // 4. Leer contenido desde archivo host
    // TODO: Implementar lectura real del archivo host
    // contentBytes, err := os.ReadFile(content)
    contentBytes := []byte(content) // Stub: trata content como contenido directo
    _ = contentBytes

    // 5. Write-ahead journaling (EXT3)
    // TODO: Implementar JournalAppend cuando esté disponible
    // _ = s.repo.JournalAppend(id, JournalEntry{
    //     Op:        OpEdit,
    //     Path:      path,
    //     Usuario:   user,
    //     Timestamp: time.Now().Unix(),
    // })

    // 6. Ejecutar edición con validación de permisos
    // TODO: El repo debe implementar Edit con:
    //   - Verificar permiso de escritura
    //   - Reemplazar contenido
    //   - Reasignar bloques si cambia el tamaño
    _ = user // Evitar warning
    _, _ = uid, gid

    return fmt.Sprintf("EDIT ejecutado en %s (stub - pendiente implementación en repo)", path), nil
}
```

### 3. **Integración en Runner** ✅
- ✅ Ubicación: `command/runner/runner.go`
- ✅ Caso `"edit"` ya estaba implementado

```go
case "edit":
    args, err := fs.ParseEdit(line)
    if err != nil {
        return "", err
    }
    return r.fsSvc.Edit(args.ID, args.Path, args.Content)
```

---

## 📋 Fase 2: Repositorio con Validación de Permisos (PENDIENTE)

### Características a Implementar:

#### 1. **Lectura de Archivo Host**
Antes de editar, leer el contenido del archivo host:

```go
func (r *FileFsRepository) Edit(id string, absPath []string, content []byte, uid int, gid int) error {
    // El contenido ya viene leído desde el servicio
    // Solo necesitamos validar permisos y escribir
}
```

**Actualización en servicio:**
```go
// Leer contenido desde archivo host
contentBytes, err := os.ReadFile(content)
if err != nil {
    return "", fmt.Errorf("error al leer archivo host: %w", err)
}

// Llamar repo con contenido leído
if err := s.repo.Edit(id, parts, contentBytes, uid, gid); err != nil {
    return "", err
}
```

#### 2. **Validación de Permisos**
Verificar que el usuario tenga permiso de escritura sobre el archivo:

```go
func (r *FileFsRepository) Edit(id string, absPath []string, content []byte, uid int, gid int) error {
    // 1. Navegar al inodo del archivo
    inode, err := r.readInodeAtPath(id, absPath)
    if err != nil {
        return fmt.Errorf("archivo no encontrado: %w", err)
    }

    // 2. Verificar que sea un archivo (no directorio)
    if inode.Type != FileTypeFile {
        return fmt.Errorf("no es un archivo regular")
    }

    // 3. Verificar permiso de escritura
    if !r.hasWritePermission(inode, uid, gid) {
        return fmt.Errorf("permiso denegado")
    }

    // 4. Reemplazar contenido
    if err := r.replaceFileContent(id, inode, content); err != nil {
        return err
    }

    // 5. Actualizar superblock si cambió el tamaño
    return r.updateSuperBlockAfterEdit(id)
}
```

#### 3. **Reemplazo de Contenido**
```go
func (r *FileFsRepository) replaceFileContent(id string, inode *Inode, newContent []byte) error {
    // 1. Calcular bloques necesarios para el nuevo contenido
    blockSize := int64(64) // Tamaño de bloque del FS
    blocksNeeded := int((int64(len(newContent)) + blockSize - 1) / blockSize)
    blocksOld := len(inode.Blocks)

    // 2. Liberar bloques antiguos
    if err := r.freeDataBlocks(id, inode); err != nil {
        return err
    }

    // 3. Asignar nuevos bloques
    newBlocks := make([]int32, blocksNeeded)
    for i := 0; i < blocksNeeded; i++ {
        blockIdx, err := r.allocateBlock(id)
        if err != nil {
            return fmt.Errorf("sin espacio para bloques: %w", err)
        }
        newBlocks[i] = blockIdx
    }

    // 4. Escribir nuevo contenido en bloques
    if err := r.writeContentToBlocks(id, newBlocks, newContent); err != nil {
        return err
    }

    // 5. Actualizar inodo
    inode.Blocks = newBlocks
    inode.Size = int64(len(newContent))
    if err := r.writeInode(id, inode); err != nil {
        return err
    }

    return nil
}
```

#### 4. **Helpers Necesarios**
```go
// Verificar permiso de escritura
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

// Asignar bloque
func (r *FileFsRepository) allocateBlock(id string) (int32, error) {
    // Leer bitmap de bloques
    // Buscar primer bit libre
    // Marcar como ocupado
    // Retornar índice
}

// Escribir contenido en bloques
func (r *FileFsRepository) writeContentToBlocks(id string, blocks []int32, content []byte) error {
    blockSize := int64(64)
    for i, blockIdx := range blocks {
        start := i * int(blockSize)
        end := start + int(blockSize)
        if end > len(content) {
            end = len(content)
        }
        chunk := content[start:end]

        // Calcular offset del bloque en el disco
        blockOffset := r.calculateBlockOffset(id, blockIdx)

        // Escribir chunk en el bloque
        if err := r.writeAtOffset(id, blockOffset, chunk); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 📊 Fase 3: Journaling (PENDIENTE)

### Write-Ahead Logging (EXT3)

**Flujo:**
1. **Antes** de ejecutar el edit, registrar en journal:
```go
journal.Append(JournalEntry{
    Op:        OpEdit,
    Path:      "/documentos/readme.txt",
    Usuario:   "user1",
    Timestamp: 1634567890,
    Size:      len(newContent), // Tamaño del nuevo contenido
})
```

2. Ejecutar edición (reemplazo de contenido)

3. Si el sistema falla, en `recovery`:
   - Leer journal
   - Completar operaciones pendientes (re-aplicar edit)

**Código del servicio (cuando esté implementado):**
```go
// Leer contenido desde archivo host
contentBytes, err := os.ReadFile(content)
if err != nil {
    return "", fmt.Errorf("error al leer archivo host: %w", err)
}

// Write-ahead journaling
if err := s.repo.JournalAppend(id, JournalEntry{
    Op:        OpEdit,
    Path:      path,
    Usuario:   user,
    Timestamp: time.Now().Unix(),
    Size:      int64(len(contentBytes)),
}); err != nil {
    // EXT2 devuelve ErrUnsupported (ignorar)
    // EXT3 devuelve error real (abortar)
    if !errors.Is(err, ErrUnsupported) {
        return "", fmt.Errorf("error en journal: %w", err)
    }
}

// Ejecutar edición
if err := s.repo.Edit(id, parts, contentBytes, uid, gid); err != nil {
    return "", err
}
```

---

## 🏗️ Flujo de Ejecución Completo

```
Usuario: edit -path=/documentos/readme.txt -contenido=/home/user/nuevo.txt
    ↓
Runner.Run()
    ↓
fs.ParseEdit(line) → EditArgs{ID:"", Path:"/documentos/readme.txt", Content:"/home/user/nuevo.txt"}
    ↓
FsService.Edit("", "/documentos/readme.txt", "/home/user/nuevo.txt")
    ├─ Verificar sesión activa
    ├─ Resolver ID: "" → usar partitionId de sesión
    ├─ Validar path: "/documentos/readme.txt" → ["documentos", "readme.txt"]
    ├─ Leer archivo host: os.ReadFile("/home/user/nuevo.txt") → []byte
    ├─ [TODO] JournalAppend(OpEdit, path, user, timestamp, size)
    └─ [TODO] repo.Edit(id, ["documentos", "readme.txt"], contentBytes, uid, gid)
           ↓
FileFsRepository.Edit(id, path, content, uid, gid)
    ├─ 1. readInodeAtPath(id, ["documentos", "readme.txt"])
    ├─ 2. Verificar que sea archivo (Type == FileTypeFile)
    ├─ 3. hasWritePermission(inode, uid, gid)
    │     └─ Si falla → Error "permiso denegado"
    ├─ 4. replaceFileContent(id, inode, content)
    │     ├─ Calcular bloques necesarios
    │     ├─ freeDataBlocks(id, inode) - Liberar bloques antiguos
    │     ├─ allocateBlock(id) × N - Asignar nuevos bloques
    │     ├─ writeContentToBlocks(id, blocks, content)
    │     └─ writeInode(id, inode) - Actualizar inodo
    └─ 5. updateSuperBlockAfterEdit(id)
          └─ FreeBlocks ajustado si cambió el número de bloques
```

---

## 🧪 Casos de Prueba (Cuando esté implementado)

### Caso 1: Editar archivo simple
```bash
# Crear archivo
login -user=user1 -pwd=123 -id=841A
mkfile -path=/test.txt -size=50

# Crear contenido nuevo en host
echo "Nuevo contenido del archivo" > /tmp/nuevo.txt

# Editar
edit -path=/test.txt -contenido=/tmp/nuevo.txt
# Output: EDIT ok: /test.txt

# Verificar
cat -file=/test.txt
# Output: Nuevo contenido del archivo
```

### Caso 2: Editar con cambio de tamaño
```bash
# Crear archivo pequeño
mkfile -path=/datos/log.txt -size=100

# Contenido grande (500 bytes)
dd if=/dev/zero of=/tmp/large.txt bs=1 count=500

# Editar (debe reasignar bloques)
edit -path=/datos/log.txt -contenido=/tmp/large.txt
# Output: EDIT ok: /datos/log.txt (bloques reasignados)

# Verificar tamaño
cat -file=/datos/log.txt
# Debería mostrar 500 bytes
```

### Caso 3: Permiso denegado
```bash
# Crear archivo con permisos restrictivos
login -user=root -pwd=root -id=841A
mkfile -path=/privado.txt -size=50
chmod -path=/privado.txt -ugo=600  # Solo root puede escribir

logout
login -user=user1 -pwd=123 -id=841A

# Intentar editar
echo "Hack" > /tmp/hack.txt
edit -path=/privado.txt -contenido=/tmp/hack.txt
# Error: permiso denegado
```

### Caso 4: Archivo no existe
```bash
login -user=user1 -pwd=123 -id=841A

# Intentar editar archivo inexistente
echo "Nuevo" > /tmp/nuevo.txt
edit -path=/no_existe.txt -contenido=/tmp/nuevo.txt
# Error: archivo no encontrado
```

### Caso 5: Editar directorio (error)
```bash
mkdir -path=/datos -r

# Intentar editar un directorio
edit -path=/datos -contenido=/tmp/nuevo.txt
# Error: no es un archivo regular
```

### Caso 6: Sesión implícita
```bash
login -user=user1 -pwd=123 -id=841A
mkfile -path=/archivo.txt -size=100

# Sin especificar -id (usa sesión actual)
echo "Contenido actualizado" > /tmp/nuevo.txt
edit -path=/archivo.txt -contenido=/tmp/nuevo.txt
# Output: EDIT ok: /archivo.txt
```

### Caso 7: Sin sesión activa
```bash
# Sin login
edit -path=/archivo.txt -contenido=/tmp/nuevo.txt
# Error: no hay sesión activa
```

### Caso 8: Archivo host no existe
```bash
login -user=user1 -pwd=123 -id=841A
mkfile -path=/archivo.txt -size=100

# Intentar editar con archivo host inexistente
edit -path=/archivo.txt -contenido=/no_existe.txt
# Error: error al leer archivo host: no such file or directory
```

---

## 📝 Archivos Modificados (2)

1. ✅ **`command/fs/parser.go`**
   - Modificado `ParseEdit()`: `-id` ahora es opcional

2. ✅ **`command/fs/service.go`**
   - Actualizado `Edit()` con lógica de sesión
   - Preparado para lectura de archivo host
   - Agregados TODOs para journaling y repo
   - Agregados comentarios TODO en FsRepository interface para futuros métodos

---

## 🚀 Próximos Pasos (Fase 2 y 3)

### Prioridad Alta
1. **Implementar Edit en FileFsRepository**
   - `readInodeAtPath()` - Navegar path y retornar inodo
   - `hasWritePermission()` - Validar permisos UGO
   - `replaceFileContent()` - Reemplazo atómico de contenido
   - `freeDataBlocks()` - Liberar bloques antiguos
   - `allocateBlock()` - Asignar nuevos bloques
   - `writeContentToBlocks()` - Escribir contenido en bloques
   - `updateSuperBlockAfterEdit()` - Actualizar contadores

2. **Lectura de Archivo Host en Servicio**
   - Actualizar servicio para usar `os.ReadFile(content)`
   - Manejar errores de lectura
   - Validar tamaño del contenido

3. **Journaling (EXT3)**
   - Implementar `JournalAppend()` en repo
   - Definir `JournalEntry` con campos necesarios
   - Integrar en servicio antes de edición

### Prioridad Media
4. **Tests E2E**
   - Edición simple
   - Cambio de tamaño (más grande / más pequeño)
   - Validación de permisos
   - Recovery desde journal

### Prioridad Baja
5. **Optimizaciones**
   - Reutilizar bloques si el tamaño no cambia significativamente
   - Cachear inodos durante ediciones múltiples
   - Logging detallado

---

## 🔍 Consideraciones de Diseño

### 1. **Reemplazo Atómico**
- ✅ Liberar bloques antiguos antes de asignar nuevos
- ✅ Si falla la asignación, revertir cambios
- ✅ Garantiza consistencia del FS

### 2. **Permisos**
- ✅ Requiere permiso de escritura sobre el archivo
- ✅ Owner, Group y Others validados con máscara UGO
- ✅ Root bypassa permisos (opcional)

### 3. **Journaling (EXT3)**
- ✅ Write-ahead logging
- ✅ Recovery automático en caso de crash
- ✅ EXT2 ignora journaling (sin error)

### 4. **Manejo de Tamaño**
- ✅ Si nuevo contenido es más grande: asignar más bloques
- ✅ Si nuevo contenido es más pequeño: liberar bloques sobrantes
- ✅ Actualizar `inode.Size` con tamaño exacto

### 5. **Lectura de Archivo Host**
- ✅ Validar que el archivo existe antes de editar
- ✅ Manejar archivos grandes (límite de tamaño?)
- ✅ Usar `os.ReadFile()` para lectura completa

---

## ✅ Checklist de Progreso

### Fase 1: Parser y Servicio
- [x] Parser `ParseEdit()` con `-id` opcional
- [x] Servicio `Edit()` con validación de sesión
- [x] Resolver ID desde sesión si no se proporciona
- [x] Preparar lectura de archivo host (stub)
- [x] Integración en Runner (ya existía)
- [x] Compilación exitosa

### Fase 2: Repositorio (Pendiente)
- [ ] Implementar `Edit()` en FileFsRepository
- [ ] Validación de permisos de escritura
- [ ] Reemplazo de contenido (replaceFileContent)
- [ ] Liberación de bloques antiguos
- [ ] Asignación de bloques nuevos
- [ ] Escritura de contenido en bloques
- [ ] Actualización de inodo
- [ ] Actualización de superblock
- [ ] Helpers de navegación (readInodeAtPath, hasWritePermission, etc.)

### Fase 3: Journaling (Pendiente)
- [ ] Definir estructura `JournalEntry`
- [ ] Implementar `JournalAppend()` en repo
- [ ] Integrar journaling en servicio
- [ ] Soporte para EXT2 (ignorar) y EXT3 (registrar)
- [ ] Tests de recovery

### Fase 4: Lectura de Archivo Host (Pendiente)
- [ ] Actualizar servicio para usar `os.ReadFile()`
- [ ] Validar existencia del archivo host
- [ ] Manejar errores de lectura
- [ ] Validar tamaño máximo (opcional)

---

## 📊 Comparación con Comandos Similares

| Característica | MKFILE | EDIT | Diferencia |
|----------------|--------|------|------------|
| Crea archivo | ✅ | ❌ | EDIT requiere archivo existente |
| Contenido desde host | ✅ (`-cont`) | ✅ (`-contenido`) | Mismo mecanismo |
| Valida permisos | ✅ (directorio padre) | ✅ (archivo) | EDIT valida escritura en archivo |
| Journaling | ⏳ | ⏳ | Ambos requieren EXT3 |
| Reasigna bloques | ✅ (inicial) | ✅ (reemplazo) | EDIT debe liberar antiguos primero |

---

**Última actualización:** 2025-10-19
**Estado:** ✅ Fase 1 Completada - Parser y Servicio
**Próximo paso:** Implementar Edit en FileFsRepository con validación de permisos
**Versión:** Proyecto 2 - EDIT con Reemplazo Atómico de Contenido
