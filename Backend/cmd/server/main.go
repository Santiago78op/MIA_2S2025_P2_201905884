package main

import (
	"Backend/command/disk"
	"Backend/command/fs"
	"Backend/command/reports"
	"Backend/command/runner"
	"Backend/command/users"
	"Backend/config"
	"Backend/controllers"
	"Backend/pkg/logger"
	"Backend/router"
	"Backend/storage/adapters"
	"Backend/storage/diskio"
	"Backend/storage/graphviz"
	"Backend/storage/mounts"
)

// arranque, DI wiring, router, CORS
func main() {
	// puerto, base paths, etc.
	cfg := config.Load()
	log := logger.New(cfg.Debug)

	log.Info("Iniciando servidor MIA en puerto %s", cfg.Port)

	// === Capa de infraestructura (storage) ===
	diskRepo := diskio.NewFileDiskRepository()
	mStore := mounts.NewState()
	fsRepoBase := diskio.NewFileFsRepository(adapters.NewPortsMountStore(mStore), diskRepo)
	gv := graphviz.New()

	// === Capa de adaptadores (adapta storage -> command interfaces) ===
	diskRepoAdapter := adapters.NewDiskAdapter(diskRepo)
	mountStoreAdapter := adapters.NewMountAdapter(mStore)
	fsRepoAdapter := adapters.NewFsAdapter(fsRepoBase)
	usersRepoAdapter := adapters.NewUsersAdapter(fsRepoBase)
	sessionAdapter := adapters.NewSessionAdapterFromMemory()
	reportGenerator := adapters.NewReportGenerator(gv, fsRepoBase, diskRepo, adapters.NewPortsMountStore(mStore))

	// === Capa de aplicación (servicios de dominio) ===
	diskSvc := disk.NewDiskService(diskRepoAdapter, mountStoreAdapter, cfg.CarnetLastTwo)
	fsSvc := fs.NewFsService(fsRepoAdapter, sessionAdapter)
	usersSvc := users.NewUserService(usersRepoAdapter, sessionAdapter)
	reportSvc := reports.NewReportService(reportGenerator)

	// === Runner central (ejecuta comandos y scripts) ===
	cmdRunner := runner.New(diskSvc, fsSvc, usersSvc, reportSvc, log)

	// === Capa de presentación (controllers HTTP) ===
	cs := controllers.NewCommandsController(cmdRunner)
	ss := controllers.NewScriptController(&scriptRunnerAdapter{cmdRunner}) // Adaptador para ScriptRunner
	rs := controllers.NewReportsController(reportSvc, cfg.ReportsPath)
	vc := controllers.NewViewerController() // NUEVO P2: ViewerController (stub por ahora)

	// === Router y servidor HTTP ===
	r := router.SetupRouter(cfg, cs, ss, rs, vc)

	log.Info("Servidor listo. Escuchando en http://localhost:%s", cfg.Port)
	log.Info("Healthcheck: http://localhost:%s/health", cfg.Port)
	log.Info("API: http://localhost:%s/api/*", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Error iniciando servidor: %v", err)
	}
}

// scriptRunnerAdapter adapta Runner.RunScript para cumplir con controllers.ScriptRunner
type scriptRunnerAdapter struct {
	runner *runner.Runner
}

func (s *scriptRunnerAdapter) Run(script string, stopOnError bool) (string, error) {
	return s.runner.RunScript(script, stopOnError)
}
