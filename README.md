# CDR-Pusher

CDR-Pusher is a Go Application that will push your CDRs (Call Detail Record)
from local storage (See list of supported storage) to a centralized
PostgreSQL Database or to a Riak Cluster.

This can be used to centralize your CDRs or simply to safely back them up.

Unifying your CDRs makes it easy for call Analysts to do their job. Software
like CDR-Stats (http://www.cdr-stats.org/) can efficiently provide Call &
Billing reporting independently of the type of switches you have in your
infrastructure, so you can do aggregation and mediation on CDRs coming from a
variety of communications platform such as Asterisk, FreeSWITCH, Kamailio & others.

[![circleci](https://circleci.com/gh/areski/cdr-pusher.png)](https://circleci.com/gh/areski/cdr-pusher)

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/areski/cdr-pusher)


## Install / Run

Install Golang dependencies (Debian/Ubuntu):

    $ apt-get -y install mercurial git bzr bison
    $ apt-get -y install bison


Install GVM to select which version of Golang you want to install:

    $ bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
    $ source /root/.gvm/scripts/gvm
    $ gvm install go1.4.2 --binary
    $ gvm use go1.4.2 --default

Make sure you are running by default Go version >= 1.4.2, check by typing the following:

    $ go version


To install and run the cdr-pusher application, follow those steps:

    $ mkdir /opt/app
    $ cd /opt/app
    $ git clone https://github.com/areski/cdr-pusher.git
    $ cd cdr-pusher
    $ export GOPATH=`pwd`
    $ make build
    $ ./bin/cdr-pusher


If you have some issues with the build, it's possible that you don't have a
recent version of Git, we need Git version >= 1.7.4.
On CentOS 6.X, upgrade Git as follow: http://tecadmin.net/how-to-upgrade-git-version-1-7-10-on-centos-6/

The config file [cdr-pusher.yaml](https://raw.githubusercontent.com/areski/cdr-pusher/master/cdr-pusher.yaml)
and is installed at the following location: /etc/cdr-pusher.yaml


## Configuration file

Config file `/etc/cdr-pusher.yaml`:

    # CDR FETCHING - SOURCE
    # ---------------------

    # storage_source_type: DB backend type where CDRs are stored
    # (accepted values: "sqlite3" and "mysql")
    storage_source: "sqlite3"

    # db_file: specify the database path and name
    db_file: "./sqlitedb/cdr.db"

    # Database DNS
    # Use this with Mysql
    db_dns: ""

    # db_table: the DB table name
    db_table: "cdr"

    # db_flag_field defines the field that will be used as table id (PK) (not used with Sqlite3)
    db_id_field: "id"

    # db_flag_field defines the table field that will be added/used to track the import
    db_flag_field: "flag_imported"

    # max_fetch_batch: Max amoun to CDR to push in batch (value: 1-1000)
    max_fetch_batch: 100

    # heartbeat: Frequency of check for new CDRs in seconds
    heartbeat: 1

    # cdr_fields is list of fields that will be fetched (from SQLite3) and pushed (to PostgreSQL)
    # - if dest_field is callid, it will be used in riak as key to insert
    cdr_fields:
        - orig_field: uuid
          dest_field: callid
          type_field: string
        - orig_field: caller_id_name
          dest_field: caller_id_name
          type_field: string
        - orig_field: caller_id_number
          dest_field: caller_id_number
          type_field: string
        - orig_field: destination_number
          dest_field: destination_number
          type_field: string
        - orig_field: hangup_cause_q850
          dest_field: hangup_cause_id
          type_field: int
        - orig_field: duration
          dest_field: duration
          type_field: int
        - orig_field: billsec
          dest_field: billsec
          type_field: int
        # - orig_field: account_code
        #   dest_field: accountcode
        #   type_field: string
        - orig_field: "datetime(start_stamp)"
          dest_field: starting_date
          type_field: date
        # - orig_field: "strftime('%s', answer_stamp)" # convert to epoch
        - orig_field: "datetime(answer_stamp)"
          dest_field: extradata
          type_field: jsonb
        - orig_field: "datetime(end_stamp)"
          dest_field: extradata
          type_field: jsonb


    # CDR PUSHING - DESTINATION
    # -------------------------

    # storage_dest_type defines where push the CDRs (accepted values: "postgres" or "riak")
    storage_destination: "postgres"

    # Used when storage_dest_type = postgres
    # datasourcename: connect string to connect to PostgreSQL used by sql.Open
    pg_datasourcename: "user=postgres password=password host=localhost port=5432 dbname=cdr-pusher sslmode=disable"

    # Used when storage_dest_type = postgres
    # pg_store_table: the DB table name to store CDRs in Postgres
    table_destination: "cdr_import"

    # Used when storage_dest_type = riak
    # riak_connect: connect string to connect to Riak used by riak.ConnectClient
    riak_connect: "127.0.0.1:8087"

    # Used when storage_dest_type = postgres
    # riak_bucket: the bucket name to store CDRs in Riak
    riak_bucket: "cdr_import"

    # switch_ip: leave this empty to default to your external IP (accepted value: ""|"your IP")
    switch_ip: ""

    # cdr_source_type: write the id of the cdr sources type
    # (accepted value: unknown: 0, csv: 1, api: 2, freeswitch: 3, asterisk: 4, yate: 5, kamailio: 6, opensips: 7, sipwise: 8, veraz: 9)
    cdr_source_type: 0


    # SETTINGS FOR FAKE GENERATOR
    # ---------------------------

    # fake_cdr will populate the SQLite database with fake CDRs for testing (accepted value: "yes|no")
    fake_cdr: "no"

    # fake_amount_cdr is the number of CDRs to generate into the SQLite database for testing purposes (value: 1-1000)
    # this number of CDRs will be created every second
    fake_amount_cdr: 1000



## Deployment

This application aims to be run as Service, it can easily be run by Supervisord.


### Install Supervisord


#### Via Distribution Package

Some Linux distributions offer a version of Supervisor that is installable through the system package manager. These packages may include distribution-specific changes to Supervisor:

    apt-get install supervisor


#### Creating a Configuration File

Follow these steps if you don't have config file for supervisord.
Once you see the file echoed to your terminal, reinvoke the command as:

    echo_supervisord_conf > /etc/supervisor/supervisord.conf

This won’t work if you do not have root access, then make sure a `.conf.d` run:

    mkdir /etc/supervisord.conf.d


### Configure CDR-Pusher with Supervisord

Create an Supervisor conf file for cdr-pusher:

    vim /etc/supervisord.conf.d/cdr-pusher-prog.conf

A supervisor configuration could look as follow:

    [program:cdr-pusher]
    autostart=true
    autorestart=true
    startretries=10
    startsecs = 5
    directory = /opt/app/cdr-pusher/bin
    command = /opt/app/cdr-pusher/bin/cdr-pusher
    user = root
    redirect_stderr = true
    stdout_logfile = /var/log/cdr-pusher/cdr-pusher.log
    stdout_logfile_maxbytes=50MB
    stdout_logfile_backups=10


Make sure the director to store the logs is created, in this case you should
create '/var/log/cdr-pusher':

    mkdir /var/log/cdr-pusher

### Supervisord Manage

Supervisord provides 2 commands, supervisord and supervisorctl:

    supervisord: Initialize Supervisord, run configed processes
    supervisorctl stop programX: Stop process programX. programX is config name in [program:mypkg].
    supervisorctl start programX: Run the process.
    supervisorctl restart programX: Restart the process.
    supervisorctl stop groupworker: Restart all processes in group groupworker
    supervisorctl stop all: Stop all processes. Notes: start, restart and stop won’t reload the latest configs.
    supervisorctl reload: Reload the latest configs.
    supervisorctl update: Reload all the processes whoes config changed.


### Supervisord Service

You can also use supervisor using the supervisor service:

    /etc/init.d/supervisor start


## Configure FreeSWITCH

FreeSWITCH mod_cdr_sqlite is used to store locally the CDRs prior being fetch and send by cdr_pusher: https://wiki.freeswitch.org/wiki/Mod_cdr_sqlite

Some customization can be achieved by editing the config file `cdr-pusher.yaml` and by tweaking the config of Mod_cdr_sqlite `cdr_sqlite.conf.xml`, for instance if you want to same custom fields in your CDRs, you will need to change both configuration files and ensure that the custom field are properly stored in SQLite, then CDR-Pusher offer enough flexibility to push any custom field.

Here an example of 'cdr_sqlite.conf':

    <configuration name="cdr_sqlite.conf" description="SQLite CDR">
      <settings>
        <!-- SQLite database name (.db suffix will be automatically appended) -->
        <!-- <param name="db-name" value="cdr"/> -->
        <!-- CDR table name -->
        <!-- <param name="db-table" value="cdr"/> -->
        <!-- Log a-leg (a), b-leg (b) or both (ab) -->
        <param name="legs" value="a"/>
        <!-- Default template to use when inserting records -->
        <param name="default-template" value="example"/>
        <!-- This is like the info app but after the call is hung up -->
        <!--<param name="debug" value="true"/>-->
      </settings>
      <templates>
        <!-- Note that field order must match SQL table schema, otherwise insert will fail -->
        <template name="example">"${caller_id_name}","${caller_id_number}","${destination_number}","${context}","${start_stamp}","${answer_stamp}","${end_stamp}",${duration},${billsec},"${hangup_cause}", "${hangup_cause_q850}","${uuid}","${bleg_uuid}","${accountcode}"</template>
      </templates>
    </configuration>


## GoLint

http://go-lint.appspot.com/github.com/areski/cdr-pusher

http://goreportcard.com/report/areski/cdr-pusher


## Testing

To run the tests, follow this step:

    $ go test .


## Test Coverage

Visit gocover for the test coverage: http://gocover.io/github.com/areski/cdr-pusher


## License

CDR-Pusher is licensed under MIT, see `LICENSE` file.

Created by Areski Belaid [@areskib](http://twitter.com/areskib).


## Roadmap

Our first focus was to support FreeSWITCH CDRs, that's why we decided to support
the SQLite backend, it's also the less invasive and one of the easiest to configure.
SQLite also gives the posibility to mark/track the pushed records which is safer
than importing them from CSV files.

We are planning to implement the following very soon:

- Extra DB backend for FreeSWITCH: Mysql, CSV, etc...
- Add support to fetch Asterisk CDRs
- Add support to fetch Kamailio CDRs (Mysql)
