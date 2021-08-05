// +build test

package db_e2e

import (
	"testing"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/mysql.yaml
//go:generate kubectl -n argo-dataflow-system wait pod -l statefulset.kubernetes.io/pod-name --for condition=ready

func TestDBSink(t *testing.T) {
	defer Setup(t)()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "db"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks: []Sink{
						{Name: "db", DB: &DBSink{
							Database: Database{
								Driver: "mysql",
								DataSource: &DBDataSource{
									ValueFrom: &DBDataSourceFrom{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "dataSource",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "mysql-data-source",
											},
										},
									},
								},
							},
							Actions: []SQLAction{
								{
									SQLStatement: SQLStatement{
										SQL: `CREATE TABLE IF NOT EXISTS test_table (
											id INT auto_increment,
											msg VARCHAR(255),
											int_col INT,
											primary key (id)
										);`,
									},
								},
								{
									SQLStatement: SQLStatement{
										SQL:  "insert into test_table values (null, ?, ?)",
										Args: []string{"object(msg).message", "object(msg).number"},
									},
								},
								{
									SQLStatement: SQLStatement{
										SQL: "update test_table set int_col = 2 where msg = 'notexisting'",
									},
									OnRecordNotFound: &SQLStatement{
										SQL:  "insert into test_table values (null, ?, 200)",
										Args: []string{"object(msg).message"},
									},
								},
								{
									SQLStatement: SQLStatement{
										SQL: "insert into test_table values(1, 'error', 1)",
									},
									OnError: &SQLStatement{
										SQL:  "insert into test_table values (null, ?, 300)",
										Args: []string{"object(msg).message"},
									},
								},
							},
						}},
					},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("db-main-0")()
	SendMessageViaHTTP(`{"message": "hello", "number": 100}`)

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	// TODO: verify the table records.

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
