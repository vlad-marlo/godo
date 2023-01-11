package app

import "golang.org/x/sync/errgroup"

type App struct {
}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	gr := errgroup.Group{}
	gr.Go(func() error {
		return nil
	})

	return gr.Wait()
}

func (a *App) startHTTP() error {
	return nil
}
