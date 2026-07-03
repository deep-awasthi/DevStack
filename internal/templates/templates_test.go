package templates

import "testing"

func TestGetTemplate(t *testing.T) {
	template, ok := Get("developer-essentials")
	if !ok {
		t.Fatal("expected developer essentials template")
	}
	if len(template.Services) == 0 {
		t.Fatal("expected services in template")
	}
}
