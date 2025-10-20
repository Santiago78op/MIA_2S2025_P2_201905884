# Estado del Proyecto P2 - Checklist de Verificación

Fecha: 2025-10-19

---

## ✅ 1) Arquitectura & Despliegue (AWS)

### EC2 (backend Go):
- [ ] Instancia Linux (ej. Ubuntu) corriendo el binario del backend como servicio (systemd)
- [ ] Security Groups abiertos solo a los puertos necesarios (HTTP p.ej. 80/8080)
- [ ] Variables de entorno / config para paths (`Discos/`, `Reports/`) y CORS del frontend
- [x] Logs accesibles (journalctl o archivo) y healthcheck `GET /health` responde 200
  - ✅ Healthcheck implementado en `/health`

### S3 (frontend):
- [ ] Bucket configurado como **sitio web estático** (index/error)
- [ ] Política y permisos S3 válidos (página es accesible por link público)
- [ ] UI consume la API del backend (dominio/URL correctos, CORS OK)

### Pruebas rápidas:
- [ ] Abrir URL del S3 y ver **Home con terminal** y botón de **Login**
- [ ] Ejecutar un comando trivial (ej. `mkdisk …`) y ver respuesta en la salida

**Status**: ⚠️ PENDIENTE - Despliegue AWS no realizado

---

## 🖥️ 2) Frontend (flujo del visualizador)

### Login UI (no por comando):
- [ ] Pantalla de login con user/pass y selección de `-id` de la partición montada
- [ ] Al iniciar sesión, la sesión se usa para comandos que lo requieren
- [ ] Botón de **Cerrar Sesión** visible y funcional

### Visualizador de FS (solo lectura):
- [ ] Vista 1: **Selección de Disco** con: nombre, tamaño total, fit, **particiones montadas**
- [ ] Vista 2: **Selección de Partición** con: nombre, tamaño, fit, estado (montada/ID)
- [ ] Vista 3: **Árbol/navegador** desde `/`:
  - [ ] Listado de archivos/carpetas con datos básicos: **permiso UGO, owner, group, size, fecha**
  - [ ] Se puede **entrar a carpetas** y **ver contenido** de archivos de texto
- [ ] Vista de **Journaling** (tabla): Operación, Path, Contenido (preview), Fecha/Hora

### Endpoints mínimos que la UI debe consumir (GET):
- [x] `/api/journal/:id` (lista entradas del journal - formato crudo)
  - ✅ Implementado en `viewer_controller.go:122`
- [x] `/api/journal/:id/table` (lista entradas del journal - formato tabla)
  - ✅ Implementado en `viewer_controller.go:174`
- [ ] `/api/disks` (lista discos + resumen)
  - ⚠️ Stub en `viewer_controller.go:30` - IMPLEMENTAR
- [ ] `/api/disks/:disk/partitions`
  - ⚠️ Stub en `viewer_controller.go:37` - IMPLEMENTAR
- [ ] `/api/fs/:id/tree?path=/…` (entradas con metadatos)
  - ⚠️ Stub en `viewer_controller.go:41` - IMPLEMENTAR
- [ ] `/api/fs/:id/file?path=/…` (content/size)
  - ⚠️ Stub en `viewer_controller.go:72` - IMPLEMENTAR

**Status**: ⚠️ PARCIAL - Journal OK, otros endpoints pendientes

---

## 🧠 3) EXT3: Layout y Cálculo

### Cálculo de n (inodos) y 3n (bloques):
- [x] Implementado el despeje de `n` con `floor` usando la fórmula
  - ✅ Verificado en implementación anterior
- [x] `JournalEntries` = **50** (constante)
  - ✅ Constante definida
- [x] Estructura de la partición (offsets coherentes):
  - [x] Superblock ✅
  - [x] Journaling (50 entradas pre-inicializadas) ✅
  - [x] Bitmap de Inodos (n) ✅
  - [x] Bitmap de Bloques (3n) ✅
  - [x] Tabla de Inodos (n * sizeof(inodo)) ✅
  - [x] Área de Bloques (3n * sizeof(bloque)) ✅
- [x] `mkfs -fs=3fs` llena SB + Journal y limpia bitmaps/áreas ✅

### Pruebas:
- [ ] `mkfs -id=… -fs=3fs -type=full` → Validar:
  - [ ] `s_fs_type==3`
  - [ ] `journal_start`/`journal_count==50`
  - [ ] Bitmaps reseteados, inodos/bloques libres, `users.txt` creado

**Status**: ✅ IMPLEMENTADO - Pendiente pruebas de validación

---

## 📓 4) Journal (estructura y registro)

### Estructuras:
- [x] `JournalEntry{ Op, Path, Content, Timestamp }`
  - ✅ Definido en `core/models/journal.go`
