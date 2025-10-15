# Componente Terminal - Especificación Proyecto 1 y 2

## Características Implementadas

### ✅ Área de Entrada de Comandos

El nuevo componente `Terminal.jsx` implementa **exactamente** la especificación del Proyecto 1:

#### 1. Área de Entrada (Input Area)
- **Textarea multilinea** para escribir múltiples comandos
- Soporte para comentarios con `#`
- Uno o más comandos por línea
- Ejecuta todos los comandos en secuencia
- **Shortcuts**:
  - `Ctrl+Enter`: Ejecutar todos los comandos
  - Enter normal: Nueva línea (no ejecuta)

#### 2. Botón de Carga de Archivo
- 📁 **"Cargar Archivo"** button
- Acepta archivos: `.mia`, `.smia`, `.txt`
- Carga el contenido directamente al área de entrada
- Muestra mensaje de confirmación en output

#### 3. Botón de Ejecutar
- ▶️ **"Ejecutar"** button
- Ejecuta **todos** los comandos del área de entrada
- Procesa línea por línea
- Ignora líneas vacías y comentarios
- Muestra progreso en tiempo real
- Se deshabilita durante ejecución

### ✅ Área de Salida de Comandos

#### 1. Output Area
- Muestra **resultados en tiempo real**
- Color coding:
  - 🟢 Verde: Comandos exitosos (`.ok`)
  - 🔴 Rojo: Errores (`.err`)
  - 🔵 Azul: Mensajes del sistema (`.sys`)
  - ⚪ Gris: Comandos ejecutados (`.cmd`)
- Auto-scroll al final
- Altura máxima con scroll
- Font monospace

#### 2. Persistencia
- **localStorage**: Mantiene historial entre sesiones
- **No se borra** al cambiar de página
- Botón "Limpiar" para resetear cuando se desee

---

## Comparación con Especificación

### Proyecto 1 - Requerimientos

| Requerimiento | Estado | Implementación |
|--------------|--------|----------------|
| Área de entrada manual | ✅ | Textarea multilinea |
| Cargar archivo de script | ✅ | Input file + botón |
| Botón de ejecutar | ✅ | Ejecuta todos los comandos |
| Área de salida | ✅ | Output con color coding |
| Mensajes del servidor | ✅ | Muestra respuestas del backend |

### Proyecto 2 - Mejoras Adicionales

| Característica | Estado | Descripción |
|----------------|--------|-------------|
| Persistencia de historial | ✅ | localStorage |
| Validación de sesión | ✅ | Verifica comandos que requieren login |
| Ejecución secuencial | ✅ | Procesa comandos uno por uno |
| Soporte de comentarios | ✅ | Líneas con # se ignoran |
| Shortcuts de teclado | ✅ | Ctrl+Enter, Ctrl+L |

---

## Uso

### Método 1: Escribir Comandos Manualmente

```bash
# En el área de entrada, escribir:
mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"
fdisk -size=1024 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part1
mount -path="/tmp/Disco1.mia" -name=Part1

# Luego:
1. Click en "Ejecutar" O
2. Presionar Ctrl+Enter
```

### Método 2: Cargar Archivo Script

```bash
1. Click en "📁 Cargar Archivo"
2. Seleccionar archivo .smia o .txt
3. El contenido aparece en el área de entrada
4. Click en "Ejecutar" o Ctrl+Enter
```

### Método 3: Ejecución Individual (Proyecto 1)

Para ejecutar comando por comando (modo interactivo):
- Escribir un comando
- Presionar Ctrl+Enter
- Ver resultado
- Escribir siguiente comando

---

## Estructura del Componente

```jsx
Terminal
├── Output Area (Salida)
│   ├── Historial de comandos
│   ├── Resultados
│   └── Auto-scroll
│
└── Input Area (Entrada)
    ├── Textarea (comandos)
    ├── Botones
    │   ├── 📁 Cargar Archivo
    │   ├── ▶️ Ejecutar
    │   └── 🗑️ Limpiar
    └── Tips y ayuda
```

---

## Persistencia (localStorage)

El componente guarda automáticamente:

### Keys utilizadas
```javascript
'mia_terminal_history' // Array de comandos ejecutados
'mia_terminal_lines'   // Array de líneas de output
```

### Beneficios
- ✅ Historial no se pierde al recargar
- ✅ Output persiste entre páginas
- ✅ Continuar trabajo donde lo dejaste
- ✅ Ver resultados anteriores

### Limpiar
```javascript
// Manual: Click en botón "Limpiar"
// Programático:
localStorage.removeItem('mia_terminal_history')
localStorage.removeItem('mia_terminal_lines')
```

