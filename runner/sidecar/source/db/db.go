package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

const offsetTableSchema = `CREATE TABLE IF NOT EXISTS argo_dataflow_offsets (
	table_name VARCHAR(255) NOT NULL,
	consumer_group VARCHAR(255) NOT NULL,
	offset VARCHAR(255),
	PRIMARY KEY(table_name, consumer_group)
)`

type rowData = map[string]interface{}

type dbSource struct {
	db *sql.DB
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, clusterName, namespace, pipelineName, stepName string, replica int, sourceName string, x dfv1.DBSource, f source.Func) (source.Interface, error) {
	dataSource, err := getDataSource(ctx, secretInterface, x)
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}
	db, err := sql.Open(x.Driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)

	if x.InitSchema {
		_, err = db.ExecContext(ctx, offsetTableSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to init offsets table schema: %w", err)
		}
	}

	var offset string
	consumerGroup := fmt.Sprintf("%s.%s.%s.%s.sources.%s", clusterName, namespace, pipelineName, stepName, sourceName)
	offset, err = getOffsetFromDB(ctx, db, x.Table, consumerGroup)
	if err != nil {
		if err == sql.ErrNoRows {
			if _, err = insertOffset(ctx, db, x.Table, consumerGroup, ""); err != nil {
				return nil, fmt.Errorf("failed to initialize offset: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get offset from db: %w", err)
		}
	}

	go func() {
		for {
			time.Sleep(x.PollInterval.Duration)
			select {
			case <-ctx.Done():
				return
			default:
				if err = queryData(ctx, db, x.Table, x.OffsetColumn, offset, func(d rowData) error {
					jsonData, err := json.Marshal(d)
					if err != nil {
						return fmt.Errorf("failed to marshal to json: %w", err)
					}
					if err := f(ctx, jsonData); err != nil {
						return fmt.Errorf("failed to process data: %w", err)
					}
					offset = fmt.Sprintf("%v", d[x.OffsetColumn])
					return nil
				}); err != nil {
					logger.Error(err, "failed to process data query: %w", err)
				}
			}
		}
	}()

	// update offset in db
	go func() {
		ticker := time.NewTicker(x.CommitInterval.Duration)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if offset != "" {
					if _, err := updateOffset(ctx, db, x.Table, consumerGroup, offset); err != nil {
						logger.Error(err, "failed to update offset", "source", sourceName)
						continue
					}
				}
			}
		}
	}()

	return dbSource{
		db: db,
	}, nil
}

func (d dbSource) Close() error {
	return d.db.Close()
}

func getOffsetFromDB(ctx context.Context, db *sql.DB, tableName, consumerGroup string) (string, error) {
	stmt, err := db.Prepare("select offset from argo_dataflow_offsets where table_name=? and consumer_group=?")
	if err != nil {
		return "", fmt.Errorf("failed to prepare statement for offset query")
	}
	defer func() { _ = stmt.Close() }()
	var offset string
	if err = stmt.QueryRowContext(ctx, tableName, consumerGroup).Scan(&offset); err != nil {
		return "", err
	}
	return offset, nil
}

func updateOffset(ctx context.Context, db *sql.DB, tableName, consumerGroup, offset string) (int64, error) {
	stmt, err := db.Prepare("update argo_dataflow_offsets set offset=? where table_name=? and consumer_group=?")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare offset update statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	if result, err := stmt.ExecContext(ctx, offset, tableName, consumerGroup); err != nil {
		return 0, fmt.Errorf("failed to exec offset update statement: %w", err)
	} else {
		return result.RowsAffected()
	}
}

func insertOffset(ctx context.Context, db *sql.DB, tableName, consumerGroup, offset string) (int64, error) {
	stmt, err := db.Prepare("insert into argo_dataflow_offsets (table_name, consumer_group, offset) values (?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare offset insert statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	if result, err := stmt.ExecContext(ctx, tableName, consumerGroup, offset); err != nil {
		return 0, fmt.Errorf("failed to exec offset insert statement: %w", err)
	} else {
		return result.RowsAffected()
	}
}

func queryData(ctx context.Context, db *sql.DB, tableName, offsetColumn, offset string, f func(rowData) error) error {
	sql := "select * from " + tableName + " order by " + offsetColumn
	params := []interface{}{}
	if offset != "" {
		sql = "select * from " + tableName + " where " + offsetColumn + " > ? order by " + offsetColumn
		params = append(params, offset)
	}
	stmt, err := db.Prepare(sql)
	if err != nil {
		return fmt.Errorf("failed to prepare query statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.QueryContext(ctx, params...)
	if err != nil {
		return fmt.Errorf("failed to query data from table %s: %w", tableName, err)
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get table columns: %w", err)
	}
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row data: %w", err)
		}
		entry := make(rowData)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				entry[col] = string(b)
			} else {
				entry[col] = val
			}
		}
		if err = f(entry); err != nil {
			logger.Error(err, "failed process data: %w", err)
		}
	}
	return nil
}

func getDataSource(ctx context.Context, secretInterface corev1.SecretInterface, x dfv1.DBSource) (string, error) {
	if x.DataSource.Value != "" {
		return x.DataSource.Value, nil
	}
	if x.DataSource.ValueFrom != nil && x.DataSource.ValueFrom.SecretKeyRef != nil {
		secret, err := secretInterface.Get(ctx, x.DataSource.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to get secret %q: %w", x.DataSource.ValueFrom.SecretKeyRef.Name, err)
		}
		if d, ok := secret.Data[x.DataSource.ValueFrom.SecretKeyRef.Key]; !ok {
			return "", fmt.Errorf("can not find key %q in secret %q", x.DataSource.ValueFrom.SecretKeyRef.Key, x.DataSource.ValueFrom.SecretKeyRef.Name)
		} else {
			return string(d), nil
		}
	}
	return "", fmt.Errorf("invalid data source config")
}
