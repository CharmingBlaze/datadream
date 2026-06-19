package main

import (
	"context"
	"os"

	"datadream/internal/ide"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App exposes IDE operations to the Wails frontend.
type App struct {
	ctx context.Context
	svc *ide.Service
}

func NewApp(root string) (*App, error) {
	root = ide.EnsureDistributionRoot(root)
	_ = os.Setenv("DATADREAM_ROOT", root)
	svc, err := ide.NewService(root)
	if err != nil {
		return nil, err
	}
	return &App{svc: svc}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.WindowSetTitle(ctx, "DataDream Studio — "+a.svc.Root())
}

func (a *App) Version() ide.VersionInfo {
	return a.svc.Version()
}

func (a *App) Tree(path string) (ide.TreeNode, error) {
	return a.svc.Tree(path)
}

func (a *App) Search(q string) ide.SearchResult {
	return a.svc.Search(q)
}

func (a *App) Read(path string) (ide.FileContent, error) {
	return a.svc.Read(path)
}

func (a *App) Write(path, content string) error {
	return a.svc.Write(path, content)
}

func (a *App) NewFile(path, template string) (ide.FileContent, error) {
	return a.svc.NewFile(path, template)
}

func (a *App) Check(path, content string) (ide.CheckResult, error) {
	return a.svc.Check(path, content)
}

func (a *App) Build(path, content string, release bool) ide.CommandResult {
	return a.svc.Build(ide.BuildRequest{Path: path, Content: content, Release: release})
}

func (a *App) Run(path, content string) ide.CommandResult {
	return a.svc.Run(path, content)
}

func (a *App) Doctor() ide.DoctorStatus {
	return a.svc.Doctor()
}

func (a *App) OpenProject() (ide.VersionInfo, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open DataDream Project",
	})
	if err != nil {
		return ide.VersionInfo{}, err
	}
	if dir == "" {
		return a.svc.Version(), nil
	}
	if err := a.svc.SetRoot(dir); err != nil {
		return ide.VersionInfo{}, err
	}
	info := a.svc.Version()
	runtime.WindowSetTitle(a.ctx, "DataDream Studio — "+info.Root)
	return info, nil
}

func (a *App) SetProjectRoot(root string) (ide.VersionInfo, error) {
	if err := a.svc.SetRoot(root); err != nil {
		return ide.VersionInfo{}, err
	}
	info := a.svc.Version()
	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, "DataDream Studio — "+info.Root)
	}
	return info, nil
}

func (a *App) IsDesktop() bool {
	return true
}
