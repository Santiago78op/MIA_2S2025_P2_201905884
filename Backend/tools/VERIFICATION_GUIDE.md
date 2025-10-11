# Guía de Verificación P2 - GoDisk (Carnet 201905884)

## 📋 Resumen

Este documento explica cómo usar el script `verify_p2_84.sh` para validar que tu implementación del Proyecto 2 (P2) cumple con todos los requisitos del P1 y P2.

## 🎯 Qué Valida el Script

### Requisitos P1 (Base)
- ✅ **MKDISK/FDISK**: Creación de discos y particiones (MBR/EBR)
- ✅ **MOUNT/MOUNTED**: Montaje con IDs basados en terminación `84` (ej: `841A`, `842A`, `841B`)
- ✅ **MKFS EXT2**: Formateo con `/users.txt`, bitmaps, superblock

### Requisitos P2 (Nuevos)
- ✅ **MKFS EXT3**: Filesystem con journaling
- ✅ **Comandos EXT3**: `journaling`, `recovery`, `loss`
- ✅ **Endpoints HTTP EXT3**: `/api/ext3/{journaling,recovery,loss}`
- ✅ **Reportes DOT**: `/api/reports/{mbr,disk,tree,journal,sb}`
- ✅ **Consola HTTP**: `/api/cmd/run` con manejo de errores consistente
- ✅ **Scripts**: `/api/cmd/script` para ejecutar múltiples comandos

## 📦 Prerequisitos

### Software Requerido

1. **jq** - Parser JSON para línea de comandos
   ```bash
   sudo apt install jq
   ```

2. **curl** - Cliente HTTP
   ```bash
   sudo apt install curl
   ```

3. **Servidor corriendo** - El backend debe estar activo en `http://localhost:8080`
   ```bash
   # Desde la raíz del proyecto
   ./start.sh
   ```

## 🚀 Uso del Script

### Ejecución Básica

```bash
cd Backend/tools
./verify_p2_84.sh
```

### Output Esperado

El script mostrará output coloreado con el resultado de cada test:

```
╔═══════════════════════════════════════════════════════════╗
║   Verificador P2 - GoDisk (Carnet 201905884)             ║
║   Validación Integral de Componentes P1 y P2             ║
╚═══════════════════════════════════════════════════════════╝

========================================
Verificando Prerequisitos
========================================
[PASS] jq está instalado
[PASS] curl está instalado
[PASS] Servidor HTTP respondiendo correctamente
[PASS] Directorio de pruebas creado: /tmp/godisk_verify_84

========================================
P1: Operaciones de Disco (MKDISK/FDISK)
========================================
[INFO] Creando disco de prueba (5MB)...
[PASS] MKDISK creó el disco correctamente
[INFO] Creando partición primaria P1 (2MB)...
[PASS] FDISK agregó partición primaria
...
```

### Interpretación de Resultados

- **[PASS]** (verde): Test exitoso ✓
- **[FAIL]** (rojo): Test falló ✗
- **[WARN]** (amarillo): Advertencia, no crítico ⚠
- **[INFO]** (azul): Información del proceso ℹ

### Resumen Final

Al final verás un resumen como:

```
========================================
RESUMEN DE VERIFICACIÓN
========================================

Total de tests ejecutados: 35
Tests exitosos: 35
Tests fallidos: 0

Tasa de éxito: 100%

✓ TODOS LOS TESTS PASARON
Tu implementación P2 (Carnet 84) está completa y funcional.
```

## 🔍 Tests Incluidos

### 1. Prerequisitos (4 tests)
- jq instalado
- curl instalado
- Servidor respondiendo
- Directorio de pruebas creado

### 2. P1: Operaciones de Disco (4 tests)
- `mkdisk` crea disco de 5MB
- `fdisk add` agrega partición primaria
- `fdisk add` agrega partición extendida
- `fdisk add` agrega partición lógica (EBR)

### 3. P1: Montaje (3 tests)
- `mount` asigna ID con formato `84<n><letra>`
- `mounted` lista particiones montadas
- API `/api/disks/mounted` funciona

