# REPORTE DE COMPARACION EXHAUSTIVA: P1 vs P2

**Fecha de Analisis**: 2025-10-05
**Proyecto 1 (Referencia)**: `/home/julian/Documents/MIA_2S2025_P1_201905884/backend`
**Proyecto 2 (Actual)**: `/home/julian/Documents/MIA_2S2025_P2_201905884/Backend`

---

## RESUMEN EJECUTIVO

### Estado General: CRITICO - IMPLEMENTACION INCOMPLETA

El Proyecto 2 presenta **diferencias estructurales criticas** que lo hacen **incompatible binariamente** con el Proyecto 1. La mayoria de las estructuras de datos estan definidas pero **SIN IMPLEMENTACION FUNCIONAL**.

### Hallazgos Criticos (Bloqueantes):

1. **Estructuras EXT2 completamente ausentes** - No existen definiciones de Superblock, Inode ni Blocks en P2
2. **Algoritmos de fit (FF/BF/WF) parcialmente implementados** - Solo estructura basica, sin logica completa
3. **Particiones logicas (EBR) sin implementacion** - Funcion marcada como TODO
4. **Incompatibilidad binaria en estructuras MBR/Partition** - Diferentes tamaños y campos
5. **Implementaciones de comandos vacias** - mkfs, mkdir, mkfile retornan advertencias de no persistencia

---

## PARTE 1: COMPARACION DE ESTRUCTURAS DE DATOS

### 1.1 MBR (Master Boot Record)

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/strMBR.go
type MBR struct {
    MbrTamanio       int64        // 8 bytes
    MbrFechaCreacion int64        // 8 bytes
    MbrDiskSignature int64        // 8 bytes
    MbrFit           byte         // 1 byte
    MbrParticiones   [4]Partition // 4 * sizeof(Partition)
}

const MBR_SIZE = int(unsafe.Sizeof(MBR{}))

// Metodos implementados:
- NewMBR()
- WriteMBR()
- ReadMBR()
- SerializeMBR()
- DeserializeMBR()
- ValidarFit()
- GetParticionLibre()
- GetParticionByName()
- HasExtendedPartition()
- CountActivePartitions()
- GetFreeSpace()
- FindBestFitPosition()  // *** IMPLEMENTA ALGORITMOS FF/BF/WF ***
- getFreeSpaces()        // *** CALCULA ESPACIOS LIBRES ***
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/disk/mbr.go
type MBR struct {
    SizeBytes int64           // 8 bytes
    CreatedAt int64           // 8 bytes (unix seconds)
    DiskSig   int32           // *** 4 bytes (DIFERENTE!) ***
    Fit       byte            // 1 byte
    _         [7]byte         // *** 7 bytes padding (NUEVO) ***
    Parts     [MaxPrimaries]Partition // 4 * sizeof(Partition)
}

// Metodos: NewMBR() - SOLO CONSTRUCTOR BASICO
// *** FALTA: Todos los metodos de operacion (read, write, serialize, etc.) ***
```

#### DIFERENCIAS CRITICAS:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **DiskSignature** | `int64` (8 bytes) | `int32` (4 bytes) | **CRITICO** - Incompatibilidad binaria |
| **Padding** | Sin padding explicito | `[7]byte` padding | **CRITICO** - Tamaño total diferente |
| **Campos** | `MbrTamanio`, `MbrFechaCreacion` | `SizeBytes`, `CreatedAt` | **MODERADO** - Solo nombres |
| **Metodos** | 15+ metodos completos | Solo constructor | **CRITICO** - Sin funcionalidad |
| **Algoritmos Fit** | Implementados en `FindBestFitPosition()` | **AUSENTES** | **CRITICO** - No puede asignar particiones |

**Calculo de Tamaño:**
- P1: `MBR_SIZE = unsafe.Sizeof(MBR{})` - Calculado dinamicamente
- P2: No hay calculo de tamaño - **FALTA**

---

### 1.2 PARTITION (Particion)

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/strPartition.go
type Partition struct {
    PartStatus      byte         // 1 byte - 0=Inactiva, 1=Activa
    PartType        byte         // 1 byte - P/E/L
    PartFit         byte         // 1 byte - B/F/W
    PartStart       int64        // 8 bytes
    PartSize        int64        // 8 bytes
    PartName        [16]byte     // 16 bytes
    PartCorrelativo int64        // 8 bytes - Para montaje
    PartID          [4]byte      // 4 bytes - ID generado al montar
}

// Total: 1 + 1 + 1 + 8 + 8 + 16 + 8 + 4 = 47 bytes base
// (mas padding del compilador)

// Metodos: 25+ metodos incluyendo:
- NewPartition(), NewEmptyPartition()
- GetName(), SetName(), GetID(), SetID()
- IsEmpty(), IsActive(), IsMounted()
- IsPrimary(), IsExtended(), IsLogical()
- GetEndPosition(), Contains(), Overlaps()
- ValidatePartition()
- Mount(), Unmount(), Delete(), Clone()
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/disk/mbr.go
type Partition struct {
    Status byte       // 1 byte - 0=libre, 1=usada
    Type   byte       // 1 byte - P/E
    Fit    byte       // 1 byte
    Start  int64      // 8 bytes
    Size   int64      // 8 bytes
    Name   [NameLen]byte // 16 bytes (NameLen=16)
    _      [8]byte    // *** 8 bytes padding ***
}

// Total: 1 + 1 + 1 + 8 + 8 + 16 + 8 = 43 bytes
// *** FALTAN: PartCorrelativo (8 bytes), PartID (4 bytes) ***

// Metodos: NINGUNO - Solo definicion de estructura
```

#### DIFERENCIAS CRITICAS:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **PartCorrelativo** | Presente (int64) | **AUSENTE** | **CRITICO** - No puede trackear montajes |
| **PartID** | Presente ([4]byte) | **AUSENTE** | **CRITICO** - No puede generar IDs |
| **Padding** | Implicito | Explicito [8]byte | **MODERADO** - Diferentes tamaños |
| **Nombres de campos** | Part* prefix | Sin prefijo | **MENOR** - Solo estilo |
| **Metodos** | 25+ metodos | **NINGUNO** | **CRITICO** - Sin funcionalidad |
| **Tipo Logica** | Soporta 'L' | Solo P/E | **CRITICO** - No maneja logicas en struct |

**Implicacion:** P2 no puede:
- Trackear que particiones estan montadas (sin PartCorrelativo)
- Generar IDs de montaje (sin PartID)
- Validar estado de particiones (sin metodos)
- Montar/desmontar particiones programaticamente

---

