package sctx

import (
	"flag"
	"fmt"
	logger "inventory-movement-processing/pkg/components/loggerc"
	zaplogger "inventory-movement-processing/pkg/components/loggerc/zap"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	DevEnv = "dev"
	StgEnv = "stg"
	PrdEnv = "prd"
)

type Component interface {
	ID() string
	InitFlags()
	Activate(ServiceContext) error
	Stop() error
}

type ServiceContext interface {
	Load() error
	Stop() error

	Logger(prefix string) logger.Logger
	LogLevel() string

	EnvName() string
	GetName() string
	Get(id string) (interface{}, bool)
	MustGet(id string) interface{}

	OutEnv()
}

type serviceCtx struct {
	name       string
	env        string
	components []Component
	store      map[string]Component
	logger     logger.Logger
}

var defaultLogger = zaplogger.NewZapLogger()

type Option func(*serviceCtx)

func WithName(name string) Option {
	return func(s *serviceCtx) {
		s.name = name
	}
}

func WithComponent(c Component) Option {
	return func(s *serviceCtx) {
		if _, ok := s.store[c.ID()]; !ok {
			s.components = append(s.components, c)
			s.store[c.ID()] = c
		}
	}
}

func NewServiceContext(opts ...Option) ServiceContext {
	s := &serviceCtx{
		store: make(map[string]Component),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.initFlags()
	s.parseFlags()

	return s
}

func (s *serviceCtx) initFlags() {
	flag.StringVar(&s.env, "app-env", DevEnv, "Env for service: dev | stg | prd")
	defaultLogger.InitFlags()
	for _, c := range s.components {
		c.InitFlags()
	}
}

func (s *serviceCtx) Load() error {
	if defaultLogger.GetLevel() == "" {
		switch s.env {
		case DevEnv:
			_ = defaultLogger.SetLevel("debug")
		case StgEnv:
			_ = defaultLogger.SetLevel("info")
		case PrdEnv:
			_ = defaultLogger.SetLevel("warn")
		default:
			log.Println("here")
			_ = defaultLogger.SetLevel("info")
		}
	}

	if err := defaultLogger.Activate(); err != nil {
		return err
	}

	s.logger = s.Logger("service-context")
	s.logger.Infof(
		"service starting env=%s log_level=%s",
		s.env,
		defaultLogger.GetLevel(),
	)

	for _, c := range s.components {
		s.logger.Infof("activating component: %s", c.ID())
		if err := c.Activate(s); err != nil {
			return fmt.Errorf("activate %s: %v", c.ID(), err)
		}
	}
	return nil
}

func (s *serviceCtx) Stop() error {
	s.logger.Info("stopping service context")
	for _, c := range s.components {
		if err := c.Stop(); err != nil {
			return fmt.Errorf("stop %s: %v", c.ID(), err)
		}
	}
	_ = defaultLogger.Stop()
	return nil
}

func (s *serviceCtx) Logger(prefix string) logger.Logger {
	return defaultLogger.GetLogger(prefix)
}

func (s *serviceCtx) Get(id string) (interface{}, bool) {
	c, ok := s.store[id]
	return c, ok
}

func (s *serviceCtx) MustGet(id string) interface{} {
	c, ok := s.Get(id)
	if !ok {
		panic(fmt.Sprintf("cannot get component %s", id))
	}
	return c
}

func (s *serviceCtx) LogLevel() string {
	return defaultLogger.GetLevel()
}

func (s *serviceCtx) EnvName() string { return s.env }
func (s *serviceCtx) GetName() string { return s.name }

func (s *serviceCtx) parseFlags() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			log.Fatalf("load env file %s: %v", envFile, err)
		}
	}

	// parse flag to env format
	flag.VisitAll(func(f *flag.Flag) {
		envKey := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if v := os.Getenv(envKey); v != "" {
			flag.Set(f.Name, v)
		}
	})

	flag.Parse()
}

func (s *serviceCtx) OutEnv() {
	fmt.Println("Resolved environment variables:")

	flag.VisitAll(func(f *flag.Flag) {
		envKey := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))

		value := f.Value.String()
		source := "default"
		if value != f.DefValue {
			source = "override"
		}

		fmt.Printf(
			"%-30s = %-20s (%s)\n    ↳ %s\n\n",
			envKey,
			value,
			source,
			f.Usage,
		)
	})
}
