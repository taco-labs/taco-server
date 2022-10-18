package service

import "github.com/panjf2000/ants/v2"

type WorkerPoolService interface {
	Submit(func()) error
}

type antWorkerPoolService struct {
	pool *ants.Pool
}

func (a antWorkerPoolService) Submit(fn func()) error {
	return a.pool.Submit(fn)
}

func NewAntWorkerPoolService(pool *ants.Pool) antWorkerPoolService {
	return antWorkerPoolService{
		pool: pool,
	}
}