### 1.3 EBR (Extended Boot Record)

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/strEBR.go
type EBR struct {
    PartMount byte         // 1 byte - Indica si esta montada
    PartFit   byte         // 1 byte - B/F/W
    PartStart int64        // 8 bytes - Inicio de datos (no del EBR)
    PartSize  int64        // 8 bytes
    PartNext  int64        // 8 bytes - Posicion del siguiente EBR (-1 si no hay)
    PartName  [16]byte     // 16 bytes
}

const EBR_SIZE = int(unsafe.Sizeof(EBR{}))

// Metodos: 20+ metodos incluyendo:
- NewEBR(), NewEmptyEBR()
- GetName(), SetName()
- IsEmpty(), IsMounted(), HasNext()
- GetEndPosition(), GetFitString()
- Mount(), Unmount(), ValidateEBR()
- SerializeEBR(), DeserializeEBR()
- WriteEBR(), ReadEBR()
- ReadAllEBRs()     // *** Lee cadena completa ***
- FindEBRByName()   // *** Busca en cadena ***
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/disk/mbr.go
type EBR struct {
    Status byte       // 1 byte
    Fit    byte       // 1 byte
    Start  int64      // 8 bytes - inicio del EBR (posicion propia)
    Size   int64      // 8 bytes
    Next   int64      // 8 bytes
    Name   [NameLen]byte // 16 bytes
    _      [8]byte    // 8 bytes padding
}

// Metodos: NINGUNO
// Funciones auxiliares en alloc.go:
func listEBRs(f *os.File, extStart, extEnd int64) ([]EBR, error) {
    return nil, errors.New("TODO")  // *** NO IMPLEMENTADO ***
}
```

#### DIFERENCIAS CRITICAS:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **PartMount** | Presente | `Status` (semantica diferente) | **CRITICO** - Confusion de proposito |
| **PartStart** | Inicio de DATOS | Inicio del EBR mismo | **CRITICO** - Logica incompatible |
| **Metodos** | 20+ metodos | **NINGUNO** | **CRITICO** - Sin funcionalidad |
| **listEBRs()** | ReadAllEBRs() implementado | return errors.New("TODO") | **CRITICO** - No puede leer cadenas |
| **Padding** | Sin padding explicito | [8]byte padding | **MODERADO** - Tamaño diferente |

**Implicacion:** P2 **NO PUEDE** trabajar con particiones logicas:
- No puede crear cadenas de EBRs
- No puede leer EBRs existentes
- No puede buscar particiones logicas
- Funcion marcada explicitamente como TODO

---

### 1.4 SUPERBLOCK EXT2

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/systemFileExt2/strSuperbloque.go
type Superblock struct {
    S_filesystem_type   int32  // 2 = EXT2
    S_inodes_count      int32
    S_blocks_count      int32
    S_free_blocks_count int32
    S_free_inodes_count int32
    S_mtime             int64  // Unix timestamp
    S_umtime            int64  // Unix timestamp
    S_mnt_count         int32
    S_magic             int32  // 0xEF53
    S_inode_size        int32
    S_block_size        int32
    S_first_ino         int32
    S_first_blo         int32
    S_bm_inode_start    int32
    S_bm_block_start    int32
    S_inode_start       int32
    S_block_start       int32
}

const (
    EXT2_FILESYSTEM_TYPE = 2
    EXT2_MAGIC           = 0xEF53
    SUPERBLOCK_SIZE      = int(unsafe.Sizeof(Superblock{}))
    INODE_SIZE           = int(unsafe.Sizeof(Inode{}))
    BLOCK_SIZE           = 64
)

// Funciones implementadas:
- NewSuperblock()          // *** CALCULA n SEGUN FORMULA ***
- SerializeSuperblock()
- DeserializeSuperblock()
- CalculateEXT2Structures() // *** FORMULA: n = (size - sb) / (1+3+inode+3*block) ***
- ValidateEXT2Structures()  // *** VALIDA MAGIC, RATIOS, ETC ***
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/fs/ext2/ext2.go
const (
    EXT2_MAGIC       = 0xEF53  // *** CORRECTO ***
    SUPERBLOCK_SIZE  = 512     // *** HARDCODED (deberia ser sizeof) ***
    INODE_SIZE       = 128     // *** HARDCODED (deberia ser sizeof) ***
    BLOCK_SIZE       = 64      // *** CORRECTO ***
)

// *** NO HAY DEFINICION DE STRUCT Superblock ***
// *** NO HAY DEFINICION DE STRUCT Inode ***
// *** NO HAY DEFINICION DE STRUCT FolderBlock ***
// *** NO HAY DEFINICION DE STRUCT FileBlock ***
// *** NO HAY DEFINICION DE STRUCT PointerBlock ***

// Solo hay calculo parcial en Mkfs():
func (e *FS2) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
    partitionSize := int64(64 * 1024 * 1024) // *** HARDCODED! ***

    available := partitionSize - int64(SUPERBLOCK_SIZE)
    denominator := int64(1 + 3 + INODE_SIZE + 3*BLOCK_SIZE)
    n := available / denominator

    // *** NO ESCRIBE NADA AL DISCO ***
    e.logger.Printf("Advertencia: Mkfs no persistente todavia")
}
```

#### DIFERENCIAS CRITICAS:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **Struct Superblock** | Completo (17 campos) | **AUSENTE** | **CRITICO** - No puede formatear |
| **Struct Inode** | Completo | **AUSENTE** | **CRITICO** - No puede crear archivos |
| **Struct Blocks** | 3 tipos (Folder/File/Pointer) | **AUSENTE** | **CRITICO** - No puede almacenar datos |
| **Constantes** | Calculadas con sizeof | Hardcoded | **MODERADO** - Inconsistencia |
| **Formula n** | Implementada correctamente | Implementada pero **no persiste** | **CRITICO** - Solo calcula, no escribe |
| **Validacion** | ValidateEXT2Structures() completo | **AUSENTE** | **CRITICO** - No valida integridad |
| **Serializacion** | Funciones completas | **AUSENTE** | **CRITICO** - No puede escribir a disco |

---

### 1.5 INODE EXT2

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/systemFileExt2/strInodo.go
type Inode struct {
    IUid   int32         // UID propietario
    IGid   int32         // GID grupo
    IS     int32         // Tamaño en bytes
    IAtime int64         // Ultimo acceso
    ICtime int64         // Creacion
    IMtime int64         // Modificacion
    IBlock [15]int32     // 12 directos + 3 indirectos
    IType  byte          // 0=carpeta, 1=archivo
    IPerm  [3]byte       // Permisos UGO
}

