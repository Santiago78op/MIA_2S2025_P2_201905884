#!/bin/bash

# Agregar lógica para usar partitionId en los métodos de usuarios

# Para cada método en users/service.go que use 'id' como parámetro,
# agregar la línea: if id == "" { id = partitionId }

# Lista de métodos: Rmgrp, Mkusr, Rmusr, Chgrp
# (Mkgrp ya tiene la lógica)

# Rmgrp
sed -i '/func (u \*UserService) Rmgrp/,/if err := u.repo.Rmgrp/ {
    /solo el usuario root puede eliminar grupos/a\
\
	// Si no se proporciona id, usar el de la sesión actual\
	if id == "" {\
		id = partitionId\
	}
}' command/users/service.go

# Mkusr
sed -i '/func (u \*UserService) Mkusr/,/if err := u.repo.Mkusr/ {
    /solo el usuario root puede crear usuarios/a\
\
	// Si no se proporciona id, usar el de la sesión actual\
	if id == "" {\
		id = partitionId\
	}
}' command/users/service.go

# Rmusr
sed -i '/func (u \*UserService) Rmusr/,/if err := u.repo.Rmusr/ {
    /solo el usuario root puede eliminar usuarios/a\
\
	// Si no se proporciona id, usar el de la sesión actual\
	if id == "" {\
		id = partitionId\
	}
}' command/users/service.go

# Chgrp
sed -i '/func (u \*UserService) Chgrp/,/if err := u.repo.Chgrp/ {
    /solo el usuario root puede cambiar grupos de usuarios/a\
\
	// Si no se proporciona id, usar el de la sesión actual\
	if id == "" {\
		id = partitionId\
	}
}' command/users/service.go

echo "Lógica de partitionId agregada"
