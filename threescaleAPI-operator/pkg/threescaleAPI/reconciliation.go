package threescaleAPI

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/3scale/ostia/threescaleAPI-operator/pkg/threescale/system_client"
	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/getkin/kin-openapi/openapi3"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ostiaExtensions struct {
	openapi3.ExtensionProps
}

func (ostiaExtensions *ostiaExtensions) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(ostiaExtensions)
}

func (ostiaExtensions *ostiaExtensions) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, ostiaExtensions)
}

func reconcilePlansAndLimits(c *client.ThreeScaleClient, service client.Service, accessToken string, desiredPlans Plans) {
	var planFetchErr error

	type mappingRuleStore map[string]struct {
		id       string
		metricId string
	}

	svcId := service.ID

	mrStoreC := make(chan mappingRuleStore)
	go func() {
		mappingRules, err := c.ListMappingRule(accessToken, svcId)
		if err != nil {
			fmt.Println("error reading mapping rules " + err.Error())
		}
		mrMap := make(mappingRuleStore, len(mappingRules.MappingRules))
		for _, mr := range mappingRules.MappingRules {
			store := fmt.Sprintf("%s-%s", mr.HTTPMethod, mr.Pattern)
			mrMap[store] = struct {
				id       string
				metricId string
			}{id: mr.ID, metricId: mr.MetricID}
		}
		mrStoreC <- mrMap
	}()

	appPlanIdMapC := make(chan map[string]string)
	go func() {
		havePlans, err := getAppPlans(c, svcId, accessToken)
		appPlanIdMapC <- havePlans
		planFetchErr = err
	}()

	havePlans := <-appPlanIdMapC
	if planFetchErr != nil {
		fmt.Println("Cannot continue")
		os.Exit(1)
	}

	//delete unwanted plans in background
	go func() {
		for k, v := range havePlans {
			keep := false
			for _, plan := range desiredPlans.Plans {
				if plan.Name == k {
					keep = true
					break
				}
			}
			if !keep {
				err := c.DeleteAppPlan(accessToken, svcId, v)
				if err != nil {
					fmt.Println("error deleting plan - " + err.Error())
				}
			}
		}
	}()

	//mappingRuleValues := <-mrStoreC

	// for all the plans we want
	for _, wantedPlan := range desiredPlans.Plans {
		var id string
		// Have you read this plan name previously? If its in the map then yes, set the id
		if val, ok := havePlans[wantedPlan.Name]; ok {
			id = val
		} else {
			// if not create a new one, wait for success and set the id
			fmt.Println("Attempting to create new plan " + wantedPlan.Name)
			newPlan, err := c.CreateAppPlan(accessToken, svcId, wantedPlan.Name, "publish")
			if err != nil {
				// Perhaps some retries here depending on the error
				fmt.Println("failed to create the application plan " + wantedPlan.Name)
				// Not much we can do having failed, move on and try the next desired plan
				continue
			}
			// safely store the id at this point and add the name --> id to map
			id = newPlan.ID
			havePlans[newPlan.PlanName] = id

		}

		// In background, get a list of limits associated with the current plan
		haveLimitsChan := make(chan client.LimitList)
		go func() {
			// list the limits associated with this application plan
			hl, err := c.ListLimitsPerAppPlan(accessToken, id)
			if err != nil {
				// some retry here based on err
				fmt.Println("cant get limits. exiting")
				os.Exit(1)
			}
			haveLimitsChan <- hl
		}()

		// In background, get a list of metrics associated with this service
		// Map the name of the metric to the metric id
		metricIdMap := make(chan map[string]string)
		//var existingMetrics client.MetricList
		go func() {
			existingMetrics, err := c.ListMetrics(accessToken, svcId)
			if err != nil {
				// some retry here based on err
				fmt.Println("error getting metrics")
				os.Exit(1)
			}
			nameToIds := make(map[string]string, len(existingMetrics.Metrics))
			for _, existingMetric := range existingMetrics.Metrics {
				nameToIds[existingMetric.SystemName] = existingMetric.ID
			}
			metricIdMap <- nameToIds
		}()

		limitsToKeep := client.LimitList{}

		// wait for async calls ready
		haveLimits := <-haveLimitsChan
		metricMap := <-metricIdMap

		for _, wantedLimit := range wantedPlan.Limits {
			dealtWith := false
			// for each wanted limit, look at the metric and see if its in our metric map
			if metricID, ok := metricMap[wantedLimit.Metric]; ok {
				// the metric exists and we have the id
				// iterate over all the appPlanIdMapC limits and see if one matches the id
				for _, limit := range haveLimits.Limits {
					if limit.MetricID == metricID {
						// we have a match
						// check if it matches the period we are looking for
						if limit.Period == wantedLimit.Period {
							// we have a match
							// check if the values are the same
							strVal := strconv.FormatInt(wantedLimit.Max, 10)
							if limit.Value != strVal {
								// patch it
								p := client.NewParams()
								p.AddParam("value", strVal)
								go c.UpdateLimitPerAppPlan(accessToken, id, metricID, limit.ID, p)
							}
							// nothing to do its in a good state
							dealtWith = true
							limitsToKeep.Limits = append(limitsToKeep.Limits, limit)
							break
						}
					}
				}

				if !dealtWith {
					fmt.Printf("[+] Attempting to create limit %s with period %s\n", wantedLimit.Metric, wantedLimit.Period)
					_, err := c.CreateLimitAppPlan(accessToken, id, metricID, wantedLimit.Period, int(wantedLimit.Max))
					if err != nil {
						fmt.Println("Tried to create limit and failed")
						fmt.Println(err)
					}
				}
			} else {
				fmt.Println("Required metric doesnt exist")
				continue
			}
		}

		// delete the limits we dont want
		go func() {
			for _, limit := range haveLimits.Limits {
				want := false
				for _, l := range limitsToKeep.Limits {
					if reflect.DeepEqual(limit, l) {
						want = true
						break
					}
				}

				if !want {
					go c.DeleteLimitPerAppPlan(accessToken, id, limit.MetricID, limit.ID)
				}
			}
		}()
	}
}

