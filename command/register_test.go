package command_test

import (
    "errors"

    "code.cloudfoundry.org/cli/plugin/models"
    "github.com/pivotal-cf/metric-registrar-cli/command"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/onsi/gomega/types"
)

var _ = Describe("Register", func() {
    Context("RegisterLogFormat", func() {
        It("creates a service", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterLogFormat(cliConnection, "app-name", "format-name")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
                "structured-format-format-name",
                "-l",
                "structured-format://format-name",
            ))

            Expect(cliConnection.cliCommandsCalled).To(receiveBindService(
                "app-name",
                "structured-format-format-name",
            ))
        })

        It("doesn't create a service if service already present", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesResult = []plugin_models.GetServices_Model{
                {Name: "structured-format-config"},
            }

            err := command.RegisterLogFormat(cliConnection, "app-name", "config")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if getting the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesError = errors.New("error")

            Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "create-user-provided-service"

            Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if binding fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "bind-service"

            Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService())
            Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
        })
    })

    Context("RegisterMetricsEndpoint", func() {
        It("creates a service given a path", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
                "metrics-endpoint-metrics",
                "-l",
                "metrics-endpoint:///metrics",
            ))

            Expect(cliConnection.cliCommandsCalled).To(receiveBindService(
                "app-name",
                "metrics-endpoint-metrics",
            ))
        })

        It("creates a service given a route", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "app-host.app-domain/app-path/metrics")
            Expect(err).ToNot(HaveOccurred())
            serviceName, endpoint := expectToReceiveCupsArgs(cliConnection.cliCommandsCalled)
            Expect(serviceName).To(HavePrefix("metrics-endpoint-"))
            Expect(endpoint).To(Equal("metrics-endpoint://app-host.app-domain/app-path/metrics"))

            Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
        })

        It("checks the route", func() {
            cliConnection := newMockCliConnection()
            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "not-app-host.app-domain/app-path/metrics")

            Expect(err).To(MatchError("route 'not-app-host.app-domain/app-path/metrics' is not bound to app 'app-name'"))
        })

        It("does not use service names longer than 50 characters", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "very-long-app-name-with-many-characters", "/metrics")
            Expect(err).ToNot(HaveOccurred())
            serviceName, _ := expectToReceiveCupsArgs(cliConnection.cliCommandsCalled)
            Expect(len(serviceName)).To(BeNumerically("<=", 50))

            Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
        })

        It("doesn't create a service if service already present", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesResult = []plugin_models.GetServices_Model{
                {Name: "metrics-endpoint-metrics"},
            }

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("replaces slashes in the service name", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/v2/path/")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
                "metrics-endpoint-v2-path",
                "-l",
                "metrics-endpoint:///v2/path/",
            ))
        })

        It("parses routes without hosts correctly", func() {
            cliConnection := newMockCliConnection()

            cliConnection.getAppResult.Routes = []plugin_models.GetApp_RouteSummary{{
                Host: "",
                Domain: plugin_models.GetApp_DomainFields{
                    Name: "tcp.app-domain",
                },
            }}

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "tcp.app-domain/v2/path/")).To(Succeed())
            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
                "metrics-endpoint-tcp.app-domain-v2-path",
                "-l",
                "metrics-endpoint://tcp.app-domain/v2/path/",
            ))
        })

        It("returns error if getting the app fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getAppError = errors.New("error")

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "app-host.app-domain/app-path/metrics")).ToNot(Succeed())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns an error if parsing the route fails", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "#$%#$%#")
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(HavePrefix("unable to parse requested route:"))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if getting the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesError = errors.New("error")

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics")).ToNot(Succeed())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "create-user-provided-service"

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if binding fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "bind-service"

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService())
            Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
        })
    })
})

func expectToReceiveCupsArgs(called chan []string) (string, string) {
    var args []string
    Expect(called).To(Receive(&args))
    Expect(args).To(HaveLen(4))
    Expect(args[0]).To(Equal("create-user-provided-service"))
    Expect(args[2]).To(Equal("-l"))
    return args[1], args[3]
}

func receiveCreateUserProvidedService(args ...string) types.GomegaMatcher {
    if len(args) == 0 {
        return Receive(ContainElement("create-user-provided-service"))
    }

    return Receive(Equal(append([]string{"create-user-provided-service"}, args...)))
}

func receiveBindService(args ...string) types.GomegaMatcher {
    if len(args) == 0 {
        return Receive(ContainElement("bind-service"))
    }
    return Receive(Equal(append([]string{"bind-service"}, args...)))
}
