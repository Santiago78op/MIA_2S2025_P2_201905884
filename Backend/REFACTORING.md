# Refactorización del Backend - MIA Proyecto 2

## Resumen de Cambios

Se ha realizado una refactorización completa del backend, mejorando la estructura modular y la separación de responsabilidades.

## Estructura Mejorada

### 1. **Comandos Modularizados** (`internal/commands/`)

#### Antes
- Un solo archivo `adapter.go` con toda la lógica de comandos (~334 líneas)
- Switch case gigante con toda la lógica inline
- Difícil de mantener y extender

#### Después
Se dividió en 4 archivos especializados:

- **`types.go`**: Define todos los structs de comandos (MkdiskCommand, FdiskCommand, etc.)
  - Cada comando es un struct independiente
  - Implementa la interfaz `CommandHandler`
  - Validaciones específicas por comando

- **`parser.go`**: Manejo de parseo de comandos
  - Parseo de líneas de comando con soporte para argumentos
  - Tokenización respetando comillas
  - Generación de objetos de comando tipados

- **`handlers.go`**: Lógica de ejecución de cada comando
  - Método `Execute()` para cada tipo de comando
  - Funciones helper (`toBytes`, `parsePermissions`)
  - Código más limpio y testeable

- **`adapter.go`**: Coordinador simplificado (~47 líneas)
  - Solo coordina entre parser, validador y ejecutor
  - Selección automática de filesystem (EXT2/EXT3)
  - Mucho más simple y mantenible

### 2. **Patrón Command**

Cada comando ahora es una entidad independiente que:
- Se parsea a sí mismo
- Se valida a sí mismo
- Se ejecuta a sí mismo

```go
type CommandHandler interface {
    Execute(ctx context.Context, adapter *Adapter) (string, error)
    Validate() error
    Name() CommandName
}
```

### 3. **Comandos Implementados**

#### Disco/Particiones
- `MkdiskCommand`: Crear discos
- `FdiskCommand`: Gestionar particiones (add/delete)
- `MountCommand`: Montar particiones
- `UnmountCommand`: Desmontar particiones

#### Formateo
- `MkfsCommand`: Formatear particiones (2fs/3fs)

#### Archivos/Directorios
- `MkdirCommand`: Crear directorios
- `MkfileCommand`: Crear archivos
- `RemoveCommand`: Eliminar archivos/directorios
- `EditCommand`: Editar archivos
- `RenameCommand`: Renombrar
- `CopyCommand`: Copiar
- `MoveCommand`: Mover
- `FindCommand`: Buscar archivos
- `ChownCommand`: Cambiar propietario
- `ChmodCommand`: Cambiar permisos

#### EXT3 Específicos
- `JournalingCommand`: Ver journal
- `RecoveryCommand`: Recuperar datos
- `LossCommand`: Simular pérdida

### 4. **Correcciones Realizadas**

1. **disk/alloc.go**: Corregido tipo `*File` → `*os.File`
2. **fs/ext3/**: Eliminado archivo duplicado `calc.go`
3. **fs/ext3/ext3.go**: Simplificado Mkfs para compilación
4. **cmd/server/**: Actualizado para usar nueva estructura de comandos
5. **main.go**: Inicializado correctamente `disk.Manager`

### 5. **Mejoras de Código**

- **Separación de responsabilidades**: Cada archivo tiene un propósito claro
- **Testabilidad**: Comandos aislados son fáciles de testear
- **Extensibilidad**: Agregar nuevos comandos es trivial
- **Mantenibilidad**: Código más limpio y organizado
- **Type Safety**: Uso de structs tipados en lugar de map[string]interface{}

## Estructura de Archivos

```
Backend/
├── cmd/
│   └── server/
│       ├── main.go          # Inicialización (mejorado)
│       ├── server.go         # HTTP handlers (actualizado)
│       ├── types.go          # Request/Response types
│       └── cors.go
├── internal/
│   ├── commands/
│   │   ├── adapter.go        # Coordinador (simplificado)
│   │   ├── types.go          # Definición de comandos (NUEVO)
│   │   ├── parser.go         # Parseo de comandos (NUEVO)
│   │   ├── handlers.go       # Ejecución de comandos (NUEVO)
│   │   └── mount_index.go
│   ├── disk/
│   │   ├── manager.go
│   │   ├── mbr.go
│   │   ├── alloc.go          # (corregido)
│   │   └── ...
│   └── fs/
│       ├── ext2/
│       ├── ext3/             # (corregido)
│       └── ...
└── bin/
    └── server                # Binario compilado (8.6MB)
```

## Compilación

```bash
cd Backend
go build -o bin/server ./cmd/server
```

## Uso

```bash
# Iniciar servidor
./bin/server

# El servidor escucha en puerto 8080 por defecto
# Configurar con variables de entorno:
PORT=3000 ALLOW_ORIGIN="*" ./bin/server
```

## API Endpoints

- `GET /healthz` - Health check
- `GET /api/version` - Información de versión
- `POST /api/cmd/run` - Ejecutar comando

### Ejemplo de Request

```json
POST /api/cmd/run
{
  "line": "mkdisk -path /tmp/disk1.mia -size 10 -unit m"
}
```

### Ejemplo de Response

```json
{
  "ok": true,
  "output": "mkdisk OK path=/tmp/disk1.mia size=10m fit=ff",
  "input": "mkdisk -path /tmp/disk1.mia -size 10 -unit m"
}
```

## Próximos Pasos (TODOs)

1. **Implementar particiones lógicas**: Completar soporte EBR
2. **Journal Store**: Implementar persistencia real del journal
3. **Filesystem Operations**: Completar operaciones EXT2/EXT3
4. **Testing**: Agregar tests unitarios para cada comando
5. **Validaciones**: Mejorar validaciones de parámetros
6. **Documentación**: Agregar ejemplos de uso para cada comando

## Beneficios de la Refactorización

✅ **Código más limpio**: De 1 archivo monolítico a 4 archivos especializados
✅ **Fácil de extender**: Agregar comandos nuevos es simple
✅ **Mejor testabilidad**: Comandos aislados y testeables
✅ **Type safety**: Structs tipados en lugar de maps
✅ **Mantenible**: Cada archivo tiene una responsabilidad clara
✅ **Compila sin errores**: Todas las dependencias resueltas

## Comandos de Ejemplo

```bash
# Crear disco
mkdisk -path /tmp/disk.mia -size 100 -unit m -fit ff

# Crear partición
fdisk -path /tmp/disk.mia -mode add -name Part1 -size 50 -unit m -type p -fit bf

# Montar partición
mount -path /tmp/disk.mia -name Part1

# Formatear
mkfs -id vd12345678 -fs 2fs

# Crear directorio
mkdir -id vd12345678 -path /home/user -p

# Crear archivo
mkfile -id vd12345678 -path /home/user/file.txt -cont "Hello World"
```

---

**Refactorización completada exitosamente** ✨
