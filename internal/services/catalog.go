package services

import (
	"sort"
	"strings"
)

type Catalog struct {
	services map[string]Service
	order    []string
}

func NewCatalog() Catalog {
	c := Catalog{services: map[string]Service{}}
	for _, service := range builtinServices() {
		c.Add(service)
	}
	return c
}

func (c *Catalog) Add(service Service) {
	service.ID = strings.ToLower(service.ID)
	if service.DefaultVersion == "" {
		service.DefaultVersion = "latest"
	}
	if len(service.Versions) == 0 {
		service.Versions = []string{service.DefaultVersion}
	}
	c.services[service.ID] = service
	c.order = append(c.order, service.ID)
}

func (c Catalog) Get(id string) (Service, bool) {
	service, ok := c.services[strings.ToLower(id)]
	return service, ok
}

func (c Catalog) MustGet(id string) Service {
	service, ok := c.Get(id)
	if !ok {
		panic("unknown service: " + id)
	}
	return service
}

func (c Catalog) All() []Service {
	services := make([]Service, 0, len(c.order))
	for _, id := range c.order {
		services = append(services, c.services[id])
	}
	return services
}

func (c Catalog) Search(query string) []Service {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return c.All()
	}
	var matches []Service
	for _, service := range c.All() {
		haystack := strings.ToLower(service.ID + " " + service.Name + " " + string(service.Category) + " " + service.Description + " " + service.Image)
		if strings.Contains(haystack, query) {
			matches = append(matches, service)
		}
	}
	return matches
}

func (c Catalog) Categories() []Category {
	seen := map[Category]bool{}
	var categories []Category
	for _, service := range c.All() {
		if !seen[service.Category] {
			categories = append(categories, service.Category)
			seen[service.Category] = true
		}
	}
	return categories
}

func builtinServices() []Service {
	services := []Service{
		postgres(),
		mysql(),
		mariadb(),
		sqlserver(),
		oracle(),
		db2(),
		mongo(),
		redis(),
		couchdb(),
		cassandra(),
		scylladb(),
		neo4j(),
		arango(),
		couchbase(),
		dynamodb(),
		hbase(),
		elasticsearch(),
		opensearch(),
		meilisearch(),
		typesense(),
		solr(),
		kafka(),
		rabbitmq(),
		artemis(),
		pulsar(),
		rocketmq(),
		nats(),
		redpanda(),
		memcached(),
		minio(),
		kong(),
		apisix(),
		traefik(),
		envoy(),
		nginx(),
		caddy(),
		prometheus(),
		grafana(),
		loki(),
		tempo(),
		jaeger(),
		zipkin(),
		alertmanager(),
		otelCollector(),
		logstash(),
		kibana(),
		fluentbit(),
		fluentd(),
		vector(),
		cadvisor(),
		nodeExporter(),
		vault(),
		keycloak(),
		oryHydra(),
		oryKratos(),
		consul(),
		etcd(),
		temporal(),
		airflow(),
		prefect(),
		localstack(),
		azurite(),
		gcloudEmulators(),
		ollama(),
		openWebUI(),
		influxdb(),
		victoriametrics(),
		clickhouse(),
		druid(),
		eventstoredb(),
		zookeeper(),
		unleash(),
		mailpit(),
		mailhog(),
		registry(),
		harbor(),
		sftpgo(),
		pgadmin(),
		adminer(),
		mongoExpress(),
		redisInsight(),
		kafkaUI(),
		wiremock(),
		mockserver(),
		toxiproxy(),
	}
	sort.SliceStable(services, func(i, j int) bool {
		if services[i].Category == services[j].Category {
			return services[i].Name < services[j].Name
		}
		return services[i].Category < services[j].Category
	})
	return services
}

