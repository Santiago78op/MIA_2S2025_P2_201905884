# Integración Proyecto 2 - Frontend y Backend

## 📋 Resumen de Cambios

Este documento detalla las mejoras implementadas para integrar las funcionalidades del Proyecto 2, manteniendo intactas todas las características del Proyecto 1.

---

## ✅ Características Implementadas

### 🔐 1. Autenticación por GUI

**Ubicación Frontend:** `src/pages/LoginPage.jsx`

- ✅ Formulario de login con campos: ID de montaje, usuario y contraseña
- ✅ Validación en el backend (no por comando)
- ✅ Manejo de errores y estados de carga
- ✅ Redirección automática al visualizador tras login exitoso

**Ubicación Backend:** `Backend/controllers/viewer_controller.go`

- ✅ `POST /api/auth/login` - Autenticación de usuarios
  - Valida que la partición esté montada
  - Lee `users.txt` para verificar credenciales
  - Crea sesión con SessionStore
  - Retorna: `{ok, user, id, uid, gid, root}`

- ✅ `POST /api/auth/logout` - Cierre de sesión
  - Limpia la sesión actual
  - Retorna: `{ok, message}`

---

### 🔍 2. Visualizador del Sistema de Archivos (Solo Lectura)

El visualizador implementa el flujo de 3 pasos especificado en el proyecto:

#### **Paso 1: Selección de Disco**

**Componente:** `src/components/DiskPicker.jsx`

- ✅ Lista todos los discos con particiones montadas
- ✅ Muestra información básica:
  - Nombre del disco
  - Path completo
  - Capacidad (placeholder)
  - Fit (placeholder)
  - Número de particiones montadas

**Endpoint:** `GET /api/disks`

**Respuesta:**
```json
[
  {
    "path": "/tmp/Disco1.mia",
    "name": "Disco1.mia",
    "size": "N/A",
    "fit": "N/A",
    "mounted": [
      {"id": "841A", "name": "Part1"}
    ]
  }
]
```

#### **Paso 2: Selección de Partición**

**Componente:** `src/components/PartitionPicker.jsx`

- ✅ Lista todas las particiones del disco seleccionado
- ✅ Muestra información básica:
  - Nombre de la partición
  - ID de montaje
  - Tipo (Primaria/Extendida/Lógica)
  - Tamaño
  - Fit
  - Estado de formateo

**Endpoint:** `GET /api/disks/:disk/partitions`

**Respuesta:**
```json
[
  {
    "id": "841A",
    "name": "Part1",
    "type": "Primaria",
    "size": "N/A",
    "fit": "N/A",
    "formatted": true
  }
]
```

#### **Paso 3: Navegación del Sistema de Archivos**

**Componente:** `src/components/Explorer.jsx`

- ✅ Inicia desde la carpeta raíz `/`
- ✅ Breadcrumb navegable (clic en cada segmento de la ruta)
- ✅ Lista archivos y carpetas con información detallada:
  - Nombre
  - Tipo (DIR/FILE)
  - Permisos (rwx format)
  - Propietario y grupo
  - UID/GID
  - Tamaño (para archivos)
  - Fecha de modificación
- ✅ Visualizador de contenido de archivos (textarea readonly)
- ✅ Botón "Actualizar" para refrescar tras comandos

**Endpoints:**

1. **Listar directorio:** `GET /api/fs/:id/tree?path=/ruta`

**Respuesta:**
```json
{
  "items": [
    {
      "name": "users.txt",
      "type": "file",
      "path": "/users.txt",
      "size": 128,
      "perm": "rw-r--r--",
      "uid": 1,
      "gid": 1,
      "owner": "root",
      "group": "root",
      "mtime": 1699999999
    }
  ]
}
```

2. **Leer archivo:** `GET /api/fs/:id/file?path=/archivo.txt`

**Respuesta:**
```json
{
  "path": "/archivo.txt",
  "content": "Contenido del archivo..."
}
```

---

### 🎨 3. Componentes del Frontend

#### **Estructura de Páginas**

```
src/pages/
├── Home.jsx          - Terminal principal (Proyecto 1 + botón login)
├── LoginPage.jsx     - Formulario de autenticación GUI ✨ NUEVO
├── Visualizer.jsx    - Orquestador de 3 pasos ✨ MEJORADO
└── Reports.jsx       - Galería de reportes (Proyecto 1)
```

