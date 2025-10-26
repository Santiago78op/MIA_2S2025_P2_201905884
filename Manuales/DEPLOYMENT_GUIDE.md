# 🚀 Guía de Despliegue en AWS EC2

## 📋 Requisitos Previos

### En tu máquina local:
- Go 1.24.5 instalado
- Acceso SSH a tu instancia EC2
- Git (para clonar el proyecto)

### En AWS EC2:
- Instancia Ubuntu (20.04 LTS o superior)
- Mínimo 2 GB RAM
- 10 GB de espacio en disco
- Security Group configurado (ver paso 1)

---

## 🔧 Paso 1: Configurar Security Groups en AWS

1. Ve a tu instancia EC2 en la consola de AWS
2. Haz clic en **Security** → **Security Groups**
3. Edita las reglas de **Inbound** y agrega:

| Tipo | Protocolo | Puerto | Origen | Descripción |
|------|-----------|--------|--------|-------------|
| Custom TCP | TCP | 8080 | 0.0.0.0/0 | MIA Backend API |
| SSH | TCP | 22 | Tu IP | SSH Access |

> ⚠️ **Importante:** Para producción, restringe el origen 0.0.0.0/0 a las IPs de tu frontend

---

## 🏗️ Paso 2: Compilar el Proyecto (En tu máquina local)

```bash
# Navega al directorio del proyecto
cd /home/julian/Documents/MIA_2S2025_P2_201905884

# Compila el binario para Linux
chmod +x build.sh
./build.sh
```

Esto generará el binario en `Backend/bin/server` optimizado para AWS.

---

## 📦 Paso 3: Transferir Archivos a EC2

### Opción A: Usando SCP (Recomendado)

```bash
# Reemplaza <TU_KEY.pem> y <TU_IP_EC2> con tus datos

# 1. Copiar todo el proyecto
scp -i <TU_KEY.pem> -r Backend ubuntu@<TU_IP_EC2>:/tmp/

# 2. Copiar script de despliegue
scp -i <TU_KEY.pem> deploy-aws.sh ubuntu@<TU_IP_EC2>:/tmp/
```

### Opción B: Usando Git

```bash
# En tu EC2, clona el repositorio
ssh -i <TU_KEY.pem> ubuntu@<TU_IP_EC2>

# Una vez dentro
git clone <TU_REPOSITORIO_GIT>
cd MIA_2S2025_P2_201905884
```

---

## 🛠️ Paso 4: Instalación en EC2

```bash
# Conéctate a tu instancia EC2
ssh -i <TU_KEY.pem> ubuntu@<TU_IP_EC2>

# Instala Go (si no está instalado)
sudo apt update
sudo apt install -y golang-go

# Verifica la versión
go version

# Si necesitas Go 1.24.5 específico:
wget https://go.dev/dl/go1.24.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Ejecutar el script de despliegue:

```bash
# Navega al directorio donde copiaste los archivos
cd /tmp

# Si usaste git clone, navega al proyecto
# cd MIA_2S2025_P2_201905884

# Da permisos de ejecución
chmod +x deploy-aws.sh

# Ejecuta el script de instalación
sudo ./deploy-aws.sh
```

El script hará automáticamente:
- ✅ Crear directorios (Discos, Reports, CONT, logs)
- ✅ Copiar el binario a `/home/ubuntu/mia-backend`
- ✅ Copiar discos .mia existentes
- ✅ Configurar variables de entorno
- ✅ Crear servicio systemd
- ✅ Abrir puerto 8080 en firewall
- ✅ Iniciar el servicio

---

## ✅ Paso 5: Verificar Instalación

### Verificar que el servicio está corriendo:

```bash
sudo systemctl status mia-backend
```

Deberías ver:
```
● mia-backend.service - MIA Backend Server
   Active: active (running) since ...
```

### Probar la API:

```bash
# Desde la EC2
curl http://localhost:8080/health

