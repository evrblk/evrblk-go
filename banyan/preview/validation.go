package banyan

import (
	"fmt"
	"regexp"
	"slices"
)

const (
	maxWorkflowNameLength  = 128
	maxDescriptionLength   = 1024
	maxMetadataKeyLength   = 128
	maxMetadataValueLength = 512
	maxNumberOfMetadata    = 32
	maxNumberOfSteps       = 64
	maxStepNameLength      = 128
	maxQueueNameLength     = 128
	maxChoiceOptionLength  = 64

	nameRegex          = "^[-_0-9a-zA-Z]*$"
	metadataKeyRegex   = "^[-_0-9a-zA-Z]*$"
	metadataValueRegex = ".*"
)

// ValidateWorkflow validates the workflow:
// - Workflow name is required
// - Workflow steps are required
func ValidateWorkflow(workflow *Workflow) error {
	err := ValidateWorkflowName(workflow.Name, "Workflow.Name")
	if err != nil {
		return err
	}

	err = ValidateDescription(workflow.Description, "Workflow.Description")
	if err != nil {
		return err
	}

	err = ValidateSteps(workflow.Steps, "Workflow.Steps")
	if err != nil {
		return err
	}

	err = ValidateMetadata(workflow.Metadata, "Workflow.Metadata")
	if err != nil {
		return err
	}

	return nil
}

// ValidateSteps checks that:
// - Step names must be unique
// - Step queue name is required
// - Condition step name must be a valid step name
// - Conditions in ANY or ALL do not contain an initial condition
// - Chosen step result must be a valid option
// - Conditions do not contain a terminal step
// - No cycles in steps DAG
// - Terminal step exists
// - Terminal step is reachable from all initial steps
func ValidateSteps(steps []*Step, stepsFieldName string) error {
	if len(steps) == 0 {
		return invalid(stepsFieldName, "must have at least one step")
	}

	if len(steps) > maxNumberOfSteps {
		return invalid(stepsFieldName, fmt.Sprintf("exceeds max number of steps (%d)", maxNumberOfSteps))
	}

	stepNames := make(map[string]*Step)
	for i, step := range steps {
		stepNameFieldName := fmt.Sprintf("%s[%d].Name", stepsFieldName, i)

		err := validateStepName(step.Name, stepNameFieldName)
		if err != nil {
			return err
		}

		if _, ok := stepNames[step.Name]; ok {
			return invalid(stepNameFieldName, fmt.Sprintf("must be unique: '%s'", step.Name))
		}
		stepNames[step.Name] = step
	}

	for i, step := range steps {
		stepFieldName := fmt.Sprintf("%s[%d]", stepsFieldName, i)

		switch stepType := step.StepType.(type) {
		case *Step_Simple:
			err := validateCondition(stepType.Simple.StartsWhen, stepNames, fmt.Sprintf("%s.Simple.StartsWhen", stepFieldName))
			if err != nil {
				return err
			}

			err = ValidateQueueName(stepType.Simple.QueueName, fmt.Sprintf("%s.Simple.QueueName", stepFieldName))
			if err != nil {
				return err
			}

		case *Step_FanOut:
			err := validateCondition(stepType.FanOut.StartsWhen, stepNames, fmt.Sprintf("%s.FanOut.StartsWhen", stepFieldName))
			if err != nil {
				return err
			}

			err = ValidateQueueName(stepType.FanOut.QueueName, fmt.Sprintf("%s.FanOut.QueueName", stepFieldName))
			if err != nil {
				return err
			}

		case *Step_Choice:
			err := validateCondition(stepType.Choice.StartsWhen, stepNames, fmt.Sprintf("%s.Choice.StartsWhen", stepFieldName))
			if err != nil {
				return err
			}

			err = ValidateQueueName(stepType.Choice.QueueName, fmt.Sprintf("%s.Choice.QueueName", stepFieldName))
			if err != nil {
				return err
			}

		case *Step_Parallel:
			fanOutFromFieldName := fmt.Sprintf("%s.Parallel.FanOutFrom", stepFieldName)
			err := validateStepName(stepType.Parallel.FanOutFrom, fanOutFromFieldName)
			if err != nil {
				return err
			}

			fanOutFrom, ok := stepNames[stepType.Parallel.FanOutFrom]
			if !ok {
				return invalid(fanOutFromFieldName, fmt.Sprintf("step not found: '%s'", stepType.Parallel.FanOutFrom))
			}
			if _, ok := fanOutFrom.StepType.(*Step_FanOut); !ok {
				return invalid(fanOutFromFieldName, "step is not a fan out step")
			}

			err = ValidateQueueName(stepType.Parallel.QueueName, fmt.Sprintf("%s.Parallel.QueueName", stepFieldName))
			if err != nil {
				return err
			}

		case *Step_External:
			err := validateCondition(stepType.External.StartsWhen, stepNames, fmt.Sprintf("%s.External.StartsWhen", stepFieldName))
			if err != nil {
				return err
			}

		case *Step_Terminal:
			err := validateCondition(stepType.Terminal.StartsWhen, stepNames, fmt.Sprintf("%s.Terminal.StartsWhen", stepFieldName))
			if err != nil {
				return err
			}
		}
	}

	err := validateDAG(steps, stepNames, stepsFieldName)
	if err != nil {
		return err
	}

	return nil
}