func getAppPlans(c *client.ThreeScaleClient, svcId, accessToken string) (map[string]string, error) {
	plansList, err := c.ListAppPlanByServiceId(accessToken, svcId)
	if err != nil {
		return nil, err
	}

	nameToIds := make(map[string]string, len(plansList.Plans))
	for _, appPlan := range plansList.Plans {
		nameToIds[appPlan.PlanName] = appPlan.ID
	}
	return nameToIds, nil

}
func decodePlans(s *openapi3.Swagger) (Plans, error) {
	var desiredPlans Plans

	switch s.Extensions["x-3scale-plans"].(type) {
	case json.RawMessage:
		err := json.Unmarshal(s.Extensions["x-3scale-plans"].(json.RawMessage), &desiredPlans)
		if err != nil {
			return desiredPlans, errors.New("error calling unmarshal of plans")

		}
	default:
		return desiredPlans, errors.New("error - plans not recognised as json")

	}
	return desiredPlans, nil
}

func reconcileEndpointsWith3scaleSystem(c *client.ThreeScaleClient, accessToken string, service client.Service, existingEndpoints Endpoints, desiredEndpoints Endpoints) error {

	existingMetrics, _ := c.ListMetrics(accessToken, service.ID)

	// Purge non desired Metrics from system in background
	go func(existingMetrics client.MetricList, accessToken string, service client.Service) {
		metricsToDelete := getNonDesiredMetricsFromSystem(existingMetrics, desiredEndpoints)
		for _, deleteMetric := range metricsToDelete {
			fmt.Println("Deleting unused Metric: ", deleteMetric.SystemName)
			c.DeleteMetric(accessToken, service.ID, deleteMetric.ID)
		}
	}(existingMetrics, accessToken, service)

	// Create missing desired metric in system

	metricsToCreate := getDesiredMetricsButNonExistingInSystem(existingMetrics, desiredEndpoints)
	for _, metricToCreate := range metricsToCreate {
		fmt.Println("Creating desired Metric: ", metricToCreate.Metric)
		_, err := c.CreateMetric(accessToken, service.ID, metricToCreate.Metric, "hits")
		if err != nil {
			return err
		}
	}

	// Now let's remove not needed mapping rules in background

	go func(existingEndpoints Endpoints, desiredEndpoints Endpoints, accessToken string, service client.Service) {
		existingMappingRules, _ := c.ListMappingRule(accessToken, service.ID)
		mappingRulesToDelete := getNonDesiredMappingRulesFromSystem(existingEndpoints, existingMappingRules, desiredEndpoints)
		for _, mappingRuleToDelete := range mappingRulesToDelete {
			c.DeleteMappingRule(accessToken, service.ID, mappingRuleToDelete.ID)
		}
	}(existingEndpoints, desiredEndpoints, accessToken, service)

	// Now let's create the desired mappingRules
	existingMappingRules, _ := c.ListMappingRule(accessToken, service.ID)
	mappingRulesToCreate := getDesiredMappingRulesButNonExistingInSystem(c, accessToken, existingEndpoints, existingMappingRules, desiredEndpoints, service)
	for _, mappingRuleToCreate := range mappingRulesToCreate {

		increment, _ := strconv.ParseInt(mappingRuleToCreate.Delta, 10, 64)

		_, err := c.CreateMappingRule(accessToken, service.ID, mappingRuleToCreate.HTTPMethod, mappingRuleToCreate.Pattern, int(increment), mappingRuleToCreate.MetricID)

		if err != nil {
			return err
		}
	}

	return nil

}

