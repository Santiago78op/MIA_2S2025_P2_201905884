# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

---

## Project Overview

This is an **EXT2/EXT3 filesystem simulator** for academic purposes (MIA 2S 2025 - Proyecto 2). The project implements a complete file system management layer with:
- Virtual disk creation and partitioning (MBR/EBR)
- EXT2 and EXT3 file system formatting with journaling
- Full CRUD operations for files and directories
- User and permission management
- Journaling support for EXT3 recovery
- REST API for web-based file system visualization

**Architecture:** The backend is written in Go with hexagonal architecture. The frontend is a React+Vite application for file system visualization and management.

---

## Build & Run Commands

### Backend (Go)

```bash
# From project root, navigate to Backend
cd Backend

# Build the server
go build -o bin/server cmd/server/main.go

# Run the server (default port: 8080)
./bin/server

# Run with custom configuration via environment variables
PORT=8080 DISKS_PATH=./Discos REPORTS_PATH=./Reports DEBUG=true ./bin/server

# Run in background
nohup ./bin/server > logs/server.log 2>&1 &
echo $! > server.pid

# Stop background server
kill $(cat server.pid)
```

### Frontend (React+Vite)

```bash
# From project root, navigate to Frontend
cd Frontend

# Install dependencies (first time only)
npm install

# Development server (default port: 5173)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Full Installation

```bash
# From project root
./install.sh
```

This script installs Go 1.23.3, Node.js v20.x, builds the backend, and installs frontend dependencies.

---

## Testing

### Run All Tests

```bash
cd Backend
go test ./tests/...
```

### Run a Single Test File

```bash
cd Backend
go test ./tests/disk_service_test.go
go test ./tests/fs_service_test.go
go test ./tests/mount_id_test.go
```

### Run Script-Based E2E Tests

```bash
cd Backend
go run tests/runner.go <script_file.smia>

