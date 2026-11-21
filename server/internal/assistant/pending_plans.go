package assistant

import (
	"context"
	"errors"

	"github.com/nwverse/persona/server/internal/db"
)

type openTaskPlanStore interface {
	OpenTaskPlans(context.Context) ([]db.TaskPlanRow, error)
}

// openTaskPlans uses only this manager's store; managers have independent databases.
func (s *Service) openTaskPlans(ctx context.Context) ([]*TaskPlan, error) {
	if st := s.store(); st != nil {
		reader, ok := st.(openTaskPlanStore)
		if !ok {
			return nil, errors.New("pending plans: store does not support open-plan reads")
		}
		rows, err := reader.OpenTaskPlans(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]*TaskPlan, 0, len(rows))
		for _, row := range rows {
			out = append(out, planFromRow(row))
		}
		return out, nil
	}
	s.tasks.mu.Lock()
	defer s.tasks.mu.Unlock()
	var out []*TaskPlan
	for _, p := range s.tasks.plans {
		if p.State != planOpen {
			continue
		}
		cp := *p
		cp.Tasks = append([]PlanTask(nil), p.Tasks...)
		out = append(out, &cp)
	}
	return out, nil
}
