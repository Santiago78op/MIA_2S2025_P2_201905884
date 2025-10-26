# 📘 Manual de Despliegue en AWS EC2 (Paso a Paso)

## 🎯 Objetivo
Desplegar tu Backend de Go en AWS EC2 con las carpetas CONT, Discos y Reports funcionando correctamente.

---

## PARTE 1: Compilar el Binario (En tu máquina local)

### Paso 1: Navega al directorio Backend
```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
```

### Paso 2: Compila el binario para Linux
```bash
go build -o server ./cmd/server
```

Esto generará un archivo llamado `server` en el directorio Backend.

**Verifica que se creó:**
```bash
ls -lh server
```

Deberías ver algo como:
```
-rwxr-xr-x 1 julian julian 30M fecha server
```

---

## PARTE 2: Preparar Archivos para Transferir

### Archivos que DEBES copiar a AWS:

1. **El binario compilado:**
   - `Backend/server`

2. **Las carpetas con contenido:**
   - `Backend/Discos/` (con tus archivos .mia dentro)
   - `Backend/Reports/` (puede estar vacía)
   - `Backend/CONT/` (puede estar vacía)

---

## PARTE 3: En tu Instancia AWS EC2

### Paso 1: Conéctate a tu EC2
```bash
ssh -i TU_ARCHIVO.pem ubuntu@TU_IP_PUBLICA
```

### Paso 2: Ir a la carpeta que creaste
```bash
cd Backen
```

### Paso 3: Crear la estructura de carpetas necesarias
```bash
mkdir -p Discos
mkdir -p Reports
mkdir -p CONT
mkdir -p logs
```

**Verificar que se crearon:**
```bash
ls -la
```

Deberías ver:
```
drwxr-xr-x Discos/
drwxr-xr-x Reports/
drwxr-xr-x CONT/
drwxr-xr-x logs/
```

---

## PARTE 4: Transferir Archivos desde tu Máquina Local a AWS

### Desde tu máquina local (en otra terminal), ejecuta:

#### 1. Copiar el binario
```bash
scp -i TU_ARCHIVO.pem /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/server ubuntu@TU_IP_PUBLICA:~/Backen/
```

#### 2. Copiar la carpeta Discos (con tus archivos .mia)
```bash
scp -i TU_ARCHIVO.pem -r /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/Discos/* ubuntu@TU_IP_PUBLICA:~/Backen/Discos/
```

#### 3. Copiar Reports (si tienes reportes ya generados)
```bash
scp -i TU_ARCHIVO.pem -r /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/Reports/* ubuntu@TU_IP_PUBLICA:~/Backen/Reports/
```

---

## PARTE 5: Configurar Variables de Entorno en AWS

### Paso 1: En tu EC2, crea el archivo .env
```bash
cd ~/Backen
nano .env
```

### Paso 2: Escribe esta configuración (ajusta las rutas):
```
PORT=8080
DISKS_PATH=/home/ubuntu/Backen/Discos
REPORTS_PATH=/home/ubuntu/Backen/Reports
CARNET_LAST_TWO=84
DEBUG=false
BASE_DIR=/home/ubuntu/Backen
```

### Paso 3: Guarda el archivo
- Presiona `Ctrl + X`
- Presiona `Y`
- Presiona `Enter`

**Verificar que se creó:**
```bash
cat .env
```

---

## PARTE 6: Dar Permisos de Ejecución al Binario

```bash
cd ~/Backen
chmod +x server
```

**Verificar permisos:**
```bash
ls -la server
```

Debe mostrar `-rwxr-xr-x` (las x indican que es ejecutable)

---

## PARTE 7: Verificar Estructura Final

Ejecuta:
```bash
cd ~/Backen
tree -L 2
```

O si no tienes tree:
```bash
ls -R
```

Deberías ver algo así:
```
/home/ubuntu/Backen/
├── server              (binario ejecutable)
├── .env                (configuración)
├── Discos/
│   ├── DiscoA.mia
│   ├── DiscoB.mia
│   └── DiscoC.mia
├── Reports/            (vacía o con reportes)
├── CONT/               (vacía)
└── logs/               (vacía, se llenará con logs)
```

---

## PARTE 8: Probar el Servidor Manualmente

### Paso 1: Ejecuta el servidor
```bash
cd ~/Backen
./server
```

Deberías ver mensajes como:
```
2025-10-26 10:30:00 INFO [main.go:45] Configuración cargada
2025-10-26 10:30:00 INFO [main.go:68] Iniciando servidor MIA en puerto 8080
2025-10-26 10:30:00 INFO [main.go:70] Servidor listo. Escuchando en http://localhost:8080
```

### Paso 2: Probar desde otra terminal (nueva conexión SSH)
```bash
curl http://localhost:8080/health
```

Deberías recibir:
```json
{
  "status": "ok",
  "message": "Servidor MIA funcionando correctamente",
  "disks_path": "/home/ubuntu/Backen/Discos",
  "reports_path": "/home/ubuntu/Backen/Reports",
  "debug_mode": false
}
```

### Paso 3: Detener el servidor
Presiona `Ctrl + C` en la terminal donde está corriendo

---

## PARTE 9: Configurar Security Groups en AWS

### Para que el servidor sea accesible desde internet:

