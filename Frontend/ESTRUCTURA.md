# Estructura del Frontend - MIA Proyecto 2

## Proyecto: GoDisk 2.0 — Neo UI
**Tecnología:** React 18.3 + Vite 5.4 + React Router 6.26
**Propósito:** Frontend moderno y futurista para gestión de sistemas de archivos EXT2/EXT3

---

## Estructura de Directorios

```
Frontend/
├─ index.html                      # Punto de entrada HTML
├─ package.json                    # Dependencias NPM y scripts
├─ package-lock.json               # Lock file de dependencias
├─ vite.config.js                  # Configuración de Vite con proxy API
├─ .env                            # Variables de entorno (desarrollo)
├─ .env.example                    # Template de variables de entorno
├─ .gitignore                      # Reglas de ignore para Git
├─ README.md                       # Documentación completa del proyecto
├─ CHANGELOG.md                    # Historial de versiones
├─ RESPONSIVE.md                   # Documentación de diseño responsive
├─ TERMINAL.md                     # Especificación del componente Terminal
├─ ejemplo.smia                    # Ejemplo de script SMIA
├─ dist/                           # Build de producción
│  ├─ index.html                   # HTML compilado
│  └─ assets/
│     ├─ index-V215Q3AL.js         # JavaScript bundle (~177 KB)
│     └─ index-I4Er14iL.css        # CSS bundle (~6.7 KB)
├─ src/                            # Código fuente
│  ├─ main.jsx                     # Punto de entrada React DOM
│  ├─ App.jsx                      # Componente raíz con routing
│  ├─ styles.css                   # Estilos globales + definición de temas
│  ├─ lib/
│  │  └─ api.js                    # Cliente API (fetch wrappers)
│  ├─ components/
│  │  ├─ Topbar.jsx                # Header con sesión y toggle de tema
│  │  ├─ Terminal.jsx              # Terminal interactiva de comandos
│  │  ├─ DiskPicker.jsx            # Paso 1: Selección de disco
│  │  ├─ PartitionPicker.jsx       # Paso 2: Selección de partición
│  │  ├─ Explorer.jsx              # Paso 3: Navegador de archivos
│  │  ├─ JournalPanel.jsx          # Visor de journaling EXT3
│  │  ├─ ReportsGallery.jsx        # Galería 3D de reportes
│  │  ├─ ReportsCarousel.jsx       # Variante estándar del carrusel
│  │  ├─ ReportsCarousel3D.jsx     # Variante 3D avanzada
│  │  └─ ImageLightbox.jsx         # Visor de imágenes con zoom
│  └─ pages/
│     ├─ Home.jsx                  # Terminal principal y ejecución de comandos
│     ├─ LoginPage.jsx             # Interfaz GUI de login
│     ├─ Visualizer.jsx            # Navegador de sistema de archivos (3 pasos)
│     └─ Reports.jsx               # Interfaz de galería de reportes
└─ node_modules/                   # Dependencias (no rastreado)
```

---

## Archivos de Configuración

### `package.json`
- **Metadatos:** name "godisk-neo-ui", version 0.1.0
- **Scripts NPM:**
  - `npm run dev` - Servidor de desarrollo Vite en puerto 5173
  - `npm run build` - Build de producción a `dist/`
  - `npm run preview` - Vista previa del build de producción
- **Dependencias:**
  - React 18.3.1 - Framework UI
  - React DOM 18.3.1 - Rendering
  - React Router DOM 6.26.2 - Routing del lado del cliente
- **Dev Dependencies:**
  - @vitejs/plugin-react 4.3.1 - Fast Refresh para React
  - Vite 5.4.2 - Build tool y dev server

### `vite.config.js`
- Configura Vite con plugin de React
- Servidor de desarrollo en puerto 5173
- Configuración de proxy para backend API:
  - `/api/*` → `http://localhost:8080`
  - `/health` → `http://localhost:8080`
  - `/reports/*` → `http://localhost:8080`

### `index.html`
- Punto de entrada HTML5 con meta viewport para responsive
- Div raíz con id="root"
- Carga app React vía `/src/main.jsx` como módulo ES

---

## Puntos de Entrada del Código Fuente

