# MIA Proyecto 2 - Frontend

Frontend moderno y futurista para el sistema de administración de archivos EXT2/EXT3 (Proyecto 2 - MIA 2S 2025).

![Theme Preview](https://img.shields.io/badge/themes-2-purple) ![React](https://img.shields.io/badge/react-18.3-blue) ![Vite](https://img.shields.io/badge/vite-5.4-yellow)

## Características Principales

### Flujo de Trabajo según Especificación

El proyecto implementa el flujo exacto especificado en el enunciado:

1. **Página Principal (Home)**: Terminal para ejecutar comandos iniciales (mkdisk, fdisk, mount, mkfs)
2. **Login GUI**: Página dedicada para iniciar sesión (no por comando)
3. **Visualizador**: Navegación paso a paso por el sistema de archivos

## Características

### Temas Visuales
- **Neo Green**: Tema neón verde/azul con estética cyberpunk
- **Aurora Purple**: Tema morado/rosa con gradientes vibrantes
- Switch instantáneo entre temas con persistencia en localStorage

### Componentes Principales

#### 1. Terminal Avanzada (Especificación Proyecto 1 + 2)

**Área de Entrada de Comandos:**
- Textarea multilinea para comandos batch
- Soporte para comentarios con `#`
- 📁 **Botón "Cargar Archivo"**: Carga scripts .smia/.txt
- ▶️ **Botón "Ejecutar"**: Ejecuta todos los comandos en secuencia
- Ctrl+Enter para ejecutar rápidamente

**Área de Salida de Comandos:**
- Output en tiempo real con color coding
- Persistencia con localStorage (no se borra al cambiar de página)
- Auto-scroll automático
- Mensajes del servidor procesados
- Botón limpiar para resetear

#### 2. Sistema de Autenticación (Login GUI)
- **Página dedicada para login** (no comando)
- Autenticación por ID de montaje, usuario y contraseña
- Sesión persistente visible en topbar
- Botón de "Cerrar Sesión" disponible cuando hay sesión activa
- **Validación automática**: Comandos que requieren sesión solo se ejecutan si hay sesión activa

#### 3. Visualizador de Sistema de Archivos (3 Pasos)

**Paso 1: Selección de Disco**
- Muestra todos los discos creados
- Información básica: capacidad, fit, particiones montadas
- Click para seleccionar y pasar al siguiente paso

**Paso 2: Selección de Partición**
- Muestra información del disco seleccionado
- Lista todas las particiones del disco
- Información: tipo, tamaño, fit, estado
- Click para seleccionar partición

**Paso 3: Navegación del Sistema de Archivos**
- **Requiere sesión activa** (redirige a login si no hay sesión)
- Explorador read-only desde la raíz "/"
- Navegación con breadcrumb interactivo
- Muestra archivos y carpetas con iconos 📁/📄
- **Permisos visibles** (UGO): uid, gid, permisos
- Visualización de contenido de archivos de texto
- Panel de Journaling (EXT3) con operaciones en tiempo real

### Animaciones y Efectos

- Logo con efecto **pulse** (glow pulsante)
- Cards con **hover elevation**
- Botones con **glow effect** al hover
- Inputs con **focus glow**
- Transiciones suaves en cambio de tema (0.3s)
- Breadcrumbs interactivos con hover
- Items del explorador con border glow

## Estructura del Proyecto

```
Frontend/
├── index.html              # Punto de entrada HTML
├── package.json            # Dependencias y scripts
├── vite.config.js          # Configuración de Vite + proxy
├── .gitignore             # Archivos ignorados
└── src/
    ├── main.jsx           # Punto de entrada React
    ├── App.jsx            # Componente raíz con routing
    ├── styles.css         # Estilos globales + temas
    ├── lib/
    │   └── api.js         # Cliente API (fetch wrappers)
    ├── components/
    │   ├── Topbar.jsx            # Header con sesión y theme switch
    │   ├── Shell.jsx             # Terminal interactiva
    │   ├── LoginDialog.jsx       # Modal de login
    │   ├── DiskPicker.jsx        # Selector de discos
    │   ├── PartitionPicker.jsx   # Selector de particiones
    │   ├── Explorer.jsx          # Explorador de archivos
    │   └── JournalPanel.jsx      # Panel de journaling
    └── pages/
        ├── Home.jsx              # Página principal (terminal)
        └── Visualizer.jsx        # Página de visualización
```

## Instalación y Uso

### Prerrequisitos
- Node.js 18+
- npm o yarn
- Backend corriendo en `http://localhost:8080`

### Instalación

```bash
cd Frontend
npm install
```

### Desarrollo

```bash
npm run dev
```

La aplicación estará disponible en `http://localhost:5173`

### Build para Producción

```bash
npm run build
```

Los archivos se generarán en la carpeta `dist/`

### Preview de Producción

```bash
npm run preview
```

## Configuración del Proxy

El archivo `vite.config.js` incluye un proxy que redirige las siguientes rutas al backend:

- `/api/*` → `http://localhost:8080`
- `/health` → `http://localhost:8080`
- `/reports/*` → `http://localhost:8080`

Si tu backend corre en otro puerto, modifica la propiedad `target` en `vite.config.js`.

## API Endpoints Utilizados

### Backend Esperado

El frontend espera los siguientes endpoints en el backend:

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/health` | GET | Estado del servidor |
| `/api/commands` | POST | Ejecutar comando individual |
| `/api/script` | POST | Ejecutar script .smia |
| `/api/reports` | POST | Generar reporte |
| `/api/auth/login` | POST | Iniciar sesión |
| `/api/auth/logout` | POST | Cerrar sesión |
| `/api/disks` | GET | Listar discos |
| `/api/disks/partitions` | GET | Listar particiones (query: `?path=...`) |
| `/api/fs/list` | GET | Listar archivos/dirs (query: `?id=...&path=...`) |
| `/api/fs/file` | GET | Obtener contenido de archivo (query: `?id=...&path=...`) |
| `/api/journaling` | GET | Obtener transacciones (query: `?id=...`) |

### Ejemplo de Request/Response

**Login**
```javascript
POST /api/auth/login
{
  "id": "841A",
  "user": "root",
  "pass": "123"
}
→ { "success": true, "user": "root" }
```

**Ejecutar Comando**
```javascript
POST /api/commands
{
  "input": "mkdisk -size=10 -unit=M -path=\"/tmp/d1.mia\""
}
→ { "output": "Disco creado exitosamente" }
```

**Listar Archivos**
```javascript
GET /api/fs/list?id=841A&path=/home
→ {
  "items": [
    { "name": "docs", "type": "dir", "perm": "755", "uid": 1, "gid": 1 },
    { "name": "file.txt", "type": "file", "perm": "644", "uid": 1, "gid": 1, "size": 1024 }
  ]
}
```

## Comandos Soportados

El frontend permite ejecutar todos los comandos del proyecto:

### Gestión de Discos
- `mkdisk` - Crear disco
- `rmdisk` - Eliminar disco

### Gestión de Particiones
- `fdisk` - Crear/eliminar particiones
- `mount` - Montar partición
- `unmount` - Desmontar partición

### Sistema de Archivos
- `mkfs` - Crear filesystem (2fs/3fs)
- `login` - Iniciar sesión (también por GUI)
- `logout` - Cerrar sesión

### Operaciones de Archivos/Directorios
- `mkdir` - Crear directorio
- `mkfile` - Crear archivo
- `cat` - Ver contenido
- `remove` - Eliminar
- `edit` - Editar archivo
- `rename` - Renombrar
- `copy` - Copiar
- `move` - Mover
- `find` - Buscar
- `chown` - Cambiar propietario
- `chmod` - Cambiar permisos

### Reportes
- `rep` - Generar reportes (mbr, disk, sb, inode, block, tree, file, ls, bm_inode, bm_block)

## Flujo de Trabajo según Especificación del Proyecto

### Paso 1: Página Principal - Crear Discos y Particiones

En la terminal de la página principal, ejecute comandos para crear la infraestructura:

```bash
# Crear disco
mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"

# Crear particiones
fdisk -size=1024 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part1
fdisk -add -size=2048 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part2

# Montar partición (devuelve un ID, ej: 841A)
mount -path="/tmp/Disco1.mia" -name=Part1

# Formatear con sistema de archivos
mkfs -id=841A -type=3fs  # EXT3 con journaling
```

### Paso 2: Iniciar Sesión (GUI)

1. Click en el botón **"Iniciar Sesión"** en la página principal
2. Ingresar:
   - **ID de montaje**: El ID retornado por el comando `mount` (ej: 841A)
   - **Usuario**: root (o el usuario que haya creado)
   - **Contraseña**: 123 (contraseña del usuario)
3. Click en "Iniciar Sesión"

**Nota**: El login ahora es **exclusivamente por GUI**, no por comando.

### Paso 3: Crear Usuarios y Grupos (Requiere Sesión)

De vuelta en la terminal (ahora con sesión activa):

```bash
# Crear grupos
mkgrp -name=usuarios

# Crear usuarios
mkusr -user=user1 -pass=pass1 -grp=usuarios

# Crear archivos y carpetas
mkdir -id=841A -path=/home
mkdir -id=841A -path=/home/user1
mkfile -id=841A -size=100 -path=/home/user1/archivo.txt
cat -id=841A -file1=/home/user1/archivo.txt

# Cambiar permisos
chmod -id=841A -path=/home/user1 -ugo=777
chown -id=841A -path=/home/user1/archivo.txt -user=user1
```

### Paso 4: Explorar en el Visualizador

1. Click en **"Visualizador"**
2. **Paso 1**: Seleccionar el disco creado
3. **Paso 2**: Seleccionar la partición montada
4. **Paso 3**: Navegar por el sistema de archivos
   - Ver carpetas desde la raíz "/"
   - Abrir carpetas haciendo click en "Abrir Carpeta"
   - Ver contenido de archivos haciendo click en "Ver Contenido"
   - Observar permisos (uid, gid, permisos UGO)
   - Ver operaciones en el panel de Journaling (EXT3)

### Paso 5: Generar Reportes

```bash
rep -name=tree -id=841A -path=/reports/tree.png
rep -name=sb -id=841A -path=/reports/sb.txt
rep -name=file -id=841A -path=/reports/file.png -ruta=/home/user1/archivo.txt
rep -name=ls -id=841A -path=/reports/ls.png -ruta=/home
```

## Temas y Personalización

### Cambiar Tema
Click en el botón de tema en la topbar:
- 🟢 Neo Green
- 🟣 Aurora Purple

El tema se guarda automáticamente en `localStorage`.

### Variables CSS Disponibles

Puedes personalizar el tema editando `src/styles.css`:

```css
:root {
  --bg: #02060a;          /* Fondo principal */
  --panel: #07131a;       /* Fondo de paneles */
  --neon: #00ff95;        /* Color neón principal */
  --neo2: #57b6ff;        /* Color secundario */
  --danger: #ff5c7c;      /* Color de error */
  --txt: #d7f7ee;         /* Color de texto */
  --muted: #78b7a5;       /* Color de texto secundario */
  --border: #0a2731;      /* Color de bordes */
}
```

## Responsive Design

El layout es completamente responsive:

- **Desktop (>1100px)**: Grid de 2 columnas (1.25fr + 1fr)
- **Tablet/Mobile (<1100px)**: Colapsa a 1 columna
- Cards y listas se ajustan automáticamente
- Breadcrumbs con flex-wrap
- Inputs y botones ocupan 100% del ancho en móvil

## Troubleshooting

### El backend no responde
1. Verifica que esté corriendo en `:8080`
2. Revisa la consola del navegador (F12)
3. Verifica el proxy en `vite.config.js`

### CORS errors
El proxy de Vite debería manejar CORS. Si persiste:
- Asegúrate de acceder vía `localhost:5173` (no 127.0.0.1)
- Verifica headers CORS en el backend

### Tema no persiste
El tema usa `localStorage`. Verifica que:
- El navegador permita localStorage
- No estés en modo incógnito

### Errores de rutas en producción
Si usas React Router con build, configura tu servidor para redirect all a `index.html`:

**Nginx**
```nginx
location / {
  try_files $uri $uri/ /index.html;
}
```

## Tecnologías

- **React 18.3** - UI Library
- **Vite 5.4** - Build tool & dev server
- **React Router 6.26** - Client-side routing
- **CSS Variables** - Theming dinámico
- **LocalStorage API** - Persistencia de tema

## Estructura de Datos

### Disk Object
```typescript
{
  path: string,
  size: number,
  fit: "BF" | "FF" | "WF",
  mounted: string[]  // IDs montados
}
```

### Partition Object
```typescript
{
  name: string,
  type: "P" | "E" | "L",
  fit: "BF" | "FF" | "WF",
  size: number
}
```

### File/Dir Item
```typescript
{
  name: string,
  type: "file" | "dir",
  perm: string,  // "755", "644", etc
  uid: number,
  gid: number,
  size?: number  // solo para files
}
```

### Journal Entry
```typescript
{
  operation: string,  // "mkfile", "mkdir", "remove", etc
  path: string,
  extra?: string,     // info adicional
  time: number        // unix timestamp
}
```

## Contribuir

Este es un proyecto educativo del curso de MIA (USAC).

---

**Desarrollado con** ⚡ Vite + ⚛️ React + 🎨 CSS Variables

**Tema**: Neo Futurista 🟢🟣
