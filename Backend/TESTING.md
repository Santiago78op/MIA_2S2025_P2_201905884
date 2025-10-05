# Guía de Testing - Backend Proyecto 2

## 🧪 Pruebas Manuales del API

### 1. Iniciar el Servidor

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
./bin/server
```

Deberías ver:
```
🚀 Servidor ExtreamFS iniciado
🌐 API REST: http://localhost:8080/api
...
```

---

## 📋 Tests de Endpoints

### Health Check

```bash
curl http://localhost:8080/healthz
# Esperado: "ok"

curl http://localhost:8080/api/version
# Esperado: JSON con información del servidor
```

### Listar Comandos Soportados

```bash
curl http://localhost:8080/api/commands
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "commands": {
    "disk": ["mkdisk...", "fdisk...", "mount...", "unmount..."],
    "filesystem": ["mkfs..."],
    "files": ["mkdir...", "mkfile...", ...],
    "ext3": ["journaling...", "recovery...", "loss..."]
  }
}
```

---

## 💾 Tests de Comandos de Disco

### 1. Crear Disco

```bash
curl -X POST http://localhost:8080/api/cmd/execute \
  -H "Content-Type: application/json" \
  -d '{
    "line": "mkdisk -path /tmp/test_disk.mia -size 10 -unit m -fit ff"
  }'
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "output": "mkdisk OK path=/tmp/test_disk.mia size=10m fit=ff",
  "input": "mkdisk -path /tmp/test_disk.mia -size 10 -unit m -fit ff"
}
```

**Verificar:**
```bash
ls -lh /tmp/test_disk.mia
# Debería mostrar un archivo de ~10MB
```

### 2. Crear Partición

```bash
curl -X POST http://localhost:8080/api/cmd/execute \
  -H "Content-Type: application/json" \
  -d '{
    "line": "fdisk -path /tmp/test_disk.mia -mode add -name Part1 -size 5 -unit m -type p -fit bf"
  }'
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "output": "fdisk add OK path=/tmp/test_disk.mia name=Part1 size=5m type=p fit=bf",
  "input": "fdisk..."
}
```

### 3. Obtener Info del Disco

```bash
curl "http://localhost:8080/api/disks/info?path=/tmp/test_disk.mia"
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "path": "/tmp/test_disk.mia",
  "size": 10485760,
  "mbr_size": 10485760,
  "fit": "First Fit",
  "partitions": [
    {
      "index": 0,
      "name": "Part1",
      "type": "Primary",
      "fit": "Best Fit",
      "start": 512,
      "size": 5242880
    }
  ]
}
```

### 4. Montar Partición

```bash
curl -X POST http://localhost:8080/api/cmd/execute \
  -H "Content-Type: application/json" \
  -d '{
    "line": "mount -path /tmp/test_disk.mia -name Part1"
  }'
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "output": "mount OK id=vdXXXXXXXX path=/tmp/test_disk.mia name=Part1",
  "input": "mount..."
}
```

**Nota:** Guarda el `id` que aparece en el output (ej: vd12345678)

### 5. Listar Particiones Montadas

```bash
curl http://localhost:8080/api/mounted
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "partitions": [
    {
      "disk_path": "/tmp/test_disk.mia",
      "partition_id": "Part1",
      "mount_id": "vdXXXXXXXX"
    }
  ],
  "count": 1
}
```

---

## 📝 Tests de Validación

### Validar Comando Correcto

```bash
curl -X POST http://localhost:8080/api/cmd/validate \
  -H "Content-Type: application/json" \
  -d '{
    "line": "mkdisk -path /tmp/test.mia -size 10 -unit m"
  }'
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "output": "Command syntax is valid",
  "input": "mkdisk...",
  "command": "mkdisk"
}
```

### Validar Comando Incorrecto

```bash
curl -X POST http://localhost:8080/api/cmd/validate \
  -H "Content-Type: application/json" \
  -d '{
    "line": "mkdisk -path /tmp/test.mia"
  }'
