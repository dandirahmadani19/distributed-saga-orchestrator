package app

import "fmt"

func (a *App) Run() error {
	go func() {
		a.Log.Info().Msg(fmt.Sprintf("gRPC server started on :%d", a.Cfg.GRPC.Port))
		if err := a.GRPC.Start(); err != nil {
			a.Log.Fatal().Err(err).Msg("grpc crashed")
		}
	}()

	a.waitShutdown()
	return nil
}
