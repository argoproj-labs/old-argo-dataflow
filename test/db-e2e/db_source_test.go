// +build test

package db_e2e

import (
	"testing"
	"time"

	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	. "github.com/argoproj-labs/argo-dataflow/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate kubectl -n argo-dataflow-system apply -f ../../config/apps/mysql.yaml

func TestDBSource(t *testing.T) {
	defer Setup(t)()

	cleanupDB()

	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "db-source"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "insert",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks: []Sink{
						{DB: &DBSink{
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
										SQL: `CREATE TABLE IF NOT EXISTS test_source_table (
											id INT auto_increment,
											msg VARCHAR(255),
											int_col INT,
											msg_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
											primary key (id)
										);`,
									},
								},
								{
									SQLStatement: SQLStatement{
										SQL:  "insert into test_source_table values (null, ?, ?, null)",
										Args: []string{"object(msg).message", "object(msg).number"},
									},
								},
							},
						}},
					},
				},
				{
					Name: "main",
					Cat:  &Cat{},
					Sources: []Source{{
						DB: &DBSource{
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
							Query:          "select * from test_source_table",
							PollInterval:   metav1.Duration{Duration: time.Second},
							CommitInterval: metav1.Duration{Duration: 5 * time.Second},
							OffsetColumn:   "id",
							InitSchema:     true,
						},
					}},
					Sinks: []Sink{DefaultLogSink},
				},
			},
		},
	})

	WaitForPod()

	defer StartPortForward("db-source-insert-0")()
	SendMessageViaHTTP(`{"message": "msg1", "number": 101}`)
	SendMessageViaHTTP(`{"message": "msg2", "number": 102}`)
	SendMessageViaHTTP(`{"message": "msg3", "number": 103}`)

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(3))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}

func cleanupDB() {
	// clean up db
	CreatePipeline(Pipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "cleanup-db"},
		Spec: PipelineSpec{
			Steps: []StepSpec{
				{
					Name:    "main",
					Cat:     &Cat{},
					Sources: []Source{{HTTP: &HTTPSource{}}},
					Sinks: []Sink{
						{DB: &DBSink{
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
										SQL: `DROP TABLE IF EXISTS test_source_table;`,
									},
								},
								{
									SQLStatement: SQLStatement{
										SQL: `DROP TABLE IF EXISTS argo_dataflow_offsets;`,
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

	defer StartPortForward("cleanup-db-main-0")()
	SendMessageViaHTTP(`hello`)

	WaitForPipeline(UntilMessagesSunk)
	WaitForStep(TotalSunkMessages(1))

	DeletePipelines()
	WaitForPodsToBeDeleted()
}