### `src/main.jsx`
- Importa React DOM createRoot
- Monta app React en elemento del documento con id="root"
- Renderiza componente App

### `src/App.jsx`
- Componente raíz con configuración de BrowserRouter
- Usa React Router para navegación del lado del cliente
- Gestiona estado de sesión global con useState
- **Rutas:**
  - `/` → Página Home (terminal)
  - `/login` → Página de Login (autenticación GUI)
  - `/visualizer` → Visualizador de sistema de archivos (3 pasos)
  - `/reports` → Galería de reportes
- Provee componente Topbar (header) en todas las rutas
- Objeto de sesión: `{id, user}` para estado de autenticación

---

## Capa de Utilidades y API

### `src/lib/api.js`
Cliente API centralizado con wrappers de fetch

**Métodos:**
- `health()` - Verificar estado del backend
- `login(id, user, pass)` - Autenticar usuario
- `logout()` - Finalizar sesión
- `run(line)` - Ejecutar comando individual
- `runScript(script, stopOnError)` - Ejecutar batch de comandos
- `disks()` - Listar todos los discos
- `partitions(diskPath)` - Obtener particiones de un disco
- `list(id, path)` - Listar archivos/carpetas en directorio
- `file(id, path)` - Obtener contenido de archivo
- `journaling(id)` - Obtener entradas de journal EXT3
- `genReport(name, id, out, extra)` - Generar reporte
- `listReports()` - Listar todos los reportes generados

---

## Páginas (Rutas)

### `src/pages/Home.jsx`
**Propósito:** Interfaz principal de terminal para ejecución de comandos

**Estructura:**
1. Barra de acciones con botones de navegación:
   - "Iniciar Sesión" - solo si no hay sesión
   - "Visualizador"
   - "Reportes"
2. Información de sesión (cuando está logueado)
3. Componente Terminal (flex: 1, altura mínima 600px)
4. Card de información con categorías de comandos:
   - Gestión de discos y particiones (mkdisk, fdisk, mount, unmount)
   - Sistema de archivos (mkfs, login, logout, mkgrp, mkusr)
   - Operaciones de archivos (mkdir, mkfile, remove, edit, rename, copy, move, cat)
   - Permisos y reportes (chmod, chown, chgrp, find, rep, exec)
5. Sección de tips con atajos de teclado

### `src/pages/LoginPage.jsx`
**Propósito:** Interfaz de autenticación basada en GUI

**Campos del Formulario:**
1. ID de Montaje - ej. "841A" del comando mount
2. Usuario - ej. "root"
3. Contraseña - campo oculto

**Flujo:**
1. Usuario llena formulario y envía
2. Llama API.login(id, user, pass)
3. En éxito: Establece estado de sesión y navega a "/"
4. En error: Muestra mensaje de error

### `src/pages/Visualizer.jsx`
**Propósito:** Interfaz guiada de 3 pasos para explorar sistemas de archivos

**Flujo de 3 Pasos:**

1. **Paso 1: Selección de Disco** (DiskPicker)
   - Lista todos los discos creados
   - Muestra: path, capacidad, algoritmo fit, particiones montadas
   - Click "Explorar" para proceder

2. **Paso 2: Selección de Partición** (PartitionPicker)
   - Muestra información del disco seleccionado
   - Lista particiones con: nombre, tipo (P/E/L), fit, tamaño
   - Click "Montar / Explorar" para proceder

3. **Paso 3: Navegación del Sistema de Archivos**
   - Requiere sesión activa (redirige a login si no)
   - Componente Explorer: navegación de archivos (solo lectura)
   - Componente JournalPanel: operaciones EXT3 en tiempo real

### `src/pages/Reports.jsx`
**Propósito:** Interfaz de galería para reportes generados

**Características:**
- Componente ReportsGallery
- Navegación de vuelta a Visualizador y Home
- Auto-refresco cada 5 segundos
- Opciones de filtro: Todos, Imágenes, Texto
- Funcionalidad de descarga

---

## Componentes

### `src/components/Topbar.jsx`
**Propósito:** Header mostrando título de app, sesión, tema y estado del backend

**Elementos:**
1. Sección de marca:
   - Logo animado (efecto pulse glow)
   - Título: "MIA - Proyecto 2"
   - Subtítulo: "Sistema de Archivos EXT2/EXT3"