// validateConditions calls validateCondition for each condition in the list
func validateConditions(conditions []*Condition, stepNames map[string]*Step, fieldName string) error {
	for i, condition := range conditions {
		err := validateCondition(condition, stepNames, fmt.Sprintf("%s[%d]", fieldName, i))
		if err != nil {
			return err
		}
	}
	return nil
}

// validateCondition validates a single condition
func validateCondition(condition *Condition, stepNames map[string]*Step, conditionFieldName string) error {
	switch c := condition.ConditionType.(type) {
	case *Condition_All:
		err := validateConditionsDoNotContainInitial(c.All.Conditions, fmt.Sprintf("%s.All.Conditions", conditionFieldName))
		if err != nil {
			return err
		}
		err = validateConditions(c.All.Conditions, stepNames, fmt.Sprintf("%s.All.Conditions", conditionFieldName))
		if err != nil {
			return err
		}

	case *Condition_Any:
		err := validateConditionsDoNotContainInitial(c.Any.Conditions, fmt.Sprintf("%s.Any.Conditions", conditionFieldName))
		if err != nil {
			return err
		}
		err = validateConditions(c.Any.Conditions, stepNames, fmt.Sprintf("%s.Any.Conditions", conditionFieldName))
		if err != nil {
			return err
		}

	case *Condition_Succeeded:
		err := validateConditionStepName(c.Succeeded.StepName, stepNames, fmt.Sprintf("%s.Succeeded.StepName", conditionFieldName))
		if err != nil {
			return err
		}

	case *Condition_Failed:
		err := validateConditionStepName(c.Failed.StepName, stepNames, fmt.Sprintf("%s.Failed.StepName", conditionFieldName))
		if err != nil {
			return err
		}

	case *Condition_Chosen:
		err := validateConditionStepName(c.Chosen.StepName, stepNames, fmt.Sprintf("%s.Chosen.StepName", conditionFieldName))
		if err != nil {
			return err
		}

		err = validateString(c.Chosen.Result, 1, maxChoiceOptionLength, nameRegex, fmt.Sprintf("%s.Chosen.Result", conditionFieldName))
		if err != nil {
			return err
		}

		chosenStep, ok := stepNames[c.Chosen.StepName]
		if !ok {
			return invalid(fmt.Sprintf("%s.Chosen.StepName", conditionFieldName), fmt.Sprintf("step not found: '%s'", c.Chosen.StepName))
		}
		chosenStepCasted, ok := chosenStep.StepType.(*Step_Choice)
		if !ok {
			return invalid(fmt.Sprintf("%s.Chosen.StepName", conditionFieldName), "step in condition is not a choice step")
		}
		if !slices.Contains(chosenStepCasted.Choice.Options, c.Chosen.Result) {
			return invalid(fmt.Sprintf("%s.Chosen.Result", conditionFieldName), fmt.Sprintf("chosen step result is not a valid option: '%s'", c.Chosen.Result))
		}
	}

	return nil
}

// validateConditionStepName validates a single condition step name. A step name must be a
// valid step name and not a terminal step.
func validateConditionStepName(stepName string, stepNames map[string]*Step, fieldName string) error {
	err := validateStepName(stepName, fieldName)
	if err != nil {
		return err
	}

	if _, ok := stepNames[stepName]; !ok {
		return invalid(fieldName, fmt.Sprintf("step not found: '%s'", stepName))
	}
	if _, ok := stepNames[stepName].StepType.(*Step_Terminal); ok {
		return invalid(fieldName, "step in condition is a terminal step")
	}

	return nil
}