func service(id, name string, category Category, image string, versions []string, port int, description string) Service {
	ports := []Port(nil)
	if port > 0 {
		ports = []Port{{Name: "default", Internal: port, Preferred: port}}
	}
	return Service{
		ID:             id,
		Name:           name,
		Category:       category,
		Image:          image,
		Versions:       versions,
		DefaultVersion: versions[0],
		Ports:          ports,
		Environment:    map[string]string{},
		Description:    description,
	}
}

func postgres() Service {
	return Service{
		ID:             "postgres",
		Name:           "PostgreSQL",
		Category:       CategoryRelational,
		Image:          "postgres",
		Versions:       []string{"latest", "17", "16", "15", "14"},
		DefaultVersion: "17",
		Ports:          []Port{{Name: "postgres", Internal: 5432, Preferred: 5432}},
		Volumes:        []Volume{{Name: "data", Path: "/var/lib/postgresql/data"}},
		Environment:    map[string]string{"POSTGRES_DB": "{{database}}", "POSTGRES_USER": "{{username}}", "POSTGRES_PASSWORD": "{{password}}"},
		Credentials:    CredentialSpec{Defaults: map[string]string{"username": "postgres", "database": "app"}},
		ConnectionHints: []ConnectionHint{
			{Label: "Host", Pattern: "localhost"},
			{Label: "Port", Pattern: "{{port:postgres}}"},
			{Label: "Database", Pattern: "{{database}}"},
			{Label: "Username", Pattern: "{{username}}"},
			{Label: "Password", Pattern: "{{password}}", Sensitive: true},
			{Label: "JDBC URL", Pattern: "jdbc:postgresql://localhost:{{port:postgres}}/{{database}}"},
			{Label: "URI", Pattern: "postgres://{{username}}:{{password}}@localhost:{{port:postgres}}/{{database}}?sslmode=disable", Sensitive: true},
		},
		Description: "Reliable open-source relational database.",
	}
}

func mysql() Service {
	return Service{
		ID:              "mysql",
		Name:            "MySQL",
		Category:        CategoryRelational,
		Image:           "mysql",
		Versions:        []string{"latest", "9", "8.4", "8.0"},
		DefaultVersion:  "8.4",
		Ports:           []Port{{Name: "mysql", Internal: 3306, Preferred: 3306}},
		Volumes:         []Volume{{Name: "data", Path: "/var/lib/mysql"}},
		Environment:     map[string]string{"MYSQL_DATABASE": "{{database}}", "MYSQL_USER": "{{username}}", "MYSQL_PASSWORD": "{{password}}", "MYSQL_ROOT_PASSWORD": "{{root_password}}"},
		Credentials:     CredentialSpec{Defaults: map[string]string{"username": "app", "database": "app"}},
		ConnectionHints: []ConnectionHint{{Label: "Host", Pattern: "localhost"}, {Label: "Port", Pattern: "{{port:mysql}}"}, {Label: "Database", Pattern: "{{database}}"}, {Label: "Username", Pattern: "{{username}}"}, {Label: "Password", Pattern: "{{password}}", Sensitive: true}, {Label: "DSN", Pattern: "{{username}}:{{password}}@tcp(localhost:{{port:mysql}})/{{database}}", Sensitive: true}},
		Description:     "Popular relational database for application development.",
	}
}

func mariadb() Service {
	s := mysql()
	s.ID = "mariadb"
	s.Name = "MariaDB"
	s.Image = "mariadb"
	s.Versions = []string{"latest", "11", "10"}
	s.DefaultVersion = "11"
	s.Description = "Community-driven MySQL-compatible relational database."
	return s
}

