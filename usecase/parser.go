package usecase

import (
	"fmt"

	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
)

func parseStatus(s int) (model.Status, error) {
	status := model.ToStatus(s)
	if status == model.StatusUnknown {
		err := fmt.Errorf(
			"status must be %d to %d, but %d",
			int(model.StatusNotReady), int(model.StatusDone), s,
		)
		return status, err
	}
	return status, nil
}

func parsePriority(p int) (model.Priority, error) {
	priority := model.ToPriority(p)
	if priority == model.PriorityUnknown {
		err := fmt.Errorf(
			"priority must be %d to %d, but %d",
			int(model.PriorityHigh), int(model.PriorityLow), p,
		)
		return priority, err
	}
	return priority, nil
}
