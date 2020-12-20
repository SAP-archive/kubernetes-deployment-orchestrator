package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/google/uuid"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/spf13/pflag"
	"github.com/wonderix/shalm/pkg/shalm"
	"github.com/wonderix/shalm/pkg/starutils"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

// OsbConfig -
type OsbConfig struct {
	configfile string
	client     osb.Client
}

// osbClientConfig -
type osbClientConfig struct {
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type osbClientFactory = func() (osb.Client, error)

// AddFlags -
func (o *OsbConfig) AddFlags(flagsSet *pflag.FlagSet) {
	flagsSet.StringVarP(&o.configfile, "osbconfig", "o", "", "OSB configuration file, which includes url, username and password as JSON")
}

func (o *OsbConfig) factory() (osb.Client, error) {
	if o.client != nil {
		return o.client, nil
	}
	if len(o.configfile) == 0 {
		return nil, errors.New("No OSB configuration given")
	}
	content, err := ioutil.ReadFile(o.configfile)
	if err != nil {
		return nil, err
	}
	var config osbClientConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	client, err := osb.NewClient(&osb.ClientConfiguration{
		Name:       "shalm",
		URL:        config.URL,
		APIVersion: osb.Version2_14(),
		AuthConfig: &osb.AuthConfig{
			BasicAuthConfig: &osb.BasicAuthConfig{
				Username: config.Username,
				Password: config.Password,
			}},
		TimeoutSeconds: 60,
		Verbose:        true,
	})
	if err != nil {
		return nil, err
	}
	o.client = client
	return o.client, nil
}

type osbBindingBackend struct {
	clientFactory     osbClientFactory
	service           string
	plan              string
	parameters        map[string]interface{}
	bindingParameters map[string]interface{}
}

var (
	_         shalm.ComplexJewelBackend = (*osbBindingBackend)(nil)
	orgGUID                             = uuid.New().String()
	spaceGUID                           = uuid.New().String()
)

func (v *osbBindingBackend) Name() string {
	return "binding"
}

func (v *osbBindingBackend) Keys() map[string]shalm.JewelValue {
	return map[string]shalm.JewelValue{
		"credentials": {Name: "credentials", Converter: func(data []byte) (starlark.Value, error) {
			credentials := map[string]interface{}{}
			err := json.Unmarshal(data, &credentials)
			if err != nil {
				return starlark.None, err
			}
			return starutils.ToStarlark(credentials), nil
		}},
	}
}

func (v *osbBindingBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	_, ok := m["credentials"]
	if ok {
		return m, nil
	}
	client, err := v.clientFactory()
	if err != nil {
		return nil, err
	}
	catalog, err := client.GetCatalog()
	if err != nil {
		return nil, fmt.Errorf("Error reading OSB catalog:'%s'", err.Error())
	}
	var selectedService *osb.Service
	var selectedPlan *osb.Plan
	for _, service := range catalog.Services {
		if service.Name == v.service {
			selectedService = &service
			for _, plan := range service.Plans {
				if plan.Name == v.plan {
					selectedPlan = &plan
				}
			}
		}
	}
	if selectedService == nil {
		return nil, fmt.Errorf("Service '%s' not found", v.service)
	}
	if selectedPlan == nil {
		return nil, fmt.Errorf("Plan '%s' not found for service '%s'", v.plan, v.service)
	}
	instanceID := uuid.New().String()
	provisionResponse, err := client.ProvisionInstance(&osb.ProvisionRequest{
		InstanceID:          instanceID,
		ServiceID:           selectedService.ID,
		PlanID:              selectedPlan.ID,
		AcceptsIncomplete:   true,
		OrganizationGUID:    orgGUID,
		SpaceGUID:           spaceGUID,
		Parameters:          v.parameters,
		Context:             nil,
		OriginatingIdentity: nil,
	})
	if err != nil {
		return nil, err
	}
	if provisionResponse.Async {
		for {
			pollResponse, err := client.PollLastOperation(&osb.LastOperationRequest{
				InstanceID:          instanceID,
				ServiceID:           &selectedService.ID,
				PlanID:              &selectedPlan.ID,
				OperationKey:        provisionResponse.OperationKey,
				OriginatingIdentity: nil,
			})
			if err != nil {
				return nil, fmt.Errorf("Provisioning of service '%s' with plan '%s' failed: %s", v.plan, v.service, err.Error())
			}
			switch pollResponse.State {
			case osb.StateSucceeded:
				break
			case osb.StateFailed:
				return nil, fmt.Errorf("Provisioning of service '%s' with plan '%s' failed", v.plan, v.service)
			case osb.StateInProgress:
				delay := 10 * time.Second
				if pollResponse.PollDelay != nil {
					delay = *pollResponse.PollDelay
				}
				time.Sleep(delay)
			}
		}
	}
	bindingID := uuid.New().String()
	bindResponse, err := client.Bind(&osb.BindRequest{
		BindingID:           bindingID,
		InstanceID:          instanceID,
		ServiceID:           selectedService.ID,
		PlanID:              selectedPlan.ID,
		AcceptsIncomplete:   false,
		Parameters:          v.bindingParameters,
		Context:             nil,
		OriginatingIdentity: nil,
	})
	if err != nil {
		return nil, err
	}
	credentials, err := json.Marshal(bindResponse.Credentials)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		"instance-id": []byte(instanceID),
		"binding-id":  []byte(bindingID),
		"service-id":  []byte(selectedService.ID),
		"plan-id":     []byte(selectedPlan.ID),
		"credentials": []byte(credentials),
	}, nil
}