```

**Respuesta esperada:**
```json
{
  "ok": false,
  "error": "mkdisk: 'size' debe ser > 0",
  "usage": "Uso: mkdisk -path <ruta> -size <tamaño> [-unit b|k|m] [-fit bf|ff|wf]",
  "input": "mkdisk -path /tmp/test.mia"
}
```

---

## 📜 Tests de Scripts

### Ejecutar Script Completo

```bash
curl -X POST http://localhost:8080/api/cmd/script \
  -H "Content-Type: application/json" \
  -d '{
    "script": "mkdisk -path /tmp/script_test.mia -size 20 -unit m\nfdisk -path /tmp/script_test.mia -mode add -name P1 -size 10 -unit m -type p\nmount -path /tmp/script_test.mia -name P1"
  }'
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "results": [
    {
      "line": 1,
      "input": "mkdisk...",
      "output": "mkdisk OK...",
      "success": true
    },
    {
      "line": 2,
      "input": "fdisk...",
      "output": "fdisk add OK...",
      "success": true
    },
    {
      "line": 3,
      "input": "mount...",
      "output": "mount OK...",
      "success": true
    }
  ],
  "total_lines": 3,
  "executed": 3,
  "success_count": 3,
  "error_count": 0
}
```

---

## 🔍 Tests de Listado de Discos

```bash
curl "http://localhost:8080/api/disks?path=/tmp"
```

**Respuesta esperada:**
```json
{
  "ok": true,
  "disks": [
    {
      "name": "test_disk.mia",
      "path": "/tmp/test_disk.mia",
      "size": 10485760,
      "modified": "2025-10-05T14:30:00Z"
    },
    {
      "name": "script_test.mia",
      "path": "/tmp/script_test.mia",
      "size": 20971520,
      "modified": "2025-10-05T14:35:00Z"
    }
  ],
  "count": 2,
  "search_path": "/tmp"
}
```

---

## 🧹 Limpieza

```bash
# Eliminar archivos de prueba
rm /tmp/test_disk.mia
rm /tmp/script_test.mia

# Verificar
curl "http://localhost:8080/api/disks?path=/tmp"
# Debería mostrar count: 0 o no incluir los discos eliminados
```

---

## ✅ Checklist de Validación

Marca las pruebas completadas:

- [ ] Health check responde "ok"
- [ ] Version endpoint retorna JSON
- [ ] Commands endpoint lista todos los comandos
- [ ] Crear disco funciona
- [ ] Archivo .mia se crea con tamaño correcto
- [ ] Crear partición funciona
- [ ] Info de disco muestra MBR y particiones
- [ ] Montar partición genera ID
- [ ] Listar montados muestra la partición
- [ ] Validar comando correcto retorna ok
- [ ] Validar comando incorrecto muestra error y uso
- [ ] Ejecutar script procesa múltiples comandos
- [ ] Listar discos encuentra archivos .mia
- [ ] CORS permite peticiones desde frontend

---

## 🐛 Troubleshooting

### El servidor no inicia
```bash
# Verificar que el puerto 8080 esté libre
sudo lsof -i :8080

# Usar puerto diferente
PORT=3000 ./bin/server
```

### Error "disk not found"
```bash
# Verificar que el archivo existe
ls -l /tmp/test_disk.mia

# Usar ruta absoluta completa
curl "http://localhost:8080/api/disks/info?path=/tmp/test_disk.mia"
```

### Error "cannot read MBR"
El disco puede no tener un MBR válido. Créalo primero con `mkdisk`.

### Particiones no aparecen en info
Asegúrate de que `fdisk` se ejecutó exitosamente y que el tipo sea 'p', 'e' o 'l'.

---

## 📊 Tests Automatizados (Opcional)

Crear archivo `test_api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== Test 1: Health Check ==="
curl -s $BASE_URL/healthz
echo -e "\n"

echo "=== Test 2: Commands List ==="
curl -s $BASE_URL/api/commands | jq '.ok'
echo -e "\n"

echo "=== Test 3: Create Disk ==="
curl -s -X POST $BASE_URL/api/cmd/execute \
  -H "Content-Type: application/json" \
  -d '{"line":"mkdisk -path /tmp/auto_test.mia -size 5 -unit m"}' | jq '.ok'
echo -e "\n"

echo "=== Test 4: List Disks ==="
curl -s "$BASE_URL/api/disks?path=/tmp" | jq '.count'
echo -e "\n"

echo "=== Test 5: Cleanup ==="
rm -f /tmp/auto_test.mia
echo "Tests completed!"
```

Ejecutar:
```bash
chmod +x test_api.sh
./test_api.sh
```

---

## 🎯 Resultados Esperados

Todos los tests deben pasar con:
- `"ok": true` en respuestas
- Códigos HTTP 200
- Archivos creados correctamente
- Información de MBR legible
- Comandos ejecutados sin errores

**Si todos los tests pasan**: ✅ El backend está funcionando correctamente!