- [x] Las mutaciones hacen **write-ahead**: primero append al journal, luego aplicar
  - ✅ Implementado en CHMOD, CHOWN, COPY, MOVE

### Verificaciones:
- [x] `/api/journal/:id/table` muestra filas (op, path, contenido, fecha)
  - ✅ Endpoint implementado
- [ ] Tras `mkdir` y `mkfile`, journal muestra entradas
  - ⚠️ VERIFICAR - Necesita prueba
- [ ] `edit` registra operación
  - ⚠️ PENDIENTE - EDIT no está implementado
- [x] `chmod/chown` registran `ugo/usuario`
  - ✅ Implementado

**Status**: ⚠️ PARCIAL - Journal OK, faltan comandos que lo usen

---

## 🧾 5) Comandos nuevos (parámetros, permisos, errores)

### 5.1 FDISK (ADD, DELETE)
- [ ] `-add=±N -unit=B|K|M`
  - [ ] Si positivo: hay espacio **libre posterior** suficiente
  - [ ] Si negativo: no deja tamaño < espacio **ya usado**
  - [ ] Actualiza MBR/EBR correctamente
- [ ] `-delete=fast|full`
  - [ ] `fast`: limpia entrada de tabla (MBR/EBR)
  - [ ] `full`: además rellena la región con `\0`
  - [ ] Borrar **extendida** elimina sus **lógicas**
- [ ] Mensajes de error claros

**Status**: ❌ NO IMPLEMENTADO

### 5.2 UNMOUNT
- [ ] `unmount -id=XXXX` desmonta si existe; si no, error
- [ ] Resetea **correlativo** de la partición a **0**

**Status**: ⚠️ PARCIAL - Verificar reset de correlativo

### 5.3 MKFS (fs=2fs|3fs)
- [x] Por defecto **ext2** si no se pasa `-fs`
  - ✅ Implementado
- [x] `-type=full` hace formateo completo
  - ✅ Implementado
- [x] EXT3 inicializa journal (50), SB y estructuras
  - ✅ Implementado

**Status**: ✅ IMPLEMENTADO

### 5.4 REMOVE
- [ ] `remove -path=/…`
- [ ] Usuario debe tener **escritura**
- [ ] Carpeta: **verificación previa recursiva**
- [ ] Libera inodo y bloques en bitmaps
- [ ] Registra en journal (EXT3)

**Status**: ❌ NO IMPLEMENTADO
- Stub en `storage/adapters/fs_adapter.go:45`

### 5.5 EDIT
- [ ] `edit -path=/… -contenido=/ruta/host`
- [ ] Requiere **lectura + escritura** sobre el archivo
- [ ] Reemplaza contenido → reasigna bloques, actualiza tamaño
- [ ] Journal (con snapshot/preview)

**Status**: ❌ NO IMPLEMENTADO
- Stub en `storage/adapters/fs_adapter.go:49`

### 5.6 RENAME
- [ ] `rename -path=/… -name=nuevo`
- [ ] Requiere **escritura** en directorio **padre**
- [ ] Error si ya existe homónimo

**Status**: ❌ NO IMPLEMENTADO
- Stub en `storage/adapters/fs_adapter.go:54`

### 5.7 COPY
- [x] `copy -path=/origen -destino=/dest`
  - ✅ Implementado en `storage/diskio/file_repo_copy.go`
- [x] Origen: **lectura**; Destino: **escritura**
  - ✅ Validación de permisos implementada
- [x] Directorios: copia **recursiva**
  - ✅ Implementado
- [x] Omite los que no pueda leer (sin fallar todo)
  - ✅ Implementado (devuelve count de skipped)
- [x] Journal (EXT3)
  - ✅ Implementado

**Status**: ✅ IMPLEMENTADO

### 5.8 MOVE
- [x] `move -path=/origen -destino=/dest`
  - ✅ Implementado en `storage/diskio/file_repo_move.go`
- [x] Requiere **escritura** en origen y destino
  - ✅ Validación implementada
- [x] **Misma partición**: solo **re-enlaza** (no copia bloques)
  - ✅ Implementado
- [x] Journal (EXT3)
  - ✅ Implementado

**Status**: ✅ IMPLEMENTADO

### 5.9 FIND
- [x] `find -path=/dir -name=patrón` con `?` y `*`
  - ✅ Implementado en `storage/diskio/file_repo_find.go`
- [x] Busca **recursivo**, respetando **lectura**
  - ✅ Implementado con validación de permisos

**Status**: ✅ IMPLEMENTADO

### 5.10 CHOWN
- [x] `chown -path=/… -usuario=… [-r]`
  - ✅ Implementado en `storage/diskio/file_repo_chown_chmod.go:23`
