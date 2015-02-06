# CDR-Pusher

CDR-Pusher is a Go Application that will push your stored CDRs (Call Detail
Record) from your local storage (See list of supported storage) to a distant
PostGreSQL or Riak Cluster.

This can be used to centralize your CDRs or simply to safely backup them in
realtime. Software, like CDR-Stats, can then be used to provide CDR reporting.

[![circleci](https://circleci.com/gh/areski/cdr-pusher.png)](https://circleci.com/gh/areski/cdr-pusher)

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/areski/cdr-pusher)


## Roadmap

Our first focus was to support FreeSWITCH CDRs, we decided to go for SQLite
backend as it seems to be the less invasive and easy enough to configure,
plus SQLite give the posibility to mark/track the pushed records which is safer
than importing them from CSV files.

Next we would like to implement:

- Extra DB backend for FS: Mysql, CSV, etc...
- Add support to fetch Asterisk CDRs
- Add support to fetch Kamailio CDRs (Mysql) and CSV


## Install / Run

The config file need to be installed at the following location /etc/cdr-pusher.yaml

To install and run the cdr-pusher application, follow those steps:

    $ git clone https://github.com/areski/cdr-pusher.git
    $ cd cdr-pusher
    $ export GOPATH=`pwd`
    $ make build
    $ ./bin/cdr-pusher


## Testing

To run the tests, follow this step:

    $ cd cdr-pusher
    $ go test .


## Test Coverage

Visit gocover for the test coverage: http://gocover.io/github.com/areski/cdr-pusher


## Configuration file

Config file `/etc/cdr-pusher.yaml`:

    # storage_dest_type defines where push the CDRs (accepted values: "postgres", "riak" or "both")
    storage_destination: "both"

    # Used when storage_dest_type = postgres
    # datasourcename: connect string to connect to PostgreSQL used by sql.Open
    pg_datasourcename: "user=postgres password=password host=localhost port=5433 dbname=cdr-pusher sslmode=disable"

    # Used when storage_dest_type = postgres
    # pg_store_table: the DB table name to store CDRs in Postgres
    table_destination: "cdr_import"

    # Used when storage_dest_type = riak
    # riak_connect: connect string to connect to Riak used by riak.ConnectClient
    riak_connect: "127.0.0.1:8087"

    # Used when storage_dest_type = postgres
    # riak_bucket: the bucket name to store CDRs in Riak
    riak_bucket: "cdr_import"

    # storage_source_type: type to CDRs to push
    storage_source: "sqlite"

    # db_file: specify the database path and name
    # db_file: "/usr/local/freeswitch/cdr.db"
    db_file: "./sqlitedb/cdr.db"

    # db_table: the DB table name
    db_table: "cdr"

    # db_flag_field defines the table field that will be added/used to track the import
    db_flag_field: "flag_imported"

    # heartbeat: Frequence of check for new CDRs in seconds
    heartbeat: 1

    # max_push_batch: Max amoun to CDR to push in batch (value: 1-1000)
    max_push_batch: 1000

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
        - orig_field: duration
          dest_field: duration
          type_field: int
        - orig_field: billsec
          dest_field: billsec
          type_field: int
        - orig_field: "datetime(start_stamp)"
          dest_field: starting_date
          type_field: date
        - orig_field: "datetime(answer_stamp)"
          dest_field: extradata
          type_field: jsonb
        - orig_field: "datetime(end_stamp)"
          dest_field: extradata
          type_field: jsonb

    # switch_ip: leave this empty to default to your external IP (accepted value: ""|"your IP")
    switch_ip: ""

    # fake_cdr will populate the SQLite database with fake CDRs for test purpose (accepted value: "yes|no")
    fake_cdr: "no"

    # fake_amount_cdr is the amount of CDRs to generate into the SQLite database for test purpose (value: 1-1000)
    # this amount of CDRs will be created every second
    fake_amount_cdr: 1000


## Deployment

This application aims to be run as Service, it can easily be run by Supervisord.

### Install Supervisord

#### Via Distribution Package

Some Linux distributions offer a version of Supervisor that is installable through the system package manager. These packages may include distribution-specific changes to Supervisor:

    apt-get install supervisor


#### Creating a Configuration File

Once you see the file echoed to your terminal, reinvoke the command as:

    echo_supervisord_conf > /etc/supervisord.conf

This won’t work if you do not have root access, then make sure a `.conf.d` run:

    mkdir /etc/supervisord.conf.d

### Configure CDR-Pusher with Supervisord

Copy Supervisor conf file for cdr-pusher:

    cp ./supervisord/cdr-pusher-prog.conf /etc/supervisord.conf.d/

### Supervisord Manage

Supervisord provides 2 commands, supervisord and supervisorctl:

    supervisord: Initialize Supervisord, run configed processes
    supervisorctl stop programxxx: Stop process programxxx. programxxx is configed name in [program:beepkg]. Here is beepkg.
    supervisorctl start programxxx: Run the process.
    supervisorctl restart programxxx: Restart the process.
    supervisorctl stop groupworker: Restart all processes in group groupworker
    supervisorctl stop all: Stop all processes. Notes: start, restart and stop won’t reload the latest configs.
    supervisorctl reload: Reload the latest configs.
    supervisorctl update: Reload all the processes whoes config changed.

## Configure FreeSWITCH

A shell script is provided to install FreeSWITCH on Debian 7.x: https://github.com/areski/cdr-pusher/blob/master/install/install-freeswitch.sh

FreeSWITCH mod_cdr_sqlite is used to store locally the CDRs prior being fetch and send by cdr_pusher: https://wiki.freeswitch.org/wiki/Mod_cdr_sqlite

Some customization can be achieved by editing the config file `cdr-pusher.yaml` and by tweaking the config of Mod_cdr_sqlite `cdr_sqlite.conf.xml`, for instance if you want to send specific fields in your CDRs, you will need to change both configuration files and ensure that the custom field are properly stored in SQLite, then CDR-Pusher offer enough flexibility to push any custom field.


## GoLint

http://go-lint.appspot.com/github.com/areski/cdr-pusher

http://goreportcard.com/report/areski/cdr-pusher


## License

CDR-Pusher is licensed under MIT, see `LICENSE` file.

Created with love by Areski Belaid [@areskib](http://twitter.com/areskib).


## TODO

- [x] Fetch & Push CDRs to Postgresql
- [x] Implement using goroutine with channel to communicate between Fetcher <--> Pusher
- [x] Add logging
- [x] Add test / circle-ci / Badge
- [x] Code lint: http://go-lint.appspot.com
- [x] godoc / https://gowalker.org
- [x] Code coverage: http://gocover.io
- [x] Improve Code coverage
- [x] Add check for PG connection in goroutine (connect error + Ping)
- [x] Push CDRs to Riak
- [ ] Deploy with Supervisord
- [ ] Review install / deployment documentation
- [ ] Install script Go App
- [ ] Ansible Support
- [ ] Improve Riak store with ConnectionPool