### 4. P1: MKFS EXT2 (2 tests)
- `mkfs -fs=2fs` formatea sin errores
- Estructura EXT2 (superblock, bitmaps, /users.txt)

### 5. P2: MKFS EXT3 (2 tests)
- `mkfs -fs=3fs` formatea con journal
- Journal inicializado correctamente

### 6. P2: Comandos EXT3 (3 tests)
- `journaling -id=XXX` devuelve entradas
- `recovery -id=XXX` ejecuta sin errores
- `loss -id=XXX` ejecuta sin errores

### 7. P2: Endpoints HTTP EXT3 (3 tests)
- `GET /api/ext3/journal?id=XXX`
- `POST /api/ext3/recovery` con JSON body
- `POST /api/ext3/loss` con JSON body

### 8. P2: Reportes DOT (5 tests)
- `GET /api/reports/mbr` devuelve `digraph MBR`
- `GET /api/reports/disk` devuelve `digraph`
- `GET /api/reports/tree` devuelve `digraph`
- `GET /api/reports/sb` devuelve `digraph SuperBlock`
- `GET /api/reports/journal` devuelve `digraph Journal`

### 9. P2: Consola HTTP (4 tests)
- Comando válido ejecuta correctamente
- Comando inválido devuelve error
- Comando incompleto devuelve error
- `/api/cmd/script` ejecuta múltiples comandos

### 10. Endpoints Informativos (3 tests)
- `GET /api/commands` lista comandos
- `GET /api/disks/list` encuentra discos .mia
- `GET /api/disks/info` devuelve info de disco

### 11. Limpieza (2 tests)
- Desmonta particiones
- Elimina archivos de prueba

## 🐛 Troubleshooting

### Error: "jq no está instalado"

**Solución:**
```bash
sudo apt update
sudo apt install jq
```

### Error: "Servidor no responde en http://localhost:8080"

**Solución:**
```bash
# Verificar que el servidor esté corriendo
curl http://localhost:8080/api/health

# Si no responde, iniciar servidor
cd /home/julian/Documents/MIA_2S2025_P2_201905884
./start.sh
```

### Error: "MOUNT asignó ID pero no sigue formato 84X"

**Diagnóstico:**
El formato esperado es `84<correlativo><letra>`:
- Primera partición del Disco1: `841A`
- Segunda partición del Disco1: `842A`
- Primera partición del Disco2: `841B`

**Verificación manual:**
```bash
curl -X POST http://localhost:8080/api/cmd/run \
  -H "Content-Type: application/json" \
  -d '{"line":"mount -path=/tmp/test.mia -name=Part1"}' | jq
```

**Archivo a revisar:**
- `Backend/internal/commands/mount_index.go:89-105` (función `GenerateID`)

### Error: "MKFS EXT2 falló"

**Diagnóstico:**
1. Verificar que la partición esté montada:
   ```bash
   curl http://localhost:8080/api/disks/mounted | jq
   ```

2. Verificar logs del servidor:
   ```bash
   tail -f logs/server.log
   ```

**Archivo a revisar:**
- `Backend/internal/fs/ext2/mkfs.go`

### Error: "Reporte MBR no devuelve formato DOT"

**Diagnóstico:**
El reporte debe comenzar con `digraph MBR {` y terminar con `}`.

**Verificación manual:**
```bash
curl "http://localhost:8080/api/reports/mbr?id=841A"
```

**Archivo a revisar:**
- `Backend/cmd/server/reports_handlers.go:211-238` (función `generateMBRDot`)

### Warning: "Journal no tiene entradas registradas"

**Explicación:**
Esto es normal si acabas de formatear. El journal se llena cuando ejecutas operaciones de escritura.

**Para generar entradas:**
```bash
# 1. Login
curl -X POST http://localhost:8080/api/cmd/run \
  -H "Content-Type: application/json" \
  -d '{"line":"login -user=root -pass=123 -id=841A"}'

# 2. Crear archivo
curl -X POST http://localhost:8080/api/cmd/run \
  -H "Content-Type: application/json" \
  -d '{"line":"mkfile -id=841A -path=/test.txt -cont=hola"}'

# 3. Ver journal
curl "http://localhost:8080/api/ext3/journal?id=841A" | jq
```

