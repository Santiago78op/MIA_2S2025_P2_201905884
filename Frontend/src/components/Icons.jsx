// Componente de iconos SVG reutilizable
// Uso: <Icon name="disk" size={24} color="var(--neon)" />

export default function Icon({ name, size = 24, color = 'currentColor', className = '' }) {
  const icons = {
    // Disco duro
    disk: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="6" width="20" height="12" rx="2" />
        <line x1="2" y1="12" x2="22" y2="12" />
        <circle cx="6" cy="15" r="0.5" fill={color} />
        <circle cx="9" cy="15" r="0.5" fill={color} />
        <path d="M6 9h4" />
        <path d="M14 9h4" />
      </svg>
    ),

    // Partición
    partition: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="7" width="18" height="10" rx="1" />
        <line x1="3" y1="12" x2="21" y2="12" />
        <line x1="10" y1="7" x2="10" y2="17" />
        <line x1="15" y1="7" x2="15" y2="17" />
        <circle cx="6.5" cy="14.5" r="0.5" fill={color} />
      </svg>
    ),

    // Carpeta cerrada
    folder: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 7v10a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-7l-2-2H5a2 2 0 0 0-2 2z" fill={`${color}22`} />
      </svg>
    ),

    // Carpeta abierta
    folderOpen: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" fill={`${color}22`} />
        <path d="M2 13l2.586 2.586A2 2 0 0 0 6 16.414V19a2 2 0 0 0 2 2h11" />
      </svg>
    ),

    // Archivo de texto
    file: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill={`${color}11`} />
        <polyline points="14 2 14 8 20 8" />
        <line x1="9" y1="13" x2="15" y2="13" />
        <line x1="9" y1="17" x2="15" y2="17" />
      </svg>
    ),

    // Archivo binario/ejecutable
    fileBinary: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill={`${color}11`} />
        <polyline points="14 2 14 8 20 8" />
        <text x="8" y="16" fontSize="8" fill={color} fontFamily="monospace">10</text>
        <text x="8" y="20" fontSize="8" fill={color} fontFamily="monospace">01</text>
      </svg>
    ),

    // Archivo de código
    fileCode: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill={`${color}11`} />
        <polyline points="14 2 14 8 20 8" />
        <polyline points="10 13 7 16 10 19" />
        <polyline points="14 13 17 16 14 19" />
      </svg>
    ),

    // Sistema (root)
    system: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" fill={`${color}11`} />
        <circle cx="12" cy="12" r="6" />
        <circle cx="12" cy="12" r="2" fill={color} />
        <line x1="12" y1="2" x2="12" y2="6" />
        <line x1="12" y1="18" x2="12" y2="22" />
        <line x1="4.93" y1="4.93" x2="7.76" y2="7.76" />
        <line x1="16.24" y1="16.24" x2="19.07" y2="19.07" />
        <line x1="2" y1="12" x2="6" y2="12" />
        <line x1="18" y1="12" x2="22" y2="12" />
        <line x1="4.93" y1="19.07" x2="7.76" y2="16.24" />
        <line x1="16.24" y1="7.76" x2="19.07" y2="4.93" />
      </svg>
    ),

    // Usuario
    user: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" fill={`${color}22`} />
      </svg>
    ),

    // Grupo
    group: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" fill={`${color}22`} />
        <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    ),

    // Candado (permisos)
    lock: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="5" y="11" width="14" height="10" rx="2" ry="2" fill={`${color}11`} />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
        <circle cx="12" cy="16" r="1" fill={color} />
      </svg>
    ),

    // Candado abierto
    unlock: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="5" y="11" width="14" height="10" rx="2" ry="2" fill={`${color}11`} />
        <path d="M7 11V7a5 5 0 0 1 9.9-1" />
        <circle cx="12" cy="16" r="1" fill={color} />
      </svg>
    ),

    // Reloj (timestamp)
    clock: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" fill={`${color}11`} />
        <polyline points="12 6 12 12 16 14" />
      </svg>
    ),

    // Tamaño/peso
    size: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" fill={`${color}11`} />
        <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
        <line x1="12" y1="22.08" x2="12" y2="12" />
      </svg>
    ),

    // Actualizar/refrescar
    refresh: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="23 4 23 10 17 10" />
        <polyline points="1 20 1 14 7 14" />
        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
      </svg>
    ),

    // Checkmark
    check: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="20 6 9 17 4 12" />
      </svg>
    ),

    // Error/X
    error: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" fill={`${color}11`} />
        <line x1="15" y1="9" x2="9" y2="15" />
        <line x1="9" y1="9" x2="15" y2="15" />
      </svg>
    ),

    // Info
    info: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" fill={`${color}11`} />
        <line x1="12" y1="16" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12.01" y2="8" />
      </svg>
    ),

    // Servidor/database
    server: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="2" width="20" height="8" rx="2" ry="2" fill={`${color}11`} />
        <rect x="2" y="14" width="20" height="8" rx="2" ry="2" fill={`${color}11`} />
        <line x1="6" y1="6" x2="6.01" y2="6" strokeWidth="3" />
        <line x1="6" y1="18" x2="6.01" y2="18" strokeWidth="3" />
      </svg>
    ),

    // Montaje
    mount: (
      <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 2v6" />
        <path d="M12 18v4" />
        <circle cx="12" cy="12" r="4" fill={`${color}22`} />
        <path d="M4.93 4.93l4.24 4.24" />
        <path d="M14.83 14.83l4.24 4.24" />
        <path d="M2 12h6" />
        <path d="M16 12h6" />
        <path d="M4.93 19.07l4.24-4.24" />
        <path d="M14.83 9.17l4.24-4.24" />
      </svg>
    ),
  }

  return (
    <div
      className={`icon ${className}`}
      style={{
        width: size,
        height: size,
        display: 'inline-block',
        verticalAlign: 'middle',
        flexShrink: 0
      }}
    >
      {icons[name] || icons.file}
    </div>
  )
}

// Componentes específicos para uso más fácil
export const DiskIcon = (props) => <Icon name="disk" {...props} />
export const PartitionIcon = (props) => <Icon name="partition" {...props} />
export const FolderIcon = (props) => <Icon name="folder" {...props} />
export const FolderOpenIcon = (props) => <Icon name="folderOpen" {...props} />
export const FileIcon = (props) => <Icon name="file" {...props} />
export const FileBinaryIcon = (props) => <Icon name="fileBinary" {...props} />
export const FileCodeIcon = (props) => <Icon name="fileCode" {...props} />
export const SystemIcon = (props) => <Icon name="system" {...props} />
export const UserIcon = (props) => <Icon name="user" {...props} />
export const GroupIcon = (props) => <Icon name="group" {...props} />
export const LockIcon = (props) => <Icon name="lock" {...props} />
export const UnlockIcon = (props) => <Icon name="unlock" {...props} />
export const ClockIcon = (props) => <Icon name="clock" {...props} />
export const SizeIcon = (props) => <Icon name="size" {...props} />
export const RefreshIcon = (props) => <Icon name="refresh" {...props} />
export const CheckIcon = (props) => <Icon name="check" {...props} />
export const ErrorIcon = (props) => <Icon name="error" {...props} />
export const InfoIcon = (props) => <Icon name="info" {...props} />
export const ServerIcon = (props) => <Icon name="server" {...props} />
export const MountIcon = (props) => <Icon name="mount" {...props} />
