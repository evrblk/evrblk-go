package banyan

// func TaskHandler(input) {

// }

// func FanOutTaskHandler() {

// }

// func ConditionHandler() {

// }

type ChainStepBuilder struct {
	name       string
	queueName  string
	startsWhen *Condition
}

type FanOutStepBuilder struct {
	name       string
	queueName  string
	startsWhen *Condition
}

type ChoiceStepBuilder struct {
	name       string
	queueName  string
	startsWhen *Condition
	options    []string
}

type MultipleStepBuilder struct {
	name                            string
	queueName                       string
	fanOutFrom                      *FanOutStepBuilder
	startOnlyWhenAllSubtasksCreated bool
}

type StepBuilder interface {
	Name() string
	build() *Step
	isStepBuilder()
}

func (b *ChainStepBuilder) isStepBuilder() {}

func (b *FanOutStepBuilder) isStepBuilder() {}

func (b *ChoiceStepBuilder) isStepBuilder() {}

func (b *MultipleStepBuilder) isStepBuilder() {}

func (b *ChainStepBuilder) Name() string {
	return b.name
}

func (b *FanOutStepBuilder) Name() string {
	return b.name
}

func (b *ChoiceStepBuilder) Name() string {
	return b.name
}

func (b *MultipleStepBuilder) Name() string {
	return b.name
}

func (b *ChainStepBuilder) build() *Step {
	return &Step{
		Name:      b.name,
		QueueName: b.queueName,
		StepType: &Step_Chain{Chain: &ChainStep{
			StartsWhen: b.startsWhen,
		}},
	}
}

func (b *FanOutStepBuilder) build() *Step {
	return &Step{
		Name:      b.name,
		QueueName: b.queueName,
		StepType: &Step_FanOut{FanOut: &FanOutStep{
			StartsWhen: b.startsWhen,
		}},
	}
}

func (b *ChoiceStepBuilder) build() *Step {
	return &Step{
		Name:      b.name,
		QueueName: b.queueName,
		StepType: &Step_Choice{Choice: &ChoiceStep{
			StartsWhen: b.startsWhen,
			Options:    b.options,
		}},
	}
}

func (b *MultipleStepBuilder) build() *Step {
	return &Step{
		Name:      b.name,
		QueueName: b.queueName,
	}
}

type WorkflowBuilder struct {
	name        string
	description string
	steps       []StepBuilder
	metadata    []*Metadata
}

func (b *WorkflowBuilder) Metadata(meta map[string]string) {
	for key, value := range meta {
		b.metadata = append(b.metadata, &Metadata{
			Key:   key,
			Value: value,
		})
	}
}

func (b *WorkflowBuilder) Succeeded(step StepBuilder) *Condition {
	return &Condition{
		Condition: &Condition_Succeeded{
			Succeeded: &PredicateSucceeded{
				StepName: step.Name(),
			},
		},
	}
}

func (b *WorkflowBuilder) Failed(step StepBuilder) *Condition {
	return &Condition{
		Condition: &Condition_Failed{
			Failed: &PredicateFailed{
				StepName: step.Name(),
			},
		},
	}
}

func (b *WorkflowBuilder) Chosen(step *ChoiceStepBuilder, result string) *Condition {
	return &Condition{
		Condition: &Condition_Chosen{
			Chosen: &PredicateChosen{
				StepName: step.Name(),
				Result:   result,
			},
		},
	}
}

func (b *WorkflowBuilder) All(conditions ...*Condition) *Condition {
	return &Condition{
		Condition: &Condition_All{
			All: &PredicateAll{
				Conditions: conditions,
			},
		},
	}
}

func (b *WorkflowBuilder) Any(conditions ...*Condition) *Condition {
	return &Condition{
		Condition: &Condition_Any{
			Any: &PredicateAny{
				Conditions: conditions,
			},
		},
	}
}

func (b *WorkflowBuilder) ChainStep(name string, startsWhen *Condition, queueName string) *ChainStepBuilder {
	step := &ChainStepBuilder{
		name:       name,
		queueName:  queueName,
		startsWhen: startsWhen,
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) FanOutStep(name string, startsWhen *Condition, queueName string) *FanOutStepBuilder {
	step := &FanOutStepBuilder{
		name:       name,
		queueName:  queueName,
		startsWhen: startsWhen,
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) ChoiceStep(name string, startsWhen *Condition, queueName string, options []string) *ChoiceStepBuilder {
	step := &ChoiceStepBuilder{
		name:       name,
		queueName:  queueName,
		startsWhen: startsWhen,
		options:    options,
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) MultipleStep(name string, fanOutFrom *FanOutStepBuilder, queueName string) *MultipleStepBuilder {
	step := &MultipleStepBuilder{
		name:       name,
		queueName:  queueName,
		fanOutFrom: fanOutFrom,
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) MustBuild() *Workflow {
	workflow, err := b.Build()
	if err != nil {
		panic(err)
	}

	return workflow
}

func (b *WorkflowBuilder) Build() (*Workflow, error) {
	workflow := &Workflow{
		Name:        b.name,
		Description: b.description,
		Steps:       make([]*Step, len(b.steps)),
		Metadata:    b.metadata,
	}

	for i, step := range b.steps {
		workflow.Steps[i] = step.build()
	}

	return workflow, nil
}

func NewWorkflowBuilder(name string, description string) *WorkflowBuilder {
	return &WorkflowBuilder{
		name:        name,
		description: description,
	}
}

// type TaskHandler func()
