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
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Initial{
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
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
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

func TestValidator_InitialStepMustNotBeInsideAllOrAnyCondition(t *testing.T) {
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name: "task1",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "task2",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_All{
								All: &PredicateAll{
									Conditions: []*Condition{
										{
											ConditionType: &Condition_Initial{
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
				Name: "task1",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "task2",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task1",
								},
							},
						},
					},
				},
			},
			{
				Name: "terminal",
				StepType: &Step_Terminal{
					Terminal: &TerminalStep{
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task2",
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

func TestValidator_CycleDetection(t *testing.T) {
	// Create a cycle: task1 (initial) -> task2 -> task3 -> task2 (cycle between B and C)
	// task2 depends on both task1 (to be reachable) and task3 (to create the cycle)
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name: "task1",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "task2",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Any{
								Any: &PredicateAny{
									Conditions: []*Condition{
										{
											ConditionType: &Condition_Succeeded{
												Succeeded: &PredicateSucceeded{
													StepName: "task1",
												},
											},
										},
										{
											ConditionType: &Condition_Succeeded{
												Succeeded: &PredicateSucceeded{
													StepName: "task3",
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
			{
				Name: "task3",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task2",
								},
							},
						},
					},
				},
			},
			{
				Name: "terminal",
				StepType: &Step_Terminal{
					Terminal: &TerminalStep{
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task2",
								},
							},
						},
					},
				},
			},
		},
	}
	// This creates a cycle: taskB -> taskC -> taskB (taskB depends on taskC, taskC depends on taskB)
	err := ValidateWorkflow(workflow)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cycle")
}

func TestValidator_ReachabilityFromInitialStep(t *testing.T) {
	workflow := &Workflow{
		Name: "test_workflow",
		Steps: []*Step{
			{
				Name: "task1",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "task2",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task1",
								},
							},
						},
					},
				},
			},
			{
				Name: "orphan_initial",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Initial{
								Initial: &PredicateInitial{
									IsInitial: true,
								},
							},
						},
					},
				},
			},
			{
				Name: "orphan_task",
				StepType: &Step_Simple{
					Simple: &SimpleStep{
						QueueName: "main_queue",
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "orphan_initial",
								},
							},
						},
					},
				},
			},
			{
				Name: "terminal",
				StepType: &Step_Terminal{
					Terminal: &TerminalStep{
						StartsWhen: &Condition{
							ConditionType: &Condition_Succeeded{
								Succeeded: &PredicateSucceeded{
									StepName: "task2",
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
	require.Contains(t, err.Error(), "terminal step is not reachable from initial step")
	require.Contains(t, err.Error(), "orphan_initial")
}