2. Sección derecha:
   - Botón toggle de tema ("Neo Green" / "Aurora Purple")
   - Badge de estado del backend
   - Info de sesión:
     - Badge de usuario (solo si está logueado)
     - Badge de ID (solo si está logueado)
     - Botón Logout (solo si está logueado)

**Gestión de Estado:**
- Estado de tema: lee de localStorage, persiste en cambio
- Health check: llamada única al montar
- Actualiza atributo data-theme en raíz del documento

### `src/components/Terminal.jsx`
**Propósito:** Interfaz avanzada de ejecución de comandos con procesamiento batch

**Características Principales:**

1. **Área de Output:**
   - Historial de comandos scrollable
   - Header de prompt: "julian@pop-os:~/home $"
   - Estadísticas en tiempo real: Total, OK (verde), ERR (rojo)
   - Botón resetear estadísticas
   - Auto-scroll al final
   - Líneas con código de color

2. **Área de Input:**
   - Textarea multilínea para comandos
   - Botones:
     - "Cargar Archivo" - Cargar scripts .mia/.smia/.txt
     - "Ejecutar" - Ejecutar todos los comandos
     - "Limpiar" - Limpiar output
   - Atajo de teclado: Ctrl+Enter para ejecutar
   - Soporte para comentarios (líneas que inician con #)

3. **Persistencia** (localStorage):
   - `mia_terminal_history` - Historial de comandos
   - `mia_terminal_lines` - Líneas de output
   - `mia_terminal_stats` - Estadísticas de ejecución

4. **Procesamiento Batch:**
   - Divide input por líneas nuevas
   - Filtra comentarios y líneas vacías
   - Ejecuta secuencialmente
   - Rastrea conteos de éxito/error

### `src/components/DiskPicker.jsx`
**Propósito:** Paso 1 del visualizador - seleccionar un disco para explorar

**Visualización de Datos:**
- Lista todos los discos disponibles
- Muestra por cada disco: Path, Capacidad, Algoritmo Fit, Particiones montadas
- Click "Explorar" para seleccionar disco
- Procede al Paso 2

### `src/components/PartitionPicker.jsx`
**Propósito:** Paso 2 del visualizador - seleccionar una partición del disco seleccionado

**Visualización de Datos:**
- Lista todas las particiones del disco seleccionado
- Muestra por partición: Nombre, Tipo, Algoritmo Fit, Tamaño
- Click botón para seleccionar partición
- Procede al Paso 3

### `src/components/Explorer.jsx`
**Propósito:** Paso 3 del visualizador - navegar sistema de archivos (solo lectura)

**Características:**

1. **Navegación Breadcrumb:**
   - Muestra ruta actual
   - Click en cualquier parte para saltar a ese directorio
   - Botón root "/" siempre disponible

2. **Listado de Archivos/Directorios:**
   - Muestra tipo (DIR/FILE)
   - Nombre de archivo
   - Permisos (formato UGO)
   - User ID (uid)
   - Group ID (gid)
   - Tamaño de archivo (solo para archivos)
   - Botones de acción:
     - "Abrir Carpeta" para directorios
     - "Ver Contenido" para archivos

3. **Visor de Contenido de Archivos:**
   - Vista modal dentro de la misma card
   - Muestra contenido en monospace
   - Botón volver al listado

### `src/components/JournalPanel.jsx`
**Propósito:** Visor de operaciones de journaling EXT3 en tiempo real

**Características:**
- Intervalo de polling: 3 segundos
- Muestra entradas de journal con:
  - Tipo de operación (mkfile, mkdir, remove, etc.)
  - Path afectado
  - Info extra (ej. nuevo nombre en rename)
  - Timestamp (formateado con toLocaleString)

### `src/components/ReportsGallery.jsx`
**Propósito:** Galería carousel 3D avanzada para reportes

**Características:**

1. **Controles de Galería:**
   - Filtros: Todos / Imágenes / Texto
   - Botón refrescar
   - Botones de navegación manual (< >)
   - Flechas de teclado (izquierda/derecha)
   - Scroll con rueda del mouse
   - Indicadores de puntos para cada item

2. **Visualización de Cards:**
   - Efectos de transformación 3D
   - Card activa en centro (scale 1, opacity 1)
   - Cards adyacentes escaladas (0.85), atenuadas (0.7 opacity)
   - Transiciones suaves (0.5s cubic-bezier)

3. **Vista Previa de Archivos:**
   - **Imágenes:** Thumbnail, click para lightbox
   - **Texto:** Ícono de archivo, metadata, botón descarga
   - Auto-refresco: polling cada 5 segundos

### `src/components/ImageLightbox.jsx`
**Propósito:** Visor de imágenes en pantalla completa con zoom

**Características:**
- Modal overlay de pantalla completa
- Controles de zoom (-, 100%, +)
- Rango de escala: 0.25x a 5x
- Click fuera para cerrar

---

## Estilos y Tematización

### `src/styles.css` (~474 líneas)

**Variables de Tema:**

**Neo Green (default):**
- `--bg: #02060a`
- `--panel: #07131a`
- `--neon: #00ff95` (acento primario)
- `--neo2: #57b6ff` (acento secundario)
- `--danger: #ff5c7c`
- `--txt: #d7f7ee`

**Aurora Purple:**
- `--bg: #0a0214`
- `--panel: #13061f`
- `--neon: #b98aff` (acento primario)
- `--neo2: #ff6ec7` (acento secundario)
- `--txt: #f4e8ff`

**Clases de Componentes:**
- `.card` - Contenedor con gradiente y efectos hover
- `.btn` / `.btn.alt` - Estilos de botón con brillo en hover
- `.input`, `.select`, `.textarea` - Inputs de formulario
- `.badge` - Badges de estado con bordes
- `.topbar` - Gradiente y sombra del header
- `.logo` - Efecto de pulso animado
- `.explorer` - Layout del navegador de archivos
- `.journalRow` - Visualización de entradas de journal
- `.cflow` - Contenedor de carousel 3D

**Breakpoints Responsive:**
1. **Desktop** (> 1100px): Layout completo de 2 columnas
2. **Tablet** (768px - 1100px): 1 columna, espaciado ajustado
3. **Mobile** (< 480px): Layout compacto, tamaños de fuente reducidos

---

## Integración con API

**Endpoints del Backend Esperados:**

| Endpoint | Método | Body | Respuesta |
|----------|--------|------|-----------|
| `/health` | GET | - | `{status: "ok"}` |
| `/api/auth/login` | POST | `{id, user, pass}` | `{success, user}` |
| `/api/auth/logout` | POST | - | - |
| `/api/commands` | POST | `{input}` | `{output}` |
| `/api/script` | POST | `{script, stopOnError}` | `{output}` |
| `/api/disks` | GET | - | `{disks: []}` |
| `/api/disks/{path}/partitions` | GET | - | `{partitions: []}` |
| `/api/fs/{id}/tree?path=...` | GET | - | `{items: []}` |
| `/api/fs/{id}/file?path=...` | GET | - | `{content}` |
| `/api/journal/{id}/table` | GET | - | `{entries: []}` |
| `/api/reports` | POST | `{name, id, out, extra}` | `{path}` |
| `/api/reports/list` | GET | - | `{files: []}` |

---

## Build y Despliegue

**Desarrollo:**
```bash
npm install
npm run dev
# Se ejecuta en http://localhost:5173
```

**Build de Producción:**
```bash
npm run build
# Output: directorio dist/ con assets empaquetados
```

**Output del Build:**
- `dist/index.html` - Archivo HTML principal
- `dist/assets/index-[hash].js` - Bundle JavaScript (~177 KB, gzipped: 56 KB)
- `dist/assets/index-[hash].css` - Bundle CSS (~6.7 KB, gzipped: 1.9 KB)

---

## Estadísticas del Proyecto

- **Total de Componentes:** 11 componentes React
- **Total de Páginas:** 4 rutas
- **Líneas de CSS:** ~474 con diseño responsive
- **Temas:** 2 (Neo Green, Aurora Purple)
- **Breakpoints:** 3 (Desktop, Tablet, Mobile)
- **Métodos del Wrapper API:** 13
- **Gestión de Estado:** Nivel de componente + localStorage
- **Tamaño del Bundle:** ~177 KB (gzipped: 56 KB)
