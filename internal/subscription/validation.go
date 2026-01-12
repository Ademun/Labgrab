package subscription

func ValidateLabType(labType LabType) bool {
	return labType == LabTypeDefence || labType == LabTypePerformance
}

func ValidateLabTopic(labTopic LabTopic) bool {
	return labTopic == LabTopicVirtual || labTopic == LabTopicElectricity || labTopic == LabTopicMechanics
}

func ValidateLabNumber(labNumber int) bool {
	return labNumber > 0 && labNumber < 256
}
