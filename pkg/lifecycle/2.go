package lifecycle

/*
// Service provides abstract methods to control the lifecycle of a service
type Service interface {
	Name() string
	Start() error
	Stop() error
}

// ServiceLifecycle is a lifecycle of one service
type ServiceLifecycle struct {
	services   []Service
	serviceMap sync.Map
	sigChan    chan os.Signal
	stopChan   chan bool
	timeout    time.Duration
}

// NewService returns an initialized service lifecycle
func NewService(timeout time.Duration) *ServiceLifecycle {
	s := &ServiceLifecycle{
		stopChan: make(chan bool, 1),
		sigChan:  make(chan os.Signal, 1),
		timeout:  timeout,
	}
	s.configureSignals()
	return s
}

func (s *ServiceLifecycle) configureSignals() {
	signal.Notify(s.sigChan, syscall.SIGUSR1)
}

// RegisterServices
func (s *ServiceLifecycle) RegisterServices(services ...Service) *ServiceLifecycle {
	for _, service := range services {
		s.serviceMap.Store(service.Name(), service)
	}
	return s
}

// StartServices starts running services
func (s *ServiceLifecycle) StartServices(ctx context.Context, services ...Service) {
	go func() {
		<-ctx.Done()
		s.StopServices()
	}()
	s.services = append(s.services, services...)
	s.start()
	//s.serviceMap.Range(func(key, value any) bool {
	//	// name := key.(string)
	//	srv := value.(Service)
	//	s.start(srv)
	//	return true
	//})
}

func (s *ServiceLifecycle) start() {
	for _, service := range s.services {
		go func(service Service) {
			log.Infow("start func", "name", service.Name())
			if err := service.Start(); err != nil {
				log.Errorf("Service %s starts error: %v", service.Name(), err)
				if err1 := service.Stop(); err1 != nil {
					log.Errorf("Service %s stops failure: %v", service.Name(), err1)
				}
			} else {
				log.Infof("Service %s starts successfully", service.Name())
			}
		}(service)
	}
}

// Wait blocks until service shutdown
func (s *ServiceLifecycle) Wait() {
	<-s.stopChan
}

// StopServices stops services
func (s *ServiceLifecycle) StopServices() {
	defer log.Info("Services stopped")
	//s.serviceMap.Range(func(key, value any) bool {
	//	srv := value.(Service)
	//	s.stop(srv)
	//	return true
	//})
	go s.stop()
	s.stopChan <- true
}

func (s *ServiceLifecycle) stop() {
	var wg sync.WaitGroup
	for _, service := range s.services {
		wg.Add(1)
		go func(service Service) {
			defer wg.Done()
			if err := service.Stop(); err != nil {
				log.Errorf("Service %s stops failure: %v", service.Name(), err)
			} else {
				log.Errorf("Service %s stops successfully!", service.Name())
			}
		}(service)
	}
	wg.Wait()
	//if err := service.Stop(); err != nil {
	//	log.Errorf("Service %s stops failure: %v", service.Name(), err)
	//} else {
	//	log.Infof("Service %s stops successfully!", service.Name())
	//}
}

// Close destroy services
func (s *ServiceLifecycle) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	go func() {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.Canceled) {
			log.Infow("Services stop working", "service config timeout", s.timeout)
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Panic("Timeout while closing service, killing instance manually")
		}
	}()

	log.Info("close services")
	signal.Stop(s.sigChan)
	close(s.sigChan)
	close(s.stopChan)

	// this func would add some other stopFunc services such as stopMetricsClient
	cancel()
}

*/
