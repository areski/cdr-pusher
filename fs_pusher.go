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
	log "github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Wait time for results in goroutine
const WAITTIME = 60

// Fetch CDRs from datasource
func gofetcher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	for {
		log.Debug("gofetcher sending to chan_sync")
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
		log.Debug("Sleep for " + strconv.Itoa(config.Heartbeat) + " seconds!")
		time.Sleep(time.Second * time.Duration(config.Heartbeat))
	}
}

// Push CDRs to storage
func gopusher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	for {
		log.Debug("gopusher waiting for chan_sync")
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
			log.Debug("Nothing received yet...")
		}
	}
}

func run_app() (string, error) {
	LoadConfig(defaultConf)
	if err := ValidateConfig(config); err != nil {
		panic(err)
	}

	chan_sync := make(chan bool, 1)
	chan_res := make(chan map[int][]string, 1)

	// Start coroutines
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
			log.Warn("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Service was interruped by system signal", nil
			}
			return "Service was killed", nil
		}
	}
	return "", nil
}

func main() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Use the Airbrake hook to report errors that have Error severity or above to
	// an exception tracker. You can create custom hooks, see the Hooks section.
	// log.AddHook(&logrus_airbrake.AirbrakeHook{})

	setlogfile := false
	if setlogfile {
		// backendlog := logging.NewLogBackend(os.Stderr, "", 0)
		f, err := os.OpenFile("fs-pusher.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err.Error())
		}
		defer f.Close()
		// Output to stderr instead of stdout, could also be a file.
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stderr)
	}

	// Only log the warning severity or above.
	// log.SetLevel(log.WarnLevel)
	log.SetLevel(log.DebugLevel)

	log.Info("StartTime: " + time.Now().Format("Mon Jan _2 2006 15:04:05"))
	run_app()
	log.Info("StopTime: " + time.Now().Format("Mon Jan _2 2006 15:04:05"))
}
