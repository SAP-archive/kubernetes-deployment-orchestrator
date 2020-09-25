package extensions

import (
	"context"
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain"
	"github.com/pivotal-cf/brokerapi/fakes"

	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

var _ = Describe("osb binding", func() {
	serviceID := uuid.New().String()
	planID := uuid.New().String()
	broker := &fakes.AutoFakeServiceBroker{
		ServicesStub: func(context.Context) ([]domain.Service, error) {
			return []domain.Service{{
				ID:   serviceID,
				Name: "service",
				Plans: []domain.ServicePlan{{
					ID:   planID,
					Name: "plan",
				}},
			}}, nil
		},
		ProvisionStub: func(context.Context, string, domain.ProvisionDetails, bool) (domain.ProvisionedServiceSpec, error) {
			return domain.ProvisionedServiceSpec{
				IsAsync:       false,
				OperationData: "",
			}, nil
		},
		DeprovisionStub: func(context.Context, string, domain.DeprovisionDetails, bool) (domain.DeprovisionServiceSpec, error) {
			return domain.DeprovisionServiceSpec{
				IsAsync:       false,
				OperationData: "",
			}, nil
		},
		BindStub: func(context.Context, string, string, domain.BindDetails, bool) (domain.Binding, error) {
			return domain.Binding{
				IsAsync:       false,
				OperationData: "",
				Credentials:   map[string]string{"password": "password", "username": "username"},
			}, nil
		},
		UnbindStub: func(context.Context, string, string, domain.UnbindDetails, bool) (domain.UnbindSpec, error) {
			return domain.UnbindSpec{
				IsAsync:       false,
				OperationData: "",
			}, nil
		},
	}
	handler := brokerapi.New(broker, lager.NewLogger("test"), brokerapi.BrokerCredentials{Username: "username", Password: "password"})
	srv := http.Server{
		Addr:    "localhost:8465",
		Handler: handler,
	}

	BeforeEach(func() {

		go func() {
			_ = srv.ListenAndServe()
		}()
		for {
			time.Sleep(100 * time.Millisecond)

			_, err := http.Get("http://localhost:8465/v2/catalog")
			if err == nil {
				break
			}
		}
	})
	AfterEach(func() {
		err := srv.Shutdown(context.Background())
		Expect(err).NotTo(HaveOccurred())
	})

	It("binding works", func() {
		client, err := osb.NewClient(&osb.ClientConfiguration{
			Name:       "test",
			URL:        "http://localhost:8465",
			APIVersion: osb.Version2_14(),
			AuthConfig: &osb.AuthConfig{
				BasicAuthConfig: &osb.BasicAuthConfig{
					Username: "username",
					Password: "password",
				}},
			TimeoutSeconds: 60,
			Verbose:        false,
		})
		Expect(err).NotTo(HaveOccurred())
		binding := osbBindingBackend{
			clientFactory: func() (osb.Client, error) { return client, nil },
			service:       "service",
			plan:          "plan",
			parameters:    map[string]interface{}{"param1": "value1"},
		}
		m, err := binding.Apply(make(map[string][]byte))
		Expect(broker.ProvisionCallCount()).To(Equal(1))
		Expect(broker.BindCallCount()).To(Equal(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(m["credentials"]).To(Equal([]byte(`{"password":"password","username":"username"}`)))
		err = binding.Delete(m)
		Expect(broker.UnbindCallCount()).To(Equal(1))
		Expect(broker.DeprovisionCallCount()).To(Equal(1))
		Expect(err).NotTo(HaveOccurred())
	})
})