1. Ve a la **Consola de AWS** → **EC2** → **Instancias**
2. Selecciona tu instancia
3. Ve a la pestaña **Security**
4. Haz clic en tu **Security Group**
5. Edita **Inbound Rules**
6. Agrega esta regla:

```
Type: Custom TCP
Port range: 8080
Source: 0.0.0.0/0
Description: MIA Backend API
```

7. Guarda los cambios

---

## PARTE 10: Ejecutar el Servidor de Forma Permanente

### Opción A: Usando nohup (Simple)

```bash
cd ~/Backen
nohup ./server > logs/server.log 2>&1 &
```

**Verificar que está corriendo:**
```bash
ps aux | grep server
```

**Ver logs:**
```bash
tail -f ~/Backen/logs/server.log
```

**Detener el servidor:**
```bash
pkill -f server
```

### Opción B: Usando systemd (Recomendado para producción)

#### 1. Crear el archivo de servicio
```bash
sudo nano /etc/systemd/system/mia-backend.service
```

#### 2. Pegar esta configuración:
```ini
[Unit]
Description=MIA Backend Server
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/Backen
EnvironmentFile=/home/ubuntu/Backen/.env
ExecStart=/home/ubuntu/Backen/server
Restart=on-failure
RestartSec=5s
StandardOutput=append:/home/ubuntu/Backen/logs/server.log
StandardError=append:/home/ubuntu/Backen/logs/error.log

[Install]
WantedBy=multi-user.target
```

#### 3. Guardar (Ctrl+X, Y, Enter)

#### 4. Habilitar e iniciar el servicio
```bash
sudo systemctl daemon-reload
sudo systemctl enable mia-backend
sudo systemctl start mia-backend
```

#### 5. Verificar estado
```bash
sudo systemctl status mia-backend
```

Deberías ver:
```
● mia-backend.service - MIA Backend Server
   Active: active (running) since ...
```

#### Comandos útiles:
```bash
# Ver logs en tiempo real
sudo journalctl -u mia-backend -f

# Reiniciar servicio
sudo systemctl restart mia-backend

# Detener servicio
sudo systemctl stop mia-backend

# Ver estado
sudo systemctl status mia-backend
```

---

## PARTE 11: Verificar desde Internet

Desde tu navegador o terminal en tu máquina local:

```bash
curl http://TU_IP_PUBLICA:8080/health
```

O abre en tu navegador:
```
http://TU_IP_PUBLICA:8080/health
```

---

## 🔍 Verificación de Archivos Importantes

### Verificar que los discos están accesibles:
```bash
curl http://TU_IP_PUBLICA:8080/api/disks
```

### Verificar que las particiones se leen:
```bash
curl "http://TU_IP_PUBLICA:8080/api/disks/partitions?path=/home/ubuntu/Backen/Discos/DiscoA.mia"
```

---

## ⚠️ Solución de Problemas Comunes

### Error: "No such file or directory" al ejecutar ./server
**Solución:**
```bash
chmod +x ~/Backen/server
```

### Error: "Permission denied" en carpetas
**Solución:**
```bash
sudo chown -R ubuntu:ubuntu ~/Backen
chmod -R 755 ~/Backen
```

### No puedo acceder desde internet
**Solución:**
1. Verifica Security Groups (Parte 9)
2. Verifica que el servidor está corriendo:
   ```bash
   sudo netstat -tuln | grep 8080
   ```

### El servidor no encuentra los discos .mia
**Solución:**
1. Verifica que las rutas en `.env` son absolutas:
   ```bash
   cat ~/Backen/.env
   ```
2. Verifica que los archivos existen:
   ```bash
   ls -la ~/Backen/Discos/
   ```

---

## 📊 Endpoints de la API Disponibles

Una vez todo esté funcionando:

- **Health:** `GET http://TU_IP:8080/health`
- **Listar discos:** `GET http://TU_IP:8080/api/disks`
- **Particiones:** `GET http://TU_IP:8080/api/disks/partitions?path=/ruta/disco.mia`
- **Login:** `POST http://TU_IP:8080/api/auth/login`
- **Comando:** `POST http://TU_IP:8080/api/commands`
- **Script:** `POST http://TU_IP:8080/api/script`
- **Árbol archivos:** `GET http://TU_IP:8080/api/fs/:id/tree?path=/`
- **Ver archivo:** `GET http://TU_IP:8080/api/fs/:id/file?path=/archivo.txt`
- **Reportes estáticos:** `GET http://TU_IP:8080/reports/static/nombre.jpg`

---

## ✅ Checklist Final

- [ ] Go instalado en AWS
- [ ] Carpeta `Backen` creada
- [ ] Subcarpetas creadas: Discos, Reports, CONT, logs
- [ ] Binario `server` compilado y copiado
- [ ] Archivos .mia copiados a Discos/
- [ ] Archivo `.env` creado con rutas correctas
- [ ] Permisos de ejecución dados al binario
- [ ] Security Group con puerto 8080 abierto
- [ ] Servidor ejecutándose (nohup o systemd)
- [ ] Health check respondiendo desde internet

---

**¡Listo! Tu backend debería estar funcionando en AWS. 🚀**

Si tienes algún error, revisa la sección de "Solución de Problemas" o verifica los logs.
