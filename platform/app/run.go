package app

func (a *App) Run() error {
	go func() {
		a.Log.Info().Msg("gRPC server started")
		if err := a.GRPC.Start(); err != nil {
			a.Log.Fatal().Err(err).Msg("grpc crashed")
		}
	}()

	a.waitShutdown()
	return nil
}