- [x] **root** puede todo; otros solo sobre **sus** archivos
  - ✅ Validación implementada (línea 72-74)
- [x] Si `-r`, aplica recursivo en el subárbol
  - ✅ Implementado con `chownRecursive` (línea 106)
- [x] Journal (EXT3)
  - ✅ Implementado (línea 44-56)

**Status**: ✅ IMPLEMENTADO

### 5.11 CHMOD
- [x] `chmod -path=/… -ugo=XYZ [-r]`, con `X,Y,Z` en `0..7`
  - ✅ Implementado en `storage/diskio/file_repo_chown_chmod.go:173`
- [x] **root o propietario** pueden cambiar permisos
  - ✅ Validación implementada (línea 216-218)
- [x] Si `-r`, aplica recursivo
  - ✅ Implementado con `chmodRecursive` (línea 249)
- [x] Journal (EXT3)
  - ✅ Implementado (línea 194-206)

**Status**: ✅ IMPLEMENTADO

---

## 🧨 6) LOSS & 🔁 RECOVERY (EXT3)

### LOSS:
- [x] `loss -id=…` limpia con `\0`:
  - [x] **bitmap inodos** ✅
  - [x] **bitmap bloques** ✅
  - [x] **área inodos** ✅
  - [x] **área bloques** ✅
- [x] **No** toca SB ni Journal
  - ✅ Implementado en `storage/diskio/file_repo_loss_recovery.go:27`

**Status**: ✅ IMPLEMENTADO

### RECOVERY:
- [x] `recovery -id=…` lee **Journal + SB**
  - ✅ Implementado en `storage/diskio/file_repo_loss_recovery.go:102`
- [x] Re-aplica en orden por **timestamp**
  - ✅ Sort implementado (línea 134-136)
- [x] Best effort (continúa si algunas fallan)
  - ✅ Implementado (línea 155-161)
- [x] Limpia el journal al final
  - ✅ Implementado (línea 164-166)

**Status**: ✅ IMPLEMENTADO

### Prueba integral:
- [ ] Ejecutar escenario completo LOSS → RECOVERY
- [ ] Verificar reconstrucción de estructura
- [ ] Validar journal antes y después

**Status**: ✅ IMPLEMENTADO - Pendiente pruebas E2E

---

## 🔐 7) Sesiones & Permisos

- [x] Login por UI requiere `id`, `user` y `pass` válidos
  - ✅ Endpoint implementado en `viewer_controller.go:228`
- [ ] Comandos que **requieren sesión** fallan con error claro
  - ⚠️ VERIFICAR - Depende de implementación de servicios
- [x] `utils/permissions`: verificación UGO
  - ✅ Implementado en validaciones de permisos
- [x] `root` bypass donde aplica
  - ✅ Implementado (uid=1 bypass)

**Status**: ⚠️ PARCIAL - Login OK, verificar integración

---

## 🧪 8) Script E2E (debe correr completo)

**Status**: ⚠️ PENDIENTE - Crear script de prueba completo

Crear archivo: `test_e2e.smia` con todos los comandos

---

## 📚 9) Documentación

- [x] **Documentación técnica de comandos:**
  - [x] COPY_IMPLEMENTATION.md ✅
  - [x] MOVE_IMPLEMENTATION.md ✅
  - [x] FIND_IMPLEMENTATION.md ✅
  - [x] CHOWN_CHMOD_LOSS_RECOVERY_IMPLEMENTATION.md ✅
  - [x] JOURNAL_VIEWER_IMPLEMENTATION.md ✅
- [ ] **Manual técnico completo** con:
  - [ ] Arquitectura completa + despliegue AWS (diagramas)
  - [ ] Estructuras internas (MBR, EBR, inodos, bloques, **EXT3 + Journal**)
  - [ ] Explicación y ejemplos de **todos** los comandos
- [ ] **Manual de usuario** (cómo usar UI y scripts, capturas)
- [ ] Link del **sitio web (S3)** y evidencia de backend (EC2)

**Status**: ⚠️ PARCIAL - Documentación técnica OK, manuales pendientes

---

## 🧩 10) Mapeo rápido a la rúbrica

### Parte 1 (5 pts): mkfs, login, logout, mkdir, mkfile
- [x] mkfs ✅
- [x] login (endpoint) ✅
- [x] logout (endpoint) ✅
- [x] mkdir ✅
- [x] mkfile ✅

**Score estimado**: 5/5

### Parte 2 (40 pts): EC2/S3 + UI
- [ ] EC2 deployment ❌
- [ ] S3 deployment ❌
- [ ] UI: Selección disco ❌
- [ ] UI: Selección partición ❌
- [ ] UI: Navegador FS ❌
- [ ] UI: Ver contenido ❌
- [ ] UI: Manejo de sesión ❌
- [x] UI: Endpoints journal ✅