// validateConditionsDoNotContainInitial validates that the conditions in ANY or ALL do not contain an
// initial condition.
func validateConditionsDoNotContainInitial(conditions []*Condition, fieldName string) error {
	for i, condition := range conditions {
		conditionFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
		switch c := condition.ConditionType.(type) {
		case *Condition_Initial:
			return invalid(conditionFieldName, "initial step cannot be inside ALL or ANY condition")
		case *Condition_All:
			return validateConditionsDoNotContainInitial(c.All.Conditions, fmt.Sprintf("%s.All.Conditions", conditionFieldName))
		case *Condition_Any:
			return validateConditionsDoNotContainInitial(c.Any.Conditions, fmt.Sprintf("%s.Any.Conditions", conditionFieldName))
		}
	}

	return nil
}

// validateDAG checks that steps form a valid DAG:
// - No cycles exist
// - Exactly one terminal step exists
// - From each initial step, the terminal step is reachable
func validateDAG(steps []*Step, stepNames map[string]*Step, stepsFieldName string) error {
	// Build dependency graph: graph[from] = []to
	graph := make(map[string][]string)

	// Find terminal step
	var terminalStepName string
	var initialSteps []string

	for _, step := range steps {
		// Check if this is a terminal step
		if _, ok := step.StepType.(*Step_Terminal); ok {
			if terminalStepName != "" {
				return invalid(stepsFieldName, "must have exactly one terminal step")
			}
			terminalStepName = step.Name
		}

		// Check if this is an initial step
		if isInitialStep(step) {
			initialSteps = append(initialSteps, step.Name)
		}

		// Build edges from this step's dependencies
		dependencies := extractStepDependencies(step, stepNames)
		for _, dep := range dependencies {
			graph[dep] = append(graph[dep], step.Name)
		}

		// Handle Parallel step's FanOutFrom dependency
		if parallel, ok := step.StepType.(*Step_Parallel); ok {
			fanOutFrom := parallel.Parallel.FanOutFrom
			graph[fanOutFrom] = append(graph[fanOutFrom], step.Name)
		}
	}

	if terminalStepName == "" {
		return invalid(stepsFieldName, "must have exactly one terminal step")
	}

	if len(initialSteps) == 0 {
		return invalid(stepsFieldName, "must have at least one initial step")
	}

	// Check for cycles using DFS
	err := checkForCycles(graph, stepNames, stepsFieldName)
	if err != nil {
		return err
	}

	// Check reachability from each initial step to terminal step
	for _, initialStep := range initialSteps {
		if !isReachable(graph, initialStep, terminalStepName) {
			return invalid(stepsFieldName, fmt.Sprintf("terminal step is not reachable from initial step: '%s'", initialStep))
		}
	}

	return nil
}

func validateStepName(stepName string, stepNameFieldName string) error {
	return validateString(stepName, 1, maxStepNameLength, nameRegex, stepNameFieldName)
}

// isInitialStep checks if a step is marked as initial
func isInitialStep(step *Step) bool {
	var condition *Condition
	switch stepType := step.StepType.(type) {
	case *Step_Simple:
		condition = stepType.Simple.StartsWhen
	case *Step_FanOut:
		condition = stepType.FanOut.StartsWhen
	case *Step_Choice:
		condition = stepType.Choice.StartsWhen
	case *Step_External:
		condition = stepType.External.StartsWhen
	case *Step_Terminal:
		condition = stepType.Terminal.StartsWhen
	}

	if condition == nil {
		return false
	}

	_, ok := condition.ConditionType.(*Condition_Initial)
	return ok
}

// extractStepDependencies extracts all step names that this step depends on from its condition
func extractStepDependencies(step *Step, stepNames map[string]*Step) []string {
	var condition *Condition
	switch stepType := step.StepType.(type) {
	case *Step_Simple:
		condition = stepType.Simple.StartsWhen
	case *Step_FanOut:
		condition = stepType.FanOut.StartsWhen
	case *Step_Choice:
		condition = stepType.Choice.StartsWhen
	case *Step_External:
		condition = stepType.External.StartsWhen
	case *Step_Terminal:
		condition = stepType.Terminal.StartsWhen
	}

	if condition == nil {
		return nil
	}

	return extractStepNamesFromCondition(condition)
}

