package banyan

import "time"

// func TaskHandler(input) {

// }

// func FanOutTaskHandler() {

// }

// func ConditionHandler() {

// }

type StepBuilder interface {
	Name() string
	build() *Step
	isStepBuilder()
}

type TerminalStepBuilder struct {
	startsWhen *Condition
}

var _ StepBuilder = (*TerminalStepBuilder)(nil)

func (b *TerminalStepBuilder) Name() string {
	return "terminal"
}

func (b *TerminalStepBuilder) StartWhen(condition *Condition) *TerminalStepBuilder {
	b.startsWhen = condition
	return b
}

func (b *TerminalStepBuilder) build() *Step {
	return &Step{
		Name: "terminal",
		StepType: &Step_Terminal{Terminal: &TerminalStep{
			StartsWhen: b.startsWhen,
		}},
	}
}

func (b *TerminalStepBuilder) isStepBuilder() {}

type ChainStepBuilder struct {
	name       string
	queueName  string
	startsWhen *Condition
	delayBy    time.Duration
}

var _ StepBuilder = (*ChainStepBuilder)(nil)

func (b *ChainStepBuilder) Name() string {
	return b.name
}

func (b *ChainStepBuilder) QueueTo(queueName string) *ChainStepBuilder {
	b.queueName = queueName
	return b
}

func (b *ChainStepBuilder) StartWhen(condition *Condition) *ChainStepBuilder {
	b.startsWhen = condition
	return b
}

func (b *ChainStepBuilder) DelayBy(delay time.Duration) *ChainStepBuilder {
	b.delayBy = delay
	return b
}

func (b *ChainStepBuilder) IsInitial() *ChainStepBuilder {
	b.startsWhen = &Condition{
		Condition: &Condition_Initial{
			Initial: &PredicateInitial{
				IsInitial: true,
			},
		},
	}
	return b
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

func (b *ChainStepBuilder) isStepBuilder() {}

type FanOutStepBuilder struct {
	name       string
	queueName  string
	startsWhen *Condition
	delayBy    time.Duration
}

var _ StepBuilder = (*FanOutStepBuilder)(nil)

func (b *FanOutStepBuilder) Name() string {
	return b.name
}

func (b *FanOutStepBuilder) QueueTo(queueName string) *FanOutStepBuilder {
	b.queueName = queueName
	return b
}

func (b *FanOutStepBuilder) StartWhen(condition *Condition) *FanOutStepBuilder {
	b.startsWhen = condition
	return b
}

func (b *FanOutStepBuilder) DelayBy(delay time.Duration) *FanOutStepBuilder {
	b.delayBy = delay
	return b
}

func (b *FanOutStepBuilder) IsInitial() *FanOutStepBuilder {
	b.startsWhen = &Condition{
		Condition: &Condition_Initial{
			Initial: &PredicateInitial{
				IsInitial: true,
			},
		},
	}
	return b
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

func (b *FanOutStepBuilder) isStepBuilder() {}

type ChoiceStepBuilder struct {
	name       string
	queueName  string
	startsWhen *Condition
	options    []string
	delayBy    time.Duration
}

var _ StepBuilder = (*ChoiceStepBuilder)(nil)

func (b *ChoiceStepBuilder) Name() string {
	return b.name
}

func (b *ChoiceStepBuilder) QueueTo(queueName string) *ChoiceStepBuilder {
	b.queueName = queueName
	return b
}

func (b *ChoiceStepBuilder) StartWhen(condition *Condition) *ChoiceStepBuilder {
	b.startsWhen = condition
	return b
}

func (b *ChoiceStepBuilder) DelayBy(delay time.Duration) *ChoiceStepBuilder {
	b.delayBy = delay
	return b
}

func (b *ChoiceStepBuilder) IsInitial() *ChoiceStepBuilder {
	b.startsWhen = &Condition{
		Condition: &Condition_Initial{
			Initial: &PredicateInitial{
				IsInitial: true,
			},
		},
	}
	return b
}

func (b *ChoiceStepBuilder) WithOptions(options ...string) *ChoiceStepBuilder {
	b.options = options
	return b
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

func (b *ChoiceStepBuilder) isStepBuilder() {}

type ParallelStepBuilder struct {
	name                            string
	queueName                       string
	fanOutFrom                      *FanOutStepBuilder
	startOnlyWhenAllSubtasksCreated bool
	delayBy                         time.Duration
}

var _ StepBuilder = (*ParallelStepBuilder)(nil)

func (b *ParallelStepBuilder) Name() string {
	return b.name
}

func (b *ParallelStepBuilder) QueueTo(queueName string) *ParallelStepBuilder {
	b.queueName = queueName
	return b
}

func (b *ParallelStepBuilder) FanOutFrom(fanOutFrom *FanOutStepBuilder) *ParallelStepBuilder {
	b.fanOutFrom = fanOutFrom
	return b
}

func (b *ParallelStepBuilder) DelayBy(delay time.Duration) *ParallelStepBuilder {
	b.delayBy = delay
	return b
}

func (b *ParallelStepBuilder) StartOnlyWhenAllSubtasksCreated() *ParallelStepBuilder {
	b.startOnlyWhenAllSubtasksCreated = true
	return b
}

func (b *ParallelStepBuilder) build() *Step {
	return &Step{
		Name:      b.name,
		QueueName: b.queueName,
		StepType: &Step_Parallel{Parallel: &ParallelStep{
			FanOutFrom:                      b.fanOutFrom.Name(),
			StartOnlyWhenAllSubtasksCreated: b.startOnlyWhenAllSubtasksCreated,
		}},
	}
}

func (b *ParallelStepBuilder) isStepBuilder() {}

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

func (b *WorkflowBuilder) TerminalStep() *TerminalStepBuilder {
	step := &TerminalStepBuilder{}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) ChainStep(name string) *ChainStepBuilder {
	step := &ChainStepBuilder{
		name:      name,
		queueName: "default",
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) FanOutStep(name string) *FanOutStepBuilder {
	step := &FanOutStepBuilder{
		name:      name,
		queueName: "default",
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) ChoiceStep(name string) *ChoiceStepBuilder {
	step := &ChoiceStepBuilder{
		name:      name,
		queueName: "default",
	}
	b.steps = append(b.steps, step)
	return step
}

func (b *WorkflowBuilder) ParallelStep(name string) *ParallelStepBuilder {
	step := &ParallelStepBuilder{
		name:                            name,
		queueName:                       "default",
		startOnlyWhenAllSubtasksCreated: false,
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