// Metodos:
- NewInode()  // Inicializa todos los bloques a -1
```

#### PROYECTO 2 (Actual)
```go
// *** NO EXISTE DEFINICION DE INODE ***
```

---

### 1.6 BLOCKS EXT2

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/systemFileExt2/strBlock.go

// Content dentro de FolderBlock
type Content struct {
    BName  [12]byte  // Nombre archivo/carpeta
    BInodo int32     // Indice del inodo
}

// Bloque carpeta
type FolderBlock struct {
    BContent [4]Content  // 4 entradas de 16 bytes = 64 bytes
}

// Bloque archivo
type FileBlock struct {
    BContent [64]byte    // 64 bytes de datos
}

// Bloque apuntadores
type PointerBlock struct {
    BPointers [16]int32  // 16 apuntadores de 4 bytes = 64 bytes
}

// Metodos:
- NewContent(), NewFolderBlock(), NewFileBlock(), NewPointerBlock()
- AddContent()  // *** Maneja logica de carpetas llenas ***
- GetContent(), GetName(), IsEmpty()
```

#### PROYECTO 2 (Actual)
```go
// *** NO EXISTEN DEFINICIONES DE BLOQUES ***
```

---

## PARTE 2: COMPARACION DE ALGORITMOS

### 2.1 Algoritmos de Fit (First/Best/Worst Fit)

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/strMBR.go
func (m *MBR) FindBestFitPosition(size int64) (int64, error) {
    freeSpaces := m.getFreeSpaces()

    switch m.MbrFit {
    case PartitionFitFirst:  // First Fit
        for _, space := range freeSpaces {
            if space.Size >= size {
                return space.Start, nil
            }
        }

    case PartitionFitBest:  // Best Fit
        bestSpace := FreeSpace{Size: m.MbrTamanio + 1}
        found := false
        for _, space := range freeSpaces {
            if space.Size >= size && space.Size < bestSpace.Size {
                bestSpace = space
                found = true
            }
        }
        if found {
            return bestSpace.Start, nil
        }

    case PartitionFitWorst:  // Worst Fit
        worstSpace := FreeSpace{Size: -1}
        for _, space := range freeSpaces {
            if space.Size >= size && space.Size > worstSpace.Size {
                worstSpace = space
            }
        }
        if worstSpace.Size >= size {
            return worstSpace.Start, nil
        }
    }

    return -1, fmt.Errorf("no se encontro espacio suficiente")
}

// *** ALGORITMO DE ESPACIOS LIBRES ***
func (m *MBR) getFreeSpaces() []FreeSpace {
    // 1. Crear lista de espacios ocupados
    occupiedSpaces := []FreeSpace{{Start: 0, Size: int64(MBR_SIZE)}}
    for i := range m.MbrParticiones {
        if m.MbrParticiones[i].PartStatus == StatusActiva {
            occupiedSpaces = append(occupiedSpaces, FreeSpace{
                Start: m.MbrParticiones[i].PartStart,
                Size:  m.MbrParticiones[i].PartSize,
            })
        }
    }

    // 2. Ordenar por Start
    // ... (bubble sort implementation)

    // 3. Encontrar espacios libres entre ocupados
    currentPos := int64(0)
    for _, occupied := range occupiedSpaces {
        if currentPos < occupied.Start {
            spaces = append(spaces, FreeSpace{
                Start: currentPos,
                Size:  occupied.Start - currentPos,
            })
        }
        currentPos = occupied.Start + occupied.Size
    }

    // 4. Espacio libre al final
    if currentPos < m.MbrTamanio {
        spaces = append(spaces, FreeSpace{
            Start: currentPos,
            Size:  m.MbrTamanio - currentPos,
        })
    }

    return spaces
}
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/disk/alloc.go
func buildFreePrimaries(mbr *MBR) []seg {
    used := make([]seg, 0, MaxPrimaries)
    for _, p := range mbr.Parts {
        if p.Status == PartStatusUsed &&
           (p.Type == PartTypePrimary || p.Type == PartTypeExtended) {
            used = append(used, seg{p.Start, p.Start + p.Size})
        }
    }
    sort.Slice(used, func(i, j int) bool { return used[i].start < used[j].start })

    free := []seg{}
    cur := int64(binarySizeOf(mbr))  // *** USA HELPER MANUAL ***
    limit := mbr.SizeBytes

    for _, u := range used {
        if u.start > cur {
            free = append(free, seg{cur, u.start})
        }
        if u.end > cur {
            cur = u.end
        }
    }
    if cur < limit {
        free = append(free, seg{cur, limit})
    }
    return free
}

func pickByFit(free []seg, need int64, fit byte) (seg, bool) {
    candidates := []seg{}
    for _, s := range free {
        if s.size() >= need {
            candidates = append(candidates, s)
        }
    }
    if len(candidates) == 0 {
        return seg{}, false
    }

    switch fit {
    case FitFF:  // First Fit
        return candidates[0], true

    case FitBF:  // Best Fit
        sort.Slice(candidates, func(i, j int) bool {
            return candidates[i].size() < candidates[j].size()
        })
        return candidates[0], true

    case FitWF:  // Worst Fit
        sort.Slice(candidates, func(i, j int) bool {
            return candidates[i].size() > candidates[j].size()
        })
        return candidates[0], true

    default:
        return candidates[0], true
    }
}
```

#### COMPARACION:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **Estructura** | Metodo de MBR | Funciones auxiliares separadas | **MODERADO** - Mismo resultado |
| **Ordenamiento** | Bubble sort manual | sort.Slice() de stdlib | **MENOR** - P2 es mejor |
| **First Fit** | Correcto | Correcto | **OK** |
| **Best Fit** | Correcto | Correcto | **OK** |
| **Worst Fit** | Correcto | Correcto | **OK** |
| **Integracion** | Usado en FindBestFitPosition() | **NO USADO** en ningun lado | **CRITICO** - No integrado |
| **Calculo MBR size** | unsafe.Sizeof() | binarySizeOf() manual | **MODERADO** - Propenso a errores |

**Conclusion:** Los algoritmos estan implementados correctamente en P2, pero **no estan integrados** con el resto del sistema. P1 los usa en fdisk, P2 no los llama desde ningun comando.

---

### 2.2 Algoritmos de Particiones Logicas (EBR)

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/command/disk/fdisk.go
func createLogicalPartition(path, name string, sizeInBytes int64, fit string) error {
    // 1. Buscar particion extendida
    extendedPartition, _, err := estructuras.GetExtendedPartition(path)

    // 2. Leer todos los EBRs existentes
    ebrs, err := estructuras.ReadAllEBRs(path, extendedPartition.PartStart)

    // 3. Calcular espacios ocupados
    occupiedSpaces := []estructuras.FreeSpace{
        {Start: extendedPartition.PartStart, Size: int64(estructuras.EBR_SIZE)},
    }
    for _, ebr := range ebrs {
        if !ebr.IsEmpty() {
            totalSize := int64(estructuras.EBR_SIZE) + ebr.PartSize
            occupiedSpaces = append(occupiedSpaces, estructuras.FreeSpace{
                Start: ebr.PartStart - int64(estructuras.EBR_SIZE),
                Size:  totalSize,
            })
        }
    }

    // 4. Encontrar espacios libres
    freeSpaces := findFreeSpacesInExtended(extendedPartition, occupiedSpaces)

    // 5. Aplicar algoritmo de ajuste (FF/BF/WF)
    neededSize := sizeInBytes + int64(estructuras.EBR_SIZE)
    // ... (implementacion igual que para primarias)

    // 6. Crear nuevo EBR
    newEBR := estructuras.NewEBR(fitByte, logicalStart+EBR_SIZE, sizeInBytes, name, nextPos)

    // 7. Escribir EBR al disco
    estructuras.WriteEBR(path, newEBR, logicalStart)

    // 8. Actualizar cadena de EBRs
    updateEBRChain(path, extendedPartition.PartStart, logicalStart)
}

// *** ALGORITMO DE CADENA DE EBRs ***
func updateEBRChain(path string, extendedStart, newEBRPosition int64) error {
    currentPos := extendedStart
    var previousPos int64 = -1

    // Encontrar el EBR que debe apuntar al nuevo
    for currentPos != -1 && currentPos < newEBRPosition {
        ebr, _ := estructuras.ReadEBR(path, currentPos)

        if ebr.PartNext == -1 || ebr.PartNext > newEBRPosition {
            previousPos = currentPos
            break
        }

        previousPos = currentPos
        currentPos = ebr.PartNext
    }

    // Actualizar puntero del EBR anterior
    if previousPos != -1 {
        prevEBR, _ := estructuras.ReadEBR(path, previousPos)
        oldNext := prevEBR.PartNext
        prevEBR.PartNext = newEBRPosition
        estructuras.WriteEBR(path, prevEBR, previousPos)

        // Hacer que el nuevo apunte al siguiente en la cadena
        newEBR, _ := estructuras.ReadEBR(path, newEBRPosition)
        newEBR.PartNext = oldNext
        estructuras.WriteEBR(path, newEBR, newEBRPosition)
    }
}
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/disk/alloc.go
func listEBRs(f *os.File, extStart, extEnd int64) ([]EBR, error) {
    return nil, errors.New("TODO")  // *** NO IMPLEMENTADO ***
}
```

