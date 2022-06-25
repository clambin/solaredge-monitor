package sampler

import "context"

type Sampler interface {
	Sample(ctx context.Context) (Sample, error)
}
