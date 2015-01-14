# FS-Pusher

FS-Pusher is a Go Application aims to run as a service that push CDRs from
local DB storage (ie: SQLite) to a PostGreSQL or Riak Cluster.

[![circleci](https://circleci.com/gh/areski/fs-pusher.png)](https://circleci.com/gh/areski/fs-pusher)

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/areski/fs-pusher)


## Usage

You may find Pinguino useful if you want to activate/deactivate some services or run custom actions on your computer/server based on the output of webservices and surroundings.


## Install / Run

The config file need to be installed at the following location /etc/fs-pusher.yaml

To install and run the fs-pusher application, follow those steps:

    $ git clone https://github.com/areski/fs-pusher.git
    $ cd fs-pusher
    $ export GOPATH=`pwd`
    $ make build
    $ ./bin/fs-pusher


## Configuration file

Config file `/etc/fs-pusher.yaml`:

    # storage_dest_type: accepted value "postgres" or "riak"
    storage_destination: "postgres"

    # Used when storage_dest_type = postgres
    # datasourcename: connect string to connect to PostgreSQL used by sql.Open
    pg_datasourcename: "host=localhost dbname=testdb sslmode=disable"

    # Used when storage_dest_type = riak
    # riak_connect: connect string to connect to Riak used by riak.ConnectClient
    riak_connect: "127.0.0.1:8087"

    # storage_source_type: type to CDRs to push
    storage_source: "sqlite"

    # db_file: specify the database path and name
    db_file: "/usr/local/freeswitch/cdr.db"

    # db_table: the DB table name
    db_table: "cdr"

    # heartbeat: Frequence of check for new CDRs in seconds
    heartbeat: 15

    # max_push_batch: Max amoun to CDR to push in batch (value: 1-1000)
    max_push_batch: 200

    # cdr_fields: list of fields with type to transit - format is "original_field:destination_field:type, ..."
    # ${caller_id_name}","${caller_id_number}","${destination_number}","${context}","${start_stamp}","${answer_stamp}","${end_stamp}",${duration},${billsec},"${hangup_cause}","${uuid}","${bleg_uuid}","${accountcode}
    cdr_fields: "caller_id_name:caller_id_name:string,caller_id_number:caller_id_number:string,destination_number:destination_number:string,context:context:string,start_stamp:start_stamp:date,answer_stamp:answer_stamp:date,end_stamp:end_stamp:date,duration:duration:integer,billsec:billsec:integer,hangup_cause:hangup_cause:integer,uuid:uuid:string,bleg_uuid:bleg_uuid:string,accountcode:accountcode:string"

    # switch_ip: leave this empty to default to your external IP (accepted value: ""|"your IP")
    switch_ip: ""


## Deployment

This application aims to be run as Go Service, it can be run by Supervisord.

### Install Supervisord

#### Via Distribution Package

Some Linux distributions offer a version of Supervisor that is installable through the system package manager. These packages may include distribution-specific changes to Supervisor:

    apt-get install supervisor


#### Creating a Configuration File

Once you see the file echoed to your terminal, reinvoke the command as:

    echo_supervisord_conf > /etc/supervisord.conf

This won’t work if you do not have root access, then make sure a `.conf.d` run:

    mkdir /etc/supervisord.conf.d

### Configure FS-Pusher with Supervisord

Copy Supervisor conf file for fs-pusher:

    cp ./supervisord/fs-pusher-prog.conf /etc/supervisord.conf.d/

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

## License

FS-pusher is licensed under MIT, see `LICENSE` file.


## TODO

- [ ] Fetch & Push CDRs to Riak
- [ ] Add logging
- [ ] Deploy with Supervisord
- [ ] Add test / travis-ci / Badge
- [ ] godoc / https://gowalker.org
- [ ] Review install / deployment documentation
- [ ] Install script Go App
- [ ] Ansible Support