#### DIFERENCIAS CRITICAS:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **Creacion de logicas** | Implementado completo | **TODO** | **CRITICO** - No funciona |
| **Lectura de EBRs** | ReadAllEBRs() | errors.New("TODO") | **CRITICO** - No puede leer |
| **Escritura de EBRs** | WriteEBR() | **AUSENTE** | **CRITICO** - No puede escribir |
| **Cadena de EBRs** | updateEBRChain() | **AUSENTE** | **CRITICO** - No mantiene lista |
| **Busqueda por nombre** | FindEBRByName() | **AUSENTE** | **CRITICO** - No puede buscar |
| **Algoritmos fit en EBR** | Mismo que primarias | **AUSENTE** | **CRITICO** - No aplica fit |

**Conclusion:** P2 **NO PUEDE** trabajar con particiones logicas en absoluto.

---

### 2.3 Formula de Calculo de Estructuras (n)

#### PROYECTO 1 (Referencia)
```go
// Archivo: /backend/struct/systemFileExt2/strSuperbloque.go
func CalculateEXT2Structures(partitionSize int64) (int32, int32) {
    available := partitionSize - int64(SUPERBLOCK_SIZE)
    denominator := int64(1 + 3 + INODE_SIZE + 3*BLOCK_SIZE)
    n := available / denominator

    inodesCount := int32(n)
    blocksCount := int32(3 * n)  // Relacion 3:1

    return inodesCount, blocksCount
}

// Usado en NewSuperblock():
func NewSuperblock(partitionSize int64) *Superblock {
    available := partitionSize - int64(SUPERBLOCK_SIZE)
    denominator := int64(1 + 3 + INODE_SIZE + 3*BLOCK_SIZE)
    n := available / denominator

    inodesCount := int32(n)
    blocksCount := int32(3 * n)

    // ... crea superblock completo
    return &Superblock{
        S_inodes_count: inodesCount,
        S_blocks_count: blocksCount,
        S_bm_inode_start: int32(SUPERBLOCK_SIZE),
        S_bm_block_start: int32(SUPERBLOCK_SIZE) + inodesCount,
        S_inode_start: int32(SUPERBLOCK_SIZE) + inodesCount + blocksCount,
        S_block_start: int32(SUPERBLOCK_SIZE) + inodesCount + blocksCount +
                       inodesCount*int32(INODE_SIZE),
        // ...
    }
}
```

#### PROYECTO 2 (Actual)
```go
// Archivo: /internal/fs/ext2/ext2.go
func (e *FS2) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
    // *** TAMAÑO HARDCODED ***
    partitionSize := int64(64 * 1024 * 1024) // 64MB por defecto

    available := partitionSize - int64(SUPERBLOCK_SIZE)
    denominator := int64(1 + 3 + INODE_SIZE + 3*BLOCK_SIZE)
    n := available / denominator

    inodesCount := int32(n)
    blocksCount := int32(3 * n)

    // *** NO CREA SUPERBLOCK REAL ***
    // *** SOLO GUARDA EN MEMORIA ***
    meta := fs.Meta{
        FSKind:   "2fs",
        BlockSz:  BLOCK_SIZE,
        InodeSz:  INODE_SIZE,
        JournalN: 0,
    }
    e.state.Set(req.MountID, meta)

    // *** NO ESCRIBE NADA AL DISCO ***
    e.logger.Printf("Formateo EXT2 completado: %d inodos, %d bloques",
                    inodesCount, blocksCount)
    return nil
}

// Helper separado (no usado):
func calculateStructures(partitionSize int64) (inodes, blocks int32) {
    available := partitionSize - int64(SUPERBLOCK_SIZE)
    denominator := int64(1 + 3 + INODE_SIZE + 3*BLOCK_SIZE)
    n := available / denominator
    return int32(n), int32(3 * n)
}
```

#### COMPARACION:

| Aspecto | P1 | P2 | Impacto |
|---------|----|----|---------|
| **Formula matematica** | Correcta | Correcta | **OK** |
| **Uso de formula** | Crea estructuras reales | Solo calcula | **CRITICO** - No persiste |
| **Tamaño de particion** | Leido de MBR | **HARDCODED 64MB** | **CRITICO** - Ignora tamaño real |
| **Creacion de Superblock** | Completo con 17 campos | **NO CREA** | **CRITICO** - No formatea |
| **Calculo de offsets** | Todos los campos calculados | **AUSENTE** | **CRITICO** - No sabe donde escribir |
| **Escritura a disco** | WriteToDisk() | **AUSENTE** | **CRITICO** - No persiste |
| **Validacion** | ValidateEXT2Structures() | **AUSENTE** | **CRITICO** - No valida |

