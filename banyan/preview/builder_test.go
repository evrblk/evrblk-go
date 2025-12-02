package banyan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	b := NewWorkflowBuilder("test_workflow", "This is a test workflow")
	b.Metadata(map[string]string{
		"owner":   "evrblk",
		"version": "1",
	})

	step1 := b.SimpleStep("task1").
		IsInitial().
		QueueTo("main_queue")
	step2 := b.SimpleStep("task2").
		StartWhen(b.Succeeded(step1)).
		QueueTo("main_queue")
	step3 := b.SimpleStep("task3").
		StartWhen(b.Succeeded(step2)).
		QueueTo("main_queue").
		DelayBy(10 * time.Second)
	step4 := b.SimpleStep("task4").
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
	step8 := b.SimpleStep("task8").
		StartWhen(b.Chosen(step7, "option1")).
		QueueTo("main_queue")
	step9 := b.SimpleStep("task9").
		StartWhen(b.Any(b.Chosen(step7, "option2"), b.Succeeded(step8))).
		QueueTo("main_queue")
	_ = b.TerminalStep().
		StartWhen(b.Succeeded(step9))

	workflow, err := b.Build()
	require.NoError(t, err)
	require.NotNil(t, workflow)

	require.Equal(t, "test_workflow", workflow.Name)
	require.Equal(t, "This is a test workflow", workflow.Description)
	require.Contains(t, workflow.Metadata, &Metadata{
		Key:   "owner",
		Value: "evrblk",
	})
	require.Contains(t, workflow.Metadata, &Metadata{
		Key:   "version",
		Value: "1",
	})
	require.Len(t, workflow.Steps, 10)

	// step1
	require.Equal(t, "task1", workflow.Steps[0].Name)
	require.Equal(t, &Step_Simple{
		Simple: &SimpleStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Initial{
					Initial: &PredicateInitial{
						IsInitial: true,
					},
				},
			},
			QueueName:      "main_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[0].StepType)

	// step2
	require.Equal(t, "task2", workflow.Steps[1].Name)
	require.Equal(t, &Step_Simple{
		Simple: &SimpleStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task1",
					},
				},
			},
			QueueName:      "main_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[1].StepType)

	// step3
	require.Equal(t, &Step_Simple{
		Simple: &SimpleStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task2",
					},
				},
			},
			QueueName:      "main_queue",
			DelayBySeconds: 10,
		},
	}, workflow.Steps[2].StepType)

	// step4
	require.Equal(t, "task4", workflow.Steps[3].Name)
	require.Equal(t, &Step_Simple{
		Simple: &SimpleStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task2",
					},
				},
			},
			QueueName:      "rate_limited_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[3].StepType)

	// step5
	require.Equal(t, "task5", workflow.Steps[4].Name)
	require.Equal(t, &Step_FanOut{
		FanOut: &FanOutStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_All{
					All: &PredicateAll{
						Conditions: []*Condition{
							{
								ConditionType: &Condition_Succeeded{
									Succeeded: &PredicateSucceeded{StepName: "task3"},
								},
							},
							{
								ConditionType: &Condition_Succeeded{
									Succeeded: &PredicateSucceeded{StepName: "task4"},
								},
							},
						},
					},
				},
			},
			QueueName:      "main_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[4].StepType)

	// step6
	require.Equal(t, "task6", workflow.Steps[5].Name)
	require.Equal(t, &Step_Parallel{
		Parallel: &ParallelStep{
			FanOutFrom:                      "task5",
			StartOnlyWhenAllSubtasksCreated: true,
			QueueName:                       "high_queue",
			DelayBySeconds:                  0,
		},
	}, workflow.Steps[5].StepType)

	// step7
	require.Equal(t, "task7", workflow.Steps[6].Name)
	require.Equal(t, &Step_Choice{
		Choice: &ChoiceStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task6",
					},
				},
			},
			Options:        []string{"option1", "option2"},
			QueueName:      "main_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[6].StepType)

	// step8
	require.Equal(t, "task8", workflow.Steps[7].Name)
	require.Equal(t, &Step_Simple{
		Simple: &SimpleStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Chosen{
					Chosen: &PredicateChosen{
						StepName: "task7",
						Result:   "option1",
					},
				},
			},
			QueueName:      "main_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[7].StepType)

	// step9
	require.Equal(t, "task9", workflow.Steps[8].Name)
	require.Equal(t, &Step_Simple{
		Simple: &SimpleStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Any{
					Any: &PredicateAny{
						Conditions: []*Condition{
							{
								ConditionType: &Condition_Chosen{
									Chosen: &PredicateChosen{
										StepName: "task7",
										Result:   "option2",
									},
								},
							},
							{
								ConditionType: &Condition_Succeeded{
									Succeeded: &PredicateSucceeded{
										StepName: "task8",
									},
								},
							},
						},
					},
				},
			},
			QueueName:      "main_queue",
			DelayBySeconds: 0,
		},
	}, workflow.Steps[8].StepType)

	// step10
	require.Equal(t, "terminal", workflow.Steps[9].Name)
	require.Equal(t, &Step_Terminal{
		Terminal: &TerminalStep{
			StartsWhen: &Condition{
				ConditionType: &Condition_Succeeded{
					Succeeded: &PredicateSucceeded{
						StepName: "task9",
					},
				},
			},
		},
	}, workflow.Steps[9].StepType)
}