func sqlserver() Service {
	return Service{ID: "sqlserver", Name: "Microsoft SQL Server Developer", Category: CategoryRelational, Image: "mcr.microsoft.com/mssql/server", Versions: []string{"2022-latest", "2019-latest"}, DefaultVersion: "2022-latest", Ports: []Port{{Name: "sqlserver", Internal: 1433, Preferred: 1433}}, Environment: map[string]string{"ACCEPT_EULA": "Y", "MSSQL_SA_PASSWORD": "{{password}}"}, Credentials: CredentialSpec{Defaults: map[string]string{"username": "sa"}}, ConnectionHints: []ConnectionHint{{Label: "Host", Pattern: "localhost"}, {Label: "Port", Pattern: "{{port:sqlserver}}"}, {Label: "Username", Pattern: "sa"}, {Label: "Password", Pattern: "{{password}}", Sensitive: true}, {Label: "JDBC URL", Pattern: "jdbc:sqlserver://localhost:{{port:sqlserver}};encrypt=false"}}, Description: "SQL Server Developer Edition for local use."}
}

func oracle() Service {
	return service("oracle", "Oracle Database Free", CategoryRelational, "gvenzl/oracle-free", []string{"latest", "23-slim"}, 1521, "Oracle Database Free container image.")
}

func db2() Service {
	return service("db2", "IBM Db2 Community", CategoryRelational, "icr.io/db2_community/db2", []string{"latest"}, 50000, "IBM Db2 Community Edition database.")
}

func mongo() Service {
	return Service{ID: "mongo", Name: "MongoDB", Category: CategoryNoSQL, Image: "mongo", Versions: []string{"latest", "8", "7", "6"}, DefaultVersion: "8", Ports: []Port{{Name: "mongo", Internal: 27017, Preferred: 27017}}, Volumes: []Volume{{Name: "data", Path: "/data/db"}}, Environment: map[string]string{"MONGO_INITDB_ROOT_USERNAME": "{{username}}", "MONGO_INITDB_ROOT_PASSWORD": "{{password}}"}, Credentials: CredentialSpec{Defaults: map[string]string{"username": "root", "database": "app"}}, ConnectionHints: []ConnectionHint{{Label: "Host", Pattern: "localhost"}, {Label: "Port", Pattern: "{{port:mongo}}"}, {Label: "Username", Pattern: "{{username}}"}, {Label: "Password", Pattern: "{{password}}", Sensitive: true}, {Label: "Connection URI", Pattern: "mongodb://{{username}}:{{password}}@localhost:{{port:mongo}}/{{database}}?authSource=admin", Sensitive: true}}, Description: "Document database with flexible JSON-style data model."}
}

func redis() Service {
	return Service{ID: "redis", Name: "Redis", Category: CategoryCache, Image: "redis", Versions: []string{"latest", "8", "7"}, DefaultVersion: "8", Ports: []Port{{Name: "redis", Internal: 6379, Preferred: 6379}}, Volumes: []Volume{{Name: "data", Path: "/data"}}, Command: []string{"redis-server", "--requirepass", "{{password}}", "--appendonly", "yes"}, Credentials: CredentialSpec{Defaults: map[string]string{}}, ConnectionHints: []ConnectionHint{{Label: "Host", Pattern: "localhost"}, {Label: "Port", Pattern: "{{port:redis}}"}, {Label: "Password", Pattern: "{{password}}", Sensitive: true}, {Label: "Redis URL", Pattern: "redis://:{{password}}@localhost:{{port:redis}}/0", Sensitive: true}}, Description: "In-memory data store commonly used for cache and queues."}
}