**Conclusion:** La formula esta bien implementada pero **no se usa** para formatear realmente. P2 solo hace el calculo y guarda metadatos en memoria, nunca escribe al disco.

---

## PARTE 3: COMPARACION DE IMPLEMENTACIONES

### 3.1 Comando MKDISK

| Aspecto | P1 | P2 |
|---------|----|----|
| Archivo | `/backend/command/disk/mkdisk.go` | Handler en `/internal/commands/handlers.go` |
| Validaciones | Completas (size>0, fit valido, etc) | Basicas |
| Creacion de archivo | Crea archivo binario lleno de ceros | Via DM.Mkdisk() |
| Escritura de MBR | WriteMBR() completo | **Implementacion desconocida** |
| Resultado | Disco .mia funcional | **Por verificar** |

### 3.2 Comando FDISK

| Aspecto | P1 | P2 |
|---------|----|----|
| Archivo | `/backend/command/disk/fdisk.go` | Handler en `/internal/commands/handlers.go` |
| Lineas de codigo | 664 lineas | ~20 lineas (solo delegacion) |
| **Primarias** | Implementado completo | Via DM.FdiskAdd() - **verificar** |
| **Extendidas** | Implementado completo | Via DM.FdiskAdd() - **verificar** |
| **Logicas** | Implementado completo (EBR chain) | **TODO en listEBRs()** |
| Algoritmos Fit | Implementados y usados | **Implementados pero no usados** |
| Validaciones | 15+ validaciones | **Basicas** |
| Espacios libres | getFreeSpaces() completo | buildFreePrimaries() - **no integrado** |

### 3.3 Comando MKFS

| Aspecto | P1 | P2 |
|---------|----|----|
| Archivo | `/backend/command/adminSistemFile/mkfs.go` | `/internal/fs/ext2/ext2.go` |
| Lineas de codigo | 451 lineas | ~70 lineas |
| **Superblock** | Crea y escribe completo | **NO CREA** |
| **Bitmaps** | Inicializa y escribe | **NO CREA** |
| **Inodo raiz** | Crea y escribe | **NO CREA** |
| **users.txt** | Crea inodo y bloque | **NO CREA** |
| **Tabla de inodos** | Inicializa completa | **NO CREA** |
| **Tabla de bloques** | Inicializa completa | **NO CREA** |
| Persistencia | Escribe todo a disco | **ADVERTENCIA: "no persistente todavia"** |
| Validacion | ValidateEXT2Structures() | **AUSENTE** |

**LOG ACTUAL DE P2:**
```
e.logger.Printf("Advertencia: WriteFile no persistente todavía")
e.logger.Printf("Advertencia: Mkdir no persistente todavía")
e.logger.Printf("Advertencia: Remove no persistente todavía")
e.logger.Printf("Advertencia: Rename no persistente todavía")
e.logger.Printf("Advertencia: Copy no persistente todavía")
e.logger.Printf("Advertencia: Move no persistente todavía")
e.logger.Printf("Advertencia: Chown no persistente todavía")
e.logger.Printf("Advertencia: Chmod no persistente todavía")
```

### 3.4 Operaciones de Archivos (mkdir, mkfile, etc.)

| Comando | P1 | P2 |
|---------|----|----|
| **mkdir** | Implementado (crea inodo + bloque) | "Advertencia: no persistente" |
| **mkfile** | Implementado (crea inodo + bloques) | "Advertencia: no persistente" |
| **remove** | Implementado (actualiza bitmaps) | "Advertencia: no persistente" |
| **edit** | Implementado (modifica bloques) | "Advertencia: no persistente" |
| **rename** | Implementado (modifica FolderBlock) | "Advertencia: no persistente" |
| **copy** | Implementado (duplica inodo+bloques) | "Advertencia: no persistente" |
| **move** | Implementado (combina copy+remove) | "Advertencia: no persistente" |
| **find** | Implementado (recorre arbol) | Retorna hardcoded ["/users.txt"] |
| **cat** | Implementado (lee bloques) | **NO IMPLEMENTADO** |

---

## PARTE 4: ANALISIS DE INCOMPATIBILIDADES BINARIAS

### 4.1 Tamaños de Estructuras

#### MBR
```go
// P1:
MBR_SIZE = sizeof(int64 + int64 + int64 + byte + [4]Partition)
         = 8 + 8 + 8 + 1 + 4*sizeof(Partition)

Partition_P1 = 1 + 1 + 1 + 8 + 8 + 16 + 8 + 4 = 47 bytes
              (+ padding del compilador)

// P2:
MBR_P2 = sizeof(int64 + int64 + int32 + byte + [7]byte + [4]Partition)
       = 8 + 8 + 4 + 1 + 7 + 4*sizeof(Partition)

Partition_P2 = 1 + 1 + 1 + 8 + 8 + 16 + 8 = 43 bytes
```

**PROBLEMA:** Si P2 lee un disco creado por P1:
- DiskSignature: P2 espera 4 bytes, P1 escribio 8 bytes → **DESALINEACION**
- Partition.PartCorrelativo: P2 no lo lee → **PIERDE INFORMACION**
- Partition.PartID: P2 no lo lee → **PIERDE INFORMACION**

#### EBR
```go
// P1:
EBR_SIZE = 1 + 1 + 8 + 8 + 8 + 16 = 42 bytes (+ padding)
PartStart = inicio de DATOS (EBR position + EBR_SIZE)

// P2:
EBR_SIZE = 1 + 1 + 8 + 8 + 8 + 16 + 8 = 50 bytes
Start = inicio del EBR mismo
```

**PROBLEMA:** Interpretacion diferente de Start:
- P1: `PartStart = 1024` significa datos en 1024, EBR en 1024-42
- P2: `Start = 1024` significa EBR en 1024, datos en 1024+50
- **INCOMPATIBILIDAD TOTAL** al leer/escribir EBRs

---

## PARTE 5: RECOMENDACIONES PRIORIZADAS

### PRIORIDAD CRITICA (Bloqueantes - Sin estos NO funciona)

#### 1. Definir Estructuras EXT2 Completas
**Archivos a crear:**
- `/internal/fs/ext2/superblock.go`
- `/internal/fs/ext2/inode.go`
- `/internal/fs/ext2/blocks.go`

**Codigo necesario:**
```go
// superblock.go
package ext2

type Superblock struct {
    FilesystemType   int32
    InodesCount      int32
    BlocksCount      int32
    FreeBlocksCount  int32
    FreeInodesCount  int32
    Mtime            int64
    Umtime           int64
    MntCount         int32
    Magic            int32  // 0xEF53
    InodeSize        int32
    BlockSize        int32
    FirstIno         int32
    FirstBlo         int32
    BmInodeStart     int32
    BmBlockStart     int32
    InodeStart       int32
    BlockStart       int32
}

func NewSuperblock(partSize int64) *Superblock {
    // Copiar implementacion de P1
}
```