func comparePlans(planA Plans, planB Plans) bool {

	A, _ := json.Marshal(planA.Sort())
	B, _ := json.Marshal(planB.Sort())

	return reflect.DeepEqual(A, B)
}
func compareEndpoints(endpointsA Endpoints, endpointsB Endpoints) bool {

	A, _ := json.Marshal(endpointsA.Sort())
	B, _ := json.Marshal(endpointsB.Sort())

	return reflect.DeepEqual(A, B)

}
func compareEndpoint(endpointA Endpoint, endpointB Endpoint) bool {

	A, _ := json.Marshal(endpointA.Sort())
	B, _ := json.Marshal(endpointB.Sort())

	return reflect.DeepEqual(A, B)

}

// TODO: MOVE TO THREESCALE PACKAGE
func getServiceFromServiceSystemName(c *client.ThreeScaleClient, accessToken string, serviceName string) (client.Service, error) {
	services, _ := c.ListServices(accessToken)
	for _, service := range services.Services {
		if service.SystemName == serviceName {
			return service, nil
		}
	}
	return client.Service{}, errors.New("not found")
}
func getMetricFromMetricID(c *client.ThreeScaleClient, accessToken string, serviceID string, metricID string) (client.Metric, error) {

	metrics, err := c.ListMetrics(accessToken, serviceID)

	if err != nil {
		return client.Metric{}, err
	}

	for _, metric := range metrics.Metrics {
		if metric.ID == metricID {
			return metric, nil
		}
	}

	return client.Metric{}, errors.New("NotFound")
}
func getMetricFromMetricName(c *client.ThreeScaleClient, accessToken string, serviceID string, metricName string) (client.Metric, error) {

	metrics, err := c.ListMetrics(accessToken, serviceID)

	if err != nil {
		return client.Metric{}, err
	}

	for _, metric := range metrics.Metrics {
		if metric.SystemName == metricName {
			return metric, nil
		}
	}

	return client.Metric{}, errors.New("not found")
}
func getEndpointsFrom3scaleSystem(c *client.ThreeScaleClient, accessToken string, serviceName string) (Endpoints, error) {

	service, err := getServiceFromServiceSystemName(c, accessToken, serviceName)
	if err != nil {
		return Endpoints{}, err
	}

	mappingRules, _ := c.ListMappingRule(accessToken, service.ID)

	var endpoints Endpoints

	for _, mappingRule := range mappingRules.MappingRules {
		var endpoint Endpoint
		var operationID OperationID
		var metric Metric
		existingEndpoint := -1
		existingOperationID := -1

		for k, v := range endpoints.Endpoint {
			if v.Path == mappingRule.Pattern {
				existingEndpoint = k
				for k, v := range endpoints.Endpoint[existingEndpoint].OperationIDs {
					if v.Name == strings.Join([]string{mappingRule.HTTPMethod, mappingRule.Pattern}, "-") {
						existingOperationID = k
						break
					}
				}
				break
			}
		}

		metrics, _ := c.ListMetrics(accessToken, service.ID)
		var metricName string
		for _, v := range metrics.Metrics {
			if mappingRule.MetricID == v.ID {
				metricName = v.SystemName
			}
		}

		metric.Metric = metricName
		metric.Increment, _ = strconv.ParseInt(mappingRule.Delta, 10, 64)

		if existingEndpoint != -1 {
			if existingOperationID != -1 {
				endpoints.Endpoint[existingEndpoint].OperationIDs[existingOperationID].Metrics = append(endpoints.Endpoint[existingEndpoint].OperationIDs[existingOperationID].Metrics, metric)
			} else {
				operationID.Method = mappingRule.HTTPMethod
				operationID.Name = strings.Join([]string{mappingRule.HTTPMethod, mappingRule.Pattern}, "-")
				operationID.Metrics = append(operationID.Metrics, metric)
				endpoints.Endpoint[existingEndpoint].OperationIDs = append(endpoints.Endpoint[existingEndpoint].OperationIDs, operationID)
			}
		} else {
			endpoint.Path = mappingRule.Pattern
			operationID.Method = mappingRule.HTTPMethod
			operationID.Name = strings.Join([]string{mappingRule.HTTPMethod, mappingRule.Pattern}, "-")
			operationID.Metrics = append(operationID.Metrics, metric)
			endpoint.OperationIDs = append(endpoint.OperationIDs, operationID)
			endpoints.Endpoint = append(endpoints.Endpoint, endpoint)
		}

	}

	return endpoints, nil
}
func getEndpointsFromSwagger(swagger *openapi3.Swagger) (Endpoints, error) {

	var endpoints Endpoints

	for name, path := range swagger.Paths {
		var endpoint Endpoint
		endpoint.Path = name

		for method, v := range path.Operations() {
			var operationID OperationID
			// Mapping Rules in system don't have a name.. let's use the method+path way.
			// operationID.Name = v.OperationID
			operationID.Name = strings.Join([]string{method, name}, "-")
			operationID.Method = method

			s, err := json.Marshal(v.Extensions["x-3scale-metrics"])
			if err != nil {
				return endpoints, err
			}

			err = json.Unmarshal(s, &operationID.Metrics)
			if err != nil {
				return endpoints, err
			}

			endpoint.OperationIDs = append(endpoint.OperationIDs, operationID)
		}
		endpoints.Endpoint = append(endpoints.Endpoint, endpoint)
	}

	return endpoints, nil
}
func getPlansFrom3scaleSystem(c *client.ThreeScaleClient, accessToken string, serviceName string) (Plans, error) {

	var plans Plans

	service, err := getServiceFromServiceSystemName(c, accessToken, serviceName)
	if err != nil {
		return Plans{}, err
	}

	appPlans, _ := c.ListAppPlanByServiceId(accessToken, service.ID)

	for _, v := range appPlans.Plans {

		var plan Plan

		limits, _ := c.ListLimitsPerAppPlan(accessToken, v.ID)

		plan.Name = v.PlanName
		metrics, _ := c.ListMetrics(accessToken, v.ServiceID)
		for _, v := range limits.Limits {
			var limit Limit
			metricID := v.MetricID
			var metricName string
			for _, v := range metrics.Metrics {
				if metricID == v.ID {
					metricName = v.SystemName
				}
			}

			limit.Metric = metricName
			limit.Period = v.Period
			limit.Max, _ = strconv.ParseInt(v.Value, 10, 64)
			plan.Limits = append(plan.Limits, limit)

		}
		plans.Plans = append(plans.Plans, plan)
	}

	return plans, nil
}
func getPlansFromSwagger(swagger *openapi3.Swagger) (Plans, error) {
	var desiredPlans Plans

	s, _ := json.Marshal(swagger.Extensions["x-3scale-plans"])
	err := json.Unmarshal(s, &desiredPlans)
	if err != nil {
		return desiredPlans, err
	}

	return desiredPlans, nil
}