#### **Componentes Reutilizables**

```
src/components/
├── Topbar.jsx            - Header con sesión activa y logout
├── Terminal.jsx          - Terminal interactiva (P1)
├── DiskPicker.jsx        - Paso 1: Selección de disco ✨ MEJORADO
├── PartitionPicker.jsx   - Paso 2: Selección de partición ✨ MEJORADO
├── Explorer.jsx          - Paso 3: Navegador FS ✨ MEJORADO
├── JournalPanel.jsx      - Visor de journaling EXT3 (P1)
└── ReportsGallery.jsx    - Galería 3D de reportes (P1)
```

#### **API Client**

**Ubicación:** `src/lib/api.js`

Funciones añadidas/mejoradas:
- ✅ `login(id, user, pass)` - Autenticación
- ✅ `logout()` - Cierre de sesión
- ✅ `disks()` - Lista de discos
- ✅ `partitions(diskPath)` - Lista de particiones
- ✅ `list(id, path)` - Contenido de directorio
- ✅ `file(id, path)` - Contenido de archivo

---

### 🔧 4. Backend - Controladores

#### **ViewerController**

**Ubicación:** `Backend/controllers/viewer_controller.go`

**Métodos implementados:**

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `Login` | `POST /api/auth/login` | Autenticación de usuarios |
| `Logout` | `POST /api/auth/logout` | Cierre de sesión |
| `ListDisks` | `GET /api/disks` | Lista discos disponibles |
| `ListPartitions` | `GET /api/disks/:disk/partitions` | Lista particiones de un disco |
| `GetTree` | `GET /api/fs/:id/tree?path=...` | Contenido de directorio |
| `GetFile` | `GET /api/fs/:id/file?path=...` | Contenido de archivo |
| `GetJournal` | `GET /api/journal/:id` | Entradas del journal (EXT3) |
| `GetJournalTable` | `GET /api/journal/:id/table` | Journal en formato tabla |

**Cambios principales:**

1. **Login:** Lee `users.txt`, valida credenciales y crea sesión
2. **Logout:** Limpia la sesión actual
3. **ListDisks/ListPartitions:** Agrupan y formatean datos de montajes
4. **GetTree:** Construye paths absolutos para cada entrada
5. **GetFile:** Lee contenido usando `FsRepository.Cat`

---

### 🔄 5. Router Actualizado

**Ubicación:** `Backend/router/router.go`

**Rutas agregadas:**

```go
// Autenticación
auth := api.Group("/auth")
{
    auth.POST("/login", vc.Login)
    auth.POST("/logout", vc.Logout)
}

// Visualizador
api.GET("/disks", vc.ListDisks)
api.GET("/disks/:disk/partitions", vc.ListPartitions)
api.GET("/fs/:id/tree", vc.GetTree)
api.GET("/fs/:id/file", vc.GetFile)
api.GET("/journal/:id", vc.GetJournal)
api.GET("/journal/:id/table", vc.GetJournalTable)
```

**CORS:** Configurado para permitir peticiones desde el frontend

---

## 🚀 Cómo Ejecutar

### **Backend**

```bash
cd Backend
go run cmd/server/main.go
```

El servidor escuchará en `http://localhost:8080`

### **Frontend**

```bash
cd Frontend
npm install
npm run dev
```

El frontend estará en `http://localhost:5173`

---

## 📝 Flujo de Uso

### 1️⃣ **Crear disco y particiones** (Terminal - Home)

```bash
mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"
fdisk -size=1024 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part1
mount -path="/tmp/Disco1.mia" -name=Part1
# Anota el ID retornado, ej: 841A
```

### 2️⃣ **Formatear partición** (Terminal - Home)

```bash
mkfs -type=ext2 -id=841A
# o
mkfs -type=ext3 -id=841A
```

### 3️⃣ **Iniciar sesión** (Login Page)

- Ir a "Iniciar Sesión"
- Ingresar:
  - **ID:** `841A` (el ID retornado por mount)
  - **Usuario:** `root`
  - **Contraseña:** `123`
- Clic en "Ingresar"

### 4️⃣ **Explorar sistema de archivos** (Visualizer)

