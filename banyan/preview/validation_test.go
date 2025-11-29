package banyan

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidator_NamesMustBeUnique(t *testing.T) {
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name: "task1",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "task1",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task1",
								},
							},
						},
					},
				},
			},
		},
	}
	err := ValidateWorkflow(workflow)
	require.Error(t, err)
}

func TestValidator_InitialStepMustBeUnique(t *testing.T) {
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name:      "task1",
				QueueName: "main_queue",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name:      "task2",
				QueueName: "main_queue",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
		},
	}
	err := ValidateWorkflow(workflow)
	require.Error(t, err)
}

func TestValidator_InitialStepMustBeInsideAllOrAnyCondition(t *testing.T) {
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name:      "task1",
				QueueName: "main_queue",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name:      "task2",
				QueueName: "main_queue",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_All{
								All: &PredicateAll{
									Conditions: []*Condition{
										{
											Condition: &Condition_Initial{
												Initial: &PredicateInitial{
													IsInitial: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	err := ValidateWorkflow(workflow)
	require.Error(t, err)
}

func TestValidator_ValidWorkflow(t *testing.T) {
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name:      "task1",
				QueueName: "main_queue",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name:      "task2",
				QueueName: "main_queue",
				StepType: &Step_Chain{
					Chain: &ChainStep{
						StartsWhen: &Condition{
							Condition: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task1",
								},
							},
						},
					},
				},
			},
		},
	}
	err := ValidateWorkflow(workflow)
	require.NoError(t, err)
}
