package closer

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

type CloserFunc func() error

type Closer struct {
	mu     sync.Mutex
	funcs  []CloserFunc
	logger *zap.Logger
}

func NewCloser(logger *zap.Logger) *Closer {
	return &Closer{
		logger: logger,
	}
}

func (c *Closer) Add(fn CloserFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, fn)
}

func (c *Closer) CloseAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.funcs) - 1; i >= 0; i-- {
		if err := c.funcs[i](); err != nil {
			c.logger.Error("Error during shutdown", zap.Error(err))
		}
	}
}

func (c *Closer) Wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	c.logger.Info("Shutting down...")
	c.CloseAll()
}