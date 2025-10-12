package main

import (
	"Backend/config"
)

// arranque, DI wiring, router, CORS
func main() {
	// puerto, base paths, etc.
	cfg := config.Load()
	logger := logger.New()
	// wiring (DI): repos + stores + services + controllers
	diskRepo := storage.NewFileDiskRepository(logger)
	fsRepo := storage.NewFileFSRepository(logger)
	mStore := mounts.NewState()
	sess := session.NewMemory()
	gv := graphviz.New()

	dSvc := command.NewDiskService(diskRepo, mStore, logger)
	fSvc := command.NewFsService(fsRepo, sess, logger)
	uSvc := command.NewUserService(fsRepo, sess, logger)
	rSvc := command.NewReportService(gv, logger)

	cs := controllers.NewCommands(dSvc, fSvc, uSvc, rSvc)
	ss := controllers.NewScript(dSvc, fSvc, uSvc, rSvc)
	rs := controllers.NewReports(rSvc)

	r := router.NewRouter(cs, ss, rs)
	r.Run(":" + cfg.Port)
}