```go
// inode.go
package ext2

type Inode struct {
    Uid   int32
    Gid   int32
    Size  int32
    Atime int64
    Ctime int64
    Mtime int64
    Block [15]int32
    Type  byte
    Perm  [3]byte
}

func NewInode(uid, gid, size int32, typ byte, perm [3]byte) *Inode {
    inode := &Inode{
        Uid: uid, Gid: gid, Size: size,
        Atime: time.Now().Unix(),
        Ctime: time.Now().Unix(),
        Mtime: time.Now().Unix(),
        Type: typ, Perm: perm,
    }
    for i := range inode.Block {
        inode.Block[i] = -1
    }
    return inode
}
```

```go
// blocks.go
package ext2

type Content struct {
    Name  [12]byte
    Inode int32
}

type FolderBlock struct {
    Content [4]Content
}

type FileBlock struct {
    Content [64]byte
}

type PointerBlock struct {
    Pointers [16]int32
}
```

**Estimado:** 200-300 lineas

---

#### 2. Implementar Persistencia de MKFS
**Archivo a modificar:** `/internal/fs/ext2/ext2.go`

**Cambios necesarios:**
```go
func (e *FS2) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
    // 1. Obtener particion real del disco
    ref, ok := e.state.GetRef(req.MountID)
    if !ok {
        return errors.New("mount not found")
    }

    // 2. Leer MBR y encontrar particion
    mbr, err := ReadMBR(ref.DiskPath)
    partition := mbr.FindPartitionByName(ref.PartitionID)
    partSize := partition.Size
    partStart := partition.Start

    // 3. Calcular estructuras
    sb := NewSuperblock(partSize)

    // 4. Escribir superblock
    sbBytes := SerializeSuperblock(sb)
    WriteToDisk(ref.DiskPath, sbBytes, partStart)

    // 5. Inicializar bitmaps
    InitializeBitmaps(ref.DiskPath, partStart, sb)

    // 6. Crear estructuras iniciales (raiz + users.txt)
    CreateInitialStructures(ref.DiskPath, partStart, sb)

    return nil
}
```

**Funciones auxiliares necesarias (copiar de P1):**
- `WriteToDisk(path string, data []byte, offset int64) error`
- `ReadFromDisk(path string, offset int64, size int) ([]byte, error)`
- `SerializeSuperblock(sb *Superblock) ([]byte, error)`
- `DeserializeSuperblock(data []byte) (*Superblock, error)`
- `InitializeBitmaps()`
- `CreateInitialStructures()`

**Estimado:** 400-500 lineas

---

#### 3. Implementar Operaciones con EBR
**Archivo a crear:** `/internal/disk/ebr.go`

**Funciones criticas:**
```go
func ReadEBR(f *os.File, position int64) (*EBR, error) {
    data := make([]byte, unsafe.Sizeof(EBR{}))
    _, err := f.ReadAt(data, position)
    if err != nil {
        return nil, err
    }

    ebr := &EBR{}
    buf := bytes.NewReader(data)
    binary.Read(buf, binary.LittleEndian, ebr)
    return ebr, nil
}

func WriteEBR(f *os.File, ebr *EBR, position int64) error {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, ebr)
    _, err := f.WriteAt(buf.Bytes(), position)
    return err
}

func ReadAllEBRs(f *os.File, extStart, extEnd int64) ([]EBR, error) {
    ebrs := []EBR{}
    current := extStart

    for current != -1 && current < extEnd {
        ebr, err := ReadEBR(f, current)
        if err != nil {
            return nil, err
        }

        if ebr.Size > 0 {  // No vacio
            ebrs = append(ebrs, *ebr)
        }

        current = ebr.Next

        if len(ebrs) > 100 {  // Prevenir loops infinitos
            return nil, errors.New("too many EBRs")
        }
    }

    return ebrs, nil
}

func FindEBRByName(f *os.File, extStart, extEnd int64, name string) (*EBR, int64, error) {
    current := extStart

    for current != -1 && current < extEnd {
        ebr, _ := ReadEBR(f, current)

        ebrName := string(bytes.TrimRight(ebr.Name[:], "\x00"))
        if ebrName == name {
            return ebr, current, nil
        }

        current = ebr.Next
    }

    return nil, -1, errors.New("EBR not found")
}
```

**Estimado:** 150-200 lineas

---

#### 4. Corregir Estructura Partition
**Archivo:** `/internal/disk/mbr.go`

**Cambios:**
```go
type Partition struct {
    Status      byte
    Type        byte
    Fit         byte
    Start       int64
    Size        int64
    Name        [NameLen]byte
    Correlative int64      // *** AGREGAR ***
    ID          [4]byte    // *** AGREGAR ***
    // Remover: _ [8]byte
}
```

**Nota:** Esto rompe compatibilidad binaria con discos existentes de P2, pero es necesario para ser compatible con P1.

---

### PRIORIDAD ALTA (Funcionalidad core)

#### 5. Implementar Operaciones de Archivos
**Archivos a modificar:** `/internal/fs/ext2/ext2.go`

Para cada operacion (mkdir, mkfile, remove, etc.), cambiar de:
```go
func (e *FS2) Mkdir(...) error {
    e.logger.Printf("Advertencia: Mkdir no persistente todavía")
    return nil
}
```

A implementacion real:
```go
func (e *FS2) Mkdir(ctx context.Context, h fs.MountHandle, req fs.MkdirRequest) error {
    // 1. Leer superblock
    sb, _ := ReadSuperblock(h.DiskID, partStart)

    // 2. Buscar inodo libre
    inodeIdx := FindFreeInode(h.DiskID, partStart, sb)

    // 3. Buscar bloque libre
    blockIdx := FindFreeBlock(h.DiskID, partStart, sb)

    // 4. Crear inodo carpeta
    inode := NewInode(h.User, h.Group, 0, INODE_TYPE_FOLDER, [3]byte{'7','5','5'})
    inode.Block[0] = blockIdx

    // 5. Crear bloque carpeta
    block := NewFolderBlock()
    block.Content[0] = Content{Name: ".", Inode: inodeIdx}
    block.Content[1] = Content{Name: "..", Inode: parentInodeIdx}

    // 6. Actualizar bitmaps
    SetBit(inodeBitmap, inodeIdx)
    SetBit(blockBitmap, blockIdx)

    // 7. Escribir a disco
    WriteInode(h.DiskID, inodeTableStart, inodeIdx, inode)
    WriteFolderBlock(h.DiskID, blockTableStart, blockIdx, block)
    UpdateSuperblock(h.DiskID, partStart, sb)

    return nil
}
```