func (v *osbBindingBackend) Template() (map[string][]byte, error) {
	return map[string][]byte{
		"credentials": []byte("{}"),
	}, nil
}

func extract(m map[string][]byte, key string) string {
	data, ok := m[key]
	if ok {
		return string(data)
	}
	return ""
}

func (v *osbBindingBackend) Delete(m map[string][]byte) error {
	client, err := v.clientFactory()
	if err != nil {
		return err
	}
	instanceID := extract(m, "instance-id")
	bindingID := extract(m, "binding-id")
	serviceID := extract(m, "service-id")
	planID := extract(m, "plan-id")
	if len(instanceID) == 0 || len(bindingID) == 0 || len(serviceID) == 0 || len(planID) == 0 {
		return fmt.Errorf("Missing configuration in '%s'", v.Name())
	}
	_, err = client.Unbind(&osb.UnbindRequest{
		InstanceID:          instanceID,
		BindingID:           bindingID,
		AcceptsIncomplete:   false,
		ServiceID:           serviceID,
		PlanID:              planID,
		OriginatingIdentity: nil,
	})
	if err != nil {
		return err
	}
	deprovisionResponse, err := client.DeprovisionInstance(&osb.DeprovisionRequest{
		InstanceID:          instanceID,
		AcceptsIncomplete:   true,
		ServiceID:           serviceID,
		PlanID:              planID,
		OriginatingIdentity: nil,
	})
	if deprovisionResponse.Async {
		for {
			pollResponse, err := client.PollLastOperation(&osb.LastOperationRequest{
				InstanceID:          instanceID,
				ServiceID:           &serviceID,
				PlanID:              &planID,
				OperationKey:        deprovisionResponse.OperationKey,
				OriginatingIdentity: nil,
			})
			if err != nil {
				return fmt.Errorf("Deprovisioning of service '%s' with plan '%s' failed: %s", v.plan, v.service, err.Error())
			}
			switch pollResponse.State {
			case osb.StateSucceeded:
				break
			case osb.StateFailed:
				return fmt.Errorf("Deprovisioning of service '%s' with plan '%s' failed", v.plan, v.service)
			case osb.StateInProgress:
				delay := 10 * time.Second
				if pollResponse.PollDelay != nil {
					delay = *pollResponse.PollDelay
				}
				time.Sleep(delay)
			}
		}
	}
	return nil
}

func makeOsbBindung(clientFactory osbClientFactory) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		c := &osbBindingBackend{clientFactory: clientFactory}
		var name string
		var parameters starlark.IterableMapping
		var bindingParameters starlark.IterableMapping
		err := starlark.UnpackArgs("binding", args, kwargs, "name", &name, "service", &c.service, "plan", &c.plan, "parameters?", &parameters, "binding_parameters", &bindingParameters)
		if err != nil {
			return nil, err
		}
		if parameters != nil {
			c.parameters = starutils.ToGoMap(parameters)
		}
		if bindingParameters != nil {
			c.bindingParameters = starutils.ToGoMap(bindingParameters)
		}
		return shalm.NewJewel(c, name)
	}
}

// OsbAPI -
func OsbAPI(config OsbConfig) starlark.StringDict {
	return starlark.StringDict{
		"osb": &starlarkstruct.Module{
			Name: "osb",
			Members: starlark.StringDict{
				"binding": starlark.NewBuiltin("binding", makeOsbBindung(config.factory)),
			},
		},
	}
}
