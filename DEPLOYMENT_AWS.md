# Guía de Despliegue en AWS - Proyecto 2 MIA

## Arquitectura de Despliegue

- **Backend**: EC2 (Amazon Linux 2023 o Ubuntu)
- **Frontend**: S3 + CloudFront (opcional)
- **CORS**: Configurado entre S3 y EC2

---

## Parte 1: Despliegue del Backend en EC2

### 1.1 Crear Instancia EC2

1. Acceder a AWS Console → EC2
2. Launch Instance:
   - **Name**: `mia-p2-backend-201905884`
   - **AMI**: Ubuntu Server 22.04 LTS (free tier eligible)
   - **Instance type**: t2.micro (1 vCPU, 1 GB RAM)
   - **Key pair**: Crear o seleccionar una existente
   - **Network settings**:
     - Allow SSH (22) from your IP
     - Allow HTTP (80) from anywhere
     - Allow Custom TCP (8080) from anywhere
   - **Storage**: 8-20 GB gp3

3. Launch y esperar a que esté `running`

### 1.2 Conectar a la Instancia

```bash
# Desde tu terminal local
ssh -i /path/to/key.pem ubuntu@<EC2-PUBLIC-IP>
```

### 1.3 Instalar Dependencias en EC2

```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar Go 1.21+ (verificar última versión en https://go.dev/dl/)
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verificar instalación
go version
```

### 1.4 Subir y Compilar el Backend

**Opción A: Compilar localmente (recomendado)**

```bash
# En tu máquina local (Linux)
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend
GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go

# Copiar a EC2
scp -i /path/to/key.pem bin/server-linux ubuntu@<EC2-PUBLIC-IP>:/home/ubuntu/server
scp -r -i /path/to/key.pem Discos ubuntu@<EC2-PUBLIC-IP>:/home/ubuntu/
scp -r -i /path/to/key.pem Reports ubuntu@<EC2-PUBLIC-IP>:/home/ubuntu/
```

**Opción B: Compilar en EC2**

```bash
# En EC2
mkdir -p ~/mia-backend
cd ~/mia-backend

# Transferir código (usar scp o git clone)
# Ejemplo con scp desde local:
# scp -r -i key.pem Backend ubuntu@<IP>:/home/ubuntu/mia-backend/

cd ~/mia-backend/Backend
go build -o bin/server cmd/server/main.go
```

### 1.5 Configurar Servicio Systemd

```bash
# En EC2
sudo nano /etc/systemd/system/mia-backend.service
```

Contenido del archivo:

```ini
[Unit]
Description=MIA Proyecto 2 Backend
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/mia-backend/Backend
ExecStart=/home/ubuntu/mia-backend/Backend/bin/server
Restart=on-failure
RestartSec=10
StandardOutput=append:/home/ubuntu/backend.log
StandardError=append:/home/ubuntu/backend-error.log

# Variables de entorno
Environment="PORT=8080"
Environment="GIN_MODE=release"

[Install]
WantedBy=multi-user.target
```

### 1.6 Iniciar el Servicio

```bash
# Recargar configuración
sudo systemctl daemon-reload

# Iniciar servicio
sudo systemctl start mia-backend

# Habilitar inicio automático
sudo systemctl enable mia-backend

# Verificar estado
sudo systemctl status mia-backend

# Ver logs en tiempo real
tail -f ~/backend.log
```

### 1.7 Verificar que el Backend Funciona

```bash
# Desde EC2
curl http://localhost:8080/health

# Desde local (reemplazar <EC2-PUBLIC-IP>)
curl http://<EC2-PUBLIC-IP>:8080/health
```

---

## Parte 2: Despliegue del Frontend en S3

### 2.1 Crear Bucket S3

1. AWS Console → S3 → Create bucket
   - **Bucket name**: `mia-p2-frontend-201905884` (debe ser único globalmente)
   - **Region**: us-east-1 (o la más cercana)
   - **Desmarcar**: "Block all public access"
   - **Confirmar**: que el bucket será público
   - Create bucket

### 2.2 Habilitar Static Website Hosting

1. Seleccionar el bucket → Properties → Static website hosting
2. Enable
   - **Index document**: `index.html`
   - **Error document**: `index.html` (para React Router)
3. Save changes

### 2.3 Configurar Política de Bucket (Público)

