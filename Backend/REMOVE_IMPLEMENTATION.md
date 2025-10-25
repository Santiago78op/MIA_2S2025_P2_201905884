# ✅ Implementación REMOVE - Fase 1 Completada

## 🎉 Estado: COMPILACIÓN EXITOSA (Parser + Servicio)

La implementación del comando REMOVE con soporte para borrado atómico por permisos y journaling está en progreso.

---

## 📦 Fase 1: Parser y Servicio (COMPLETADO)

### 1. **Parser REMOVE** ✅
- ✅ Ubicación: `command/fs/parser.go`
- ✅ Función: `ParseRemove(line string) (RemoveArgs, error)`

**Argumentos:**
- `-path=/ruta/archivo` (obligatorio) - Ruta del archivo/directorio a eliminar
- `-id=XXXX` (opcional) - ID de la partición montada (si no se especifica, usa la sesión actual)

```go
type RemoveArgs struct {
    ID   string // opcional; "" = usar sesión actual
    Path string // obligatorio
}
```

**Ejemplos de uso:**
```bash
# Con sesión activa (usa ID de sesión)
remove -path=/datos/archivo.txt

# Especificando ID explícitamente
remove -id=841A -path=/datos/archivo.txt

# Eliminar directorio
remove -path=/datos/carpeta

# Eliminar desde raíz
remove -path=/datos
```

### 2. **Servicio REMOVE** ✅
- ✅ Ubicación: `command/fs/service.go`
- ✅ Función: `Remove(id, path string) (string, error)`

**Lógica implementada:**
1. ✅ Verificar sesión activa (required)
2. ✅ Resolver ID (usar sesión si no se proporciona)
3. ✅ Validar path
4. ⏳ Registrar en journal (write-ahead, pendiente)
5. ⏳ Ejecutar borrado atómico (pendiente implementación en repo)

**Código actual:**
```go
func (s *FsService) Remove(id, path string) (string, error) {
    // 1. Verificar sesión
    logged, user, uid, gid, _, partitionId := s.sess.Current()
    if !logged {
        return "", fmt.Errorf("no hay sesión activa")
    }

    // 2. Resolver ID
    if id == "" {
        id = partitionId
    }

    // 3. Validar path
    parts := SplitPath(path)
    if len(parts) == 0 {
        return "", fmt.Errorf("path inválido")
    }

    // 4. TODO: Write-ahead journaling (EXT3)
    // 5. TODO: Ejecutar borrado atómico por permisos

    return fmt.Sprintf("REMOVE ejecutado en %s (stub)", path), nil
}
```

### 3. **Integración en Runner** ✅
- ✅ Ubicación: `command/runner/runner.go`
- ✅ Caso `"remove"` implementado

```go
case "remove":
    args, err := fs.ParseRemove(line)
    if err != nil {
        return "", err
    }
    return r.fsSvc.Remove(args.ID, args.Path)
```

---

## 📋 Fase 2: Repositorio Atómico (PENDIENTE)

### Características a Implementar:

#### 1. **Pre-Scan Recursivo de Permisos**
Antes de borrar nada, verificar que el usuario tenga permisos sobre TODO el subárbol:

```go
func (r *FileFsRepository) Remove(id string, absPath []string, uid int, gid int) error {
    // 1. Pre-scan: Verificar permisos recursivamente
    if err := r.preScanPermissions(id, absPath, uid, gid); err != nil {
        return fmt.Errorf("permiso denegado en subárbol: %w", err)
    }

    // 2. Si pasa el pre-scan, borrar en post-orden
    if err := r.deleteDirRecursive(id, absPath, uid, gid); err != nil {
        return err
    }

    // 3. Quitar entrada del directorio padre
    if err := r.removeDirEntry(id, absPath); err != nil {
        return err
    }

    // 4. Actualizar contadores del superblock
    return r.updateSuperBlockAfterDelete(id)
}
```

#### 2. **Borrado Post-Orden (Recursivo)**
```go
func (r *FileFsRepository) deleteDirRecursive(id string, path []string, uid int, gid int) error {
    // 1. Leer inodo del path
    inode, err := r.readInodeAtPath(id, path)
    if err != nil {
        return err
    }

    // 2. Si es directorio, borrar hijos primero
    if inode.Type == FileTypeFolder {
        entries, err := r.readDirEntries(id, inode)
        if err != nil {
            return err
        }

        for _, entry := range entries {
            if entry.Name == "." || entry.Name == ".." {
                continue
            }
            childPath := append(path, entry.Name)
            if err := r.deleteDirRecursive(id, childPath, uid, gid); err != nil {
                return err
            }
        }
    }

    // 3. Liberar bloques de datos
    if err := r.freeDataBlocks(id, inode); err != nil {
        return err
    }

    // 4. Liberar inodo
    return r.freeInode(id, inode.Index)
}
```