**Funciones auxiliares necesarias:**
- `FindFreeInode()`, `FindFreeBlock()`
- `ReadInode()`, `WriteInode()`
- `ReadBlock()`, `WriteBlock()`
- `UpdateBitmap()`, `SetBit()`, `ClearBit()`
- `UpdateSuperblock()`

**Estimado:** 1500-2000 lineas para todas las operaciones

---

#### 6. Integrar Algoritmos de Fit con FDISK
**Archivo a crear/modificar:** `/internal/disk/manager.go`

```go
func (dm *FileManager) FdiskAdd(ctx context.Context, path, name string,
                                 size int64, typ, fit string) error {
    // 1. Abrir disco
    f, _ := os.OpenFile(path, os.O_RDWR, 0644)
    defer f.Close()

    // 2. Leer MBR
    mbr := ReadMBR(f)

    // 3. Validar tipo
    switch strings.ToUpper(typ) {
    case "P":
        return dm.createPrimary(f, mbr, name, size, fit)
    case "E":
        return dm.createExtended(f, mbr, name, size, fit)
    case "L":
        return dm.createLogical(f, mbr, name, size, fit)  // *** IMPLEMENTAR ***
    }
}

func (dm *FileManager) createPrimary(f *os.File, mbr *MBR,
                                      name string, size int64, fit string) error {
    // 1. Obtener espacios libres
    free := buildFreePrimaries(mbr)

    // 2. Aplicar algoritmo fit
    fitByte := parseFit(fit)
    seg, ok := pickByFit(free, size, fitByte)
    if !ok {
        return errors.New("no space")
    }

    // 3. Crear particion
    part := Partition{
        Status: PartStatusUsed,
        Type: PartTypePrimary,
        Fit: fitByte,
        Start: seg.start,
        Size: size,
    }
    copy(part.Name[:], name)

    // 4. Agregar a MBR
    for i := range mbr.Parts {
        if mbr.Parts[i].Status == PartStatusFree {
            mbr.Parts[i] = part
            break
        }
    }

    // 5. Escribir MBR
    WriteMBR(f, mbr)

    return nil
}
```

**Estimado:** 300-400 lineas

---

### PRIORIDAD MEDIA (Mejoras)

#### 7. Implementar Validaciones
- `ValidateSuperblock()`
- `ValidateMBR()`
- `ValidatePartition()`
- `ValidateEBR()`

#### 8. Implementar Metodos de Estructuras
Para `Partition`:
```go
func (p *Partition) GetName() string {
    return string(bytes.TrimRight(p.Name[:], "\x00"))
}

func (p *Partition) SetName(name string) {
    copy(p.Name[:], name)
}

func (p *Partition) IsExtended() bool {
    return p.Type == PartTypeExtended && p.Status == PartStatusUsed
}
```

**Estimado:** 200-300 lineas

---

## PARTE 6: EJEMPLO DE CODIGO CORREGIDO

### Ejemplo 1: Mkfs Completo

```go
// /internal/fs/ext2/ext2.go

package ext2

import (
    "bytes"
    "encoding/binary"
    "errors"
    "os"
    "time"
    "unsafe"
)

const (
    EXT2_MAGIC      = 0xEF53
    SUPERBLOCK_SIZE = int(unsafe.Sizeof(Superblock{}))
    INODE_SIZE      = int(unsafe.Sizeof(Inode{}))
    BLOCK_SIZE      = 64
)

type Superblock struct {
    FilesystemType   int32
    InodesCount      int32
    BlocksCount      int32
    FreeBlocksCount  int32
    FreeInodesCount  int32
    Mtime            int64
    Umtime           int64
    MntCount         int32
    Magic            int32
    InodeSize        int32
    BlockSize        int32
    FirstIno         int32
    FirstBlo         int32
    BmInodeStart     int32
    BmBlockStart     int32
    InodeStart       int32
    BlockStart       int32
}

type Inode struct {
    Uid   int32
    Gid   int32
    Size  int32
    Atime int64
    Ctime int64
    Mtime int64
    Block [15]int32
    Type  byte
    Perm  [3]byte
}

type Content struct {
    Name  [12]byte
    Inode int32
}

type FolderBlock struct {
    Content [4]Content
}

type FileBlock struct {
    Content [64]byte
}

func NewSuperblock(partSize int64) *Superblock {
    available := partSize - int64(SUPERBLOCK_SIZE)
    denominator := int64(1 + 3 + INODE_SIZE + 3*BLOCK_SIZE)
    n := available / denominator

    inodesCount := int32(n)
    blocksCount := int32(3 * n)

    now := time.Now().Unix()

    return &Superblock{
        FilesystemType:   2,
        InodesCount:      inodesCount,
        BlocksCount:      blocksCount,
        FreeBlocksCount:  blocksCount - 2,
        FreeInodesCount:  inodesCount - 2,
        Mtime:            now,
        Umtime:           0,
        MntCount:         1,
        Magic:            EXT2_MAGIC,
        InodeSize:        int32(INODE_SIZE),
        BlockSize:        int32(BLOCK_SIZE),
        FirstIno:         2,
        FirstBlo:         2,
        BmInodeStart:     int32(SUPERBLOCK_SIZE),
        BmBlockStart:     int32(SUPERBLOCK_SIZE) + inodesCount,
        InodeStart:       int32(SUPERBLOCK_SIZE) + inodesCount + blocksCount,
        BlockStart:       int32(SUPERBLOCK_SIZE) + inodesCount + blocksCount +
                         inodesCount*int32(INODE_SIZE),
    }
}

func (e *FS2) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
    // 1. Obtener referencia a particion montada
    ref, ok := e.state.GetRef(req.MountID)
    if !ok {
        return errors.New("partition not mounted")
    }

    // 2. Abrir disco
    f, err := os.OpenFile(ref.DiskPath, os.O_RDWR, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    // 3. Leer MBR y encontrar particion
    mbr, err := ReadMBR(f)
    if err != nil {
        return err
    }

    var partition *Partition
    for i := range mbr.Parts {
        if mbr.Parts[i].GetName() == ref.PartitionID {
            partition = &mbr.Parts[i]
            break
        }
    }
    if partition == nil {
        return errors.New("partition not found in MBR")
    }

    partStart := partition.Start
    partSize := partition.Size

    // 4. Crear superblock
    sb := NewSuperblock(partSize)

    // 5. Escribir superblock
    sbBytes, _ := SerializeSuperblock(sb)
    f.WriteAt(sbBytes, partStart)

    // 6. Inicializar bitmaps
    inodeBitmap := make([]byte, (sb.InodesCount+7)/8)
    inodeBitmap[0] = 0x03  // bits 0 y 1 ocupados
    f.WriteAt(inodeBitmap, partStart+int64(sb.BmInodeStart))

    blockBitmap := make([]byte, (sb.BlocksCount+7)/8)
    blockBitmap[0] = 0x03  // bits 0 y 1 ocupados
    f.WriteAt(blockBitmap, partStart+int64(sb.BmBlockStart))

    // 7. Crear inodo raiz
    rootInode := &Inode{
        Uid: 1, Gid: 1, Size: 0,
        Atime: time.Now().Unix(),
        Ctime: time.Now().Unix(),
        Mtime: time.Now().Unix(),
        Type: 0,  // carpeta
        Perm: [3]byte{'7', '5', '5'},
    }
    for i := range rootInode.Block {
        rootInode.Block[i] = -1
    }
    rootInode.Block[0] = 0

    rootInodeBytes, _ := SerializeInode(rootInode)
    f.WriteAt(rootInodeBytes, partStart+int64(sb.InodeStart))

    // 8. Crear bloque carpeta raiz
    rootBlock := &FolderBlock{}
    rootBlock.Content[0] = Content{Inode: 0}
    copy(rootBlock.Content[0].Name[:], ".")
    rootBlock.Content[1] = Content{Inode: 0}
    copy(rootBlock.Content[1].Name[:], "..")
    rootBlock.Content[2] = Content{Inode: 1}
    copy(rootBlock.Content[2].Name[:], "users.txt")

    rootBlockBytes, _ := SerializeFolderBlock(rootBlock)
    f.WriteAt(rootBlockBytes, partStart+int64(sb.BlockStart))

    // 9. Crear inodo users.txt
    usersContent := "1,G,root\n1,U,root,root,123\n"
    usersInode := &Inode{
        Uid: 1, Gid: 1, Size: int32(len(usersContent)),
        Atime: time.Now().Unix(),
        Ctime: time.Now().Unix(),
        Mtime: time.Now().Unix(),
        Type: 1,  // archivo
        Perm: [3]byte{'6', '6', '4'},
    }
    for i := range usersInode.Block {
        usersInode.Block[i] = -1
    }
    usersInode.Block[0] = 1

    usersInodeBytes, _ := SerializeInode(usersInode)
    f.WriteAt(usersInodeBytes, partStart+int64(sb.InodeStart)+int64(INODE_SIZE))

    // 10. Crear bloque archivo users.txt
    usersBlock := &FileBlock{}
    copy(usersBlock.Content[:], usersContent)

    usersBlockBytes, _ := SerializeFileBlock(usersBlock)
    f.WriteAt(usersBlockBytes, partStart+int64(sb.BlockStart)+int64(BLOCK_SIZE))

    e.logger.Printf("Formateo EXT2 completado: %d inodos, %d bloques",
                    sb.InodesCount, sb.BlocksCount)

    return nil
}

func SerializeSuperblock(sb *Superblock) ([]byte, error) {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, sb)
    return buf.Bytes(), nil
}

func SerializeInode(inode *Inode) ([]byte, error) {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, inode)
    return buf.Bytes(), nil
}

func SerializeFolderBlock(fb *FolderBlock) ([]byte, error) {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, fb)
    return buf.Bytes(), nil
}

func SerializeFileBlock(fb *FileBlock) ([]byte, error) {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, fb)
    return buf.Bytes(), nil
}
```

