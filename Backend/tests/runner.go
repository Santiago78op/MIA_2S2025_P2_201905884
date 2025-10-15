package main

import (
	"Backend/command/disk"
	"Backend/command/fs"
	"Backend/command/reports"
	"Backend/command/runner"
	"Backend/command/users"
	"Backend/config"
	"Backend/pkg/logger"
	"Backend/storage/adapters"
	"Backend/storage/diskio"
	"Backend/storage/graphviz"
	"Backend/storage/mounts"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run run_test.go <script_file>")
		os.Exit(1)
	}

	scriptPath := os.Args[1]

	// Read script file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("Error reading script: %v\n", err)
		os.Exit(1)
	}

	// Setup dependencies (same as main.go)
	cfg := config.Load()
	log := logger.New(true) // Enable debug mode

	// === Capa de infraestructura (storage) ===
	diskRepo := diskio.NewFileDiskRepository()
	mStore := mounts.NewState()
	fsRepoBase := diskio.NewFileFsRepository(adapters.NewPortsMountStore(mStore), diskRepo)
	gv := graphviz.New()

	// === Capa de adaptadores ===
	diskRepoAdapter := adapters.NewDiskAdapter(diskRepo)
	mountStoreAdapter := adapters.NewMountAdapter(mStore)
	fsRepoAdapter := adapters.NewFsAdapter(fsRepoBase)
	usersRepoAdapter := adapters.NewUsersAdapter(fsRepoBase)
	sessionAdapter := adapters.NewSessionAdapterFromMemory()
	reportGenerator := adapters.NewReportGenerator(gv, fsRepoBase, diskRepo, adapters.NewPortsMountStore(mStore))

	// === Capa de aplicación ===
	diskSvc := disk.NewDiskService(diskRepoAdapter, mountStoreAdapter, cfg.CarnetLastTwo)
	fsSvc := fs.NewFsService(fsRepoAdapter, sessionAdapter)
	usersSvc := users.NewUserService(usersRepoAdapter, sessionAdapter)
	reportSvc := reports.NewReportService(reportGenerator)

	// === Runner central ===
	cmdRunner := runner.New(diskSvc, fsSvc, usersSvc, reportSvc, log)

	// Run script
	fmt.Println("=== Running test script ===")
	fmt.Println()

	result, err := cmdRunner.RunScript(string(content), false) // Don't stop on error
	if err != nil {
		fmt.Printf("Script execution error: %v\n", err)
	}

	fmt.Println(result)
	fmt.Println()
	fmt.Println("=== Script execution completed ===")
}
