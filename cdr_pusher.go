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
func gofetcher(config Config, chanRes chan map[int][]string, chanSync chan bool) {
	for {
		log.Debug("gofetcher sending to chanSync")
		// TODO: move chanSync top of f.Fetch and add a loop
		<-chanSync
		f := new(SQLFetcher)
		if config.StorageDestination == "sqlite" {
			f.Init(config.DBFile, config.DBTable, config.MaxPushBatch, config.CDRFields)
			// Fetch CDRs from SQLite
			err := f.Fetch()
			if err != nil {
				log.Error(err.Error())
				panic(err)
			}
		}
		chanRes <- f.results
		// Wait x seconds between each DB fetch | Heartbeat
		log.Debug("Sleep for " + strconv.Itoa(config.Heartbeat) + " seconds!")
		time.Sleep(time.Second * time.Duration(config.Heartbeat))
	}
}

// Push CDRs to storage
func gopusher(config Config, chanRes chan map[int][]string, chanSync chan bool) {
	for {
		log.Debug("gopusher waiting for chanSync")
		// Send signal to go_fetch to fetch
		chanSync <- true
		// waiting for CDRs on channel
		select {
		case results := <-chanRes:
			if config.StorageDestination == "postgres" {
				// Push CDRs to PostgreSQL
				p := new(PGPusher)
				p.Init(config.PGDatasourcename, config.CDRFields, config.SwitchIP, config.TableDestination)
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

func runApp() (string, error) {
	LoadConfig(defaultConf)
	if err := ValidateConfig(config); err != nil {
		panic(err)
	}

	chanSync := make(chan bool, 1)
	chanRes := make(chan map[int][]string, 1)

	// Start coroutines
	go gofetcher(config, chanRes, chanSync)
	go gopusher(config, chanRes, chanSync)

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
	runApp()
	log.Info("StopTime: " + time.Now().Format("Mon Jan _2 2006 15:04:05"))
}