**Score estimado**: 2/40 ⚠️ CRÍTICO

### Parte 3 (30 pts): Comandos avanzados
- [ ] FDISK add/delete ❌
- [x] UNMOUNT (parcial) ⚠️
- [x] MKFS 2fs/3fs ✅
- [ ] REMOVE ❌
- [ ] EDIT ❌
- [ ] RENAME ❌
- [x] COPY ✅
- [x] MOVE ✅
- [x] FIND ✅
- [x] CHOWN ✅
- [x] CHMOD ✅

**Score estimado**: 18/30

### Parte 4 (15 pts): RECOVERY, LOSS, JOURNALING
- [x] RECOVERY ✅
- [x] LOSS ✅
- [x] JOURNALING (backend) ✅
- [ ] JOURNALING (UI) ❌

**Score estimado**: 11/15

### Parte 5 (10 pts): Documentación
- [x] Documentación técnica parcial ✅
- [ ] Manual técnico completo ❌
- [ ] Manual de usuario ❌

**Score estimado**: 3/10

---

## 📊 RESUMEN EJECUTIVO

### ✅ COMPLETADO (60%)
1. EXT3 layout y journal ✅
2. COPY, MOVE, FIND ✅
3. CHOWN, CHMOD ✅
4. LOSS, RECOVERY ✅
5. Journal viewer endpoints ✅
6. Documentación técnica de comandos ✅

### ⚠️ PARCIAL (20%)
1. Viewer endpoints (stubs implementados)
2. UNMOUNT (verificar reset correlativo)
3. Sesiones (login OK, integración pendiente)
4. Documentación (técnica OK, manuales pendientes)

### ❌ PENDIENTE (20%)
1. **CRÍTICO**: Despliegue AWS (EC2 + S3) ❌
2. **CRÍTICO**: Frontend UI completo ❌
3. FDISK add/delete ❌
4. REMOVE, EDIT, RENAME ❌
5. Script E2E ❌
6. Manuales completos ❌

---

## 🎯 PLAN DE ACCIÓN PRIORITARIO

### Prioridad 1 (CRÍTICO - 40 pts en riesgo):
1. Implementar endpoints viewer restantes:
   - [ ] `/api/disks` - Lista discos con metadatos
   - [ ] `/api/disks/:disk/partitions` - Lista particiones
   - [ ] `/api/fs/:id/tree?path=/` - Navegador de archivos
   - [ ] `/api/fs/:id/file?path=/archivo` - Contenido de archivo

2. Desarrollar Frontend UI:
   - [ ] Pantalla de login
   - [ ] Selección de disco/partición
   - [ ] Navegador de archivos (árbol + tabla)
   - [ ] Visor de contenido
   - [ ] Tabla de journal

3. Despliegue AWS:
   - [ ] Backend en EC2 con systemd
   - [ ] Frontend en S3 (sitio estático)
   - [ ] Configurar CORS y security groups

### Prioridad 2 (ALTA - 15 pts en riesgo):
1. Implementar comandos faltantes:
   - [ ] REMOVE (con validación recursiva de permisos)
   - [ ] EDIT (con reasignación de bloques)
   - [ ] RENAME (con validación de duplicados)
   - [ ] FDISK add/delete

### Prioridad 3 (MEDIA - 10 pts):
1. Crear script E2E de pruebas
2. Completar manuales (técnico y usuario)
3. Agregar diagramas de arquitectura

---

## 📝 NOTAS ADICIONALES

### Fortalezas del proyecto:
- ✅ Arquitectura limpia con separación de capas
- ✅ Implementación completa de LOSS/RECOVERY
- ✅ Journal con write-ahead logging
- ✅ Permisos UGO correctamente implementados
- ✅ Comandos avanzados (COPY, MOVE, FIND, CHMOD, CHOWN)

### Riesgos identificados:
- ⚠️ **ALTO**: Sin UI ni despliegue AWS → 40 pts en riesgo
- ⚠️ **MEDIO**: Comandos REMOVE/EDIT/RENAME sin implementar → 10 pts en riesgo
- ⚠️ **BAJO**: Documentación incompleta → 7 pts en riesgo

### Recomendaciones:
1. **URGENTE**: Priorizar desarrollo de UI y despliegue AWS
2. Crear endpoints viewer faltantes (2-3 horas de trabajo estimadas)
3. Implementar REMOVE/EDIT/RENAME (4-6 horas de trabajo estimadas)
4. Crear script E2E para validación rápida
5. Completar manuales con capturas de pantalla

---

**Última actualización**: 2025-10-19 19:02:00
**Próxima revisión**: Después de implementar endpoints viewer
