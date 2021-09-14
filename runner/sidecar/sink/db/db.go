package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger = sharedutil.NewLogger()

type dbSink struct {
	sinkName string
	db       *sql.DB
	actions  []dfv1.SQLAction
	progs    map[string]*vm.Program
}

func New(ctx context.Context, sinkName string, secretInterface corev1.SecretInterface, x dfv1.DBSink) (sink.Interface, error) {
	dataSource, err := getDataSource(ctx, secretInterface, x)
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}
	db, err := sql.Open(x.Driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)

	progs := make(map[string]*vm.Program)
	for _, action := range x.Actions {
		args := action.Args
		if action.OnError != nil {
			args = append(args, action.OnError.Args...)
		}
		if action.OnRecordNotFound != nil {
			args = append(args, action.OnRecordNotFound.Args...)
		}
		for _, arg := range args {
			if _, ok := progs[arg]; ok {
				continue
			}
			prog, err := expr.Compile(arg)
			if err != nil {
				return nil, fmt.Errorf("failed to compile %q: %w", arg, err)
			}
			progs[arg] = prog
		}
	}
	return dbSink{
		sinkName,
		db,
		x.Actions,
		progs,
	}, nil
}

func (d dbSink) Sink(ctx context.Context, msg []byte) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // The rollback will be ignored if the tx has been committed later in the function.
	for _, action := range d.actions {
		rs, err := d.execStatement(ctx, tx, action.SQL, action.Args, msg)
		if err != nil {
			if action.OnError != nil {
				logger.Error(err, "failed to exec sql", "sql", action.SQL)
				_, err = d.execStatement(ctx, tx, action.OnError.SQL, action.OnError.Args, msg)
				if err != nil {
					return fmt.Errorf("failed to exec onError sql %q", action.OnError.SQL)
				}
				continue
			} else {
				return fmt.Errorf("failed to exec sql %q", action.SQL)
			}
		}
		n, err := rs.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get number of rows affected after exectuation")
		}
		if n == 0 && action.OnRecordNotFound != nil {
			_, err = d.execStatement(ctx, tx, action.OnRecordNotFound.SQL, action.OnRecordNotFound.Args, msg)
			if err != nil {
				return fmt.Errorf("failed to exec onRecordNotFound sql %q", action.OnRecordNotFound.SQL)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit db transction: %w", err)
	}
	return nil
}

func (d dbSink) execStatement(ctx context.Context, tx *sql.Tx, sql string, args []string, msg []byte) (sql.Result, error) {
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get a prepared statement: %w", err)
	}
	defer stmt.Close()
	l := []interface{}{}
	for _, arg := range args {
		prog := d.progs[arg]
		env, err := util.ExprEnv(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create expr env: %w", err)
		}
		res, err := expr.Run(prog, env)
		if err != nil {
			return nil, err
		}
		l = append(l, res)
	}
	rs, err := stmt.Exec(l...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec sql %q: %w", sql, err)
	}
	return rs, nil
}

func getDataSource(ctx context.Context, secretInterface corev1.SecretInterface, x dfv1.DBSink) (string, error) {
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

func (d dbSink) Close() error {
	return d.db.Close()
}