# Example
go run tests/runner.go test_e2e.smia
go run tests/runner.go ../test_loss_recovery.smia
```

---

## Architecture

### Hexagonal Architecture (Ports & Adapters)

The backend follows a clean hexagonal architecture with clear separation of concerns:

```
Domain Layer (core/)        → Business entities, models, and contracts (ports)
Application Layer (command/) → Use cases and command handlers
Adapter Layer (storage/)    → Infrastructure implementations
Presentation Layer (controllers/) → HTTP handlers
```

**Key Principle:** All dependencies point inward. The domain (core) never depends on infrastructure or presentation layers.

### Core Layers

#### 1. **Domain Layer (`core/`)**
- **`core/models/`**: Data structures for file systems (EXT2, EXT3, MBR, Journal, Inodes, Blocks)
  - `ext2.go`: SuperBlock, Inode, DirEntry, FileBlock, PointerBlock
  - `journal.go`: Journal entry structures for EXT3 recovery
  - `mbr.go`: MBR, EBR, Partition structures
  - `types.go`: Constants like `FSMagic = 0xEF53`, block sizes, pointers per block
  
- **`core/ports/`**: Interfaces defining contracts (repository patterns)
  - `disk_repository.go`: Disk I/O operations
  - `fs_repository.go`: File system operations (mkfs, mkdir, mkfile, chmod, etc.)
  - `mount_store.go`: Mount management
  - `session_store.go`: User session tracking

#### 2. **Application Layer (`command/`)**
Each module has its own parser and service:
- **`command/disk/`**: MKDISK, FDISK, MOUNT, UNMOUNT
- **`command/fs/`**: MKFS, MKDIR, MKFILE, CAT, REMOVE, EDIT, RENAME, COPY, MOVE, FIND
- **`command/users/`**: LOGIN, LOGOUT, MKGRP, RMGRP, MKUSR, RMUSR
- **`command/reports/`**: REP (generates Graphviz reports)
- **`command/runner/`**: Orchestrates commands and executes `.smia` scripts

#### 3. **Infrastructure Layer (`storage/`)**
- **`storage/diskio/`**: Low-level binary I/O for disks and file systems
  - `diskio.go`: Read/write MBR, EBR, partitions
  - `file_repo.go`: Implements `FsRepository` port (EXT2/EXT3 operations)
  
- **`storage/journal/`**: Journal persistence for EXT3
  - `file_journal.go`: Write-ahead logging for recovery

- **`storage/mounts/`**: Global mount state (`state.go`)
- **`storage/session/`**: In-memory user session store (`memory.go`)
- **`storage/graphviz/`**: Report generation with Graphviz (`dot.go`)
- **`storage/adapters/`**: Adapters that bridge ports to storage implementations

#### 4. **Presentation Layer (`controllers/`)**
- `commands_controller.go`: Single command execution via `/api/commands`
- `script_controller.go`: Script execution via `/api/script`
- `reports_controller.go`: Report generation and listing
- `viewer_controller.go`: **P2 Addition** - File system browser, login, journal viewer

---

## Key Architectural Patterns

### 1. **Dependency Injection**
All dependencies are wired in `cmd/server/main.go`. Each layer receives its dependencies through constructors, making testing and swapping implementations easy.

### 2. **Repository Pattern**
All data access goes through repository interfaces defined in `core/ports/`. This allows mocking for tests and potential future database backends.

### 3. **Command Pattern**
Each operation (mkdisk, fdisk, login, etc.) is a discrete command with its own parser and service handler.

### 4. **State Management**
- **Mount state**: Shared via `storage/mounts/state.go`
- **Session state**: Shared via `storage/session/memory.go`
- Both are thread-safe and accessed through port interfaces

### 5. **Journaling (EXT3)**
All mutating operations (MKFILE, MKDIR, REMOVE, EDIT, RENAME, COPY, MOVE, CHMOD, CHOWN) write a journal entry **before** executing. This enables recovery after `LOSS` command via `RECOVERY`.

---

## File System Implementation Details

### EXT2 vs EXT3 Calculation

**EXT2 Layout:**
```
partition_size = sizeof(SuperBlock) + n + 3n + n*sizeof(Inode) + 3n*sizeof(Block)
n = floor((partition_size - sizeof(SuperBlock)) / (1 + 3 + sizeof(Inode) + 3*sizeof(Block)))
```

**EXT3 Layout (with Journal):**
```
partition_size = sizeof(SuperBlock) + 50*sizeof(JournalEntry) + n + 3n + n*sizeof(Inode) + 3n*sizeof(Block)
n = floor((partition_size - sizeof(SuperBlock) - journal_size) / (1 + 3 + sizeof(Inode) + 3*sizeof(Block)))
```

See `Backend/utils/calc.go` for `CalcN()` and `CalcNExt3()` implementations.

### Inode Structure
- **15 block pointers**: 12 direct, 3 indirect levels
- **Permissions**: UGO format (3 bytes, octal: 0-7 per field)
- **Type**: 0 = directory, 1 = file
- **UID/GID**: Stored as int32

### Block Types
- **Directory blocks**: 4 entries per block (20-byte name + int32 inode index)
- **File blocks**: 64 bytes of content per block
- **Pointer blocks**: 16 pointers (int32) per block

### Journal Entry Format (EXT3)
See `Backend/core/models/journal.go`:
- **Structure size**: 600 bytes fixed
- **Fields**: Operation (10 bytes), Path (32 bytes), Content (64 bytes), Date (float64)
- **Max entries**: 50 per partition

---

## API Endpoints

### Core Operations
- `POST /api/commands` - Execute single command
- `POST /api/script` - Execute `.smia` script
- `POST /api/reports` - Generate report (mbr, disk, inode, block, tree, etc.)
- `GET /api/reports/list` - List all generated reports
- `GET /reports/static/<file>` - Serve static report files

### P2 Additions (Viewer & Auth)
- `POST /api/auth/login` - Authenticate user on mounted partition
- `POST /api/auth/logout` - Close session
- `GET /api/disks` - List available `.mia` disk files
- `GET /api/disks/partitions?path=<disk>` - List partitions in disk
- `GET /api/fs/:id/tree?path=<path>` - Directory listing
- `GET /api/fs/:id/file?path=<path>` - File content
- `GET /api/journal/:id` - Journal entries (EXT3 only)
- `GET /api/journal/:id/table` - Journal table view

---

## Configuration

**Environment Variables** (see `Backend/config/config.go`):
- `PORT`: HTTP server port (default: 8080)
- `DISKS_PATH`: Directory for `.mia` disk files (default: `./Discos`)
- `REPORTS_PATH`: Directory for generated reports (default: `./Reports`)
- `CARNET_LAST_TWO`: Last 2 digits of student ID for mount IDs (default: "84")
- `DEBUG`: Enable debug logging (default: false)

**Example `.env` file:**
```bash
PORT=8080
DISKS_PATH=/home/julian/Documents/MIA_2S2025_P2_201905884/Backend/Discos
REPORTS_PATH=/home/julian/Documents/MIA_2S2025_P2_201905884/Backend/Reports
CARNET_LAST_TWO=84
DEBUG=true
```

---

## Important Implementation Notes

### Mount ID Format
Mount IDs follow the pattern: `<carnet_last_two><partition_letter><correlative>`
- Example: `841A`, `842B`, `843C`
- Correlative resets to 0 on UNMOUNT

### Permission Checking
All file operations check permissions:
- **REMOVE**: Requires write permission on file/directory
- **EDIT**: Requires read+write on file
- **RENAME**: Requires write on parent directory
- **COPY**: Requires read on source, write on destination
- **MOVE**: Requires write on both source and destination parents
- **CHMOD/CHOWN**: Only root or file owner can modify

### Journaling Behavior
- **EXT2**: No journaling; operations applied directly
- **EXT3**: Every mutation writes to journal **before** execution (write-ahead logging)
- **RECOVERY**: Replays all journal entries in chronological order
- **LOSS**: Simulates catastrophic failure by wiping bitmaps, inodes, and blocks

---

## Common Development Workflows

### Adding a New Command

1. Create parser in appropriate `command/` module (e.g., `command/fs/parser.go`)
2. Add service method in `command/fs/service.go`
3. Update `command/runner/runner.go` to route the command
4. If needed, add repository method to `core/ports/` interface
5. Implement in `storage/diskio/file_repo.go`
6. Add tests in `tests/`

### Modifying File System Layout

1. Update calculation in `utils/calc.go`
2. Modify `core/models/ext2.go` or `ext3.go` structures
3. Update serialization in `storage/diskio/`
4. Run tests: `go test ./tests/...`

### Adding a New API Endpoint

1. Add method to appropriate controller in `controllers/`
2. Register route in `router/router.go`
3. Update CORS if needed (already set to `*`)
4. Update frontend `Frontend/src/lib/api.js` if needed

---

## Frontend Structure

**Framework:** React 18 + Vite 5

**Key Files:**
- `src/main.jsx`: Entry point
- `src/App.jsx`: Root component with routing
- `src/lib/api.js`: API client wrapper
- `src/pages/`: Page components
- `src/components/`: Reusable UI components

**Proxy Configuration:** `vite.config.js` proxies `/api`, `/health`, `/reports` to `http://localhost:8080`