func couchdb() Service {
	return service("couchdb", "CouchDB", CategoryNoSQL, "couchdb", []string{"latest", "3"}, 5984, "JSON document database with HTTP API.")
}
func cassandra() Service {
	return service("cassandra", "Cassandra", CategoryNoSQL, "cassandra", []string{"latest", "5", "4"}, 9042, "Wide-column distributed database.")
}
func scylladb() Service {
	return service("scylladb", "ScyllaDB", CategoryNoSQL, "scylladb/scylla", []string{"latest", "6", "5"}, 9042, "High-performance Cassandra-compatible database.")
}
func neo4j() Service {
	return service("neo4j", "Neo4j", CategoryNoSQL, "neo4j", []string{"latest", "5"}, 7474, "Graph database with Cypher query language.")
}
func arango() Service {
	return service("arangodb", "ArangoDB", CategoryNoSQL, "arangodb", []string{"latest", "3"}, 8529, "Multi-model graph, document, and search database.")
}
func couchbase() Service {
	return service("couchbase", "Couchbase Community", CategoryNoSQL, "couchbase", []string{"community", "latest"}, 8091, "Distributed document database and cache.")
}
func dynamodb() Service {
	return service("dynamodb-local", "DynamoDB Local", CategoryCloud, "amazon/dynamodb-local", []string{"latest"}, 8000, "Local DynamoDB-compatible service.")
}
func hbase() Service {
	return service("hbase", "HBase", CategoryNoSQL, "harisekhon/hbase", []string{"latest"}, 16010, "Apache HBase local development image.")
}
func elasticsearch() Service {
	return service("elasticsearch", "Elasticsearch", CategorySearch, "docker.elastic.co/elasticsearch/elasticsearch", []string{"8.15.0", "8.14.3"}, 9200, "Distributed search and analytics engine.")
}
func opensearch() Service {
	return service("opensearch", "OpenSearch", CategorySearch, "opensearchproject/opensearch", []string{"latest", "2"}, 9200, "Open-source search and observability engine.")
}
func meilisearch() Service {
	return service("meilisearch", "Meilisearch", CategorySearch, "getmeili/meilisearch", []string{"latest", "v1.10"}, 7700, "Fast developer-friendly search engine.")
}
func typesense() Service {
	return service("typesense", "Typesense", CategorySearch, "typesense/typesense", []string{"latest", "27.1"}, 8108, "Typo-tolerant search engine.")
}
func solr() Service {
	return service("solr", "Apache Solr", CategorySearch, "solr", []string{"latest", "9"}, 8983, "Apache Lucene-based search platform.")
}
func kafka() Service {
	return service("kafka", "Apache Kafka", CategoryBroker, "apache/kafka", []string{"latest", "3.8.0"}, 9092, "Distributed event streaming platform.")
}
func rabbitmq() Service {
	return service("rabbitmq", "RabbitMQ", CategoryBroker, "rabbitmq", []string{"3-management", "latest", "4-management"}, 5672, "AMQP message broker with optional management UI.")
}
func artemis() Service {
	return service("artemis", "Apache ActiveMQ Artemis", CategoryBroker, "apache/activemq-artemis", []string{"latest"}, 61616, "Multi-protocol message broker.")
}
func pulsar() Service {
	return service("pulsar", "Apache Pulsar", CategoryBroker, "apachepulsar/pulsar", []string{"latest", "3"}, 6650, "Cloud-native distributed messaging and streaming.")
}
func rocketmq() Service {
	return service("rocketmq", "Apache RocketMQ", CategoryBroker, "apache/rocketmq", []string{"latest", "5"}, 9876, "Distributed messaging and streaming platform.")
}
func nats() Service {
	return service("nats", "NATS", CategoryBroker, "nats", []string{"latest", "2"}, 4222, "Lightweight high-performance messaging system.")
}
func redpanda() Service {
	return service("redpanda", "Redpanda", CategoryBroker, "redpandadata/redpanda", []string{"latest", "v24.2.1"}, 9092, "Kafka-compatible streaming platform.")
}
func memcached() Service {
	return service("memcached", "Memcached", CategoryCache, "memcached", []string{"latest", "1.6"}, 11211, "Simple high-performance memory object cache.")
}
func minio() Service {
	return Service{ID: "minio", Name: "MinIO", Category: CategoryObjectStorage, Image: "minio/minio", Versions: []string{"latest"}, DefaultVersion: "latest", Ports: []Port{{Name: "api", Internal: 9000, Preferred: 9000}, {Name: "console", Internal: 9001, Preferred: 9001}}, Volumes: []Volume{{Name: "data", Path: "/data"}}, Environment: map[string]string{"MINIO_ROOT_USER": "{{access_key}}", "MINIO_ROOT_PASSWORD": "{{secret_key}}"}, Command: []string{"server", "/data", "--console-address", ":9001"}, Credentials: CredentialSpec{Defaults: map[string]string{"access_key": "minioadmin"}}, ConnectionHints: []ConnectionHint{{Label: "Endpoint", Pattern: "http://localhost:{{port:api}}"}, {Label: "Console URL", Pattern: "http://localhost:{{port:console}}"}, {Label: "Access Key", Pattern: "{{access_key}}"}, {Label: "Secret Key", Pattern: "{{secret_key}}", Sensitive: true}}, Description: "S3-compatible object storage."}
}
func kong() Service {
	return service("kong", "Kong Gateway OSS", CategoryGateway, "kong", []string{"latest", "3"}, 8000, "OSS API gateway.")
}
func apisix() Service {
	return service("apisix", "Apache APISIX", CategoryGateway, "apache/apisix", []string{"latest", "3"}, 9080, "Dynamic cloud-native API gateway.")
}
func traefik() Service {
	return service("traefik", "Traefik", CategoryGateway, "traefik", []string{"latest", "3"}, 80, "Modern reverse proxy and load balancer.")
}
func envoy() Service {
	return service("envoy", "Envoy Proxy", CategoryGateway, "envoyproxy/envoy", []string{"latest", "v1.31-latest"}, 10000, "Cloud-native edge and service proxy.")
}
func nginx() Service {
	return service("nginx", "NGINX", CategoryProxy, "nginx", []string{"latest", "1.27", "alpine"}, 80, "HTTP server and reverse proxy.")
}
func caddy() Service {
	return service("caddy", "Caddy", CategoryProxy, "caddy", []string{"latest", "2"}, 80, "Automatic HTTPS web server and proxy.")
}
func prometheus() Service {
	return service("prometheus", "Prometheus", CategoryObservability, "prom/prometheus", []string{"latest", "v2.54.1"}, 9090, "Metrics and alerting toolkit.")
}
func grafana() Service {
	return service("grafana", "Grafana", CategoryObservability, "grafana/grafana", []string{"latest", "11"}, 3000, "Observability dashboards.")
}
func loki() Service {
	return service("loki", "Loki", CategoryObservability, "grafana/loki", []string{"latest", "3"}, 3100, "Log aggregation system.")
}
func tempo() Service {
	return service("tempo", "Tempo", CategoryObservability, "grafana/tempo", []string{"latest", "2"}, 3200, "Distributed tracing backend.")
}
func jaeger() Service {
	return service("jaeger", "Jaeger", CategoryObservability, "jaegertracing/all-in-one", []string{"latest", "1.60"}, 16686, "Distributed tracing UI and collector.")
}
func zipkin() Service {
	return service("zipkin", "Zipkin", CategoryObservability, "openzipkin/zipkin", []string{"latest", "3"}, 9411, "Distributed tracing system.")
}
func alertmanager() Service {
	return service("alertmanager", "Alertmanager", CategoryObservability, "prom/alertmanager", []string{"latest", "v0.27.0"}, 9093, "Prometheus alert routing service.")
}
func otelCollector() Service {
	return service("otel-collector", "OpenTelemetry Collector", CategoryObservability, "otel/opentelemetry-collector", []string{"latest", "0.109.0"}, 4317, "Vendor-neutral telemetry collector.")
}
func logstash() Service {
	return service("logstash", "Logstash", CategoryLogging, "docker.elastic.co/logstash/logstash", []string{"8.15.0"}, 5044, "Data processing pipeline for logs and events.")
}
func kibana() Service {
	return service("kibana", "Kibana", CategoryLogging, "docker.elastic.co/kibana/kibana", []string{"8.15.0"}, 5601, "Elastic search and analytics UI.")
}
func fluentbit() Service {
	return service("fluent-bit", "Fluent Bit", CategoryLogging, "fluent/fluent-bit", []string{"latest", "3"}, 2020, "Lightweight log processor.")
}
func fluentd() Service {
	return service("fluentd", "Fluentd", CategoryLogging, "fluent/fluentd", []string{"latest", "v1.17"}, 24224, "Unified logging layer.")
}
func vector() Service {
	return service("vector", "Vector", CategoryLogging, "timberio/vector", []string{"latest", "0.40.X-alpine"}, 8686, "Observability data pipeline.")
}
func cadvisor() Service {
	return service("cadvisor", "cAdvisor", CategoryMonitoring, "gcr.io/cadvisor/cadvisor", []string{"latest", "v0.49.1"}, 8080, "Container resource usage analyzer.")
}
func nodeExporter() Service {
	return service("node-exporter", "Node Exporter", CategoryMonitoring, "prom/node-exporter", []string{"latest", "v1.8.2"}, 9100, "Host metrics exporter.")
}
func vault() Service {
	return service("vault", "HashiCorp Vault", CategorySecrets, "hashicorp/vault", []string{"latest", "1.17"}, 8200, "Secrets management server.")
}
func keycloak() Service {
	return service("keycloak", "Keycloak", CategoryAuth, "quay.io/keycloak/keycloak", []string{"latest", "25"}, 8080, "Identity and access management.")
}
func oryHydra() Service {
	return service("ory-hydra", "Ory Hydra", CategoryAuth, "oryd/hydra", []string{"latest", "v2.2"}, 4444, "OAuth2 and OpenID Connect server.")
}
func oryKratos() Service {
	return service("ory-kratos", "Ory Kratos", CategoryAuth, "oryd/kratos", []string{"latest", "v1.2"}, 4433, "Identity and user management system.")
}
func consul() Service {
	return service("consul", "Consul", CategoryDiscovery, "hashicorp/consul", []string{"latest", "1.19"}, 8500, "Service discovery and configuration.")
}
func etcd() Service {
	return service("etcd", "etcd", CategoryDiscovery, "bitnami/etcd", []string{"latest", "3"}, 2379, "Distributed reliable key-value store.")
}
func temporal() Service {
	return service("temporal", "Temporal", CategoryWorkflow, "temporalio/auto-setup", []string{"latest", "1.25"}, 7233, "Durable workflow orchestration platform.")
}
func airflow() Service {
	return service("airflow", "Apache Airflow", CategoryWorkflow, "apache/airflow", []string{"latest", "2.10.0"}, 8080, "Workflow scheduling and orchestration.")
}
func prefect() Service {
	return service("prefect", "Prefect", CategoryWorkflow, "prefecthq/prefect", []string{"latest", "3"}, 4200, "Python workflow orchestration.")
}
func localstack() Service {
	return service("localstack", "LocalStack", CategoryCloud, "localstack/localstack", []string{"latest", "3"}, 4566, "Local AWS cloud emulator.")
}
func azurite() Service {
	return service("azurite", "Azurite", CategoryCloud, "mcr.microsoft.com/azure-storage/azurite", []string{"latest", "3"}, 10000, "Azure Storage emulator.")
}
func gcloudEmulators() Service {
	return service("gcloud-emulators", "Google Cloud Emulator Suite", CategoryCloud, "gcr.io/google.com/cloudsdktool/google-cloud-cli", []string{"latest", "emulators"}, 8085, "Google Cloud SDK emulator container.")
}
func ollama() Service {
	return service("ollama", "Ollama", CategoryAI, "ollama/ollama", []string{"latest"}, 11434, "Local model runtime.")
}
func openWebUI() Service {
	return service("open-webui", "Open WebUI", CategoryAI, "ghcr.io/open-webui/open-webui", []string{"main"}, 8080, "Web UI for local AI runtimes.")
}
func influxdb() Service {
	return service("influxdb", "InfluxDB", CategoryTimeSeries, "influxdb", []string{"latest", "2"}, 8086, "Time-series database.")
}
func victoriametrics() Service {
	return service("victoriametrics", "VictoriaMetrics", CategoryTimeSeries, "victoriametrics/victoria-metrics", []string{"latest", "v1.103.0"}, 8428, "Fast time-series database and monitoring solution.")
}
func clickhouse() Service {
	return service("clickhouse", "ClickHouse", CategoryAnalytics, "clickhouse/clickhouse-server", []string{"latest", "24"}, 8123, "Column-oriented analytics database.")
}
func druid() Service {
	return service("druid", "Apache Druid", CategoryAnalytics, "apache/druid", []string{"latest", "31.0.0"}, 8888, "Real-time analytics database.")
}
func eventstoredb() Service {
	return service("eventstoredb", "EventStoreDB", CategoryEventStore, "eventstore/eventstore", []string{"latest", "24.6.0"}, 2113, "Event-native database.")
}
func zookeeper() Service {
	return service("zookeeper", "ZooKeeper", CategoryCoordination, "zookeeper", []string{"latest", "3.9"}, 2181, "Distributed coordination service.")
}
func unleash() Service {
	return service("unleash", "Unleash", CategoryFeatureFlags, "unleashorg/unleash-server", []string{"latest", "5"}, 4242, "Open-source feature flag server.")
}
func mailpit() Service {
	return service("mailpit", "Mailpit", CategoryMail, "axllent/mailpit", []string{"latest"}, 8025, "Email and SMTP testing tool.")
}
func mailhog() Service {
	return service("mailhog", "MailHog", CategoryMail, "mailhog/mailhog", []string{"latest"}, 8025, "Email testing tool with web UI.")
}
func registry() Service {
	return service("registry", "Docker Registry", CategoryRegistry, "registry", []string{"latest", "2"}, 5000, "Local Docker image registry.")
}
func harbor() Service {
	return service("harbor", "Harbor", CategoryRegistry, "goharbor/harbor-core", []string{"latest", "v2.11.0"}, 8080, "Cloud-native container registry core service.")
}
func sftpgo() Service {
	return service("sftpgo", "SFTPGo", CategoryFTP, "drakkan/sftpgo", []string{"latest", "v2"}, 8080, "SFTP, FTP, WebDAV server.")
}
func pgadmin() Service {
	return service("pgadmin", "pgAdmin", CategoryDeveloperTools, "dpage/pgadmin4", []string{"latest", "8"}, 80, "PostgreSQL administration UI.")
}
func adminer() Service {
	return service("adminer", "Adminer", CategoryDeveloperTools, "adminer", []string{"latest", "4"}, 8080, "Database administration UI.")
}
func mongoExpress() Service {
	return service("mongo-express", "Mongo Express", CategoryDeveloperTools, "mongo-express", []string{"latest", "1"}, 8081, "MongoDB administration UI.")
}
func redisInsight() Service {
	return service("redisinsight", "RedisInsight", CategoryDeveloperTools, "redis/redisinsight", []string{"latest"}, 5540, "Redis management UI.")
}
func kafkaUI() Service {
	return service("kafka-ui", "Kafka UI", CategoryDeveloperTools, "provectuslabs/kafka-ui", []string{"latest"}, 8080, "Kafka cluster management UI.")
}
func wiremock() Service {
	return service("wiremock", "WireMock", CategoryTesting, "wiremock/wiremock", []string{"latest", "3"}, 8080, "HTTP API mocking server.")
}
func mockserver() Service {
	return service("mockserver", "MockServer", CategoryTesting, "mockserver/mockserver", []string{"latest", "5"}, 1080, "Mock HTTP and HTTPS services.")
}
func toxiproxy() Service {
	return service("toxiproxy", "Toxiproxy", CategoryTesting, "shopify/toxiproxy", []string{"latest", "2"}, 8474, "Network fault injection proxy.")
}
