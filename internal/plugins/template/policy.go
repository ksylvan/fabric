package template

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrRestrictedTemplateFeature = errors.New("restricted template feature")

type ApplyPolicy struct {
	AllowExtensions                 bool
	AllowedPluginNamespaces         map[string]struct{}
	RejectNestedTemplateValues      bool
	DisallowDynamicPluginNamespaces bool
}

var RemotePatternPolicy = ApplyPolicy{
	AllowExtensions: false,
	AllowedPluginNamespaces: map[string]struct{}{
		"text":     {},
		"datetime": {},
	},
	RejectNestedTemplateValues:      true,
	DisallowDynamicPluginNamespaces: true,
}

var literalPluginNamespacePattern = regexp.MustCompile(`\{\{plugin:([^:{]+):`)

func (p *ApplyPolicy) allowsPluginNamespace(namespace string) bool {
	if p == nil || p.AllowedPluginNamespaces == nil {
		return true
	}

	_, allowed := p.AllowedPluginNamespaces[namespace]
	return allowed
}

func newRestrictedTemplateError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrRestrictedTemplateFeature, fmt.Sprintf(format, args...))
}

func validateTemplateVariablesWithPolicy(variables map[string]string, policy *ApplyPolicy) error {
	if policy == nil || !policy.RejectNestedTemplateValues {
		return nil
	}

	for name, value := range variables {
		if strings.Contains(value, "{{") || strings.Contains(value, "}}") {
			return newRestrictedTemplateError(
				"template variable %q cannot contain nested template directives",
				name,
			)
		}
	}

	return nil
}

func ValidateTemplateContentWithPolicy(content string, policy *ApplyPolicy) error {
	if policy == nil {
		return nil
	}

	if !policy.AllowExtensions && strings.Contains(content, "{{ext:") {
		return newRestrictedTemplateError("template extensions are disabled for remote pattern execution")
	}

	if policy.DisallowDynamicPluginNamespaces && strings.Contains(content, "{{plugin:{{") {
		return newRestrictedTemplateError("dynamic template plugin namespaces are disabled for remote pattern execution")
	}

	for _, match := range literalPluginNamespacePattern.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		if !policy.allowsPluginNamespace(match[1]) {
			return newRestrictedTemplateError(
				"template plugin %q is disabled for remote pattern execution",
				match[1],
			)
		}
	}

	return nil
}
