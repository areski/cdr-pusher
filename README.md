# CDR-Pusher

CDR-Pusher is a Go Application that will push your CDRs (Call Detail Record)
from your local storage (See list of supported storage) to a centralized
PostgreSQL Database or to a Riak Cluster.

This can be used to centralize your CDRs or simply to safely backup them.

Unifying your CDRs make it easy for call Analysts to do their job. Software
like CDR-Stats (http://www.cdr-stats.org/) can efficiently provide Call &
Billing reporting independently of the type of switches you have in your
infrastructure, so you can centralized CDRs coming from different
communications platform such as Asterisk, FreeSWITCH, Kamailio & others.

[![circleci](https://circleci.com/gh/areski/cdr-pusher.png)](https://circleci.com/gh/areski/cdr-pusher)

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/areski/cdr-pusher)


## Install / Run

Install Golang dependencies (Debian/Ubuntu):

    $ apt-get -y install mercurial git bzr bison
    $ apt-get -y install bison


Install GVM to select which version of Golang you want to install:

    $ bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
    $ source /root/.gvm/scripts/gvm
    $ gvm install go1.4.2
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

The config file [cdr-pusher.yaml](https://raw.githubusercontent.com/areski/cdr-pusher/master/cdr-pusher.yaml)
is installed at the following location: /etc/cdr-pusher.yaml


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

    # cdr_source_type: write the id of the cdr sources type
    # (accepted value: unknown: 0, freeswitch: 1, asterisk: 2, yate: 3, kamailio: 4, opensips: 5)
    cdr_source_type: 1

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

Follow those steps if you don't have config file for supervisord.
Once you see the file echoed to your terminal, reinvoke the command as:

    echo_supervisord_conf > /etc/supervisor/supervisord.conf

This won’t work if you do not have root access, then make sure a `.conf.d` run:

    mkdir /etc/supervisord.conf.d


### Configure CDR-Pusher with Supervisord

Copy Supervisor conf file for cdr-pusher:

    cp ./supervisord/cdr-pusher-prog.conf /etc/supervisord.conf.d/

The makefile provides a function to copy the supervisor conf file:

    make install-supervisor-conf


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
        <template name="example">"${caller_id_name}","${caller_id_number}","${destination_number}","${context}","${start_stamp}","${answer_stamp}","${end_stamp}",${duration},${billsec},"${hangup_cause}","${uuid}","${bleg_uuid}","${accountcode}"</template>
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

Created with love by Areski Belaid [@areskib](http://twitter.com/areskib).


## Roadmap

Our first focus was to support FreeSWITCH CDRs, that's why we decided to support
the SQLite backend, it's also the less invasive and one of the easiest to configure.
SQLite give also the posibility to mark/track the pushed records which is safer
than importing them from CSV files.

We are planning to implement the following very soon:

- Extra DB backend for FreeSWITCH: Mysql, CSV, etc...
- Add support to fetch Asterisk CDRs
- Add support to fetch Kamailio CDRs (Mysql)