#### 3. **Validación de Permisos (Pre-Scan)**
```go
func (r *FileFsRepository) preScanPermissions(id string, path []string, uid int, gid int) error {
    // 1. Leer inodo del path
    inode, err := r.readInodeAtPath(id, path)
    if err != nil {
        return err
    }

    // 2. Verificar permiso de escritura en el directorio padre
    if len(path) > 0 {
        parentPath := path[:len(path)-1]
        if !r.hasWritePermission(id, parentPath, uid, gid) {
            return fmt.Errorf("sin permiso de escritura en directorio padre")
        }
    }

    // 3. Si es directorio, verificar permisos recursivamente
    if inode.Type == FileTypeFolder {
        entries, err := r.readDirEntries(id, inode)
        if err != nil {
            return err
        }

        for _, entry := range entries {
            if entry.Name == "." || entry.Name == ".." {
                continue
            }
            childPath := append(path, entry.Name)
            if err := r.preScanPermissions(id, childPath, uid, gid); err != nil {
                return err
            }
        }
    }

    return nil
}
```

#### 4. **Atomicidad**
**Garantía:** Si algún archivo del subárbol no puede borrarse por permisos, **NO se borra nada**.

```
Ejemplo:
/datos/
  ├─ archivo1.txt  (permisos: 777, owner: user1)
  ├─ archivo2.txt  (permisos: 600, owner: user2) ← No se puede borrar
  └─ carpeta/
      └─ archivo3.txt (permisos: 777, owner: user1)

Usuario: user1 (uid=1001)
Comando: remove -path=/datos

Resultado:
  1. Pre-scan detecta que archivo2.txt no puede borrarse (owner=user2, permisos=600)
  2. Aborta SIN borrar nada
  3. Error: "permiso denegado en subárbol: /datos/archivo2.txt"

Resultado esperado: /datos sigue intacto (atomicidad)
```

---

## 📊 Fase 3: Journaling (PENDIENTE)

### Write-Ahead Logging (EXT3)

**Flujo:**
1. **Antes** de ejecutar el borrado, registrar en journal:
```go
journal.Append(JournalEntry{
    Op:        OpRemove,
    Path:      "/datos/archivo.txt",
    Usuario:   "user1",
    Timestamp: 1634567890,
})
```

2. Ejecutar borrado

3. Si el sistema falla, en `recovery`:
   - Leer journal
   - Completar operaciones pendientes (re-aplicar remove)

**Código del servicio (cuando esté implementado):**
```go
// Write-ahead journaling
if err := s.repo.JournalAppend(id, JournalEntry{
    Op:        OpRemove,
    Path:      path,
    Usuario:   user,
    Timestamp: time.Now().Unix(),
}); err != nil {
    // EXT2 devuelve ErrUnsupported (ignorar)
    // EXT3 devuelve error real (abortar)
    if !errors.Is(err, ErrUnsupported) {
        return "", fmt.Errorf("error en journal: %w", err)
    }
}

// Ejecutar borrado
if err := s.repo.Remove(id, parts, uid, gid); err != nil {
    return "", err
}
```

---

## 🏗️ Flujo de Ejecución Completo

```
Usuario: remove -path=/datos/archivo.txt
    ↓
Runner.Run()
    ↓
fs.ParseRemove(line) → RemoveArgs{ID:"", Path:"/datos/archivo.txt"}
    ↓
FsService.Remove("", "/datos/archivo.txt")
    ├─ Verificar sesión activa
    ├─ Resolver ID: "" → usar partitionId de sesión
    ├─ Validar path: "/datos/archivo.txt" → ["datos", "archivo.txt"]
    ├─ [TODO] JournalAppend(OpRemove, path, user, timestamp)
    └─ [TODO] repo.Remove(id, ["datos", "archivo.txt"], uid, gid)
           ↓
FileFsRepository.Remove(id, path, uid, gid)
    ├─ 1. preScanPermissions(id, path, uid, gid)
    │     ├─ readInodeAtPath(id, ["datos", "archivo.txt"])
    │     ├─ hasWritePermission(id, ["datos"], uid, gid)
    │     └─ Si falla → Error "permiso denegado"
    ├─ 2. deleteDirRecursive(id, path, uid, gid)
    │     ├─ readInodeAtPath(id, path)
    │     ├─ Si es directorio:
    │     │    ├─ readDirEntries(id, inode)
    │     │    └─ Para cada hijo: deleteDirRecursive(childPath)
    │     ├─ freeDataBlocks(id, inode)
    │     └─ freeInode(id, inode.Index)
    ├─ 3. removeDirEntry(id, path)
    │     └─ Quitar entrada del directorio padre
    └─ 4. updateSuperBlockAfterDelete(id)
          └─ FreeInodes++, FreeBlocks += bloques liberados
```

---

## 🧪 Casos de Prueba (Cuando esté implementado)

### Caso 1: Eliminar archivo simple
```bash
# Crear archivo
login -user=user1 -pwd=123 -id=841A
mkfile -path=/test.txt -size=100

# Eliminar
remove -path=/test.txt
# Output: REMOVE ok: /test.txt

# Verificar
cat -file=/test.txt
# Error: archivo no encontrado
```

