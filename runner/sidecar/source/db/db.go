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
	"k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

const offsetTableSchema = `CREATE TABLE IF NOT EXISTS argo_dataflow_offsets (
	uid VARCHAR(255) NOT NULL,
	remark VARCHAR(255) NOT NULL,
	offset VARCHAR(255),
	PRIMARY KEY (uid)
)`

type rowData = map[string]interface{}

type dbSource struct {
	db *sql.DB
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, cluster, namespace, pipelineName, stepName, sourceName, sourceURN string, x dfv1.DBSource, buffer source.Buffer) (source.Interface, error) {
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
	remark := fmt.Sprintf("%s.%s.%s.%s.sources.%s", cluster, namespace, pipelineName, stepName, sourceName)
	uid := sharedutil.GetSourceUID(cluster, namespace, pipelineName, stepName, sourceName)
	offset, err = getOffsetFromDB(ctx, db, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			if _, err = insertOffset(ctx, db, uid, remark, ""); err != nil {
				return nil, fmt.Errorf("failed to initialize offset: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get offset from db: %w", err)
		}
	}

	go func() {
		defer runtime.HandleCrash()
		for {
			time.Sleep(x.PollInterval.Duration)
			select {
			case <-ctx.Done():
				return
			default:
				if err = queryData(ctx, db, x.Query, x.OffsetColumn, offset, func(d rowData) error {
					jsonData, err := json.Marshal(d)
					if err != nil {
						return fmt.Errorf("failed to marshal to json: %w", err)
					}
					id := fmt.Sprint(d[x.OffsetColumn])
					buffer <- &source.Msg{
						Meta: dfv1.Meta{
							Source: sourceURN,
							ID:     id,
							Time:   time.Now().Unix(),
						},
						Data: jsonData,
						Ack: func() error {
							offset = id
							return nil
						},
					}
					return nil
				}); err != nil {
					logger.Error(err, "failed to source data query")
				}
			}
		}
	}()

	// update offset in db
	go func() {
		defer runtime.HandleCrash()
		ticker := time.NewTicker(x.CommitInterval.Duration)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if offset != "" {
					if _, err := updateOffset(ctx, db, uid, offset); err != nil {
						logger.Error(err, "failed to update offset", "source", sourceName)
						continue
					}
				}
			}
		}
	}()

	return dbSource{db}, nil
}

func (d dbSource) Close() error {
	return d.db.Close()
}

func getOffsetFromDB(ctx context.Context, db *sql.DB, uid string) (string, error) {
	stmt, err := db.Prepare("select offset from argo_dataflow_offsets where uid=?")
	if err != nil {
		return "", fmt.Errorf("failed to prepare statement for offset query")
	}
	defer func() { _ = stmt.Close() }()
	var offset string
	if err = stmt.QueryRowContext(ctx, uid).Scan(&offset); err != nil {
		return "", err
	}
	return offset, nil
}

func updateOffset(ctx context.Context, db *sql.DB, uid, offset string) (int64, error) {
	stmt, err := db.Prepare("update argo_dataflow_offsets set offset=? where uid=? and offset<>?")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare offset update statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	if result, err := stmt.ExecContext(ctx, offset, uid, offset); err != nil {
		return 0, fmt.Errorf("failed to exec offset update statement: %w", err)
	} else {
		return result.RowsAffected()
	}
}

func insertOffset(ctx context.Context, db *sql.DB, uid, remark, offset string) (int64, error) {
	stmt, err := db.Prepare("insert into argo_dataflow_offsets (uid, remark, offset) values (?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare offset insert statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	if result, err := stmt.ExecContext(ctx, uid, remark, offset); err != nil {
		return 0, fmt.Errorf("failed to exec offset insert statement: %w", err)
	} else {
		return result.RowsAffected()
	}
}

func queryData(ctx context.Context, db *sql.DB, query, offsetColumn, offset string, f func(rowData) error) error {
	sql := fmt.Sprintf("select * from (%s) as dataflow_query_table order by %s", query, offsetColumn)
	params := []interface{}{}
	if offset != "" {
		sql = fmt.Sprintf("select * from (%s) as dataflow_query_table where %s > ? order by %s", query, offsetColumn, offsetColumn)
		params = append(params, offset)
	}
	stmt, err := db.Prepare(sql)
	if err != nil {
		return fmt.Errorf("failed to prepare query statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.QueryContext(ctx, params...)
	if err != nil {
		return fmt.Errorf("failed to execute sql query: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Err() != nil {
		return fmt.Errorf("failed to execute sql query: %w", rows.Err())
	}

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
			logger.Error(err, "failed to process message")
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