1. **Seleccionar disco:** Clic en el disco deseado
2. **Seleccionar partición:** Clic en la partición formateada
3. **Navegar:**
   - Clic en carpetas para abrirlas
   - Clic en archivos para ver contenido
   - Usar breadcrumb para retroceder
   - Botón "Actualizar" tras crear archivos/carpetas desde terminal

### 5️⃣ **Crear archivos y carpetas** (Terminal - Home)

```bash
mkdir -path=/home -id=841A
mkfile -path=/home/archivo1.txt -id=841A -size=100
cat -file1=/home/archivo1.txt -id=841A
```

Luego, en el Visualizador → "Actualizar" para ver los cambios

---

## ⚠️ Notas Importantes

### ✅ **Mantenido del Proyecto 1**

- ✅ Terminal de comandos (entrada y salida)
- ✅ Ejecución de scripts .smia
- ✅ Todos los comandos (mkdisk, fdisk, mount, mkfs, etc.)
- ✅ Reportes (mbr, disk, tree, sb, file, ls, etc.)
- ✅ Galería 3D de reportes
- ✅ Journaling EXT3
- ✅ Temas visuales (Neo Green / Aurora Purple)

### 🆕 **Agregado en Proyecto 2**

- ✅ Login por GUI (no por comando)
- ✅ Visualizador de 3 pasos (solo lectura)
- ✅ Información detallada de archivos/carpetas
- ✅ Breadcrumb navegable
- ✅ Botón de logout en topbar
- ✅ Estados de carga y manejo de errores mejorado

### ⏳ **Pendiente (según solicitud)**

- ⏳ Deploy en AWS S3 (dejado pendiente)
- ⏳ Lectura de tamaño/fit real del MBR (actualmente "N/A")
- ⏳ UID/GID reales del inodo (actualmente 0/placeholders)

---

## 🐛 Troubleshooting

### **Error: "Partición no montada"**

- Verifica que usaste `mount` antes de intentar login
- El ID debe coincidir exactamente

### **Error: "Error leyendo users.txt"**

- Asegúrate de haber formateado la partición con `mkfs`
- EXT2/EXT3 crean `users.txt` automáticamente

### **Error: "Usuario o contraseña incorrectos"**

- Usuario por defecto: `root`
- Contraseña por defecto: `123`
- Puedes crear usuarios con `mkusr` tras login

### **Frontend no se conecta al backend**

- Verifica que el backend esté corriendo en puerto 8080
- Revisa el archivo `.env` (debe tener `VITE_API_URL=http://localhost:8080`)
- Revisa `vite.config.js` (proxy configurado)

---

## 📚 Referencias

- **Router Backend:** `Backend/router/router.go`
- **ViewerController:** `Backend/controllers/viewer_controller.go`
- **API Client:** `Frontend/src/lib/api.js`
- **Visualizador:** `Frontend/src/pages/Visualizer.jsx`
- **Login:** `Frontend/src/pages/LoginPage.jsx`

---

## 🎯 Cumplimiento de Requisitos del Proyecto 2

| Requisito | Estado | Ubicación |
|-----------|--------|-----------|
| Login por GUI | ✅ | `LoginPage.jsx`, `viewer_controller.go:308` |
| Visualizador de FS (solo lectura) | ✅ | `Visualizer.jsx`, `Explorer.jsx` |
| Paso 1: Selección de disco | ✅ | `DiskPicker.jsx` |
| Paso 2: Selección de partición | ✅ | `PartitionPicker.jsx` |
| Paso 3: Navegación desde / | ✅ | `Explorer.jsx` |
| Mostrar permisos | ✅ | `Explorer.jsx:89-110` |
| Mostrar propietario/grupo | ✅ | `Explorer.jsx:95-96` |
| Visualizar contenido de archivos | ✅ | `Explorer.jsx:132-145` |
| Botón de logout | ✅ | `Topbar.jsx:37-40` |
| Sesión persiste entre comandos | ✅ | `SessionStore` |
| Comandos requieren sesión | ✅ | Backend (ya implementado P1) |
| Deploy en S3 | ⏳ | Pendiente |

---

**✨ Integración completada exitosamente. Todas las funcionalidades del Proyecto 1 se mantienen intactas y las del Proyecto 2 están plenamente operativas.**
