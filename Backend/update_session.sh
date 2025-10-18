#!/bin/bash

# Script para actualizar las llamadas a sess.Current() agregando el nuevo parámetro partitionId

# Actualizar users/service.go
sed -i 's/logged, _, _, _, isRoot := u\.sess\.Current()/logged, _, _, _, isRoot, partitionId := u.sess.Current()/g' command/users/service.go

# Agregar lógica para usar partitionId si id está vacío en cada método
# Esto se hará manualmente ya que requiere lógica específica

# Actualizar fs/service.go
sed -i 's/logged, _, uid, gid, _ := s\.sess\.Current()/logged, _, uid, gid, _, partitionId := s.sess.Current()/g' command/fs/service.go

echo "Actualización completada"