// getNonDesiredMetricsFromSystem returns a list of metrics that are in the current system but not desired
func getNonDesiredMetricsFromSystem(existingMetrics client.MetricList, desiredEndpoints Endpoints) []client.Metric {

	var metricsToDelete client.MetricList
	for _, existingMetric := range existingMetrics.Metrics {
		keepMetric := false
		for _, desiredEndpoint := range desiredEndpoints.Endpoint {
			for _, operationId := range desiredEndpoint.OperationIDs {
				for _, desiredMetric := range operationId.Metrics {
					if existingMetric.SystemName == desiredMetric.Metric {
						keepMetric = true
					}
				}
			}
		}

		if !keepMetric {
			metricsToDelete.Metrics = append(metricsToDelete.Metrics, existingMetric)
		}
	}

	return metricsToDelete.Metrics

}

// getNonDesiredMappingRulesFromSystem returns a list of mappingRules from system that are not desired but created in system
func getNonDesiredMappingRulesFromSystem(existingEndpoints Endpoints, existingMappingRules client.MappingRuleList, desiredEndpoints Endpoints) []client.MappingRule {
	var mappingRulesToDelete []client.MappingRule

	for _, existingEndpoint := range existingEndpoints.Endpoint {
		for _, operationId := range existingEndpoint.OperationIDs {
			endpointExists := false
			for _, desiredEndpoint := range desiredEndpoints.Endpoint {
				for _, desiredOperationid := range desiredEndpoint.OperationIDs {
					if existingEndpoint.Path == desiredEndpoint.Path && operationId.Method == desiredOperationid.Method {
						endpointExists = true
					}
				}
			}
			if !endpointExists {
				var mappingRuleID int

				for k, existingMappingRule := range existingMappingRules.MappingRules {
					if existingMappingRule.Pattern == existingEndpoint.Path && operationId.Method == existingMappingRule.HTTPMethod {
						mappingRuleID = k
					}
				}
				mappingRulesToDelete = append(mappingRulesToDelete, existingMappingRules.MappingRules[mappingRuleID])
			}
		}
	}

	return mappingRulesToDelete
}

