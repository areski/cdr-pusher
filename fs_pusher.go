package main

//
// Prepare PG Database:
//
// $ createdb testdb
// $ psql testdb
// testdb=#
// CREATE TABLE test
//     (id int, call_uuid text, dst text, callerid_name text, callerid_num text, duration int,
//      data jsonb, created timestamp );
//
// INSERT INTO cdr VALUES ("Outbound Call","123555555","123555555","default","2015-01-14 17:58:01","2015-01-14 17:58:01","2015-01-14 17:58:06",5,5,"NORMAL_CLEARING","2bbe83f7-5111-4b5b-9626-c5154608d4ee","","")
//

import (
	"fmt"
	// "github.com/kr/pretty"
	"github.com/op/go-logging"
	// "log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Wait time for results in goroutine
const WAITTIME = 60

var log = logging.MustGetLogger("example")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Fetch CDRs from datasource
func gofetcher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	for {
		println("gofetcher")
		// TODO: move chan_sync top of f.Fetch and add a loop
		<-chan_sync
		f := new(SQLFetcher)
		if config.Storage_destination == "sqlite" {
			f.Init(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
			// Fetch CDRs from SQLite
			err := f.Fetch()
			if err != nil {
				log.Error(err.Error())
				panic(err)
			}
		}
		chan_res <- f.results
		// Wait x seconds between each DB fetch | Heartbeat
		fmt.Printf("Sleep for %d seconds!\n", config.Heartbeat)
		time.Sleep(time.Second * time.Duration(config.Heartbeat))
	}
}

// Push CDRs to storage
func gopusher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	for {
		println("gopusher")
		// Send signal to go_fetch to fetch
		chan_sync <- true
		// waiting for CDRs on channel
		select {
		case results := <-chan_res:
			if config.Storage_destination == "postgres" {
				// Push CDRs to PostgreSQL
				p := new(PGPusher)
				p.Init(config.Pg_datasourcename, config.Cdr_fields, config.Switch_ip, config.Table_destination)
				err := p.Push(results)
				if err != nil {
					log.Error(err.Error())
					panic(err)
				}
			}
		case <-time.After(time.Second * WAITTIME):
			fmt.Println("Nothing received yet...")
		}
	}
}

func run_app() (string, error) {
	LoadConfig(Default_conf)
	// log.Printf("Loaded Config:\n%# v\n\n", pretty.Formatter(config))
	if err := ValidateConfig(config); err != nil {
		panic(err)
	}

	chan_sync := make(chan bool, 1)
	chan_res := make(chan map[int][]string, 1)

	// Start coroutines
	println("Start coroutines")
	go gofetcher(config, chan_res, chan_sync)
	go gopusher(config, chan_res, chan_sync)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// loop work cycle which listen for command or interrupt
	// by system signal
	for {
		select {
		case killSignal := <-interrupt:
			log.Error("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Service was interruped by system signal", nil
			}
			return "Service was killed", nil
		}
	}

}

type HideLogger string

func (h HideLogger) Redacted() interface{} {
	return logging.Redact(string(h))
}

func main() {
	// backendlog := logging.NewLogBackend(os.Stderr, "", 0)
	f, err := os.OpenFile("testlogfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	backendlog := logging.NewLogBackend(f, "", 0)

	backendFormatter := logging.NewBackendFormatter(backendlog, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	// Only errors and more severe messages should be sent to backend log
	backendLeveled.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backendLeveled)

	log.Debug("debug %s", HideLogger("secret message"))
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")

	fmt.Printf("StartTime: %v\n", time.Now())
	run_app()
	fmt.Printf("StopTime: %v\n", time.Now())
}
