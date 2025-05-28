package cron

import (
	"github.com/robfig/cron/v3"
)

type Service struct {
	c *cron.Cron
}

func NewService() *Service {
	return &Service{
		c: cron.New(),
	}
}

func (s *Service) Add(spec string, cmd cron.Job) error {
	_, err := s.c.AddJob(spec, cmd)
	return err
}

func (s *Service) Start() {
	s.c.Start()
}

func (s *Service) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
}
