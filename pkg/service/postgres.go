package service

import (
	"context"
	"fmt"
	gopg "github.com/go-pg/pg/v10"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type status int
type role int

const (
	started  status = iota
	stopped  status = iota
	primary  role   = iota
	standby  role   = iota
	recovery role   = iota
	virgin   role   = iota
)

const (
	defaultWaitRetry = time.Second * 2
	defaultRetries   = 3
	postgres         = "postgres"
)

var (
	log *logrus.Entry
)

// OSUser ...
type OSUser struct {
	username string
	uid      int
	gid      int
	homeDir  string
}

// DBUser ...
type DBUser struct {
	Username string
	Password string
}

// Postgres ...
type Postgres struct {
	nominee     nominee.Nominee
	cluster     string
	domain      string
	stopCh      chan error
	osUser      OSUser
	replicaUser DBUser
	dbaUser     DBUser
	pgdata      string
	status      status
	role        role
	db          *gopg.DB
	leader      nominee.Nominee
}

// NewPostgres ...
func NewPostgres(config *PGConfig) *Postgres {
	osu, _ := user.Lookup(postgres)
	pg := &Postgres{
		stopCh:  make(chan error),
		nominee: config.Nominee,
		cluster: config.Cluster,
		domain:  config.Domain,
		osUser: OSUser{
			username: postgres,
			homeDir:  osu.HomeDir,
		},
		replicaUser: config.Replicator,
		dbaUser: DBUser{
			Username: postgres,
			Password: config.Postgres.Password,
		},
		pgdata: os.Getenv("PGDATA"),
		status: stopped,
	}

	_ = os.Setenv("POSTGRES_PASSWORD", config.Postgres.Password)

	pg.nominee.Name = fmt.Sprintf("%s-%d", pg.nominee.Name, time.Now().Nanosecond())
	pg.role = pg.lookupCurrentRole()
	pg.osUser.uid, _ = strconv.Atoi(osu.Uid)
	pg.osUser.gid, _ = strconv.Atoi(osu.Gid)
	pg.db = gopg.Connect(&gopg.Options{User: pg.dbaUser.Username, Password: pg.dbaUser.Password})

	_ = pg.createPgPassFile()

	log = logrus.WithFields(logrus.Fields{
		"service": pg.ServiceName(),
		"node":    pg.NomineeName(),
	})
	return pg
}

// ServiceName ...
func (pg *Postgres) ServiceName() string {
	return "postgres"
}

// NomineeName ...
func (pg *Postgres) NomineeName() string {
	return pg.nominee.Name
}

// NomineeAddress ...
func (pg *Postgres) NomineeAddress() string {
	return pg.nominee.Address
}

// Nominee ...
func (pg *Postgres) Nominee() nominee.Nominee {
	return pg.nominee
}

// Lead ...
func (pg *Postgres) Lead(context context.Context, myself nominee.Nominee) error {
	log.Infof("postgres: promote to primary as %v ...\n", myself.Name)
	pg.leader = myself
	defer pg.db.Close()

	if err := pg.start(context); err != nil {
		return err
	}

	if pg.role == virgin {

		if err := pg.createReplicaUser(context); err != nil {
			return err
		}

		if err := pg.authorizeReplication(context); err != nil {
			return err
		}

		if err := pg.reloadConf(context); err != nil {
			return err
		}

	} else if pg.role == standby {

		if err := pg.execOSCmd(context, "pg_ctl promote", 0); err != nil {
			return err
		}

	}

	pg.role = primary
	return nil
}

// Follow ...
func (pg *Postgres) Follow(ctx context.Context, leader nominee.Nominee) error {
	log.Infof("postgres: following the new leader: %v \n", leader.Name)
	pg.leader = leader

	if pg.role == virgin {

		if err := pg.baseBackup(ctx, leader); err != nil {
			return err
		}

		if err := pg.start(ctx); err != nil {
			return err
		}

	} else if pg.role == primary {

		if err := pg.execOSCmd(ctx, fmt.Sprintf("pg_rewind --source-server='host=%s port=5432 user=%s' --target-pgdata=%s", pg.leader.Address, pg.dbaUser.Username, pg.pgdata), 3); err != nil {
			return err
		}
		_ = pg.execOSCmd(ctx, fmt.Sprintf("touch %s/standby.signal", pg.pgdata), 0)
		_ = pg.start(ctx)
		_ = pg.setPrimaryConnInfo(ctx)
		_ = pg.reloadConf(ctx)

	} else {
		if pg.status == stopped {
			_ = pg.start(ctx)
		}
		_ = pg.setPrimaryConnInfo(ctx)
		_ = pg.reloadConf(ctx)
	}

	pg.role = standby
	return nil
}

// Stonith ...
func (pg *Postgres) Stonith(context context.Context) error {
	log.Infof("postgres: stonithing... \n")
	_ = pg.execOSCmd(context, "pg_ctl stop", 0)
	return nil
}

// StopChan ...
func (pg *Postgres) StopChan() nominee.StopChan {
	return pg.stopCh
}

func (pg *Postgres) start(context context.Context) error {
	if pg.status == started {
		return nil
	}

	go func() {
		if pg.role == virgin {
			_ = os.Setenv("POSTGRES_INITDB_ARGS", fmt.Sprintf("--data-checksums %s", os.Getenv("POSTGRES_INITDB_ARGS")))
		}
		start := exec.CommandContext(context, "/docker-entrypoint.sh", pg.osUser.username)
		start.Stdout, start.Stderr = log.Writer(), log.Writer()
		pg.stopCh <- start.Run()
		pg.status = stopped //When the Run returns it means the service is stopped.
	}()

	if err := pg.warmUp(context, defaultRetries); err != nil {
		return err
	}

	pg.status = started
	return nil
}

func (pg *Postgres) createPgPassFile() error {
	postgresPgPass := fmt.Sprintf("%s/.pgpass", pg.osUser.homeDir)
	if _, err := os.Stat(postgresPgPass); os.IsNotExist(err) {
		replicator := fmt.Sprintf("*:*:*:%s:%s", pg.replicaUser.Username, pg.replicaUser.Password)
		dba := fmt.Sprintf("*:*:*:%s:%s", pg.dbaUser.Username, pg.dbaUser.Password)
		lines := []byte(replicator + "\n" + dba)
		if err := ioutil.WriteFile(postgresPgPass, lines, 0600); err != nil {
			return err
		}

		if err := os.Chown(postgresPgPass, pg.osUser.uid, pg.osUser.gid); err != nil {
			return err
		}
	}

	return nil
}

func (pg *Postgres) authorizeReplication(context context.Context) error {
	if err := pg.execDBCmd(context, "ALTER SYSTEM SET listen_addresses TO '*'"); err != nil {
		return err
	}
	path := fmt.Sprintf("%s/pg_hba.conf", pg.pgdata)
	line := fmt.Sprintf("host replication %s 0.0.0.0/0 md5", pg.replicaUser.Username)
	content, _ := ioutil.ReadFile(path)
	if contains := strings.Contains(string(content), line); contains {
		return nil
	}
	return pg.execOSCmd(context, fmt.Sprintf("echo '%s' >> %s", line, path), 0)
}

func (pg *Postgres) baseBackup(context context.Context, leader nominee.Nominee) error {
	cmd := fmt.Sprintf("pg_basebackup -h %s -U %s -p 5432 -D %s -Fp -Xs -P -R", leader.Address, pg.replicaUser.Username, pg.pgdata)
	return pg.execOSCmd(context, cmd, defaultRetries)
}

func (pg *Postgres) createReplicaUser(context context.Context) error {
	if err := pg.execDBCmd(context, fmt.Sprintf("DROP USER IF EXISTS %s", pg.replicaUser.Username)); err != nil {
		return err
	}

	if err := pg.execDBCmd(context, fmt.Sprintf("CREATE USER %s WITH REPLICATION ENCRYPTED PASSWORD '%s'", pg.replicaUser.Username, pg.replicaUser.Password)); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) setPrimaryConnInfo(ctx context.Context) error {
	return pg.execDBCmd(ctx, fmt.Sprintf("ALTER SYSTEM SET primary_conninfo TO "+
		"'user=replicator "+
		"passfile=''/var/lib/postgresql/.pgpass'' "+
		"channel_binding=prefer "+
		"host=%s "+
		"port=5432 "+
		"sslmode=prefer "+
		"sslcompression=0 "+
		"ssl_min_protocol_version=TLSv1.2 "+
		"gssencmode=prefer "+
		"krbsrvname=postgres "+
		"target_session_attrs=any'", pg.leader.Address))
}

func (pg *Postgres) reloadConf(context context.Context) error {
	return pg.execDBCmd(context, "select pg_reload_conf()")
}

func (pg *Postgres) lookupCurrentRole() role {
	if pg.isPgDataEmpty() {
		return virgin
	} else if _, err := os.Stat(fmt.Sprintf("%s/standby.signal", pg.pgdata)); err == nil {
		return standby
	} else if _, err := os.Stat(fmt.Sprintf("%s/recovery.signal", pg.pgdata)); err == nil {
		return recovery
	} else {
		return primary
	}
}

func (pg *Postgres) execOSCmd(context context.Context, cmd string, retries int) error {
	for i := retries; ; i-- {
		command := exec.CommandContext(context, "su", "-c", cmd, pg.osUser.username)
		command.Stdout, command.Stderr = log.Writer(), log.Writer()

		if err := command.Run(); err != nil {
			if i >= 0 {
				time.Sleep(defaultWaitRetry)
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}

func (pg *Postgres) execDBCmd(context context.Context, cmd string) error {
	_, err := pg.db.ExecContext(context, cmd)
	return err
}

func (pg *Postgres) warmUp(context context.Context, retries int) error {
	for i := retries; ; i-- {
		if err := pg.db.Ping(context); err != nil {
			if i >= 0 {
				time.Sleep(defaultWaitRetry)
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}

func (pg *Postgres) isPgDataEmpty() bool {
	entries, _ := ioutil.ReadDir(pg.pgdata)
	return len(entries) == 0
}