// extractStepNamesFromCondition recursively extracts step names from a condition
func extractStepNamesFromCondition(condition *Condition) []string {
	if condition == nil {
		return nil
	}

	var stepNames []string
	switch c := condition.ConditionType.(type) {
	case *Condition_Succeeded:
		if c.Succeeded != nil && c.Succeeded.StepName != "" {
			stepNames = append(stepNames, c.Succeeded.StepName)
		}
	case *Condition_Failed:
		if c.Failed != nil && c.Failed.StepName != "" {
			stepNames = append(stepNames, c.Failed.StepName)
		}
	case *Condition_Chosen:
		if c.Chosen != nil && c.Chosen.StepName != "" {
			stepNames = append(stepNames, c.Chosen.StepName)
		}
	case *Condition_All:
		if c.All != nil {
			for _, subCondition := range c.All.Conditions {
				stepNames = append(stepNames, extractStepNamesFromCondition(subCondition)...)
			}
		}
	case *Condition_Any:
		if c.Any != nil {
			for _, subCondition := range c.Any.Conditions {
				stepNames = append(stepNames, extractStepNamesFromCondition(subCondition)...)
			}
		}
	}

	return stepNames
}

// checkForCycles uses DFS to detect cycles in the graph
func checkForCycles(graph map[string][]string, stepNames map[string]*Step, stepsFieldName string) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var visit func(string) error
	visit = func(stepName string) error {
		if recStack[stepName] {
			return invalid(stepsFieldName, fmt.Sprintf("contains a cycle involving step: '%s'", stepName))
		}
		if visited[stepName] {
			return nil
		}

		visited[stepName] = true
		recStack[stepName] = true

		for _, neighbor := range graph[stepName] {
			if err := visit(neighbor); err != nil {
				return err
			}
		}

		recStack[stepName] = false
		return nil
	}

	// Visit all steps
	for stepName := range stepNames {
		if !visited[stepName] {
			if err := visit(stepName); err != nil {
				return err
			}
		}
	}

	return nil
}

// isReachable checks if target is reachable from source using BFS
func isReachable(graph map[string][]string, source, target string) bool {
	if source == target {
		return true
	}

	visited := make(map[string]bool)
	queue := []string{source}
	visited[source] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range graph[current] {
			if neighbor == target {
				return true
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return false
}

func ValidateMetadata(value []*Metadata, fieldName string) error {
	if len(value) > maxNumberOfMetadata {
		return invalid(fieldName, fmt.Sprintf("exceeds max number of metadata (%d)", maxNumberOfMetadata))
	}

	keys := make(map[string]struct{})
	for i, metadata := range value {
		if _, ok := keys[metadata.Key]; ok {
			return invalid(fieldName, fmt.Sprintf("duplicate metadata key: '%s'", metadata.Key))
		}
		keys[metadata.Key] = struct{}{}

		if err := validateString(metadata.Key, 1, maxMetadataKeyLength, metadataKeyRegex, fmt.Sprintf("%s[%d].Key", fieldName, i)); err != nil {
			return err
		}

		if err := validateString(metadata.Value, 1, maxMetadataValueLength, metadataValueRegex, fmt.Sprintf("%s[%d].Value", fieldName, i)); err != nil {
			return err
		}
	}

	return nil
}

func ValidateQueueName(value string, fieldName string) error {
	return validateString(value, 1, maxQueueNameLength, nameRegex, fieldName)
}

func ValidateWorkflowName(value string, fieldName string) error {
	return validateString(value, 1, maxWorkflowNameLength, nameRegex, fieldName)
}

func ValidateDescription(value string, fieldName string) error {
	if len(value) > maxDescriptionLength {
		return invalid(fieldName, fmt.Sprintf("exceeds max length (%d)", maxDescriptionLength))
	}

	return nil
}

func validateString(value string, minLength int, maxLength int, regex string, fieldName string) error {
	if len(value) > maxLength || len(value) < minLength {
		return invalid(fieldName, fmt.Sprintf("length must be between %d and %d characters", minLength, maxLength))
	}

	if m, err := regexp.MatchString(regex, value); err != nil || !m {
		return invalid(fieldName, "must match regex pattern "+regex)
	}

	return nil
}

func invalid(fieldName string, details string) error {
	if details == "" {
		return fmt.Errorf("Invalid %s", fieldName)
	} else {
		return fmt.Errorf("Invalid %s: %s", fieldName, details)
	}
}
