package source

const (
	// Annotation for setting Policy name
	policyServiceAnnotationKey = "port-forward.firewall.hknrd.io/policy"
	// Annotation for setting External Zone name
	externalZoneServiceAnnotationKey = "port-forward.firewall.hknrd.io/external-zone"
	// Annotation for setting Internal Zone name
	internalZoneServiceAnnotationKey = "port-forward.firewall.hknrd.io/internal-zone"
	// Annotation for setting Rule Set name
	ruleSetServiceAnnotationKey = "port-forward.firewall.hknrd.io/rule-set"
	// Annotation for Port Forward name
	portForwardNameAnnotationKey = "port-forward.firewall.hknrd.io/forwarding-name"
)

func getPolicyFromAnnotation(annotations map[string]string) string {
	return annotations[policyServiceAnnotationKey]
}

func getExternalZoneFromAnnotation(annotations map[string]string) string {
	return annotations[externalZoneServiceAnnotationKey]
}

func getInternalZoneFromAnnotation(annotations map[string]string) string {
	return annotations[internalZoneServiceAnnotationKey]
}

func getRuleSetFromAnnotation(annotations map[string]string) string {
	return annotations[ruleSetServiceAnnotationKey]
}

func getNameFromAnnotation(annotations map[string]string) string {
	return annotations[portForwardNameAnnotationKey]
}
