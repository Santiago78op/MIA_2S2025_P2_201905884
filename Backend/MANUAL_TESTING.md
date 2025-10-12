# Manual de Testing con Scripts

Este manual te guía paso a paso para ejecutar scripts de prueba en el sistema MIA sin necesidad de asistencia de Claude.

## Tabla de Contenidos
1. [Preparación del Entorno](#preparación-del-entorno)
2. [Estructura de Directorios](#estructura-de-directorios)
3. [Crear Archivos de Contenido](#crear-archivos-de-contenido)
4. [Escribir Scripts de Prueba](#escribir-scripts-de-prueba)
5. [Compilar y Ejecutar el Servidor](#compilar-y-ejecutar-el-servidor)
6. [Ejecutar Scripts via API](#ejecutar-scripts-via-api)
7. [Verificar Resultados](#verificar-resultados)
8. [Ejemplos de Scripts](#ejemplos-de-scripts)
9. [Solución de Problemas](#solución-de-problemas)

---

## Preparación del Entorno

### Requisitos previos
- Go instalado (version 1.16+)
- Graphviz instalado (para reportes gráficos)
- curl o herramienta similar para hacer peticiones HTTP

### Instalar Graphviz
```bash
# En Ubuntu/Debian
sudo apt-get install graphviz

# En macOS
brew install graphviz

# Verificar instalación
dot -V
```

---

## Estructura de Directorios

Antes de ejecutar pruebas, crea los directorios necesarios:

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Crear directorios necesarios
mkdir -p Discos      # Para almacenar archivos .mia de discos
mkdir -p Reports     # Para almacenar reportes generados
mkdir -p CONT        # Para archivos de contenido de prueba
```

---

## Crear Archivos de Contenido

Algunos comandos como `mkfile` necesitan archivos de contenido. Créalos antes de ejecutar el script:

### Ejemplo: Archivo NAME.txt
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Crear directorio CONT si no existe
mkdir -p CONT

# Crear archivo con contenido
echo "Julian" > CONT/NAME.txt
```

### Ejemplo: Múltiples archivos de contenido
```bash
# Archivo de texto simple
echo "Este es un contenido de prueba" > CONT/test1.txt

# Archivo multilinea
cat > CONT/users_data.txt << EOF
Usuario: admin
Password: 12345
Grupo: root
EOF

# Archivo con contenido largo
cat > CONT/large_file.txt << EOF
Línea 1
Línea 2
Línea 3
...
EOF
```

---

## Escribir Scripts de Prueba

Los scripts usan la sintaxis del lenguaje SMIA. Cada comando va en una línea.

### Sintaxis básica

```bash
# Comentario: líneas que empiezan con # son ignoradas

# Comando simple
mkdisk -size=50 -unit=M -path=/ruta/disco.mia

# Comando con múltiples parámetros
fdisk -size=10 -unit=M -path=/ruta/disco.mia -type=P -fit=BF -name=Particion1

# Pausa (útil para debugging)
pause
```

### Crear un script de prueba

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Crear script de prueba básico
cat > test_basico.smia << 'EOF'
# Script de prueba básico

# 1. Crear disco
mkdisk -size=10 -unit=M -path=./Discos/TestDisco.mia

# 2. Crear partición primaria
fdisk -size=5 -unit=M -path=./Discos/TestDisco.mia -type=P -fit=BF -name=Part1

# 3. Montar partición
mount -path=./Discos/TestDisco.mia -name=Part1

# 4. Formatear con EXT2
mkfs -type=full -id=841A

# 5. Login como root
login -user=root -pass=123 -id=841A

# 6. Crear directorio
mkdir -path=/home -id=841A

# 7. Crear archivo
mkfile -path=/home/test.txt -id=841A -size=100

# 8. Logout
logout

# 9. Generar reporte de disco
rep -name=disk -path=./Discos/TestDisco.mia -id=841A -ruta=./Reports/test_disk.jpg

# 10. Desmontar
umount -id=841A
EOF
```

---

## Compilar y Ejecutar el Servidor

### Paso 1: Compilar
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend

# Compilar el servidor
go build -o server ./cmd/server

# Verificar que se creó el binario
ls -lh server
```

### Paso 2: Ejecutar el servidor

**Opción A: En primer plano (para ver logs)**
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
./server
```

**Opción B: En segundo plano**
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
./server &

# Guardar el PID para detenerlo después
echo $! > server.pid
```

**Opción C: Con logs en archivo**
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
./server > server.log 2>&1 &
echo $! > server.pid

# Ver logs en tiempo real
tail -f server.log
```

### Paso 3: Verificar que el servidor está corriendo
```bash
# Verificar que escucha en puerto 8080
curl http://localhost:8080/health

# O con netstat
netstat -tuln | grep 8080
```

### Paso 4: Detener el servidor (cuando termines)
```bash
# Si usaste la opción B o C
kill $(cat server.pid)

# O encuentra el proceso
ps aux | grep server
kill <PID>
```

---

## Ejecutar Scripts via API

### Método 1: Con archivo JSON temporal

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Leer el script y crear JSON
SCRIPT_CONTENT=$(cat test_basico.smia)

# Crear archivo JSON temporal
cat > request.json << EOF
{
  "script": "$SCRIPT_CONTENT"
}
EOF

# Ejecutar via API
curl -X POST http://localhost:8080/api/script \
  -H "Content-Type: application/json" \
  -d @request.json

# Limpiar
rm request.json
```

### Método 2: Con jq (más limpio)

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Crear JSON con jq
jq -n --rawfile script test_basico.smia '{script: $script}' > request.json

# Ejecutar
curl -X POST http://localhost:8080/api/script \
  -H "Content-Type: application/json" \
  -d @request.json

# Limpiar
rm request.json
```

### Método 3: Script bash automatizado

Crea un script helper:

```bash
cat > run_test.sh << 'EOF'
#!/bin/bash

# Script para ejecutar tests fácilmente
# Uso: ./run_test.sh <nombre_script.smia>

if [ -z "$1" ]; then
    echo "Uso: $0 <script.smia>"
    exit 1
fi

SCRIPT_FILE=$1

if [ ! -f "$SCRIPT_FILE" ]; then
    echo "Error: Archivo $SCRIPT_FILE no encontrado"
    exit 1
fi

echo "Ejecutando script: $SCRIPT_FILE"

# Crear JSON
jq -n --rawfile script "$SCRIPT_FILE" '{script: $script}' > /tmp/request.json

# Ejecutar
curl -X POST http://localhost:8080/api/script \
  -H "Content-Type: application/json" \
  -d @/tmp/request.json \
  | jq '.'

# Limpiar
rm /tmp/request.json

echo ""
echo "Script ejecutado. Revisa la salida arriba."
EOF

chmod +x run_test.sh
```

Uso del helper:
```bash
./run_test.sh test_basico.smia
```

---

## Verificar Resultados

### Ver reportes generados
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Listar reportes
ls -lh Reports/

# Ver reporte de texto
cat Reports/p4_r3_bm_inode.txt

# Ver imagen (si tienes entorno gráfico)
xdg-open Reports/test_disk.jpg
# o en macOS: open Reports/test_disk.jpg
```

### Ver discos creados
```bash
# Listar discos
ls -lh Discos/

# Ver tamaño de disco
du -h Discos/TestDisco.mia
```

### Ver logs del servidor
```bash
# Si ejecutaste con logs
tail -f server.log

# O ver últimas líneas
tail -n 50 server.log
```

---

## Ejemplos de Scripts

### Script 1: Prueba completa de disco

```bash
cat > test_disco_completo.smia << 'EOF'
# Prueba completa de operaciones de disco

# Crear disco de 50MB
mkdisk -size=50 -unit=M -path=./Discos/Disco1.mia

# Crear partición primaria de 10MB
fdisk -size=10 -unit=M -path=./Discos/Disco1.mia -type=P -fit=BF -name=Part1

# Crear partición primaria de 15MB
fdisk -size=15 -unit=M -path=./Discos/Disco1.mia -type=P -fit=BF -name=Part2

# Crear partición extendida con el resto
fdisk -path=./Discos/Disco1.mia -type=E -fit=BF -name=Extended

# Crear partición lógica dentro de la extendida
fdisk -size=5 -unit=M -path=./Discos/Disco1.mia -type=L -fit=BF -name=Logic1

# Generar reporte MBR
rep -name=mbr -path=./Discos/Disco1.mia -id=841A -ruta=./Reports/test_mbr.txt

# Generar reporte de disco gráfico
rep -name=disk -path=./Discos/Disco1.mia -id=841A -ruta=./Reports/test_disk.jpg
EOF
```

### Script 2: Prueba de filesystem y usuarios

```bash
cat > test_filesystem.smia << 'EOF'
# Prueba de filesystem y usuarios

# Crear y preparar disco
mkdisk -size=20 -unit=M -path=./Discos/DiskFS.mia
fdisk -size=10 -unit=M -path=./Discos/DiskFS.mia -type=P -fit=BF -name=PartFS
mount -path=./Discos/DiskFS.mia -name=PartFS
mkfs -type=full -id=841A

# Login como root
login -user=root -pass=123 -id=841A

# Crear estructura de directorios
mkdir -path=/home -id=841A
mkdir -path=/etc -id=841A
mkdir -path=/var -id=841A

# Crear archivos
mkfile -path=/home/test.txt -id=841A -size=100
mkfile -path=/home/archivo.txt -id=841A -cont=./CONT/NAME.txt

# Leer archivo
cat -file=/home/archivo.txt -id=841A

# Generar reportes
rep -name=inode -path=./Discos/DiskFS.mia -id=841A -ruta=./Reports/fs_inode.txt
rep -name=block -path=./Discos/DiskFS.mia -id=841A -ruta=./Reports/fs_block.txt
rep -name=bm_inode -path=./Discos/DiskFS.mia -id=841A -ruta=./Reports/fs_bm_inode.txt
rep -name=bm_block -path=./Discos/DiskFS.mia -id=841A -ruta=./Reports/fs_bm_block.txt
rep -name=sb -path=./Discos/DiskFS.mia -id=841A -ruta=./Reports/fs_sb.txt
rep -name=tree -path=./Discos/DiskFS.mia -id=841A -ruta=./Reports/fs_tree.png

# Logout
logout
umount -id=841A
EOF
```

### Script 3: Prueba de errores (testing negativo)

```bash
cat > test_errores.smia << 'EOF'
# Prueba de manejo de errores

# Intentar montar disco que no existe (debe fallar)
mount -path=./Discos/NoExiste.mia -name=Part1

# Intentar crear partición en disco que no existe (debe fallar)
fdisk -size=10 -unit=M -path=./Discos/NoExiste.mia -type=P -fit=BF -name=Part1

# Crear disco válido
mkdisk -size=10 -unit=M -path=./Discos/TestErrors.mia

# Intentar crear partición más grande que el disco (debe fallar)
fdisk -size=100 -unit=M -path=./Discos/TestErrors.mia -type=P -fit=BF -name=Part1

# Crear partición válida
fdisk -size=5 -unit=M -path=./Discos/TestErrors.mia -type=P -fit=BF -name=Part1

# Intentar crear partición con mismo nombre (debe fallar)
fdisk -size=2 -unit=M -path=./Discos/TestErrors.mia -type=P -fit=BF -name=Part1

# Limpiar
rmdisk -path=./Discos/TestErrors.mia
EOF
```

---

## Solución de Problemas

### Problema 1: Servidor no inicia

**Error**: "address already in use"

**Solución**:
```bash
# Encontrar proceso usando puerto 8080
lsof -i :8080
# o
netstat -tuln | grep 8080

# Matar proceso
kill <PID>

# Reintentar
./server
```

### Problema 2: Script falla con "falta id"

**Causa**: Comandos después de login necesitan el parámetro `-id=841A` explícitamente.

**Solución**: Agregar `-id=841A` a todos los comandos después del login:
```bash
# Correcto
login -user=root -pass=123 -id=841A
mkdir -path=/home -id=841A
mkfile -path=/home/test.txt -id=841A
```

### Problema 3: Reportes gráficos no se generan

**Causa**: Graphviz no está instalado o no está en PATH.

**Solución**:
```bash
# Verificar Graphviz
which dot
dot -V

# Si no está instalado
sudo apt-get install graphviz  # Ubuntu/Debian
brew install graphviz          # macOS
```

### Problema 4: "unit=b no soportado"

**Causa**: El sistema solo soporta unidades K (kilobytes) y M (megabytes).

**Solución**: Cambiar `-unit=b` a `-unit=K` o `-unit=M`:
```bash
# Incorrecto
mkdisk -size=1024 -unit=b -path=./Discos/disk.mia

# Correcto
mkdisk -size=1 -unit=K -path=./Discos/disk.mia
```

### Problema 5: Archivo de contenido no encontrado

**Error**: "no se pudo leer el archivo de contenido"

**Solución**: Verificar que el archivo existe y la ruta es correcta:
```bash
# Verificar archivo
ls -l ./CONT/NAME.txt

# Si no existe, crearlo
echo "Contenido" > ./CONT/NAME.txt

# Usar ruta absoluta si es necesario
mkfile -path=/test.txt -id=841A -cont=/home/julian/Documents/.../CONT/NAME.txt
```

### Problema 6: Permisos denegados

**Error**: "permission denied"

**Solución**:
```bash
# Dar permisos al directorio
chmod -R 755 ./Discos ./Reports ./CONT

# Dar permisos al binario
chmod +x ./server

# Si es problema con script helper
chmod +x ./run_test.sh
```

---

## Workflow Completo Recomendado

### Preparación (una sola vez)
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# 1. Crear directorios
mkdir -p Discos Reports CONT

# 2. Crear archivos de contenido
echo "Julian" > CONT/NAME.txt

# 3. Compilar
cd ../..
go build -o server ./cmd/server
```

### Para cada sesión de testing

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend

# 1. Iniciar servidor
./server > server.log 2>&1 &
echo $! > server.pid

# 2. Esperar que inicie
sleep 2

# 3. Verificar que está corriendo
curl http://localhost:8080/health

# 4. Ejecutar tu script
cd cmd/server
jq -n --rawfile script test_basico.smia '{script: $script}' > request.json
curl -X POST http://localhost:8080/api/script \
  -H "Content-Type: application/json" \
  -d @request.json | jq '.'

# 5. Ver resultados
ls -lh Reports/
ls -lh Discos/

# 6. Cuando termines, detener servidor
cd ../..
kill $(cat server.pid)
```

### Limpiar entre pruebas

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Eliminar discos de prueba
rm -rf Discos/*

# Eliminar reportes
rm -rf Reports/*

# Mantener archivos de contenido
# NO eliminar CONT/
```

---

## Tips y Mejores Prácticas

1. **Siempre verifica que el servidor está corriendo** antes de ejecutar scripts
2. **Usa rutas relativas** en los scripts (ej: `./Discos/` en lugar de rutas absolutas)
3. **Comenta tus scripts** para saber qué hace cada sección
4. **Guarda tus scripts de prueba** con nombres descriptivos
5. **Revisa los logs** si algo falla
6. **Limpia entre pruebas** para evitar conflictos de nombres
7. **Usa el helper script** `run_test.sh` para facilitar la ejecución
8. **Ejecuta pruebas pequeñas primero** antes de scripts grandes

---

## Comandos Disponibles

### Gestión de Discos
- `mkdisk` - Crear disco
- `rmdisk` - Eliminar disco
- `fdisk` - Crear/eliminar particiones
- `mount` - Montar partición
- `umount` - Desmontar partición

### Sistema de Archivos
- `mkfs` - Formatear partición
- `login` - Iniciar sesión
- `logout` - Cerrar sesión
- `mkdir` - Crear directorio
- `mkfile` - Crear archivo
- `cat` - Leer archivo
- `remove` - Eliminar archivo/directorio
- `edit` - Editar archivo
- `rename` - Renombrar archivo/directorio
- `copy` - Copiar archivo
- `move` - Mover archivo
- `find` - Buscar archivo
- `chown` - Cambiar propietario
- `chmod` - Cambiar permisos

### Usuarios y Grupos
- `mkgrp` - Crear grupo
- `rmgrp` - Eliminar grupo
- `mkusr` - Crear usuario
- `rmusr` - Eliminar usuario
- `chgrp` - Cambiar grupo de usuario

### Reportes
- `rep -name=mbr` - Reporte MBR
- `rep -name=disk` - Reporte de disco (gráfico)
- `rep -name=inode` - Reporte de inodo
- `rep -name=block` - Reporte de bloque
- `rep -name=bm_inode` - Bitmap de inodos
- `rep -name=bm_block` - Bitmap de bloques
- `rep -name=sb` - SuperBlock
- `rep -name=tree` - Árbol de directorios (gráfico)
- `rep -name=file` - Contenido de archivo
- `rep -name=ls` - Listado de directorio

---

¡Con este manual deberías poder ejecutar todas tus pruebas de manera independiente!
