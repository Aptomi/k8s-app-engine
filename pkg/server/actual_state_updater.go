package server

import (
	log "github.com/Sirupsen/logrus"
	"time"
)

func (server *Server) actualStateUpdateLoop() error {
	for {
		err := server.actualStateUpdate()
		if err != nil {
			log.Errorf("error while updating actual state: %s", err)
		}

		// sleep for a specified time or wait until policy has changed, whichever comes first
		timer := time.NewTimer(server.cfg.Updater.Interval)
		select {
		case <-server.runActualStateUpdate:
			break // nolint: megacheck
		case <-timer.C:
			break // nolint: megacheck
		}
		timer.Stop()
	}
}

func (server *Server) actualStateUpdate() error {
	server.actualStateUpdateIdx++

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic while updating actual state: %s", err)
		}
	}()

	log.Infof("(update-%d) Actual state updated", server.actualStateUpdateIdx)

	return nil
}