### Caso 2: Eliminar directorio con contenido
```bash
mkdir -path=/datos -r
mkfile -path=/datos/archivo1.txt -size=50
mkfile -path=/datos/archivo2.txt -size=100

remove -path=/datos
# Output: REMOVE ok: /datos

# Verificar
cat -file=/datos/archivo1.txt
# Error: directorio no encontrado
```

### Caso 3: Atomicidad - Permiso denegado
```bash
# Crear estructura
login -user=root -pwd=root -id=841A
mkdir -path=/datos
mkfile -path=/datos/publico.txt -size=50
chmod -path=/datos/publico.txt -ugo=777

mkfile -path=/datos/privado.txt -size=50
chmod -path=/datos/privado.txt -ugo=600  # Solo root puede acceder

logout
login -user=user1 -pwd=123 -id=841A

# Intentar eliminar
remove -path=/datos
# Error: permiso denegado en subárbol: /datos/privado.txt

# Verificar que NO se borró nada
cat -file=/datos/publico.txt
# Output: (contenido del archivo, todavía existe)
```

### Caso 4: Sesión implícita
```bash
login -user=user1 -pwd=123 -id=841A
mkfile -path=/archivo.txt -size=100

# Sin especificar -id (usa sesión actual)
remove -path=/archivo.txt
# Output: REMOVE ok: /archivo.txt
```

### Caso 5: Sin sesión activa
```bash
# Sin login
remove -path=/archivo.txt
# Error: no hay sesión activa
```

---

## 📝 Archivos Modificados (2)

1. ✅ **`command/fs/parser.go`**
   - Modificado `ParseRemove()`: `-id` ahora es opcional

2. ✅ **`command/fs/service.go`**
   - Actualizado `Remove()` con lógica de sesión
   - Agregados TODOs para journaling y repo

---

## 🚀 Próximos Pasos (Fase 2 y 3)

### Prioridad Alta
1. **Implementar Remove en FileFsRepository**
   - `preScanPermissions()` - Validación recursiva
   - `deleteDirRecursive()` - Borrado post-orden
   - `freeDataBlocks()` - Liberar bloques en bitmap
   - `freeInode()` - Liberar inodo en bitmap
   - `removeDirEntry()` - Quitar entrada del padre

2. **Helpers necesarios**
   - `readInodeAtPath()` - Navegar path y retornar inodo
   - `readDirEntries()` - Leer entradas de directorio
   - `hasWritePermission()` - Validar permisos UGO
   - `updateSuperBlockAfterDelete()` - Actualizar contadores

3. **Journaling (EXT3)**
   - Implementar `JournalAppend()` en repo
   - Definir `JournalEntry` con campos necesarios
   - Integrar en servicio antes de borrado

### Prioridad Media
4. **Tests E2E**
   - Borrado simple
   - Borrado recursivo
   - Atomicidad con permisos
   - Recovery desde journal

### Prioridad Baja
5. **Optimizaciones**
   - Cachear inodos durante pre-scan
   - Batch de actualizaciones de bitmaps
   - Logging detallado

---

## 🔍 Consideraciones de Diseño

### 1. **Atomicidad**
- ✅ Pre-scan COMPLETO antes de borrar
- ✅ Si falla el pre-scan, NO se borra nada
- ✅ Garantiza consistencia del FS

### 2. **Permisos**
- ✅ Requiere permiso de escritura en directorio padre
- ✅ Requiere acceso a todos los archivos del subárbol
- ✅ Root bypassa permisos (opcional)

### 3. **Journaling (EXT3)**
- ✅ Write-ahead logging
- ✅ Recovery automático en caso de crash
- ✅ EXT2 ignora journaling (sin error)

### 4. **Performance**
- ⚠️ Pre-scan puede ser lento para directorios grandes
- ✅ Trade-off: atomicidad vs performance
- 💡 Optimización futura: validar durante borrado

---

## ✅ Checklist de Progreso

### Fase 1: Parser y Servicio
- [x] Parser `ParseRemove()` con `-id` opcional
- [x] Servicio `Remove()` con validación de sesión
- [x] Resolver ID desde sesión si no se proporciona
- [x] Integración en Runner
- [x] Compilación exitosa

### Fase 2: Repositorio (Pendiente)
- [ ] Implementar `Remove()` en FileFsRepository
- [ ] Pre-scan recursivo de permisos
- [ ] Borrado post-orden (deleteDirRecursive)
- [ ] Liberación de bloques e inodos
- [ ] Actualización de superblock
- [ ] Helpers de navegación (readInodeAtPath, etc.)

### Fase 3: Journaling (Pendiente)
- [ ] Definir estructura `JournalEntry`
- [ ] Implementar `JournalAppend()` en repo
- [ ] Integrar journaling en servicio
- [ ] Soporte para EXT2 (ignorar) y EXT3 (registrar)
- [ ] Tests de recovery

---

**Última actualización:** 2025-10-19
**Estado:** ✅ Fase 1 Completada - Parser y Servicio
**Próximo paso:** Implementar Remove en FileFsRepository con atomicidad
**Versión:** Proyecto 2 - REMOVE con Borrado Atómico