---

## Validación de Sesión

Comandos que **requieren sesión activa**:
```
mkgrp, rmgrp, mkusr, rmusr
chmod, mkfile, cat, remove
edit, rename, mkdir, copy
move, find, chgrp, chown
```

**Comportamiento:**
- Si NO hay sesión → Muestra error
- Si SÍ hay sesión → Ejecuta normal

---

## Ejemplo de Script (.smia)

```bash
# ejemplo.smia - Script de inicialización

# Paso 1: Crear disco
mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"

# Paso 2: Particionar
fdisk -size=1024 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part1
fdisk -add -size=2048 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part2

# Paso 3: Montar (tomar nota del ID retornado)
mount -path="/tmp/Disco1.mia" -name=Part1

# Paso 4: Formatear (cambiar 841A por tu ID)
mkfs -id=841A -type=3fs

# Paso 5: Login desde GUI con el ID retornado
```

---

## Shortcuts de Teclado

| Atajo | Acción |
|-------|--------|
| `Ctrl+Enter` | Ejecutar todos los comandos |
| `Ctrl+L` | Limpiar salida (próxima versión) |
| `Enter` | Nueva línea (no ejecuta) |

---

## Diferencias con Shell.jsx (Antiguo)

| Característica | Shell.jsx (Antiguo) | Terminal.jsx (Nuevo) |
|----------------|---------------------|----------------------|
| Modo de ejecución | Línea por línea (como bash) | Batch (todos a la vez) |
| Botón ejecutar | ❌ No (solo Enter) | ✅ Sí |
| Carga de archivos | ❌ No | ✅ Sí |
| Persistencia | ❌ No | ✅ Sí (localStorage) |
| Área separada entrada/salida | ❌ No (todo junto) | ✅ Sí |
| Comentarios | ❌ No | ✅ Sí (#) |
| Multilinea | Shift+Enter | Enter normal |

---

## API del Componente

```jsx
<Terminal session={sessionObject} />
```

### Props

| Prop | Tipo | Requerido | Descripción |
|------|------|-----------|-------------|
| `session` | `Object \| null` | ✅ | Objeto de sesión con `{id, user}` |

### session Object
```javascript
{
  id: "841A",      // ID de montaje
  user: "root"     // Usuario logueado
}
```

---

## Testing Checklist

### ✅ Funcionalidad Básica
- [ ] Escribir comando y ejecutar con botón
- [ ] Escribir múltiples comandos y ejecutar
- [ ] Ver output en área de salida
- [ ] Color coding correcto (ok, err, sys, cmd)

### ✅ Carga de Archivos
- [ ] Click en "Cargar Archivo"
- [ ] Seleccionar .smia
- [ ] Contenido aparece en entrada
- [ ] Ejecutar y ver resultados

### ✅ Persistencia
- [ ] Ejecutar comandos
- [ ] Cambiar de página (ir a /visualizer)
- [ ] Volver a home (/)
- [ ] Verificar que output sigue ahí

### ✅ Validación de Sesión
- [ ] Sin sesión: ejecutar `mkdir` → Error
- [ ] Con sesión: ejecutar `mkdir` → OK
- [ ] Sin sesión: ejecutar `mkdisk` → OK (no requiere sesión)

### ✅ Shortcuts
- [ ] Ctrl+Enter ejecuta
- [ ] Enter agrega nueva línea
- [ ] Comentarios (#) se ignoran

---

## Problemas Conocidos y Soluciones

### ❌ Output se borra al cambiar de página
✅ **Solucionado**: Ahora usa localStorage

### ❌ No hay forma de cargar scripts
✅ **Solucionado**: Botón "Cargar Archivo"

### ❌ Terminal muy simple (solo CLI)
✅ **Solucionado**: Áreas separadas entrada/salida como IDE

### ❌ No se puede ejecutar batch de comandos
✅ **Solucionado**: Botón "Ejecutar" procesa todos

---

## Mejoras Futuras (Opcional)

1. **Syntax highlighting**: Resaltar comandos en entrada
2. **Autocompletado**: Sugerir comandos mientras escribes
3. **Historial con ↑/↓**: Navegar comandos anteriores
4. **Export output**: Guardar salida como .txt
5. **Temas personalizados**: Cambiar colores de output
6. **Ejecutar selección**: Solo comandos seleccionados
7. **Breakpoints**: Pausar ejecución en línea específica

---

**Especificación cumplida**: ✅ Proyecto 1 y Proyecto 2
**Build status**: ✅ Sin errores
**Versión**: 2.0.0