---

## Utilities

### `utils/calc.go`
- `CalcN()`: Calculate number of inodes/blocks for EXT2
- `CalcNExt3()`: Calculate number for EXT3 with journal
- `ComputeLayout()`: Generate SuperBlock with absolute offsets

### `utils/bytes.go`
- Binary serialization helpers for disk structures

### `utils/permissions.go`
- UGO permission checking and manipulation

### `utils/smia.go`
- Parser for `.smia` script files (comment stripping, line processing)

---

## Testing Scripts

**Script Format:** `.smia` files contain one command per line. Lines starting with `#` are comments.

**Location:** 
- `Backend/test_e2e.smia`: End-to-end test script
- `test_loss_recovery.smia`: LOSS and RECOVERY testing

**Run via:**
```bash
cd Backend
go run tests/runner.go test_e2e.smia
```

---

## Troubleshooting

### "partition not found" errors
- Ensure the partition is mounted with `MOUNT`
- Check mount ID format: `<carnet><partition><correlative>`

### "permission denied" errors
- Verify user is logged in: `LOGIN -user=... -pass=... -id=...`
- Check file permissions with `REP -name=ls -id=... -path=/`

### Journal errors on EXT2 partitions
- Journal operations only work on EXT3 (`-fs=3fs`)
- RECOVERY and journal viewer require EXT3

### Port already in use
```bash
# Kill existing process
lsof -ti:8080 | xargs kill -9
lsof -ti:5173 | xargs kill -9
```

---

## Project Structure Summary

```
MIA_2S2025_P2_201905884/
├── Backend/
│   ├── cmd/server/main.go          # Entry point, DI wiring
│   ├── command/                     # Use cases (disk, fs, users, reports, runner)
│   ├── core/                        # Domain (models, ports)
│   ├── storage/                     # Infrastructure (diskio, journal, adapters)
│   ├── controllers/                 # HTTP handlers
│   ├── router/                      # Route configuration
│   ├── utils/                       # Shared utilities
│   ├── config/                      # Configuration loading
│   ├── tests/                       # Test files and runner
│   ├── Discos/                      # Generated .mia disk files
│   └── Reports/                     # Generated reports (Graphviz, etc.)
│
├── Frontend/
│   ├── src/
│   │   ├── main.jsx                 # React entry point
│   │   ├── App.jsx                  # Root component
│   │   ├── lib/api.js               # API client
│   │   ├── pages/                   # Page components
│   │   └── components/              # UI components
│   └── vite.config.js               # Vite configuration with proxy
│
├── install.sh                       # Automated installation script
└── WARP.md                          # This file
```

---

## Additional Documentation

- `README`: Project planning and architecture overview
- `INSTALL.md`: Detailed installation instructions
- `Backend/IMPLEMENTATION_CHECKLIST.md`: Feature implementation status
- `Backend/PROJECT_STATUS_CHECKLIST.md`: Project completion tracking
- Various `*_IMPLEMENTATION.md` files in Backend/: Detailed design docs per feature