---

## PARTE 7: RESUMEN DE ACCIONES CORRECTIVAS PRIORIZADAS

### CRITICAS (Deben hacerse PRIMERO)

1. **Definir estructuras EXT2** (Superblock, Inode, Blocks)
   - Tiempo estimado: 2-3 horas
   - Impacto: Sin esto, mkfs no puede funcionar
   - Archivos: crear `/internal/fs/ext2/superblock.go`, `inode.go`, `blocks.go`

2. **Implementar persistencia en Mkfs**
   - Tiempo estimado: 3-4 horas
   - Impacto: Sin esto, formateo solo simula
   - Archivo: modificar `/internal/fs/ext2/ext2.go`

3. **Implementar operaciones con EBR**
   - Tiempo estimado: 2-3 horas
   - Impacto: Sin esto, particiones logicas no funcionan
   - Archivo: crear `/internal/disk/ebr.go`

4. **Corregir estructura Partition**
   - Tiempo estimado: 1 hora
   - Impacto: Sin esto, incompatibilidad binaria con P1
   - Archivo: modificar `/internal/disk/mbr.go`

**Total CRITICO: 8-11 horas**

---

### ALTAS (Funcionalidad core)

5. **Implementar operaciones de archivos** (mkdir, mkfile, remove, etc.)
   - Tiempo estimado: 8-10 horas
   - Impacto: Sin esto, sistema de archivos es read-only
   - Archivo: modificar `/internal/fs/ext2/ext2.go`

6. **Integrar algoritmos de fit con fdisk**
   - Tiempo estimado: 3-4 horas
   - Impacto: Sin esto, asignacion de particiones ineficiente
   - Archivo: modificar `/internal/disk/manager.go`

**Total ALTO: 11-14 horas**

---

### MEDIAS (Mejoras)

7. **Implementar validaciones**
   - Tiempo estimado: 2-3 horas
   - Archivos: varios

8. **Implementar metodos de estructuras**
   - Tiempo estimado: 2-3 horas
   - Archivos: `/internal/disk/mbr.go`, etc.

**Total MEDIO: 4-6 horas**

---

## TIEMPO TOTAL ESTIMADO: 23-31 horas

---

## CONCLUSIONES FINALES

### Estado Actual
El Proyecto 2 es una **reimplementacion arquitectonica** (patron clean architecture, separacion de capas) pero **sin implementacion funcional completa**. Tiene:
- Estructura de codigo organizada
- Definiciones de tipos basicos
- Algoritmos de fit correctos (pero no integrados)
- Servidor HTTP funcionando
- Parser de comandos

Pero **NO tiene**:
- Estructuras EXT2 definidas
- Persistencia de datos a disco
- Soporte para particiones logicas
- Validaciones completas
- Compatibilidad binaria con P1

### Recomendacion Principal
**NO intentar** hacer compatible binariamente con P1. En su lugar:

1. Completar la implementacion funcional de P2 siguiendo la logica de P1
2. Mantener la arquitectura limpia de P2
3. Copiar la logica de negocio de P1 (especialmente EXT2)
4. Implementar las funciones marcadas como TODO
5. Reemplazar logs de "no persistente" con implementaciones reales

### Siguientes Pasos
1. Empezar por las tareas CRITICAS en orden
2. Usar P1 como referencia para la logica, no para la arquitectura
3. Probar cada componente individualmente antes de integrar
4. Validar con discos de prueba pequeños (1-5 MB)

---

**FIN DEL REPORTE**
