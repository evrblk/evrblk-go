package banyan

import (
	"errors"
	"slices"
)

// ValidateWorkflow validates the workflow:
// - Workflow name is required
// - Workflow steps are required
// - Step names must be unique
// - Step queue name is required
// - Condition step name must be a valid step name
// - Conditions in ANY or ALL do not contain an initial condition
// - Chosen step result must be a valid option
// - Conditions do not contain a terminal step
// - No cycles in steps DAG
// - Terminal step exists
// - Terminal step is reachable from all initial steps
func ValidateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return errors.New("workflow name is required")
	}

	if len(workflow.Steps) == 0 {
		return errors.New("workflow steps are required")
	}

	stepNames := make(map[string]*Step)
	for _, step := range workflow.Steps {
		if step.Name == "" {
			return errors.New("step name is required")
		}

		if _, ok := stepNames[step.Name]; ok {
			return errors.New("step name must be unique")
		}
		stepNames[step.Name] = step
	}

	for _, step := range workflow.Steps {
		switch stepType := step.StepType.(type) {
		case *Step_Simple:
			err := validateCondition(stepType.Simple.StartsWhen, stepNames)
			if err != nil {
				return err
			}

			if stepType.Simple.QueueName == "" {
				return errors.New("step queue name is required")
			}

		case *Step_FanOut:
			err := validateCondition(stepType.FanOut.StartsWhen, stepNames)
			if err != nil {
				return err
			}

			if stepType.FanOut.QueueName == "" {
				return errors.New("step queue name is required")
			}

		case *Step_Choice:
			err := validateCondition(stepType.Choice.StartsWhen, stepNames)
			if err != nil {
				return err
			}

			if stepType.Choice.QueueName == "" {
				return errors.New("step queue name is required")
			}

		case *Step_Parallel:
			if stepType.Parallel.FanOutFrom == "" {
				return errors.New("parallel step fan out from is required")
			}

			fanOutFrom, ok := stepNames[stepType.Parallel.FanOutFrom]
			if !ok {
				return errors.New("parallel step fan out from step not found")
			}
			if _, ok := fanOutFrom.StepType.(*Step_FanOut); !ok {
				return errors.New("parallel step fan out from step is not a fan out step")
			}

			if stepType.Parallel.QueueName == "" {
				return errors.New("step queue name is required")
			}

		case *Step_External:
			err := validateCondition(stepType.External.StartsWhen, stepNames)
			if err != nil {
				return err
			}

		case *Step_Terminal:
			err := validateCondition(stepType.Terminal.StartsWhen, stepNames)
			if err != nil {
				return err
			}
		}
	}

	err := validateDAG(workflow, stepNames)
	if err != nil {
		return err
	}

	return nil
}

// validateConditions calls validateCondition for each condition in the list
func validateConditions(conditions []*Condition, stepNames map[string]*Step) error {
	for _, condition := range conditions {
		err := validateCondition(condition, stepNames)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateCondition validates a single condition
func validateCondition(condition *Condition, stepNames map[string]*Step) error {
	switch c := condition.ConditionType.(type) {
	case *Condition_All:
		err := validateConditionsDoNotContainInitial(c.All.Conditions)
		if err != nil {
			return err
		}
		err = validateConditions(c.All.Conditions, stepNames)
		if err != nil {
			return err
		}

	case *Condition_Any:
		err := validateConditionsDoNotContainInitial(c.Any.Conditions)
		if err != nil {
			return err
		}
		err = validateConditions(c.Any.Conditions, stepNames)
		if err != nil {
			return err
		}

	case *Condition_Succeeded:
		err := validateConditionStepName(c.Succeeded.StepName, stepNames)
		if err != nil {
			return err
		}

	case *Condition_Failed:
		err := validateConditionStepName(c.Failed.StepName, stepNames)
		if err != nil {
			return err
		}

	case *Condition_Chosen:
		err := validateConditionStepName(c.Chosen.StepName, stepNames)
		if err != nil {
			return err
		}

		if c.Chosen.Result == "" {
			return errors.New("chosen step result is required")
		}
		chosenStep, ok := stepNames[c.Chosen.StepName]
		if !ok {
			return errors.New("chosen step not found")
		}
		chosenStepCasted, ok := chosenStep.StepType.(*Step_Choice)
		if !ok {
			return errors.New("step in condition is not a choice step")
		}
		if !slices.Contains(chosenStepCasted.Choice.Options, c.Chosen.Result) {
			return errors.New("chosen step result is not a valid option")
		}
	}

	return nil
}

// validateConditionStepName validates a single condition step name. A step name must be a
// valid step name and not a terminal step.
func validateConditionStepName(stepName string, stepNames map[string]*Step) error {
	if stepName == "" {
		return errors.New("step name is required")
	}
	if _, ok := stepNames[stepName]; !ok {
		return errors.New("step not found")
	}
	if _, ok := stepNames[stepName].StepType.(*Step_Terminal); ok {
		return errors.New("step in condition is a terminal step")
	}

	return nil
}

// validateConditionsDoNotContainInitial validates that the conditions in ANY or ALL do not contain an
// initial condition.
func validateConditionsDoNotContainInitial(conditions []*Condition) error {
	for _, condition := range conditions {
		switch c := condition.ConditionType.(type) {
		case *Condition_Initial:
			return errors.New("initial step cannot be inside ALL or ANY condition")
		case *Condition_All:
			return validateConditionsDoNotContainInitial(c.All.Conditions)
		case *Condition_Any:
			return validateConditionsDoNotContainInitial(c.Any.Conditions)
		}
	}

	return nil
}

// validateDAG checks that the workflow forms a valid DAG:
// - No cycles exist
// - Exactly one terminal step exists
// - From each initial step, the terminal step is reachable
func validateDAG(workflow *Workflow, stepNames map[string]*Step) error {
	// Build dependency graph: graph[from] = []to
	graph := make(map[string][]string)

	// Find terminal step
	var terminalStepName string
	var initialSteps []string

	for _, step := range workflow.Steps {
		// Check if this is a terminal step
		if _, ok := step.StepType.(*Step_Terminal); ok {
			if terminalStepName != "" {
				return errors.New("workflow must have exactly one terminal step")
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
			// Validate that referenced step exists
			if _, ok := stepNames[dep]; !ok {
				return errors.New("step references non-existent step: " + dep)
			}
			graph[dep] = append(graph[dep], step.Name)
		}

		// Handle Parallel step's FanOutFrom dependency
		if parallel, ok := step.StepType.(*Step_Parallel); ok {
			fanOutFrom := parallel.Parallel.FanOutFrom
			if fanOutFrom != "" {
				graph[fanOutFrom] = append(graph[fanOutFrom], step.Name)
			}
		}
	}

	if terminalStepName == "" {
		return errors.New("workflow must have exactly one terminal step")
	}

	if len(initialSteps) == 0 {
		return errors.New("workflow must have at least one initial step")
	}

	// Check for cycles using DFS
	err := checkForCycles(graph, stepNames)
	if err != nil {
		return err
	}

	// Check reachability from each initial step to terminal step
	for _, initialStep := range initialSteps {
		if !isReachable(graph, initialStep, terminalStepName) {
			return errors.New("terminal step is not reachable from initial step: " + initialStep)
		}
	}

	return nil
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
func checkForCycles(graph map[string][]string, stepNames map[string]*Step) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var visit func(string) error
	visit = func(stepName string) error {
		if recStack[stepName] {
			return errors.New("workflow contains a cycle involving step: " + stepName)
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
