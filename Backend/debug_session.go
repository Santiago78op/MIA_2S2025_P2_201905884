package main

import (
	"Backend/storage/adapters"
	"Backend/storage/session"
	"fmt"
)

func main() {
	// Crear sesión
	sess := session.NewMemory()

	// Verificar estado inicial
	logged, user, uid, gid, isRoot, partID := sess.Current()
	fmt.Printf("Estado inicial: logged=%v, user=%s, uid=%d, gid=%d, isRoot=%v, partID=%s\n",
		logged, user, uid, gid, isRoot, partID)

	// Login
	sess.Login("root", 1, 1, "841A")
	fmt.Println("Login ejecutado con: user=root, uid=1, gid=1, partID=841A")

	// Verificar estado después del login
	logged, user, uid, gid, isRoot, partID = sess.Current()
	fmt.Printf("Estado después de login: logged=%v, user=%s, uid=%d, gid=%d, isRoot=%v, partID=%s\n",
		logged, user, uid, gid, isRoot, partID)

	// Probar con adaptador
	adapter := adapters.NewSessionAdapter(sess)
	logged, user, uid, gid, isRoot, partID = adapter.Current()
	fmt.Printf("Estado vía adaptador: logged=%v, user=%s, uid=%d, gid=%d, isRoot=%v, partID=%s\n",
		logged, user, uid, gid, isRoot, partID)
}
