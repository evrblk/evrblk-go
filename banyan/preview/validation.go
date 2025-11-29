package banyan

import "errors"

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

	var initialStep *Step
	for _, step := range workflow.Steps {
		if step.QueueName == "" {
			return errors.New("step queue name is required")
		}

		switch stepType := step.StepType.(type) {
		case *Step_Chain:
			switch condition := stepType.Chain.StartsWhen.Condition.(type) {
			case *Condition_Initial:
				if initialStep != nil {
					return errors.New("multiple initial steps are not allowed")
				}
				initialStep = step
			case *Condition_All:
				err := validateConditionsDoNotContainInitial(condition.All.Conditions)
				if err != nil {
					return err
				}
			case *Condition_Any:
				err := validateConditionsDoNotContainInitial(condition.Any.Conditions)
				if err != nil {
					return err
				}
			}

		case *Step_FanOut:
			switch condition := stepType.FanOut.StartsWhen.Condition.(type) {
			case *Condition_Initial:
				if initialStep != nil {
					return errors.New("multiple initial steps are not allowed")
				}
				initialStep = step
			case *Condition_All:
				err := validateConditionsDoNotContainInitial(condition.All.Conditions)
				if err != nil {
					return err
				}
			case *Condition_Any:
				err := validateConditionsDoNotContainInitial(condition.Any.Conditions)
				if err != nil {
					return err
				}
			}

		case *Step_Choice:
			switch condition := stepType.Choice.StartsWhen.Condition.(type) {
			case *Condition_Initial:
				if initialStep != nil {
					return errors.New("multiple initial steps are not allowed")
				}
				initialStep = step
			case *Condition_All:
				err := validateConditionsDoNotContainInitial(condition.All.Conditions)
				if err != nil {
					return err
				}
			case *Condition_Any:
				err := validateConditionsDoNotContainInitial(condition.Any.Conditions)
				if err != nil {
					return err
				}
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

		case *Step_External:
			switch condition := stepType.External.StartsWhen.Condition.(type) {
			case *Condition_Initial:
				if initialStep != nil {
					return errors.New("multiple initial steps are not allowed")
				}
				initialStep = step
			case *Condition_All:
				err := validateConditionsDoNotContainInitial(condition.All.Conditions)
				if err != nil {
					return err
				}
			case *Condition_Any:
				err := validateConditionsDoNotContainInitial(condition.Any.Conditions)
				if err != nil {
					return err
				}
			}
		}
	}
	if initialStep == nil {
		return errors.New("initial step is required")
	}

	return nil
}

func validateConditionsDoNotContainInitial(conditions []*Condition) error {
	for _, condition := range conditions {
		switch c := condition.Condition.(type) {
		case *Condition_Initial:
			return errors.New("initial step cannot be inside ALL or ANY condition")
		case *Condition_All:
			err := validateConditionsDoNotContainInitial(c.All.Conditions)
			if err != nil {
				return err
			}
		case *Condition_Any:
			err := validateConditionsDoNotContainInitial(c.Any.Conditions)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
