package manager

import "time"

// SPExitScheduler subscribes sp exit events and produces a gvg migrate plan.
type SPExitScheduler struct {
	spID                        uint32
	manager                     *ManageModular
	currentSubscribeBlockHeight int
	isExiting                   bool
	// spStatus                    string  // active/exiting/exited
}

// Init function is used to load db subscribe block progress and migrate gvg progress.
func (s *SPExitScheduler) Init() error {
	return nil
}

// Start function is used to subscribe sp exit event from metadata and produces a gvg migrate plan.
func (s *SPExitScheduler) Start() error {
	go s.subscribeSPExitEvents()
	return nil
}

func (s *SPExitScheduler) subscribeSPExitEvents() {
	subscribeSPExitEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSPExitEventInterval) * time.Second)
	for {
		select {
		case <-subscribeSPExitEventsTicker.C:
			// TODO: subscribe sp exit events from metadata service.
			// spExitEvent, err = s.manager.baseApp.GfSpClient().ListSPExitEvents(s.currentSubscribeBlockHeight, s.manager.baseApp.OperatorAddress())

			if s.isExiting {
				return
			}
			// TODO exit
			s.produceVirtualGroupSwapExecutePlan()
			s.isExiting = true
		}
	}
}

// TODO: proto replace it.
// bucket migrate and sp exit can reuse it.
type virtualGroupSwapExecuteUnit struct {
}

func (s *SPExitScheduler) produceVirtualGroupSwapExecutePlan() {
	// TODO

}
