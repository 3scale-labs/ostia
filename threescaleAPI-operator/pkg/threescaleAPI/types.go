package threescaleAPI

import "sort"

type Endpoints struct {
	Endpoint []Endpoint `json:"endpoints"`
}

func (e Endpoints) Sort() Endpoints {
	for _, ep := range e.Endpoint {
		for _, op := range ep.OperationIDs {
			sort.Slice(op.Metrics, func(i, j int) bool {
				if op.Metrics[i].Metric != op.Metrics[j].Metric {
					return op.Metrics[i].Metric < op.Metrics[j].Metric
				} else {
					return op.Metrics[i].Increment < op.Metrics[j].Increment
				}
			})
		}
	}

	sort.Slice(e.Endpoint, func(i, j int) bool {
		if e.Endpoint[i].Path != e.Endpoint[j].Path {
			return e.Endpoint[i].Path < e.Endpoint[j].Path
		} else {
			return len(e.Endpoint[i].OperationIDs) < len(e.Endpoint[j].OperationIDs)
		}
	})

	for _, ep := range e.Endpoint {
		sort.Slice(ep.OperationIDs, func(i, j int) bool {
			if ep.OperationIDs[i].Method != ep.OperationIDs[j].Method {
				return ep.OperationIDs[i].Name < ep.OperationIDs[j].Method
			} else {
				return len(ep.OperationIDs[i].Metrics) < len(ep.OperationIDs[j].Metrics)
			}
		})
	}

	return e
}

type Endpoint struct {
	Path         string        `json:"path"`
	OperationIDs []OperationID `json:"operationIDs"`
}

func (ep Endpoint) Sort() Endpoint {
	for _, op := range ep.OperationIDs {
		sort.Slice(op.Metrics, func(i, j int) bool {
			if op.Metrics[i].Metric != op.Metrics[j].Metric {
				return op.Metrics[i].Metric < op.Metrics[j].Metric
			} else {
				return op.Metrics[i].Increment < op.Metrics[j].Increment
			}
		})
	}
	sort.Slice(ep.OperationIDs, func(i, j int) bool {
		return len(ep.OperationIDs[i].Name) < len(ep.OperationIDs[j].Name)
	})
	return ep
}

type OperationID struct {
	Name    string   `json:"name"`
	Method  string   `json:"method"`
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	Metric    string `json:"metric"`
	Increment int64  `json:"increment"`
}