1. Bucket → Permissions → Bucket policy
2. Pegar esta política (reemplazar `YOUR-BUCKET-NAME`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::YOUR-BUCKET-NAME/*"
    }
  ]
}
```

### 2.4 Configurar CORS en el Bucket

1. Bucket → Permissions → CORS configuration
2. Pegar:

```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["GET", "HEAD"],
    "AllowedOrigins": ["*"],
    "ExposeHeaders": []
  }
]
```

### 2.5 Actualizar API URL en el Frontend

**Antes de compilar**, editar `Frontend/src/lib/api.js`:

```javascript
// Cambiar
const BASE = '' // Vite proxy → :8080

// Por
const BASE = 'http://<EC2-PUBLIC-IP>:8080'
```

O mejor aún, crear un archivo `.env.production`:

```bash
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Frontend
echo "VITE_API_URL=http://<EC2-PUBLIC-IP>:8080" > .env.production
```

Y modificar `api.js`:

```javascript
const BASE = import.meta.env.VITE_API_URL || ''
```

### 2.6 Compilar y Subir Frontend

```bash
# En tu máquina local
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Frontend

# Compilar para producción
npm run build

# Subir a S3 (requiere AWS CLI configurado)
aws s3 sync dist/ s3://mia-p2-frontend-201905884/ --delete

# O usar la consola web: Upload → arrastrar contenido de dist/
```

### 2.7 Obtener URL del Sitio

1. S3 → Bucket → Properties → Static website hosting
2. **Endpoint URL**: `http://mia-p2-frontend-201905884.s3-website-us-east-1.amazonaws.com`

---

## Parte 3: Configurar CORS en el Backend

Editar `Backend/router/router.go` para permitir requests desde S3:

```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{
        "http://localhost:5173",
        "http://mia-p2-frontend-201905884.s3-website-us-east-1.amazonaws.com",
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

Recompilar y reiniciar el backend:

```bash
# Local
GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go
scp -i key.pem bin/server-linux ubuntu@<EC2-IP>:/home/ubuntu/mia-backend/Backend/bin/server

# En EC2
sudo systemctl restart mia-backend
```

---

## Parte 4: Verificación Final

### 4.1 Endpoints a Probar

```bash
# Health check
curl http://<EC2-IP>:8080/health

# Listar discos
curl http://<EC2-IP>:8080/api/disks

# Journal (después de crear partición y montarla)
curl "http://<EC2-IP>:8080/api/journal/841A/table"
```

### 4.2 Desde el Frontend

1. Abrir `http://<S3-ENDPOINT-URL>`
2. Ir a Terminal y ejecutar:
   ```
   mkdisk -size=10 -unit=M -path=Discos/test.mia
   fdisk -size=5 -unit=M -path=Discos/test.mia -type=P -name=part1
   mount -path=Discos/test.mia -name=part1
   mkfs -id=<ID> -fs=3fs -type=full
   ```
3. Ir a Visualizer y verificar que aparece el disco

---

## Parte 5: (Opcional) CloudFront para HTTPS

### 5.1 Crear Distribución CloudFront

1. CloudFront → Create distribution
   - **Origin domain**: Seleccionar el bucket S3
   - **Origin access**: Origin access control (recommended)
   - **Viewer protocol policy**: Redirect HTTP to HTTPS
   - **Default root object**: `index.html`

2. Create distribution y esperar ~10-15 min

3. **Distribution domain name**: `d1234abcd.cloudfront.net`

### 5.2 Configurar Error Pages

1. Distribución → Error pages → Create custom error response
   - **HTTP error code**: 403 Forbidden
   - **Customize error response**: Yes
   - **Response page path**: `/index.html`
   - **HTTP Response code**: 200 OK

2. Repetir para 404 Not Found

### 5.3 Actualizar CORS en Backend

Agregar el dominio de CloudFront a `AllowOrigins`:

```go
AllowOrigins: []string{
    "http://localhost:5173",
    "http://mia-p2-frontend-201905884.s3-website-us-east-1.amazonaws.com",
    "https://d1234abcd.cloudfront.net",
},
```

---

## Parte 6: Troubleshooting

### Backend no responde

```bash
# Verificar que el servicio esté corriendo
sudo systemctl status mia-backend

# Ver logs
sudo journalctl -u mia-backend -f

# Verificar puerto
sudo netstat -tlnp | grep 8080

# Verificar Security Group en EC2 (puerto 8080 abierto)
```

### CORS Errors en Frontend

1. Verificar que `AllowOrigins` incluye la URL del frontend
2. Verificar que el backend esté reiniciado después de cambios
3. Inspeccionar Network tab en DevTools del navegador

### Frontend no carga en S3

1. Verificar que `index.html` está en la raíz del bucket
2. Verificar Bucket Policy (debe ser público)
3. Verificar que Static Website Hosting está habilitado

---

## URLs Finales para Documentación

Incluir en el informe:

- **Backend API**: `http://<EC2-PUBLIC-IP>:8080`
- **Frontend**: `http://<S3-ENDPOINT-URL>` o `https://<CLOUDFRONT-DOMAIN>`
- **Health Check**: `http://<EC2-PUBLIC-IP>:8080/health`

---

## Comandos Útiles

```bash
# Logs del backend en EC2
tail -f ~/backend.log

# Reiniciar backend
sudo systemctl restart mia-backend

# Actualizar frontend en S3
aws s3 sync dist/ s3://mia-p2-frontend-201905884/ --delete

# Conectar a EC2
ssh -i key.pem ubuntu@<EC2-IP>
```

---

## Estimación de Costos (Free Tier)

- **EC2 t2.micro**: 750 horas/mes gratis (primer año)
- **S3**: 5 GB storage + 20,000 GET requests gratis
- **CloudFront**: 1 TB de transferencia gratis (primer año)

**Total**: $0 USD si se mantiene dentro del free tier.

---

✅ **Checklist Final de Despliegue**

- [ ] Backend compilado y corriendo en EC2
- [ ] Health endpoint responde (200 OK)
- [ ] Frontend compilado con URL de API correcta
- [ ] Frontend subido a S3
- [ ] Static website hosting habilitado
- [ ] CORS configurado en backend
- [ ] Endpoints viewer funcionando (/api/disks, /api/fs/:id/tree, etc.)
- [ ] Journal endpoint funcionando (/api/journal/:id/table)
- [ ] URLs documentadas en informe

**¡Despliegue completo!** 🚀
