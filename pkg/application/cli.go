package application

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/config"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/controller"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/repository"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/server"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/service"
)

type cli struct {
	version *Version
}

func NewCLI(ver *Version) *cli {
	return &cli{
		version: ver,
	}
}

func (a *cli) Execute() error {
	var configPath string

	cmd := &cobra.Command{
		Use:   "notion-mindmap-server",
		Short: "A mindmap server application",
		Long:  `A mindmap server application that manages keyword nodes and their relationships for creating mind maps with Notion.`,
		RunE:  a.getRunErrorFn(&configPath),
	}

	cmd.Flags().
		StringVarP(&configPath, "config", "c", "config/config.json", "Path to the configuration file")

	cmd.AddCommand(a.getVersionCommand())

	return cmd.Execute()
}

func (a *cli) getRunErrorFn(configPath *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return a.execute(*configPath)
	}
}

func (a *cli) execute(configPath string) error {
	cfg, err := a.loadConfig(configPath)
	if err != nil {
		return err
	}

	server, err := server.New(cfg.Server)
	if err != nil {
		return err
	}

	mindMapRepo := repository.NewMemoryMindMapRepo()
	mindMapSvc := service.NewMindMapService(mindMapRepo)
	mindMapAPIGroup := controller.NewMindMapController(mindMapSvc)

	server.InstallAPIGroup(
		api.NewSimpleAPI("/version", a.getVersionHandler()),
		mindMapAPIGroup,
	)

	return server.Start()
}

func (a *cli) loadConfig(cfgFilePath string) (*config.AppConfig, error) {
	cfg := config.Default()
	if err := cfg.LoadConfig(cfgFilePath); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (a *cli) getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of notion-mindmap-server",
		Long:  `All software has versions. This is notion-mindmap-server's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", a.version.AppVersion)
			fmt.Printf("Commit: %s\n", a.version.Commit)
			fmt.Printf("Build Date: %s\n", a.version.BuildDate)
			fmt.Printf("Go Version: %s\n", a.version.GoVersion)
			fmt.Printf("Platform: %s\n", a.version.Platform)
		},
	}
}

func (a *cli) getVersionHandler() api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(a.version); err != nil {
			return err
		}
		return nil
	}
}
