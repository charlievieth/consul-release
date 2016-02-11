package chaperon_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/consul-release/src/confab/agent"
	"github.com/cloudfoundry-incubator/consul-release/src/confab/chaperon"
	"github.com/cloudfoundry-incubator/consul-release/src/confab/fakes"
	consulagent "github.com/hashicorp/consul/command/agent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var (
		server      chaperon.Server
		timeout     *fakes.Timeout
		agentClient *agent.Client
		controller  *fakes.Controller
		rpcClient   *consulagent.RPCClient
		rpcEndpoint string
	)

	BeforeEach(func() {
		controller = &fakes.Controller{}
		rpcClient = &consulagent.RPCClient{}
		rpcClientConstructor := func(endpoint string) (*consulagent.RPCClient, error) {
			rpcEndpoint = endpoint
			return rpcClient, nil
		}

		server = chaperon.NewServer(controller, rpcClientConstructor)

		timeout = &fakes.Timeout{}
		agentClient = &agent.Client{}
	})

	Describe("Start", func() {
		It("writes the consul configuration file", func() {
			err := server.Start(timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(controller.WriteConsulConfigCall.CallCount).To(Equal(1))
		})

		It("writes the service definitions", func() {
			err := server.Start(timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(controller.WriteServiceDefinitionsCall.CallCount).To(Equal(1))
		})

		It("boots the agent process", func() {
			err := server.Start(timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(controller.BootAgentCall.CallCount).To(Equal(1))
			Expect(controller.BootAgentCall.Receives.Timeout).To(Equal(timeout))
		})

		It("sets up an RPC client", func() {
			err := server.Start(timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(controller.ConfigureServerCall.CallCount).To(Equal(1))
			Expect(controller.ConfigureServerCall.Receives.RPCClient).To(Equal(rpcClient))
			Expect(rpcEndpoint).To(Equal("localhost:8400"))
		})

		It("configures the server", func() {
			err := server.Start(timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(controller.ConfigureServerCall.CallCount).To(Equal(1))
			Expect(controller.ConfigureServerCall.Receives.Timeout).To(Equal(timeout))
		})

		Context("failure cases", func() {
			Context("when writing the consul config file fails", func() {
				It("returns an error", func() {
					controller.WriteConsulConfigCall.Returns.Error = errors.New("failed to write consul config")

					err := server.Start(timeout)
					Expect(err).To(MatchError(errors.New("failed to write consul config")))
				})
			})

			Context("when writing the service definitions fails", func() {
				It("returns an error", func() {
					controller.WriteServiceDefinitionsCall.Returns.Error = errors.New("failed to write service definitions")

					err := server.Start(timeout)
					Expect(err).To(MatchError(errors.New("failed to write service definitions")))
				})
			})

			Context("when booting the agent fails", func() {
				It("returns an error", func() {
					controller.BootAgentCall.Returns.Error = errors.New("failed to boot agent")

					err := server.Start(timeout)
					Expect(err).To(MatchError(errors.New("failed to boot agent")))
				})
			})

			Context("when constructing an RPC client fails", func() {
				It("returns an error", func() {
					server = chaperon.NewServer(controller, func(string) (*consulagent.RPCClient, error) {
						return nil, errors.New("failed to create rpc client")
					})

					err := server.Start(timeout)
					Expect(err).To(MatchError(errors.New("failed to create rpc client")))
				})
			})

			Context("when configuring the server fails", func() {
				It("returns an error", func() {
					controller.ConfigureServerCall.Returns.Error = errors.New("failed to configure server")

					err := server.Start(timeout)
					Expect(err).To(MatchError(errors.New("failed to configure server")))
				})
			})
		})
	})

	Describe("Stop", func() {
		It("sets up an RPC client", func() {
			err := server.Stop()
			Expect(err).NotTo(HaveOccurred())
			Expect(controller.StopAgentCall.CallCount).To(Equal(1))
			Expect(controller.StopAgentCall.Receives.RPCClient).To(Equal(rpcClient))
			Expect(rpcEndpoint).To(Equal("localhost:8400"))
		})

		Context("failure cases", func() {
			Context("when constructing an RPC client fails", func() {
				It("returns an error", func() {
					server = chaperon.NewServer(controller, func(string) (*consulagent.RPCClient, error) {
						return nil, errors.New("failed to create rpc client")
					})

					err := server.Stop()
					Expect(err).To(MatchError(errors.New("failed to create rpc client")))
					Expect(controller.StopAgentCall.CallCount).To(Equal(1))
				})
			})
		})
	})
})