## 📊 Ejemplo de Ejecución Completa

```bash
# 1. Iniciar servidor
cd /home/julian/Documents/MIA_2S2025_P2_201905884
./start.sh

# 2. En otra terminal, ejecutar verificador
cd Backend/tools
./verify_p2_84.sh

# Output:
# ╔═══════════════════════════════════════════════════════════╗
# ║   Verificador P2 - GoDisk (Carnet 201905884)             ║
# ║   Validación Integral de Componentes P1 y P2             ║
# ╚═══════════════════════════════════════════════════════════╝
#
# ========================================
# Verificando Prerequisitos
# ========================================
# [PASS] jq está instalado
# [PASS] curl está instalado
# [PASS] Servidor HTTP respondiendo correctamente
# ...
#
# ========================================
# RESUMEN DE VERIFICACIÓN
# ========================================
#
# Total de tests ejecutados: 35
# Tests exitosos: 35
# Tests fallidos: 0
#
# Tasa de éxito: 100%
#
# ✓ TODOS LOS TESTS PASARON
# Tu implementación P2 (Carnet 84) está completa y funcional.
```

## 🔧 Personalización del Script

Si necesitas modificar el script para agregar más tests:

1. **Agregar nuevo test:**
   ```bash
   test_my_feature() {
       log_section "Mi Nueva Funcionalidad"

       log_info "Ejecutando test..."
       local resp
       resp=$(run_command "mi_comando")

       if check_json_ok "$resp"; then
           log_success "Test pasó"
       else
           log_fail "Test falló: $(get_json_field "$resp" "error")"
       fi
   }
   ```

2. **Agregar test a main:**
   ```bash
   main() {
       ...
       test_my_feature  # Agregar aquí
       ...
   }
   ```

## 📝 Checklist Manual (Adicional)

Estos aspectos NO son validados automáticamente por el script:

- [ ] CORS habilitado para frontend en S3/CloudFront
- [ ] Frontend integrado con Viz.js para renderizar reportes DOT
- [ ] Manejo de permisos (chmod/chown) funciona correctamente
- [ ] Login/Logout con usuarios de /users.txt
- [ ] Comandos de archivos (mkfile, remove, edit, rename, copy, move)
- [ ] Comando `find` con patrones
- [ ] Reportes visuales se ven bien en el frontend

## 📚 Referencias

- Documentación P1: `docs/P1.pdf`
- Documentación P2: `docs/P2.pdf`
- README principal: `README.md`
- Código fuente:
  - Commands: `Backend/internal/commands/`
  - Filesystems: `Backend/internal/fs/{ext2,ext3}/`
  - Disk management: `Backend/internal/disk/`
  - API handlers: `Backend/cmd/server/`

## 💡 Tips

1. **Ejecuta el script frecuentemente** durante el desarrollo para detectar regresiones temprano.

2. **Revisa los logs** si algo falla:
   ```bash
   tail -f logs/server.log
   ```

3. **Usa el modo verbose de curl** para debugging:
   ```bash
   curl -v http://localhost:8080/api/cmd/run \
     -H "Content-Type: application/json" \
     -d '{"line":"mounted"}'
   ```

4. **Valida JSON responses** manualmente:
   ```bash
   curl http://localhost:8080/api/disks/mounted | jq '.'
   ```

5. **Limpia estado** entre pruebas:
   ```bash
   # Eliminar todos los discos de prueba
   rm -rf /tmp/godisk_verify_84
   rm -f Discos/*.mia

   # Reiniciar servidor
   ./stop.sh && ./start.sh
   ```

## 🎓 Soporte

Si encuentras problemas con el script o necesitas ayuda:

1. Verifica que cumples todos los prerequisitos
2. Revisa la sección de Troubleshooting
3. Compara tu output con el ejemplo de ejecución completa
4. Revisa los archivos mencionados en "Archivo a revisar"

---

**Autor:** Script de verificación para P2 - MIA 2S2025
**Carnet:** 201905884
**Última actualización:** Octubre 2025
