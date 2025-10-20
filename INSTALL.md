# Guía de Instalación - MIA Proyecto 2

Sistema de Archivos EXT2/EXT3 con Visualizador Web

---

## 📋 Requisitos Previos

- **Sistema Operativo:** Ubuntu 20.04+ o Debian 11+
- **Privilegios:** Acceso sudo para instalar paquetes del sistema
- **Conexión a Internet:** Necesaria para descargar dependencias

---

## 🚀 Instalación Rápida

### 1. Clonar o ubicar el proyecto

```bash
cd /ruta/al/proyecto/MIA_2S2025_P2_201905884
```

### 2. Ejecutar el script de instalación

```bash
./install.sh
```

El script instalará automáticamente:
- ✅ Go v1.23.3
- ✅ Node.js v20.x
- ✅ npm (incluido con Node.js)
- ✅ Dependencias del Backend (Go modules)
- ✅ Dependencias del Frontend (npm packages)
- ✅ Compilación del Backend
- ✅ Directorios necesarios

### 3. Cerrar y reabrir la terminal

Después de la instalación, cierra y vuelve a abrir la terminal para que los cambios de PATH tomen efecto.

---

## 🔧 Instalación Manual

Si prefieres instalar manualmente cada componente:

### Backend (Go)

1. **Instalar Go:**

```bash
# Descargar Go 1.23.3
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz

# Instalar
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz

# Configurar PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verificar
go version
```

2. **Instalar dependencias del Backend:**

```bash
cd Backend
go mod download
go mod tidy
```

3. **Compilar el Backend:**

```bash
mkdir -p bin
go build -o bin/server cmd/server/main.go
```

### Frontend (Node.js)

1. **Instalar Node.js v20:**

```bash
# Agregar repositorio NodeSource
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -

# Instalar Node.js
sudo apt install -y nodejs

# Verificar
node -v
npm -v
```

2. **Instalar dependencias del Frontend:**

```bash
cd Frontend
npm install
```

---

## ▶️ Ejecutar el Proyecto

### Opción 1: Ejecución Normal

**Terminal 1 - Backend:**
```bash
cd Backend
./bin/server
```

**Terminal 2 - Frontend:**
```bash
cd Frontend
npm run dev
```

### Opción 2: Ejecución en Background

**Backend:**
```bash
cd Backend
nohup ./bin/server > logs/server.log 2>&1 &
echo $! > server.pid
```

**Frontend:**
```bash
cd Frontend
npm run dev
```

### Detener el Backend en Background

```bash
cd Backend
kill $(cat server.pid)
rm server.pid
```

---

## 🌐 Acceder a la Aplicación

Una vez iniciados ambos servicios:

- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:8080
- **Health Check:** http://localhost:8080/health

---

## 📂 Estructura del Proyecto

```
MIA_2S2025_P2_201905884/
├── Backend/
│   ├── bin/
│   │   └── server              # Ejecutable compilado
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Punto de entrada
│   ├── controllers/            # Controladores HTTP
│   ├── core/                   # Lógica de negocio
│   ├── router/                 # Rutas de la API
│   ├── Discos/                 # Archivos .mia
│   ├── Reportes/               # Reportes generados
│   └── go.mod                  # Dependencias Go
│
├── Frontend/
│   ├── src/
│   │   ├── components/         # Componentes React
│   │   ├── pages/              # Páginas
│   │   ├── lib/                # API client
│   │   └── main.jsx            # Punto de entrada
│   ├── package.json            # Dependencias npm
│   └── vite.config.js          # Configuración Vite
│
└── install.sh                  # Script de instalación
```

---

## 🔍 Verificar Instalación

Después de ejecutar el script, verifica que todo esté instalado correctamente:

```bash
# Verificar Go
go version
# Debe mostrar: go version go1.23.3 linux/amd64

# Verificar Node.js
node -v
# Debe mostrar: v20.x.x

# Verificar npm
npm -v
# Debe mostrar: 10.x.x

# Verificar Backend compilado
ls -lh Backend/bin/server
# Debe mostrar el ejecutable

# Verificar dependencias Frontend
ls Frontend/node_modules
# Debe mostrar múltiples directorios
```

---

## ❌ Solución de Problemas

### Error: "go: command not found"

**Solución:**
```bash
# Verificar que Go esté instalado
ls /usr/local/go/bin/go

# Si existe, agregar al PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Error: "node: command not found"

**Solución:**
```bash
# Reinstalar Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
```

### Error: Puerto 8080 en uso

**Solución:**
```bash
# Ver qué proceso está usando el puerto
lsof -ti:8080

# Matar el proceso
kill -9 $(lsof -ti:8080)
```

### Error: Puerto 5173 en uso

**Solución:**
```bash
# Ver qué proceso está usando el puerto
lsof -ti:5173

# Matar el proceso
kill -9 $(lsof -ti:5173)
```

### Error de permisos al compilar

**Solución:**
```bash
cd Backend
sudo chown -R $USER:$USER .
go build -o bin/server cmd/server/main.go
```

---

## 🔄 Actualizar Dependencias

### Backend (Go)

```bash
cd Backend
go get -u ./...
go mod tidy
```

### Frontend (npm)

```bash
cd Frontend
npm update
```

---

## 📝 Comandos Útiles

```bash
# Ver logs del Backend en tiempo real
tail -f Backend/logs/server.log

# Recompilar Backend después de cambios
cd Backend
go build -o bin/server cmd/server/main.go

# Limpiar caché de npm
cd Frontend
npm cache clean --force
rm -rf node_modules package-lock.json
npm install

# Ver procesos del proyecto
ps aux | grep -E '(server|vite)'

# Reiniciar todo
pkill -f "bin/server"
cd Backend && ./bin/server &
cd ../Frontend && npm run dev
```

---

## 📚 Dependencias Principales

### Backend (Go)

- **Gin** v1.10.0 - Framework web HTTP
- **Go** v1.23.3 - Lenguaje de programación

Paquetes estándar utilizados:
- `encoding/binary` - Lectura binaria de discos
- `os` - Operaciones de sistema de archivos
- `net/http` - Servidor HTTP

### Frontend (Node.js)

- **React** 18.x - Librería UI
- **Vite** 5.x - Build tool y dev server
- **React Router** 6.x - Enrutamiento

---

## ✅ Checklist Post-Instalación

- [ ] Go instalado y en PATH
- [ ] Node.js y npm instalados
- [ ] Backend compilado (`Backend/bin/server` existe)
- [ ] Frontend con dependencias (`Frontend/node_modules` existe)
- [ ] Backend se inicia sin errores
- [ ] Frontend se inicia sin errores
- [ ] Se puede acceder a http://localhost:5173
- [ ] La API responde en http://localhost:8080/health

---

## 🆘 Soporte

Si encuentras problemas no cubiertos en esta guía:

1. Verifica los logs: `Backend/logs/server.log`
2. Revisa la consola del navegador (F12)
3. Asegúrate de tener las versiones correctas de Go y Node.js
4. Verifica que los puertos 8080 y 5173 estén libres

---

## 📄 Licencia

Proyecto académico - MIA 2S 2025

---

**¡Instalación completada exitosamente! 🎉**