# Desde tu máquina local (reemplaza <TU_IP_EC2>)
curl http://<TU_IP_EC2>:8080/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "message": "Servidor MIA funcionando correctamente",
  "disks_path": "/home/ubuntu/mia-backend/Discos",
  "reports_path": "/home/ubuntu/mia-backend/Reports",
  "debug_mode": false
}
```

---

## 🔍 Paso 6: Monitoreo y Logs

### Ver logs en tiempo real:
```bash
sudo journalctl -u mia-backend -f
```

### Ver últimos 100 logs:
```bash
sudo journalctl -u mia-backend -n 100
```

### Ver logs de errores:
```bash
tail -f /home/ubuntu/mia-backend/logs/error.log
```

### Ver logs del servidor:
```bash
tail -f /home/ubuntu/mia-backend/logs/server.log
```

---

## 🔄 Comandos Útiles

### Reiniciar el servicio:
```bash
sudo systemctl restart mia-backend
```

### Detener el servicio:
```bash
sudo systemctl stop mia-backend
```

### Iniciar el servicio:
```bash
sudo systemctl start mia-backend
```

### Deshabilitar inicio automático:
```bash
sudo systemctl disable mia-backend
```

### Ver estado completo:
```bash
sudo systemctl status mia-backend -l --no-pager
```

---

## 🔧 Actualizar el Backend

Si necesitas actualizar el código:

```bash
# En tu máquina local
cd /home/julian/Documents/MIA_2S2025_P2_201905884
./build.sh

# Copiar nuevo binario
scp -i <TU_KEY.pem> Backend/bin/server ubuntu@<TU_IP_EC2>:/home/ubuntu/mia-backend/bin/

# En EC2, reiniciar servicio
ssh -i <TU_KEY.pem> ubuntu@<TU_IP_EC2>
sudo systemctl restart mia-backend
```

---

## 🌐 Configurar con Dominio (Opcional)

Si tienes un dominio, puedes usar Nginx como reverse proxy:

```bash
# Instalar Nginx
sudo apt install -y nginx

# Crear configuración
sudo tee /etc/nginx/sites-available/mia-backend << 'EOF'
server {
    listen 80;
    server_name tu-dominio.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /reports/static/ {
        proxy_pass http://localhost:8080/reports/static/;
    }
}
EOF

# Habilitar sitio
sudo ln -s /etc/nginx/sites-available/mia-backend /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 🐛 Solución de Problemas

### El servicio no inicia:
```bash
# Ver errores específicos
sudo journalctl -u mia-backend -n 50 --no-pager

# Verificar permisos
ls -la /home/ubuntu/mia-backend/bin/server

# Ejecutar manualmente para ver errores
cd /home/ubuntu/mia-backend
./bin/server
```

### Puerto 8080 no responde:
```bash
# Verificar que el servicio escucha en 8080
sudo netstat -tuln | grep 8080

# Verificar firewall local
sudo ufw status

# Verificar Security Group en AWS Console
```

### Discos no se encuentran:
```bash
# Verificar que los archivos .mia existen
ls -lh /home/ubuntu/mia-backend/Discos/

# Verificar permisos
sudo chown -R ubuntu:ubuntu /home/ubuntu/mia-backend

# Verificar variable de entorno
cat /home/ubuntu/mia-backend/.env | grep DISKS_PATH
```

---

## 📊 Endpoints de la API

Una vez desplegado, tu API estará disponible en:

- **Health Check:** `http://<TU_IP_EC2>:8080/health`
- **Ejecutar Comando:** `POST http://<TU_IP_EC2>:8080/api/commands`
- **Ejecutar Script:** `POST http://<TU_IP_EC2>:8080/api/script`
- **Listar Discos:** `GET http://<TU_IP_EC2>:8080/api/disks`
- **Listar Particiones:** `GET http://<TU_IP_EC2>:8080/api/disks/partitions?path=/ruta/disco.mia`
- **Login:** `POST http://<TU_IP_EC2>:8080/api/auth/login`
- **Árbol de Archivos:** `GET http://<TU_IP_EC2>:8080/api/fs/:id/tree?path=/`
- **Ver Archivo:** `GET http://<TU_IP_EC2>:8080/api/fs/:id/file?path=/archivo.txt`
- **Reportes:** `GET http://<TU_IP_EC2>:8080/reports/static/<nombre_reporte>.jpg`

---

## 📝 Notas Finales

- El servicio se inicia automáticamente al reiniciar la EC2
- Los logs se rotan automáticamente
- Los reportes se generan en `/home/ubuntu/mia-backend/Reports`
- Los discos .mia se almacenan en `/home/ubuntu/mia-backend/Discos`
- El servicio se reinicia automáticamente si falla

---

**¡Despliegue completado! 🎉**

Para cualquier problema, revisa los logs con:
```bash
sudo journalctl -u mia-backend -f
```
