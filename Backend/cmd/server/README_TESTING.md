# Guía Rápida de Testing

## Inicio Rápido (5 minutos)

### 1. Preparación inicial (solo la primera vez)
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend

# Compilar el servidor
go build -o server ./cmd/server

# Crear directorios necesarios
cd cmd/server
mkdir -p Discos Reports CONT

# Crear archivo de contenido de prueba
echo "Julian" > CONT/NAME.txt
```

### 2. Iniciar el servidor
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend

# Iniciar en segundo plano
./server > server.log 2>&1 &
echo $! > server.pid

# Verificar que está corriendo
curl http://localhost:8080/health
```

### 3. Ejecutar un script de prueba
```bash
cd cmd/server

# Usar el script auxiliar
./run_test.sh test_ejemplo_simple.smia

# O manualmente con curl
jq -n --rawfile script test_ejemplo_simple.smia '{script: $script}' > request.json
curl -X POST http://localhost:8080/api/script \
  -H "Content-Type: application/json" \
  -d @request.json | jq '.'
rm request.json
```

### 4. Ver resultados
```bash
# Ver discos creados
ls -lh Discos/

# Ver reportes generados
ls -lh Reports/

# Ver un reporte de texto
cat Reports/ejemplo_mbr.txt

# Ver logs del servidor
tail -f ../../server.log
```

### 5. Detener el servidor
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
kill $(cat server.pid)
```

---

## Archivos Importantes

- **MANUAL_TESTING.md** - Manual completo con todos los detalles
- **run_test.sh** - Script auxiliar para ejecutar tests fácilmente
- **test_ejemplo_simple.smia** - Script de ejemplo básico
- **test_script.smia** - Script de prueba completo (99 líneas)

---

## Comandos Más Usados

### Crear disco y partición
```
mkdisk -size=10 -unit=M -path=./Discos/MiDisco.mia
fdisk -size=5 -unit=M -path=./Discos/MiDisco.mia -type=P -fit=BF -name=Part1
mount -path=./Discos/MiDisco.mia -name=Part1
mkfs -type=full -id=841A
```

### Operaciones de archivos
```
login -user=root -pass=123 -id=841A
mkdir -path=/home -id=841A
mkfile -path=/home/test.txt -id=841A -size=100
cat -file=/home/test.txt -id=841A
logout
```

### Generar reportes
```
rep -name=disk -path=./Discos/MiDisco.mia -id=841A -ruta=./Reports/disk.jpg
rep -name=tree -path=./Discos/MiDisco.mia -id=841A -ruta=./Reports/tree.png
rep -name=mbr -path=./Discos/MiDisco.mia -id=841A -ruta=./Reports/mbr.txt
```

---

## Solución de Problemas Comunes

### El servidor no inicia
```bash
# Verificar si ya está corriendo
lsof -i :8080

# Matar proceso anterior
kill $(lsof -t -i:8080)

# Reintentar
./server &
```

### Script falla con "jq: command not found"
```bash
# Instalar jq
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS
```

### Reportes gráficos no se generan
```bash
# Instalar Graphviz
sudo apt-get install graphviz
```

### "Falta id" en comandos
Todos los comandos después de login necesitan `-id=841A`:
```
# Incorrecto
mkdir -path=/home

# Correcto
mkdir -path=/home -id=841A
```

---

## Limpiar entre pruebas

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/cmd/server

# Limpiar discos
rm -rf Discos/*

# Limpiar reportes
rm -rf Reports/*

# NO eliminar CONT/ (contiene archivos de contenido)
```

---

## Estructura de Directorios

```
Backend/cmd/server/
├── main.go                    # Código del servidor
├── server                     # Binario compilado (en Backend/)
├── server.pid                 # PID del servidor
├── server.log                 # Logs del servidor
├── run_test.sh               # Script auxiliar para tests
├── test_ejemplo_simple.smia  # Script de ejemplo
├── test_script.smia          # Script de prueba completo
├── Discos/                   # Discos .mia generados
├── Reports/                  # Reportes generados
└── CONT/                     # Archivos de contenido
    └── NAME.txt
```

---

## Ejemplo de Workflow Completo

```bash
# 1. Compilar (solo primera vez)
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
go build -o server ./cmd/server

# 2. Iniciar servidor
./server > server.log 2>&1 &
echo $! > server.pid

# 3. Ejecutar test
cd cmd/server
./run_test.sh test_ejemplo_simple.smia

# 4. Ver resultados
ls -lh Reports/
cat Reports/ejemplo_mbr.txt

# 5. Limpiar
rm -rf Discos/* Reports/*

# 6. Detener servidor
cd ../..
kill $(cat server.pid)
```

---

Para más detalles, consulta **MANUAL_TESTING.md**
