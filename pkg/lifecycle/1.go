package lifecycle

/*
// Service provides abstract methods to control the lifecycle of a service
type Service1 interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ServiceLifecycle is a lifecycle of one service
type ServiceLifecycle1 struct {
	innerCtx    context.Context
	innerCancel context.CancelFunc
	services    []Service1
	failure     bool
	timeout     time.Duration
}

// NewService returns an initialized service lifecycle
func NewService1(timeout time.Duration) *ServiceLifecycle1 {
	innerCtx, innerCancel := context.WithCancel(context.Background())
	return &ServiceLifecycle1{
		innerCtx:    innerCtx,
		innerCancel: innerCancel,
		timeout:     timeout,
	}
}

// StartServices starts running services
func (s *ServiceLifecycle1) StartServices(ctx context.Context, services ...Service1) *ServiceLifecycle1 {
	if s.failure {
		return s
	}
	s.services = append(s.services, services...)
	for _, service := range services {
		select {
		case <-s.innerCtx.Done():
			s.failure = true
			return s
		default:
		}
		go s.start(ctx, service)
	}
	return s
}

func (s *ServiceLifecycle1) start(ctx context.Context, service Service1) {
	defer s.innerCancel()
	if err := service.Start(ctx); err != nil {
		log.Panicf("Service %s starts error: %v", service.Name(), err)
	} else {
		log.Printf("Service %s starts successfully", service.Name())
	}
}

// Signals registers monitor signals
func (s *ServiceLifecycle1) Signals(sigs ...os.Signal) *ServiceLifecycle1 {
	if s.failure {
		return s
	}
	go s.signals(sigs...)
	return s
}

func (s *ServiceLifecycle1) signals(sigs ...os.Signal) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, sigs...)
	for {
		select {
		case <-s.innerCtx.Done():
			return
		case sig := <-sigCh:
			for _, j := range sigs {
				if j == sig {
					s.innerCancel()
					return
				}
			}
		}
	}
}

// Wait blocks until context is done
func (s *ServiceLifecycle1) Wait(ctx context.Context) {
	<-s.innerCtx.Done()
	s.StopServices(ctx)
}

// StopServices can stop services when context is done or timeout
func (s *ServiceLifecycle1) StopServices(ctx context.Context) {
	gCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
	go s.stop(ctx, cancel)

	<-gCtx.Done()
	if errors.Is(gCtx.Err(), context.Canceled) {
		log.Printf("Services stop working, service config timeout : %s", s.timeout)
	} else if errors.Is(gCtx.Err(), context.DeadlineExceeded) {
		log.Panic("Timeout while stopping service, killing instance manually")
	}
}

func (s *ServiceLifecycle1) stop(ctx context.Context, cancel context.CancelFunc) {
	var wg sync.WaitGroup
	for _, service := range s.services {
		wg.Add(1)
		go func(ctx context.Context, service Service1) {
			defer wg.Done()
			if err := service.Stop(ctx); err != nil {
				log.Panicf("Service %s stops failure: %v", service.Name(), err)
			} else {
				log.Printf("Service %s stops successfully!", service.Name())
			}
		}(ctx, service)
	}
	wg.Wait()
	cancel()
}

// Done check context is done
func (s *ServiceLifecycle1) Done() <-chan struct{} {
	return s.innerCtx.Done()
}

*/
