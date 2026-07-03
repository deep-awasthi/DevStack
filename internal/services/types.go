package services

import "fmt"

type Category string

const (
	CategoryRelational     Category = "Relational Databases"
	CategoryNoSQL          Category = "NoSQL Databases"
	CategorySearch         Category = "Search Engines"
	CategoryBroker         Category = "Message Brokers"
	CategoryCache          Category = "Cache"
	CategoryObjectStorage  Category = "Object Storage"
	CategoryGateway        Category = "API Gateways"
	CategoryProxy          Category = "Reverse Proxies"
	CategoryObservability  Category = "Observability"
	CategoryLogging        Category = "Logging"
	CategoryMonitoring     Category = "Monitoring"
	CategorySecrets        Category = "Secrets"
	CategoryAuth           Category = "Authentication"
	CategoryDiscovery      Category = "Service Discovery"
	CategoryWorkflow       Category = "Workflow Engines"
	CategoryCloud          Category = "Local Cloud"
	CategoryAI             Category = "AI"
	CategoryTimeSeries     Category = "Time-Series Databases"
	CategoryAnalytics      Category = "Analytics"
	CategoryEventStore     Category = "Event Store"
	CategoryCoordination   Category = "Coordination"
	CategoryFeatureFlags   Category = "Feature Flags"
	CategoryMail           Category = "Mail Testing"
	CategoryRegistry       Category = "Container Registry"
	CategoryFTP            Category = "FTP"
	CategoryDeveloperTools Category = "Developer Tools"
	CategoryTesting        Category = "Testing"
)

type CredentialSpec struct {
	UsernameEnv string
	PasswordEnv string
	DatabaseEnv string
	Defaults    map[string]string
}

type Service struct {
	ID              string
	Name            string
	Category        Category
	Image           string
	Versions        []string
	DefaultVersion  string
	Ports           []Port
	Volumes         []Volume
	Environment     map[string]string
	Command         []string
	DependsOn       []string
	Healthcheck     []string
	Credentials     CredentialSpec
	ConnectionHints []ConnectionHint
	Description     string
}

type Port struct {
	Name      string
	Internal  int
	Preferred int
	Protocol  string
}

func (p Port) Proto() string {
	if p.Protocol == "" {
		return "tcp"
	}
	return p.Protocol
}

type Volume struct {
	Name string
	Path string
}

type ConnectionHint struct {
	Label     string
	Pattern   string
	Sensitive bool
}

type ResolvedService struct {
	Service      Service
	Version      string
	Image        string
	Container    string
	Ports        map[string]int
	Credentials  map[string]string
	Network      string
	VolumeNames  map[string]string
	Dependencies []string
}

func (s Service) ImageFor(version string) string {
	if version == "" {
		version = s.DefaultVersion
	}
	return fmt.Sprintf("%s:%s", s.Image, version)
}

func (s Service) SupportsVersion(version string) bool {
	if version == "" {
		return true
	}
	for _, candidate := range s.Versions {
		if candidate == version {
			return true
		}
	}
	return false
}
