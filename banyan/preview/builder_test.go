package banyan

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	b := NewWorkflowBuilder("test_workflow", "This is a test workflow")
	b.Metadata(map[string]string{
		"owner":   "evrblk",
		"version": "1",
	})

	step1 := b.ChainStep("task1").
		IsInitial().
		QueueTo("main_queue")
	step2 := b.ChainStep("task2").
		StartWhen(b.Succeeded(step1)).
		QueueTo("main_queue")
	step3 := b.ChainStep("task3").
		StartWhen(b.Succeeded(step2)).
		QueueTo("main_queue")
	step4 := b.ChainStep("task4").
		StartWhen(b.Succeeded(step2)).
		QueueTo("rate_limited_queue")
	step5 := b.FanOutStep("task5").
		StartWhen(b.All(b.Succeeded(step3), b.Succeeded(step4))).
		QueueTo("main_queue")
	step6 := b.ParallelStep("task6").
		FanOutFrom(step5).
		StartOnlyWhenAllSubtasksCreated().
		QueueTo("high_queue")
	step7 := b.ChoiceStep("task7").
		StartWhen(b.Succeeded(step6)).
		WithOptions("option1", "option2").
		QueueTo("main_queue")
	step8 := b.ChainStep("task8").
		StartWhen(b.Chosen(step7, "option1")).
		QueueTo("main_queue")
	_ = b.ChainStep("task9").
		StartWhen(b.Any(b.Chosen(step7, "option2"), b.Succeeded(step8))).
		QueueTo("main_queue")

	workflow, err := b.Build()
	require.NoError(t, err)
	require.NotNil(t, workflow)

	require.Equal(t, "test_workflow", workflow.Name)
	require.Equal(t, "This is a test workflow", workflow.Description)
	require.Equal(t, []*Metadata{
		{
			Key:   "owner",
			Value: "evrblk",
		},
		{
			Key:   "version",
			Value: "1",
		},
	}, workflow.Metadata)
	require.Len(t, workflow.Steps, 9)

	// step1
	require.Equal(t, "task1", workflow.Steps[0].Name)
	require.Equal(t, "main_queue", workflow.Steps[0].QueueName)
	require.Equal(t, &Step_Chain{
		Chain: &ChainStep{
			StartsWhen: &Condition{
				Condition: &Condition_Initial{
					Initial: &PredicateInitial{
						IsInitial: true,
					},
				},
			},
		},
	}, workflow.Steps[0].StepType)

	// step2
	require.Equal(t, "task2", workflow.Steps[1].Name)
	require.Equal(t, "main_queue", workflow.Steps[1].QueueName)
	require.Equal(t, &Step_Chain{
		Chain: &ChainStep{
			StartsWhen: &Condition{
				Condition: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task1",
					},
				},
			},
		},
	}, workflow.Steps[1].StepType)

	// step3
	require.Equal(t, "main_queue", workflow.Steps[2].QueueName)
	require.Equal(t, &Step_Chain{
		Chain: &ChainStep{
			StartsWhen: &Condition{
				Condition: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task2",
					},
				},
			},
		},
	}, workflow.Steps[2].StepType)

	// step4
	require.Equal(t, "task4", workflow.Steps[3].Name)
	require.Equal(t, "rate_limited_queue", workflow.Steps[3].QueueName)
	require.Equal(t, &Step_Chain{
		Chain: &ChainStep{
			StartsWhen: &Condition{
				Condition: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task2",
					},
				},
			},
		},
	}, workflow.Steps[3].StepType)

	// step5
	require.Equal(t, "task5", workflow.Steps[4].Name)
	require.Equal(t, "main_queue", workflow.Steps[4].QueueName)
	require.Equal(t, &Step_FanOut{
		FanOut: &FanOutStep{
			StartsWhen: &Condition{
				Condition: &Condition_All{
					All: &PredicateAll{
						Conditions: []*Condition{
							{
								Condition: &Condition_Succeeded{
									Succeeded: &PredicateSucceeded{StepName: "task3"},
								},
							},
							{
								Condition: &Condition_Succeeded{
									Succeeded: &PredicateSucceeded{StepName: "task4"},
								},
							},
						},
					},
				},
			},
		},
	}, workflow.Steps[4].StepType)

	// step6
	require.Equal(t, "task6", workflow.Steps[5].Name)
	require.Equal(t, "high_queue", workflow.Steps[5].QueueName)
	require.Equal(t, &Step_Parallel{
		Parallel: &ParallelStep{
			FanOutFrom:                      "task5",
			StartOnlyWhenAllSubtasksCreated: true,
		},
	}, workflow.Steps[5].StepType)

	// step7
	require.Equal(t, "task7", workflow.Steps[6].Name)
	require.Equal(t, "main_queue", workflow.Steps[6].QueueName)
	require.Equal(t, &Step_Choice{
		Choice: &ChoiceStep{
			StartsWhen: &Condition{
				Condition: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task6",
					},
				},
			},
			Options: []string{"option1", "option2"},
		},
	}, workflow.Steps[6].StepType)

	// step8
	require.Equal(t, "task8", workflow.Steps[7].Name)
	require.Equal(t, "main_queue", workflow.Steps[7].QueueName)
	require.Equal(t, &Step_Chain{
		Chain: &ChainStep{
			StartsWhen: &Condition{
				Condition: &Condition_Chosen{
					Chosen: &PredicateChosen{
						StepName: "task7",
						Result:   "option1",
					},
				},
			},
		},
	}, workflow.Steps[7].StepType)

	// step9
	require.Equal(t, "task9", workflow.Steps[8].Name)
	require.Equal(t, "main_queue", workflow.Steps[8].QueueName)
	require.Equal(t, &Step_Chain{
		Chain: &ChainStep{
			StartsWhen: &Condition{
				Condition: &Condition_Any{
					Any: &PredicateAny{
						Conditions: []*Condition{
							{
								Condition: &Condition_Chosen{
									Chosen: &PredicateChosen{
										StepName: "task7",
										Result:   "option2",
									},
								},
							},
							{Condition: &Condition_Succeeded{Succeeded: &PredicateSucceeded{StepName: "task8"}}},
						},
					},
				},
			},
		},
	}, workflow.Steps[8].StepType)
}