// getDesiredMappingRulesButNonExistingInSystem returns a list of desired mappingRules that are not in System
func getDesiredMappingRulesButNonExistingInSystem(c *client.ThreeScaleClient, accessToken string, existingEndpoints Endpoints, existingMappingRules client.MappingRuleList, desiredEndpoints Endpoints, service client.Service) []client.MappingRule {

	// Now let's create the missing mapping rules
	var mappingRulesToCreate []client.MappingRule

	for _, desiredEndpoint := range desiredEndpoints.Endpoint {
		endpointExists := -1
		for i, existingEndpoint := range existingEndpoints.Endpoint {
			if existingEndpoint.Path == desiredEndpoint.Path {
				endpointExists = i
				break
			}
		}

		// Let's check the operationID info, metrics...
		if !(endpointExists != -1 && compareEndpoint(desiredEndpoint, existingEndpoints.Endpoint[endpointExists])) {

			for _, operationid := range desiredEndpoint.OperationIDs {
				for _, desiredMetric := range operationid.Metrics {
					// Get the metric id from the metric Name
					metric, err := getMetricFromMetricName(c, accessToken, service.ID, desiredMetric.Metric)
					if err != nil {
						panic(err)
					}

					// Construct the mapping rule desired.
					mappingRule := client.MappingRule{
						XMLName: xml.Name{
							Space: "",
							Local: "",
						},
						ID:         "",
						MetricID:   metric.ID,
						Pattern:    desiredEndpoint.Path,
						HTTPMethod: operationid.Method,
						Delta:      strconv.Itoa(int(desiredMetric.Increment)),
						CreatedAt:  "",
						UpdatedAt:  "",
					}
					mappingRuleAlreadyExists := false

					// Avoid duplicated mappingRules
					for _, existingMappingRule := range existingMappingRules.MappingRules {
						if (mappingRule.Pattern == existingMappingRule.Pattern && mappingRule.HTTPMethod == existingMappingRule.HTTPMethod) &&
							(mappingRule.Delta == existingMappingRule.Delta) && (mappingRule.MetricID == existingMappingRule.MetricID) {
							mappingRuleAlreadyExists = true
							break
						}
					}

					if !mappingRuleAlreadyExists {
						mappingRulesToCreate = append(mappingRulesToCreate, mappingRule)
					}
				}
			}
		}
	}
	return mappingRulesToCreate
}

// getDesiredMetricsButNonExistingInSystem returns a list of metrics that are desired but not in system
func getDesiredMetricsButNonExistingInSystem(existingMetrics client.MetricList, desiredEndpoints Endpoints) []Metric {

	var metricsToCreate []Metric

	for _, endpoint := range desiredEndpoints.Endpoint {
		for _, operationId := range endpoint.OperationIDs {
			for _, desiredMetric := range operationId.Metrics {
				metricExists := false
				for _, existingMetric := range existingMetrics.Metrics {
					if existingMetric.SystemName == desiredMetric.Metric {
						metricExists = true
						break
					}
				}

				if !metricExists {
					metricsToCreate = append(metricsToCreate, desiredMetric)
				}
			}
		}
	}

	return metricsToCreate
}
