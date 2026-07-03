package templates

import "sort"

type Template struct {
	ID          string
	Name        string
	Description string
	Services    []string
}

func Builtins() []Template {
	items := []Template{
		{ID: "spring-boot-starter", Name: "Spring Boot Starter", Description: "PostgreSQL and Redis for a typical Spring Boot API.", Services: []string{"postgres", "redis"}},
		{ID: "microservices", Name: "Microservices", Description: "Data, messaging, metrics, dashboards, and tracing.", Services: []string{"postgres", "redis", "kafka", "prometheus", "grafana", "zipkin"}},
		{ID: "node-api", Name: "Node.js API", Description: "MongoDB and Redis for Node/Nest APIs.", Services: []string{"mongo", "redis"}},
		{ID: "search-platform", Name: "Search Platform", Description: "Elasticsearch with Kibana.", Services: []string{"elasticsearch", "kibana"}},
		{ID: "authentication", Name: "Authentication", Description: "Keycloak backed by PostgreSQL.", Services: []string{"keycloak", "postgres"}},
		{ID: "cloud-development", Name: "Cloud Development", Description: "Local cloud and S3-compatible object storage.", Services: []string{"localstack", "minio"}},
		{ID: "ai-development", Name: "AI Development", Description: "Ollama runtime with Open WebUI.", Services: []string{"ollama", "open-webui"}},
		{ID: "developer-essentials", Name: "Developer Essentials", Description: "Daily backend dependencies and admin tools.", Services: []string{"postgres", "redis", "rabbitmq", "minio", "mailpit", "adminer"}},
		{ID: "observability", Name: "Observability", Description: "Metrics, dashboards, logs, and traces.", Services: []string{"prometheus", "grafana", "loki", "tempo"}},
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func Get(id string) (Template, bool) {
	for _, template := range Builtins() {
		if template.ID == id || template.Name == id {
			return template, true
		}
	}
	return Template{}, false
}
