package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	cf "github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotal-cf/brokerapi"
)

// Provision forwards on a service instance provision request to the backend Cloud Foundry API
func (bkr *MarketplaceBrokerImpl) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (spec brokerapi.ProvisionedServiceSpec, err error) {
	bkr.Logger.Info("provision.start", lager.Data{"instanceID": instanceID})
	// var params map[string]interface{}
	fmt.Printf("Parameters: %#v\n", details.RawParameters)
	req := cf.ServiceInstanceRequest{
		Name:            instanceID,
		ServicePlanGuid: details.PlanID,
		SpaceGuid:       bkr.CF.SpaceGUID,
		// Parameters:      details.RawParameters,
	}

	cfclient, err := bkr.CF.Client()
	if err != nil {
		return
	}
	svcInstance, err := cfclient.CreateServiceInstance(req)
	if err != nil {
		return
	}

	spec.DashboardURL = svcInstance.DashboardUrl
	spec.IsAsync = svcInstance.LastOperation.State == "in progress"
	spec.OperationData = "provision::" + svcInstance.Guid

	if svcInstance.LastOperation.State == "failed" {
		err = fmt.Errorf("cf failed to provision: %s", svcInstance.LastOperation.Description)
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	bkr.Logger.Info("provision.end", lager.Data{
		"instanceID":          instanceID,
		"async":               spec.IsAsync,
		"operation-data":      spec.OperationData,
		"last-op.status":      svcInstance.LastOperation.State,
		"last-op.description": svcInstance.LastOperation.Description,
		"error":               errMsg,
	})

	return
}
